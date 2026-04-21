<script lang="ts">
  import { tick, untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { Bus, ChevronLeft, ChevronRight, Search, TrainFront, TramFront, X } from 'lucide-svelte';
  import { transitStops } from '$lib/stores/transit.svelte';
  import TransitStopDetail from './TransitStopDetail.svelte';

  function modeIconFor(raw: string) {
    const m = raw.toLowerCase();
    if (m === 'bus') return Bus;
    if (m === 'subway' || m === 'metro' || m === 'rail') return TrainFront;
    if (m === 'tram' || m === 'light_rail' || m === 'lightrail' || m === 'streetcar') return TramFront;
    return null;
  }

  const rowEls = new Map<string, HTMLElement>();
  function registerRow(node: HTMLElement, stopId: string) {
    rowEls.set(stopId, node);
    return {
      destroy() {
        if (rowEls.get(stopId) === node) rowEls.delete(stopId);
      },
    };
  }

  $effect(() => {
    const currentTick = transitStops.scrollTick;
    if (currentTick === 0) return;
    untrack(async () => {
      const id = transitStops.expandedStopId;
      if (!id) return;
      await tick();
      const el = rowEls.get(id);
      el?.scrollIntoView({ block: 'center', behavior: 'smooth' });
    });
  });
</script>

<section class="transit-list-section">
  <header class="section-toolbar">
    <h3>{$t('transit.stops_heading')}</h3>
    <div class="section-toolbar-actions">
      <button
        class="toolbar-btn"
        class:active={transitStops.searchOpen}
        onclick={() => transitStops.toggleSearch()}
        title={transitStops.searchOpen ? $t('transit.close_stop_search') : $t('transit.search_stops')}
        aria-label={transitStops.searchOpen ? $t('transit.close_stop_search') : $t('transit.search_stops')}
      >
        {#if transitStops.searchOpen}
          <X size={14} />
        {:else}
          <Search size={14} />
        {/if}
      </button>
    </div>
  </header>

  {#if transitStops.searchOpen}
    <div class="search-row">
      <!-- svelte-ignore a11y_autofocus -->
      <input
        type="text"
        placeholder={$t('transit.search_stop_placeholder')}
        value={transitStops.searchQuery}
        oninput={(e) => transitStops.setSearchQuery((e.target as HTMLInputElement).value)}
        autofocus
      />
    </div>
  {/if}

  <div class="transit-list">
    {#if transitStops.loading && transitStops.items.length === 0}
      <div class="transit-list-empty">…</div>
    {:else if transitStops.error}
      <div class="transit-list-empty transit-error">{transitStops.error}</div>
    {:else if transitStops.items.length === 0}
      <div class="transit-list-empty">{$t('transit.no_stops_in_view')}</div>
    {:else}
      {#each transitStops.items as stop (stop.id)}
        <div
          class="transit-stop-wrapper"
          class:expanded={transitStops.expandedStopId === stop.id}
          use:registerRow={stop.id}
        >
          <button
            class="transit-stop-item"
            class:selected={transitStops.expandedStopId === stop.id}
            onclick={() => transitStops.toggleExpansion(stop.id)}
          >
            <span class="transit-line-modes">
              {#each stop.modes as raw}
                {@const Icon = modeIconFor(raw)}
                {#if Icon}
                  <Icon size={14} />
                {/if}
              {/each}
            </span>
            <span class="transit-stop-name">{stop.name ?? stop.id}</span>
          </button>
          {#if transitStops.expandedStopId === stop.id}
            <TransitStopDetail {stop} />
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  {#if transitStops.total > transitStops.pageSize}
    {@const totalPages = Math.max(1, Math.ceil(transitStops.total / transitStops.pageSize))}
    <div class="transit-pager">
      <button
        class="transit-pager-btn"
        disabled={transitStops.page <= 0}
        onclick={() => transitStops.setPage(transitStops.page - 1)}
        title={$t('transit.prev_page')}
        aria-label={$t('transit.prev_page')}
      >
        <ChevronLeft size={14} />
      </button>
      <span class="transit-pager-label">
        {$t('transit.stops_page_of', { values: { page: transitStops.page + 1, total: totalPages } })}
      </span>
      <button
        class="transit-pager-btn"
        disabled={transitStops.page + 1 >= totalPages}
        onclick={() => transitStops.setPage(transitStops.page + 1)}
        title={$t('transit.next_page')}
        aria-label={$t('transit.next_page')}
      >
        <ChevronRight size={14} />
      </button>
    </div>
  {/if}
</section>

<style>
  .transit-list-section {
    display: flex;
    flex-direction: column;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
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
  .transit-list-empty.transit-error { color: #b91c1c; }

  .transit-line-modes {
    display: flex;
    align-items: center;
    gap: 3px;
    opacity: 0.75;
    flex-shrink: 0;
  }

  .transit-stop-wrapper {
    display: flex;
    flex-direction: column;
  }

  .transit-stop-item {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 8px 14px;
    background: none;
    border: none;
    color: var(--color-base-content);
    text-align: left;
    cursor: pointer;
    transition: background 0.1s;
  }

  .transit-stop-item:hover {
    background: var(--color-base-200);
  }

  .transit-stop-item.selected {
    background: var(--color-base-300);
    box-shadow: inset 2px 0 0 var(--color-primary);
  }

  .transit-stop-name {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-pager {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 8px 12px;
    border-top: 1px solid var(--color-base-300);
    background: var(--color-base-100);
  }

  .transit-pager-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 4px;
    border: 1px solid var(--color-base-300);
    background: var(--color-base-200);
    color: var(--color-base-content);
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s;
  }

  .transit-pager-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-pager-btn:disabled {
    opacity: 0.3;
    cursor: default;
  }

  .transit-pager-label {
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.65);
  }
</style>
