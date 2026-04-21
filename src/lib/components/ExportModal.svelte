<script lang="ts">
  import { t } from 'svelte-i18n';
  import { X } from 'lucide-svelte';
  import { invoke } from '@tauri-apps/api/core';
  import { open } from '@tauri-apps/plugin-dialog';
  import { exportStore, status } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import type { Source } from '$lib/components/types';

  type FilenameError =
    | 'empty'
    | 'invalid_char'
    | 'reserved'
    | 'trailing_dot_or_space'
    | 'has_xml_suffix';
  type RowError = FilenameError | 'duplicate';

  const INVALID_CHARS = /[<>:"/\\|?*\x00-\x1F]/;
  const RESERVED = new Set([
    'CON','PRN','AUX','NUL',
    'COM1','COM2','COM3','COM4','COM5','COM6','COM7','COM8','COM9',
    'LPT1','LPT2','LPT3','LPT4','LPT5','LPT6','LPT7','LPT8','LPT9',
  ]);

  function validateFilename(name: string): FilenameError | null {
    const trimmed = name.trim();
    if (trimmed.length === 0) return 'empty';
    if (/\.xml$/i.test(trimmed)) return 'has_xml_suffix';
    if (INVALID_CHARS.test(trimmed)) return 'invalid_char';
    if (/[. ]$/.test(trimmed)) return 'trailing_dot_or_space';
    const base = trimmed.split('.')[0].toUpperCase();
    if (RESERVED.has(base)) return 'reserved';
    return null;
  }

  function errorMessage(err: RowError): string {
    return $t(`export.error_${err}`);
  }

  function buildTargetPath(folder: string, baseName: string): string {
    const sep = folder.includes('\\') && !folder.includes('/') ? '\\' : '/';
    const trimmedFolder = folder.replace(/[\\/]+$/, '');
    return `${trimmedFolder}${sep}${baseName.trim()}.xml`;
  }

  let exportFolder = $state('');
  let filenames = $state<Map<string, string>>(new Map());
  let selectedSourceIds = $state<Set<string>>(new Set());
  let pendingOverwrite = $state<string[] | null>(null);
  let exportError = $state<string | null>(null);

  let exportableSources = $derived(
    sources.items.filter(
      (s: Source) => s.kind === 'population' || s.kind === 'network' || s.kind === 'transit'
    )
  );

  $effect(() => {
    if (exportStore.open) {
      exportFolder = '';
      exportError = null;
      pendingOverwrite = null;
      selectedSourceIds = new Set(
        exportableSources.filter((s: Source) => s.visible).map((s: Source) => s.id)
      );
      const next = new Map<string, string>();
      for (const s of exportableSources) next.set(s.id, s.name);
      filenames = next;
    }
  });

  let rowErrors = $derived.by(() => {
    const errs = new Map<string, RowError>();
    for (const s of exportableSources) {
      if (!selectedSourceIds.has(s.id)) continue;
      const err = validateFilename(filenames.get(s.id) ?? '');
      if (err) errs.set(s.id, err);
    }
    const seen = new Map<string, string[]>();
    for (const s of exportableSources) {
      if (!selectedSourceIds.has(s.id)) continue;
      if (errs.has(s.id)) continue;
      const key = (filenames.get(s.id) ?? '').trim().toLowerCase();
      if (!seen.has(key)) seen.set(key, []);
      seen.get(key)!.push(s.id);
    }
    for (const ids of seen.values()) {
      if (ids.length > 1) for (const id of ids) errs.set(id, 'duplicate');
    }
    return errs;
  });

  let canExport = $derived(
    exportFolder.trim().length > 0 && selectedSourceIds.size > 0 && rowErrors.size === 0
  );

  function toggleSource(id: string) {
    if (selectedSourceIds.has(id)) {
      selectedSourceIds.delete(id);
    } else {
      selectedSourceIds.add(id);
    }
    selectedSourceIds = new Set(selectedSourceIds);
    pendingOverwrite = null;
  }

  function setFilename(id: string, value: string) {
    filenames.set(id, value);
    filenames = new Map(filenames);
    pendingOverwrite = null;
  }

  function onBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) exportStore.hide();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') exportStore.hide();
  }

  async function handleBrowse() {
    const folder = await open({
      directory: true,
      multiple: false,
      title: $t('export.folder_picker_title'),
    });
    if (typeof folder === 'string') {
      exportFolder = folder;
      pendingOverwrite = null;
    }
  }

  function cancelOverwrite() {
    pendingOverwrite = null;
  }

  async function handleExport() {
    if (!canExport) return;
    const selected = exportableSources.filter((s: Source) => selectedSourceIds.has(s.id));
    const targets = selected.map((s) => ({
      source: s,
      path: buildTargetPath(exportFolder, filenames.get(s.id) ?? s.name),
    }));

    if (pendingOverwrite === null) {
      const existing: string[] = await invoke('check_paths_exist', {
        paths: targets.map((t) => t.path),
      });
      if (existing.length > 0) {
        pendingOverwrite = existing;
        return;
      }
    }

    exportError = null;
    status.setOperation('exporting');
    try {
      for (const { source, path } of targets) {
        const command =
          source.kind === 'network'
            ? 'export_network'
            : source.kind === 'transit'
              ? 'export_transit'
              : 'export_population';
        await invoke(command, { sourceName: source.name, outputPath: path });
      }
      exportStore.hide();
    } catch (err) {
      const e = err as { message?: string };
      exportError = typeof err === 'string' ? err : e.message || 'Export failed.';
    } finally {
      status.setOperation(null);
      pendingOverwrite = null;
    }
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if exportStore.open}
  <div
    class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/50"
    role="button"
    tabindex="-1"
    aria-label="Close export dialog"
    onclick={onBackdropClick}
    onkeydown={onKeydown}
  >
    <div class="bg-base-100 text-base-content rounded-lg shadow-xl w-[520px] flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3">
        <h2 class="text-sm font-semibold">{$t('export.title')}</h2>
        <button class="btn btn-ghost btn-xs btn-square" onclick={() => exportStore.hide()}>
          <X size={14} />
        </button>
      </div>

      <!-- Body -->
      <div class="px-4 pb-2 space-y-4 overflow-y-auto">
        <p class="text-xs text-base-content/60">{$t('export.save_hint')}</p>

        <!-- Destination folder -->
        <div>
          <span class="text-xs font-medium text-base-content/70 mb-2 block">{$t('export.folder_label')}</span>
          <div class="flex gap-2">
            <input
              type="text"
              class="input input-sm input-bordered flex-1 text-xs"
              readonly
              value={exportFolder}
              placeholder={$t('export.folder_placeholder')}
            />
            <button class="btn btn-sm btn-ghost" onclick={handleBrowse}>{$t('export.browse_folder')}</button>
          </div>
        </div>

        <!-- Sources -->
        <div>
          <span class="text-xs font-medium text-base-content/70 mb-2 block">{$t('export.sources')}</span>
          {#if exportableSources.length === 0}
            <p class="text-xs text-base-content/40">{$t('sources.no_sources')}</p>
          {:else}
            <div class="space-y-2">
              {#each exportableSources as source (source.id)}
                {@const selected = selectedSourceIds.has(source.id)}
                {@const rowErr = rowErrors.get(source.id)}
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-2">
                    <button
                      class="flex items-center gap-2 py-1 text-left cursor-pointer flex-shrink-0"
                      style:width="180px"
                      onclick={() => toggleSource(source.id)}
                      aria-pressed={selected}
                    >
                      <span
                        class="inline-block w-3 h-3 rounded-sm flex-shrink-0 transition-all"
                        style:background={selected ? source.color : 'transparent'}
                        style:opacity={source.opacity}
                        style:border={selected ? 'none' : `2px solid ${source.color}`}
                      ></span>
                      <span class="text-xs truncate" class:opacity-40={!selected}>{source.name}</span>
                    </button>
                    <input
                      type="text"
                      class="input input-xs input-bordered flex-1 text-xs"
                      class:input-error={selected && rowErr !== undefined}
                      class:opacity-40={!selected}
                      value={filenames.get(source.id) ?? ''}
                      placeholder={$t('export.filename_label')}
                      oninput={(e) => setFilename(source.id, (e.target as HTMLInputElement).value)}
                    />
                    <span class="text-xs text-base-content/40 flex-shrink-0">.xml</span>
                  </div>
                  {#if selected && rowErr}
                    <p class="text-xs text-error pl-[188px]">{errorMessage(rowErr)}</p>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Overwrite warning -->
        {#if pendingOverwrite}
          <div class="rounded-lg bg-warning/10 px-3 py-2 text-xs space-y-1">
            <p class="font-medium text-warning">{$t('export.warn_overwrite_title')}</p>
            <ul class="list-disc pl-4 text-base-content/80">
              {#each pendingOverwrite as path (path)}
                <li class="break-all">{path}</li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>

      <!-- Error -->
      {#if exportError}
        <p class="text-xs text-error px-4 pb-2">{exportError}</p>
      {/if}

      <!-- Footer -->
      <div class="flex justify-end gap-2 px-4 py-3">
        {#if pendingOverwrite}
          <button class="btn btn-sm btn-ghost" onclick={cancelOverwrite}>
            {$t('export.cancel')}
          </button>
          <button class="btn btn-sm btn-warning" onclick={handleExport}>
            {$t('export.confirm_overwrite_btn')}
          </button>
        {:else}
          <button class="btn btn-sm btn-ghost" onclick={() => exportStore.hide()}>
            {$t('export.cancel')}
          </button>
          <button class="btn btn-sm btn-primary" disabled={!canExport} onclick={handleExport}>
            {$t('export.export_btn')}
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}
