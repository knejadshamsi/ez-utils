<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Bus, Pencil, TrainFront, TramFront, Trash2 } from 'lucide-svelte';

  import type { ListedStop } from '$lib/stores/transit/types';
  import { transitStops, transitEdit } from '$lib/stores/transit.svelte';
  import TransitFacilityEditForm from './TransitFacilityEditForm.svelte';
  import TransitStopTransfersEditor from './TransitStopTransfersEditor.svelte';

  interface Props {
    stop: ListedStop;
  }

  let { stop }: Props = $props();

  const isEditing = $derived(
    transitEdit.editStopFacilityTarget?.originalId === stop.id,
  );

  function modeIconFor(raw: string) {
    const m = raw.toLowerCase();
    if (m === 'bus') return Bus;
    if (m === 'subway' || m === 'metro' || m === 'rail') return TrainFront;
    if (m === 'tram' || m === 'light_rail' || m === 'lightrail' || m === 'streetcar') return TramFront;
    return null;
  }

</script>

<div class="transit-stop-detail">
  {#if isEditing}
    <TransitFacilityEditForm />
  {:else}
    <div class="transit-stop-detail-row">
      <span class="transit-stop-detail-key">ID</span>
      <span class="transit-stop-detail-val mono">{stop.id}</span>
    </div>
    <div class="transit-stop-detail-row">
      <span class="transit-stop-detail-key">{$t('transit.stop_coords')}</span>
      <span class="transit-stop-detail-val mono">{stop.lng.toFixed(6)}, {stop.lat.toFixed(6)}</span>
    </div>

    <div class="transit-stop-detail-actions">
      <button
        class="transit-stop-edit-inline-btn"
        onclick={() => transitEdit.beginEditStopFacility({ id: stop.id, name: stop.name, lng: stop.lng, lat: stop.lat, linkRefId: stop.linkRefId, stopAreaId: stop.stopAreaId, isBlocking: stop.isBlocking })}
        title={$t('transit.edit_stop_facility')}
        aria-label={$t('transit.edit_stop_facility')}
      >
        <Pencil size={12} />
        <span>{$t('transit.edit_stop_facility')}</span>
      </button>
      <button
        class="transit-stop-delete-btn"
        onclick={() => transitEdit.requestDeleteStopFacility(stop.id)}
        title={$t('transit.delete_stop_facility')}
        aria-label={$t('transit.delete_stop_facility')}
      >
        <Trash2 size={12} />
        <span>{$t('transit.delete_stop_facility')}</span>
      </button>
    </div>
  {/if}

  <div class="transit-stop-detail-section">
    <div class="transit-stop-detail-section-title">{$t('transit.stop_lines_used')}</div>
    {#if transitStops.detail.linesLoading}
      <div class="transit-stop-detail-muted">…</div>
    {:else if transitStops.detail.linesError}
      <div class="transit-stop-detail-muted transit-error">{transitStops.detail.linesError}</div>
    {:else if transitStops.detail.lines.length === 0}
      <div class="transit-stop-detail-muted">{$t('transit.stop_none_lines')}</div>
    {:else}
      {#each transitStops.detail.lines as line (line.lineId)}
        <div class="transit-stop-detail-item">
          <span class="transit-line-modes">
            {#each line.modes as raw}
              {@const Icon = modeIconFor(raw)}
              {#if Icon}<Icon size={12} />{/if}
            {/each}
          </span>
          <span class="transit-line-name">{line.lineName ?? line.lineId}</span>
          <span class="transit-line-count">{line.routeCount}</span>
        </div>
      {/each}
    {/if}
  </div>

  <div class="transit-stop-detail-section">
    <div class="transit-stop-detail-section-title">{$t('transit.stop_routes_used')}</div>
    {#if transitStops.detail.routesLoading}
      <div class="transit-stop-detail-muted">…</div>
    {:else if transitStops.detail.routesError}
      <div class="transit-stop-detail-muted transit-error">{transitStops.detail.routesError}</div>
    {:else if transitStops.detail.routes.length === 0}
      <div class="transit-stop-detail-muted">{$t('transit.stop_none_routes')}</div>
    {:else}
      {#each transitStops.detail.routes as r (r.lineId + '|' + r.routeId)}
        {@const Icon = modeIconFor(r.transportMode)}
        <div class="transit-stop-detail-item">
          <span class="transit-line-modes">
            {#if Icon}<Icon size={12} />{/if}
          </span>
          <span class="transit-line-name">{r.description ?? r.routeId}</span>
        </div>
      {/each}
    {/if}
  </div>

  <TransitStopTransfersEditor stopId={stop.id} />
</div>

<style>
  .transit-stop-detail {
    background: var(--color-base-200);
    padding: 10px 14px 14px;
    border-bottom: 1px solid var(--color-base-300);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .transit-stop-detail-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.72rem;
  }

  .transit-stop-detail-key {
    flex-shrink: 0;
    width: 90px;
    color: oklch(var(--bc) / 0.55);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
    font-size: 0.65rem;
  }

  .transit-stop-detail-val {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .transit-stop-detail-val.mono { font-family: monospace; }

  .transit-stop-detail-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
    padding-top: 2px;
  }

  .transit-stop-edit-inline-btn,
  .transit-stop-delete-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 8px;
    background: none;
    border: 1px solid transparent;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 600;
    cursor: pointer;
    opacity: 0.8;
    transition: border-color 0.1s, opacity 0.1s;
  }

  .transit-stop-edit-inline-btn {
    color: var(--color-base-content);
    opacity: 0.75;
  }

  .transit-stop-edit-inline-btn:hover {
    opacity: 1;
    border-color: var(--color-base-300);
  }

  .transit-stop-delete-btn {
    color: #ef4444;
  }

  .transit-stop-delete-btn:hover {
    opacity: 1;
    border-color: #ef4444;
  }


  .transit-stop-detail-section {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .transit-stop-detail-section-title {
    font-size: 0.65rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: oklch(var(--bc) / 0.55);
    margin-bottom: 2px;
  }

  .transit-stop-detail-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 0;
    font-size: 0.78rem;
  }

  .transit-stop-detail-muted {
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.5);
    font-style: italic;
  }

  .transit-stop-detail-muted.transit-error { color: #b91c1c; }

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
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-line-count {
    flex-shrink: 0;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--color-base-200);
    font-size: 0.75rem;
    font-weight: 600;
    color: oklch(var(--bc) / 0.7);
  }
</style>
