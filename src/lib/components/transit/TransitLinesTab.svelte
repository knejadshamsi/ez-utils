<script lang="ts">
  import { tick, untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { Bus, Pencil, Plus, Search, TrainFront, TramFront, Trash2, X } from 'lucide-svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { transit, transitEdit } from '$lib/stores/transit.svelte';
  import {
    loadDeletePreview,
    submitCreateLine,
    submitRenameLine,
  } from '$lib/stores/transit-actions';
  import TransitLineEditRow from './TransitLineEditRow.svelte';
  import TransitRoutesList from './TransitRoutesList.svelte';

  $effect(() => {
    const activeName = sources.active?.kind === 'transit' ? sources.activeName : null;
    untrack(() => {
      if (activeName) {
        void transit.loadLinesForSource(activeName);
      } else {
        transit.clearLines();
      }
    });
  });

  $effect(() => {
    const lineId = transit.selectedLineId;
    const activeName = sources.active?.kind === 'transit' ? sources.activeName : null;
    untrack(() => {
      if (lineId && activeName) {
        void transit.loadRoutesForLine(activeName, lineId);
      }
    });
  });

  $effect(() => {
    if (transitEdit.deleteTarget && transitEdit.deleteTarget.loadingPreview) {
      void loadDeletePreview();
    }
  });

  $effect(() => {
    const id = transit.selectedLineId;
    if (!id) return;
    const count = transit.filteredLines.length;
    untrack(() => {
      if (!transit.filteredLines.some((l) => l.id === id)) return;
      void tick().then(() => {
        const el = document.querySelector(`[data-line-id="${CSS.escape(id)}"]`);
        (el as HTMLElement | null)?.scrollIntoView({ block: 'nearest' });
      });
      void count;
    });
  });

  function modeIconFor(raw: string) {
    const m = raw.toLowerCase();
    if (m === 'bus') return Bus;
    if (m === 'subway' || m === 'metro' || m === 'rail') return TrainFront;
    if (m === 'tram' || m === 'light_rail' || m === 'lightrail' || m === 'streetcar') return TramFront;
    return null;
  }
</script>

<section class="transit-list-section">
  <header class="section-toolbar">
    <h3>{$t('transit.lines_heading')}</h3>
    <div class="section-toolbar-actions">
      <button
        class="toolbar-btn"
        class:active={transitEdit.createLineForm.open}
        onclick={() => {
          if (transitEdit.createLineForm.open) {
            transitEdit.closeCreateLineForm();
          } else {
            transit.closeSearch();
            transitEdit.openCreateLineForm();
          }
        }}
        title={$t('transit.add_line')}
        aria-label={$t('transit.add_line')}
      >
        <Plus size={14} />
      </button>
      <button
        class="toolbar-btn"
        class:active={transit.searchOpen}
        onclick={() => transit.toggleSearch()}
        title={transit.searchOpen ? $t('transit.close_line_search') : $t('transit.search_lines')}
        aria-label={transit.searchOpen ? $t('transit.close_line_search') : $t('transit.search_lines')}
      >
        {#if transit.searchOpen}
          <X size={14} />
        {:else}
          <Search size={14} />
        {/if}
      </button>
    </div>
  </header>

  {#if transit.searchOpen}
    <div class="search-row">
      <!-- svelte-ignore a11y_autofocus -->
      <input
        type="text"
        placeholder={$t('transit.search_line_placeholder')}
        value={transit.searchQuery}
        oninput={(e) => transit.setSearchQuery((e.target as HTMLInputElement).value)}
        autofocus
      />
    </div>
  {/if}

  {#if transitEdit.createLineForm.open}
    <TransitLineEditRow
      idValue={transitEdit.createLineForm.id}
      nameValue={transitEdit.createLineForm.name}
      modeValue={transitEdit.createLineForm.transportMode}
      submitting={transitEdit.createLineForm.submitting}
      error={transitEdit.createLineForm.error}
      onIdChange={(v) => transitEdit.setCreateLineFormId(v)}
      onNameChange={(v) => transitEdit.setCreateLineFormName(v)}
      onModeChange={(v) => transitEdit.setCreateLineFormMode(v)}
      onSave={() => void submitCreateLine()}
      onCancel={() => transitEdit.closeCreateLineForm()}
    />
  {/if}

  <div class="list-container">
    {#if transit.loadingLines && transit.lines.length === 0}
      <div class="list-empty">…</div>
    {:else if transit.linesError}
      <div class="list-error">{transit.linesError}</div>
    {:else if transit.filteredLines.length === 0}
      <div class="list-empty">{$t('transit.no_lines')}</div>
    {:else}
      {#each transit.filteredLines as line (line.id)}
        {#if transitEdit.renameTarget && transitEdit.renameTarget.originalId === line.id}
          <TransitLineEditRow
            idValue={transitEdit.renameTarget.id}
            nameValue={transitEdit.renameTarget.name}
            modeValue={transitEdit.renameTarget.transportMode}
            submitting={transitEdit.renameTarget.submitting}
            error={transitEdit.renameTarget.error}
            onIdChange={(v) => transitEdit.setRenameId(v)}
            onNameChange={(v) => transitEdit.setRenameName(v)}
            onModeChange={(v) => transitEdit.setRenameMode(v)}
            onSave={() => void submitRenameLine()}
            onCancel={() => transitEdit.cancelRename()}
          />
        {:else}
          <div
            class="list-row group"
            class:list-row-selected={transit.selectedLineId === line.id}
            data-line-id={line.id}
          >
            <button
              class="list-row-main"
              onclick={() => transit.selectLine(transit.selectedLineId === line.id ? null : line.id)}
            >
              <span class="list-row-label">
                <span class="list-row-modes">
                  {#each line.modes as raw}
                    {@const Icon = modeIconFor(raw)}
                    {#if Icon}
                      <Icon size={14} />
                    {/if}
                  {/each}
                </span>
                <span class="list-row-id">{line.name ?? line.id}</span>
              </span>
              <span class="list-row-meta opacity-0 group-hover:opacity-100">
                {line.routeCount === 1
                  ? $t('transit.routes_count_one')
                  : $t('transit.routes_count_many', { values: { count: line.routeCount } })}
              </span>
            </button>
            <div class="list-row-actions opacity-0 group-hover:opacity-100">
              <button
                onclick={(e) => {
                  e.stopPropagation();
                  transitEdit.beginRename(line);
                }}
                title={$t('transit.rename_line')}
                aria-label={$t('transit.rename_line')}
              >
                <Pencil size={14} />
              </button>
              <button
                onclick={(e) => {
                  e.stopPropagation();
                  transitEdit.requestDeleteLine(line.id);
                }}
                title={$t('transit.delete_line')}
                aria-label={$t('transit.delete_line')}
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        {/if}
      {/each}
    {/if}
  </div>
</section>

{#if transit.selectedLineId !== null}
  <TransitRoutesList />
{/if}

<style>
  .transit-list-section {
    display: flex;
    flex-direction: column;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
  }

  .transit-list-section:not(.transit-routes-section) {
    border-bottom: 1px solid var(--color-base-300);
  }
</style>
