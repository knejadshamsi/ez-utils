<script lang="ts">
  import { t } from 'svelte-i18n';
  import { X } from 'lucide-svelte';
  import { exportStore } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import type { Source } from '$lib/components/types';

  let exportPath = $state('');
  let selectedSourceIds = $state<Set<string>>(new Set());

  // Reset state when modal opens
  $effect(() => {
    if (exportStore.open) {
      exportPath = '';
      selectedSourceIds = new Set(sources.items.filter((s: Source) => s.visible).map((s: Source) => s.id));
    }
  });

  function toggleSource(id: string) {
    if (selectedSourceIds.has(id)) {
      selectedSourceIds.delete(id);
    } else {
      selectedSourceIds.add(id);
    }
    selectedSourceIds = new Set(selectedSourceIds);
  }

  let canExport = $derived(exportPath.trim() && selectedSourceIds.size > 0);

  function onBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) exportStore.hide();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') exportStore.hide();
  }

  function handleBrowse() {
    // TODO: open Tauri native directory picker
    console.log('Browse for export directory');
  }

  function handleExport() {
    // TODO: trigger actual export to chosen path with selected sources
    const selected = sources.items.filter((s: Source) => selectedSourceIds.has(s.id));
    console.log('Export to:', exportPath, 'sources:', selected);
    exportStore.hide();
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if exportStore.open}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/50" onclick={onBackdropClick}>
    <div class="bg-base-100 text-base-content rounded-lg shadow-xl w-[380px] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3">
        <h2 class="text-sm font-semibold">{$t('export.title')}</h2>
        <button class="btn btn-ghost btn-xs btn-square" onclick={() => exportStore.hide()}>
          <X size={14} />
        </button>
      </div>

      <!-- Body -->
      <div class="px-4 pb-2 space-y-4">
        <!-- Output path -->
        <div>
          <label class="text-xs font-medium text-base-content/70 mb-2 block">{$t('export.destination')}</label>
          <div class="flex gap-2">
            <input type="text" class="input input-sm input-bordered flex-1 text-xs" bind:value={exportPath} placeholder="Choose export folder..." />
            <button class="btn btn-sm btn-ghost" onclick={handleBrowse}>Browse</button>
          </div>
        </div>

        <!-- Sources to export -->
        <div>
          <label class="text-xs font-medium text-base-content/70 mb-2 block">{$t('export.sources')}</label>
          {#if sources.items.length === 0}
            <p class="text-xs text-base-content/40">{$t('sources.no_sources')}</p>
          {:else}
            <div class="space-y-1">
              {#each sources.items as source (source.id)}
                <button
                  class="flex items-center gap-2 py-1 w-full text-left cursor-pointer"
                  onclick={() => toggleSource(source.id)}
                >
                  <span
                    class="inline-block w-3 h-3 rounded-sm flex-shrink-0 transition-all"
                    style:background={selectedSourceIds.has(source.id) ? source.color : 'transparent'}
                    style:opacity={source.opacity}
                    style:border={selectedSourceIds.has(source.id) ? 'none' : `2px solid ${source.color}`}
                  ></span>
                  <span class="text-xs" class:opacity-40={!selectedSourceIds.has(source.id)}>{source.name}</span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-2 px-4 py-3">
        <button class="btn btn-sm btn-ghost" onclick={() => exportStore.hide()}>
          {$t('export.cancel')}
        </button>
        <button class="btn btn-sm btn-primary" disabled={!canExport} onclick={handleExport}>
          {$t('export.export_btn')}
        </button>
      </div>
    </div>
  </div>
{/if}
