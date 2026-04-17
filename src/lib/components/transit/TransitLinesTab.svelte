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
  <header class="transit-list-header">
    <h3>{$t('transit.lines_heading')}</h3>
    <div class="transit-header-actions">
      <button
        class="transit-icon-btn"
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
        class="transit-icon-btn"
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
    <div class="transit-search-row">
      <!-- svelte-ignore a11y_autofocus -->
      <input
        type="text"
        class="transit-search-input"
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

  <div class="transit-list">
    {#if transit.loadingLines && transit.lines.length === 0}
      <div class="transit-list-empty">…</div>
    {:else if transit.linesError}
      <div class="transit-list-empty transit-error">{transit.linesError}</div>
    {:else if transit.filteredLines.length === 0}
      <div class="transit-list-empty">{$t('transit.no_lines')}</div>
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
            class="transit-line-shell"
            class:selected={transit.selectedLineId === line.id}
            data-line-id={line.id}
          >
            <button
              class="transit-line-item"
              onclick={() => transit.selectLine(transit.selectedLineId === line.id ? null : line.id)}
            >
              <span class="transit-line-modes">
                {#each line.modes as raw}
                  {@const Icon = modeIconFor(raw)}
                  {#if Icon}
                    <Icon size={16} />
                  {/if}
                {/each}
              </span>
              <span class="transit-line-name">{line.name ?? line.id}</span>
            </button>
            <div class="transit-line-actions">
              <button
                class="transit-icon-btn"
                onclick={(e) => {
                  e.stopPropagation();
                  transitEdit.beginRename(line);
                }}
                title={$t('transit.rename_line')}
                aria-label={$t('transit.rename_line')}
              >
                <Pencil size={13} />
              </button>
              <button
                class="transit-icon-btn transit-icon-btn-danger"
                onclick={(e) => {
                  e.stopPropagation();
                  transitEdit.requestDeleteLine(line.id);
                }}
                title={$t('transit.delete_line')}
                aria-label={$t('transit.delete_line')}
              >
                <Trash2 size={13} />
              </button>
            </div>
            <span class="transit-line-count">{line.routeCount}</span>
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

  .transit-list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    background: var(--color-base-200);
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-list-header h3 {
    margin: 0;
    font-size: 0.78rem;
    font-weight: 700;
    letter-spacing: 0.02em;
  }

  .transit-header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .transit-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 4px;
    border: none;
    background: none;
    color: var(--color-base-content);
    opacity: 0.6;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s, color 0.1s;
  }

  .transit-icon-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-300);
  }

  .transit-icon-btn.active {
    opacity: 1;
    background: var(--color-base-300);
    color: var(--color-primary);
  }

  .transit-icon-btn-danger:hover:not(:disabled) {
    color: #ef4444;
  }

  .transit-search-row {
    padding: 6px 10px;
    background: var(--color-base-100);
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-search-input {
    width: 100%;
    padding: 4px 8px;
    font-size: 0.78rem;
    background: var(--color-base-200);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    outline: none;
  }

  .transit-search-input:focus {
    border-color: var(--color-primary);
  }

  .transit-list {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding: 4px 0;
  }

  .transit-list-empty {
    padding: 12px 16px;
    text-align: center;
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.5);
  }

  .transit-list-empty.transit-error {
    color: #b91c1c;
  }

  .transit-line-shell {
    display: flex;
    align-items: center;
    width: 100%;
    transition: background 0.1s;
  }

  .transit-line-shell:hover {
    background: var(--color-base-200);
  }

  .transit-line-shell.selected {
    background: var(--color-base-300);
    box-shadow: inset 2px 0 0 var(--color-primary);
  }

  .transit-line-item {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1 1 auto;
    min-width: 0;
    padding: 9px 14px;
    background: none;
    border: none;
    color: var(--color-base-content);
    text-align: left;
    cursor: pointer;
    transition: background 0.1s;
  }

  .transit-line-actions {
    display: none;
    align-items: center;
    gap: 2px;
    flex-shrink: 0;
  }

  .transit-line-shell:hover .transit-line-actions,
  .transit-line-shell.selected .transit-line-actions {
    display: flex;
  }

  .transit-line-modes {
    display: flex;
    align-items: center;
    gap: 3px;
    opacity: 0.75;
    flex-shrink: 0;
  }

  .transit-line-name {
    flex: 1;
    min-width: 0;
    font-size: 0.88rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-line-count {
    flex-shrink: 0;
    margin-left: 6px;
    margin-right: 14px;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--color-base-200);
    font-size: 0.75rem;
    font-weight: 600;
    color: oklch(var(--bc) / 0.7);
  }
</style>
