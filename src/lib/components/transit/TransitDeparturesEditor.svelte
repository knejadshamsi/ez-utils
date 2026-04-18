<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { Check, Plus, Trash2, Undo2 } from 'lucide-svelte';

  import { sources } from '$lib/stores/data.svelte';
  import { transit, transitDeparturesEdit } from '$lib/stores/transit.svelte';
  import {
    discardDepartures,
    loadVehiclesForPicker,
    saveDepartures,
  } from '$lib/stores/transit/departures-actions';

  $effect(() => {
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    const loading = transit.loadingDepartures;
    const items = transit.departures;
    untrack(() => {
      if (!lineId || !routeId) {
        transitDeparturesEdit.clear();
        return;
      }
      if (loading) return;
      const current = transitDeparturesEdit.draft;
      const sameRoute =
        current && current.lineId === lineId && current.routeId === routeId;
      if (!sameRoute) {
        transitDeparturesEdit.load(lineId, routeId, items);
      }
    });
  });

  // Load vehicles whenever the draft exists, AND also when the attachment
  // state of the vehicles file toggles mid-session (user attaches while the
  // route detail is already open). The action gates re-fetches by comparing
  // against the last-loaded state so this effect is cheap to re-run.
  $effect(() => {
    const draft = transitDeparturesEdit.draft;
    const hasFile = !!sources.active?.transitMetadata?.vehicles;
    if (!draft) return;
    void hasFile;
    void loadVehiclesForPicker();
  });
</script>

{#if transitDeparturesEdit.draft}
  {@const d = transitDeparturesEdit.draft}
  {@const dirty = transitDeparturesEdit.dirty}
  <div class="transit-dep-wrap">
    <div class="transit-dep-header">
      {#if d.saveState === 'error'}
        <span class="transit-dep-status transit-dep-status-err" title={d.saveError ?? ''}>
          {d.saveError ?? $t('transit.save_failed')}
        </span>
      {:else if d.saveState === 'saving'}
        <span class="transit-dep-status">{$t('transit.saving')}</span>
      {/if}

      {#if dirty || d.saveState === 'saving' || d.saveState === 'error'}
        <button
          class="transit-dep-icon-btn"
          onclick={() => discardDepartures()}
          disabled={d.saveState === 'saving'}
          title={$t('transit.discard')}
          aria-label={$t('transit.discard')}
        ><Undo2 size={13} /></button>
        <button
          class="transit-dep-icon-btn transit-dep-icon-btn-primary"
          onclick={() => void saveDepartures()}
          disabled={d.saveState === 'saving'}
          title={$t('transit.save')}
          aria-label={$t('transit.save')}
        ><Check size={14} /></button>
      {/if}

      <span class="transit-dep-spacer"></span>

      <button class="transit-dep-btn" onclick={() => transitDeparturesEdit.add()}>
        <Plus size={13} /> {$t('transit.add_departure')}
      </button>
    </div>

    {#if d.departures.length === 0}
      <div class="transit-dep-empty">{$t('transit.no_departures')}</div>
    {:else}
      <div class="transit-dep-table">
        <div class="transit-dep-head-row">
          <span class="transit-dep-col-id">ID</span>
          <span class="transit-dep-col-time">{$t('transit.departure_time')}</span>
          <span class="transit-dep-col-veh">{$t('transit.departure_veh')}</span>
          <span class="transit-dep-col-action"></span>
        </div>
        {#each d.departures as dep, index (dep.key)}
          <div class="transit-dep-data-row">
            <input
              class="transit-dep-input transit-dep-col-id"
              type="text"
              value={dep.id}
              oninput={(e) => transitDeparturesEdit.updateField(index, 'id', (e.target as HTMLInputElement).value)}
            />
            <input
              class="transit-dep-input transit-dep-col-time"
              type="text"
              placeholder="HH:MM:SS"
              value={dep.departureTime}
              oninput={(e) => transitDeparturesEdit.updateField(index, 'departureTime', (e.target as HTMLInputElement).value)}
            />
            <input
              class="transit-dep-input transit-dep-col-veh"
              type="text"
              list={d.vehicles.length > 0 ? 'transit-dep-vehicle-list' : undefined}
              placeholder={d.vehicles.length > 0 ? $t('transit.vehicle_placeholder') : $t('transit.optional')}
              value={dep.vehicleRefId}
              oninput={(e) => transitDeparturesEdit.updateField(index, 'vehicleRefId', (e.target as HTMLInputElement).value)}
            />
            <button
              class="transit-dep-icon-btn transit-dep-icon-btn-danger transit-dep-col-action"
              onclick={() => transitDeparturesEdit.remove(index)}
              title={$t('transit.remove_departure')}
              aria-label={$t('transit.remove_departure')}
            ><Trash2 size={13} /></button>
          </div>
        {/each}
      </div>
    {/if}

    {#if d.vehicles.length > 0}
      <datalist id="transit-dep-vehicle-list">
        {#each d.vehicles as v (v.id)}
          <option value={v.id}>
            {v.vehicleType ? `${v.vehicleType} · ${v.id}` : v.id}{v.referenced ? ' · in use' : ''}
          </option>
        {/each}
      </datalist>
    {/if}
  </div>
{/if}

<style>
  .transit-dep-wrap {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-height: 0;
  }

  .transit-dep-header {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-base-300);
    background: var(--color-base-100);
  }

  .transit-dep-spacer { flex: 1; }

  .transit-dep-status {
    font-size: 0.7rem;
    font-weight: 600;
    opacity: 0.75;
    white-space: nowrap;
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-dep-status-err { color: #ef4444; opacity: 1; }

  .transit-dep-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 5px 10px;
    font-size: 0.75rem;
    font-weight: 600;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
    cursor: pointer;
  }

  .transit-dep-btn:hover:not(:disabled) { background: var(--color-base-300); }

  .transit-dep-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
    opacity: 0.85;
    cursor: pointer;
  }

  .transit-dep-icon-btn:hover:not(:disabled) { opacity: 1; background: var(--color-base-300); }
  .transit-dep-icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
  .transit-dep-icon-btn-primary { color: var(--color-primary); border-color: var(--color-primary); }
  .transit-dep-icon-btn-danger:hover:not(:disabled) { color: #ef4444; border-color: #ef4444; }

  .transit-dep-empty {
    padding: 20px;
    text-align: center;
    font-size: 0.76rem;
    color: oklch(var(--bc) / 0.5);
  }

  .transit-dep-table {
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 4px 10px 10px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .transit-dep-head-row,
  .transit-dep-data-row {
    display: grid;
    grid-template-columns: 1fr 90px 1.2fr 26px;
    gap: 6px;
    align-items: center;
  }

  .transit-dep-head-row {
    padding: 6px 4px 4px;
    font-size: 0.62rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: oklch(var(--bc) / 0.55);
    border-bottom: 1px solid var(--color-base-300);
    position: sticky;
    top: 0;
    background: var(--color-base-100);
    z-index: 1;
  }

  .transit-dep-data-row {
    padding: 2px 0;
  }

  .transit-dep-input {
    padding: 3px 6px;
    font-size: 0.78rem;
    background: var(--color-base-100);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 3px;
    outline: none;
    min-width: 0;
  }
  .transit-dep-input:focus { border-color: var(--color-primary); }

  .transit-dep-col-action {
    justify-self: center;
  }
</style>
