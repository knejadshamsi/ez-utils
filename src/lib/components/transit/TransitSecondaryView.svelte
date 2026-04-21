<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { sources } from '$lib/stores/data.svelte';
  import { transit } from '$lib/stores/transit.svelte';
  import TransitDeparturesEditor from './TransitDeparturesEditor.svelte';
  import TransitProfileStopsEditor from './TransitProfileStopsEditor.svelte';

  $effect(() => {
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    const activeName = sources.active?.kind === 'transit' ? sources.activeName : null;
    untrack(() => {
      if (activeName && lineId && routeId) {
        void transit.loadProfileStopsForRoute(activeName, lineId, routeId);
        void transit.loadPathLinksForRoute(activeName, lineId, routeId);
        void transit.loadDeparturesForRoute(activeName, lineId, routeId);
      }
    });
  });
</script>

<div class="transit-secondary">
  {#if !transit.selectedRouteId}
    <div class="transit-secondary-empty">{$t('transit.no_route_selected')}</div>
  {:else}
    {@const route = transit.selectedRoute}
    <header class="transit-secondary-header">
      <h2>{route?.description ?? route?.id ?? transit.selectedRouteId}</h2>
    </header>

    <div class="transit-tab-bar">
      <button
        class="transit-tab"
        class:transit-tab-active={transit.secondaryTab === 'stops'}
        onclick={() => transit.setSecondaryTab('stops')}
      >
        {$t('transit.tab_stops')}
        <span class="transit-tab-count">{transit.profileStops.length}</span>
      </button>
      <button
        class="transit-tab"
        class:transit-tab-active={transit.secondaryTab === 'departures'}
        onclick={() => transit.setSecondaryTab('departures')}
      >
        {$t('transit.tab_departures')}
        <span class="transit-tab-count">{transit.departures.length}</span>
      </button>
    </div>

    {#if transit.secondaryTab === 'stops'}
      {#if transit.loadingProfileStops && transit.profileStops.length === 0}
        <div class="transit-secondary-empty">…</div>
      {:else if transit.profileStopsError}
        <div class="transit-secondary-empty transit-secondary-error">{transit.profileStopsError}</div>
      {:else}
        <TransitProfileStopsEditor />
      {/if}
    {:else}
      {#if transit.loadingDepartures && transit.departures.length === 0}
        <div class="transit-secondary-empty">…</div>
      {:else if transit.departuresError}
        <div class="transit-secondary-empty transit-secondary-error">{transit.departuresError}</div>
      {:else}
        <TransitDeparturesEditor />
      {/if}
    {/if}
  {/if}
</div>

<style>
  .transit-secondary {
    height: 100%;
    display: flex;
    flex-direction: column;
    color: var(--color-base-content);
  }

  .transit-secondary-header {
    padding: 14px 16px;
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-secondary-header h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 700;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-tab-bar {
    display: flex;
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-tab {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 12px;
    border: none;
    background: none;
    color: var(--color-base-content);
    font-size: 0.82rem;
    font-weight: 600;
    opacity: 0.5;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: opacity 0.1s, border-color 0.1s, background 0.1s;
  }

  .transit-tab:hover {
    opacity: 0.85;
    background: var(--color-base-200);
  }

  .transit-tab-active {
    opacity: 1;
    border-bottom-color: var(--color-primary);
  }

  .transit-tab-count {
    font-size: 0.68rem;
    padding: 1px 7px;
    border-radius: 10px;
    background: var(--color-base-200);
    color: oklch(var(--bc) / 0.7);
    font-weight: 600;
  }

  .transit-tab-active .transit-tab-count {
    background: var(--color-base-300);
  }

  .transit-secondary-empty {
    padding: 16px;
    text-align: center;
    font-size: 0.78rem;
    color: oklch(var(--bc) / 0.5);
  }

  .transit-secondary-empty.transit-secondary-error {
    color: #b91c1c;
  }
</style>
