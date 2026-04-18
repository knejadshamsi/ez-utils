<script lang="ts">
  import { t } from 'svelte-i18n';
  import { ChevronDown, ChevronUp, Pencil, Trash2 } from 'lucide-svelte';

  import type { EditableProfileStop } from '$lib/stores/transit/types';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import TransitFacilityEditForm from './TransitFacilityEditForm.svelte';

  interface Props {
    stop: EditableProfileStop;
    index: number;
    total: number;
    onMoveUp: () => void;
    onMoveDown: () => void;
    onRemove: () => void;
    onFieldChange: <K extends keyof EditableProfileStop>(field: K, value: EditableProfileStop[K]) => void;
  }

  let { stop, index, total, onMoveUp, onMoveDown, onRemove, onFieldChange }: Props = $props();

  const isEditingFacility = $derived(
    transitEdit.editStopFacilityTarget?.originalId === stop.stopRefId,
  );

  function startEditFacility() {
    if (stop.stopLng === null || stop.stopLat === null) return;
    transitEdit.beginEditStopFacility({
      id: stop.stopRefId,
      name: stop.stopName,
      lng: stop.stopLng,
      lat: stop.stopLat,
      linkRefId: stop.stopLinkRefId,
      stopAreaId: stop.stopAreaId,
      isBlocking: stop.stopIsBlocking ?? false,
    });
  }

</script>

<article class="transit-ps-card">
  <header class="transit-ps-card-head">
    <span class="transit-ps-seq">{index + 1}</span>
    <div class="transit-ps-title">
      <div class="transit-ps-name">{stop.stopName ?? stop.stopRefId}</div>
      <div class="transit-ps-meta">
        <span class="transit-ps-id" title={stop.stopRefId}>{stop.stopRefId}</span>
        {#if stop.stopLng !== null && stop.stopLat !== null}
          <span class="transit-ps-coords">
            {stop.stopLng.toFixed(5)}, {stop.stopLat.toFixed(5)}
          </span>
        {/if}
      </div>
    </div>
    <button
      class="transit-ps-icon-btn"
      class:transit-ps-icon-btn-active={isEditingFacility}
      title={$t('transit.edit_stop_facility')}
      aria-label={$t('transit.edit_stop_facility')}
      onclick={() => (isEditingFacility ? transitEdit.cancelEditStopFacility() : startEditFacility())}
    ><Pencil size={13} /></button>
    <button
      class="transit-ps-icon-btn"
      disabled={index === 0}
      title={$t('transit.move_up')}
      aria-label={$t('transit.move_up')}
      onclick={() => onMoveUp()}
    ><ChevronUp size={14} /></button>
    <button
      class="transit-ps-icon-btn"
      disabled={index === total - 1}
      title={$t('transit.move_down')}
      aria-label={$t('transit.move_down')}
      onclick={() => onMoveDown()}
    ><ChevronDown size={14} /></button>
    <button
      class="transit-ps-icon-btn transit-ps-icon-btn-danger"
      title={$t('transit.remove_from_route')}
      aria-label={$t('transit.remove_from_route')}
      onclick={() => onRemove()}
    ><Trash2 size={13} /></button>
  </header>

  <div class="transit-ps-row">
    <label class="transit-ps-field">
      <span>{$t('transit.arr')}</span>
      <input
        type="text"
        placeholder="HH:MM:SS"
        value={stop.arrivalOffset}
        oninput={(e) => onFieldChange('arrivalOffset', (e.target as HTMLInputElement).value)}
      />
    </label>
    <label class="transit-ps-field">
      <span>{$t('transit.dep')}</span>
      <input
        type="text"
        placeholder="HH:MM:SS"
        value={stop.departureOffset}
        oninput={(e) => onFieldChange('departureOffset', (e.target as HTMLInputElement).value)}
      />
    </label>
  </div>

  {#if isEditingFacility}
    <TransitFacilityEditForm />
  {/if}

  <div class="transit-ps-toggles">
    <button
      class="transit-ps-toggle"
      class:transit-ps-toggle-on={stop.allowBoarding}
      aria-pressed={stop.allowBoarding}
      onclick={() => onFieldChange('allowBoarding', !stop.allowBoarding)}
    >{$t('transit.flag_boarding')}</button>
    <button
      class="transit-ps-toggle"
      class:transit-ps-toggle-on={stop.allowAlighting}
      aria-pressed={stop.allowAlighting}
      onclick={() => onFieldChange('allowAlighting', !stop.allowAlighting)}
    >{$t('transit.flag_alighting')}</button>
    <button
      class="transit-ps-toggle"
      class:transit-ps-toggle-on={stop.awaitDeparture}
      aria-pressed={stop.awaitDeparture}
      onclick={() => onFieldChange('awaitDeparture', !stop.awaitDeparture)}
    >{$t('transit.flag_await_departure')}</button>
  </div>
</article>

<style>
  .transit-ps-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 8px 10px;
    background: var(--color-base-200);
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
  }

  .transit-ps-card-head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .transit-ps-seq {
    flex-shrink: 0;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--color-base-300);
    color: oklch(var(--bc) / 0.75);
    font-size: 0.7rem;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .transit-ps-title {
    flex: 1;
    min-width: 0;
  }

  .transit-ps-name {
    font-size: 0.82rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-ps-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.68rem;
    opacity: 0.65;
    min-width: 0;
  }

  .transit-ps-id {
    font-family: monospace;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 40%;
  }

  .transit-ps-coords {
    font-family: monospace;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .transit-ps-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    background: none;
    color: var(--color-base-content);
    border: none;
    border-radius: 4px;
    opacity: 0.6;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s, color 0.1s;
  }

  .transit-ps-icon-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-300);
  }

  .transit-ps-icon-btn:disabled {
    opacity: 0.2;
    cursor: not-allowed;
  }

  .transit-ps-icon-btn-danger:hover:not(:disabled) {
    color: #ef4444;
  }

  .transit-ps-icon-btn-active {
    opacity: 1;
    color: var(--color-primary);
    background: var(--color-base-300);
  }


  .transit-ps-row {
    display: flex;
    gap: 8px;
    align-items: flex-end;
  }

  .transit-ps-toggles {
    display: flex;
    gap: 4px;
  }

  .transit-ps-field {
    display: flex;
    flex-direction: column;
    gap: 3px;
    flex: 1 1 0;
    min-width: 0;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.7;
  }

  .transit-ps-field input {
    padding: 4px 8px;
    font-size: 0.78rem;
    font-weight: 400;
    text-transform: none;
    letter-spacing: normal;
    background: var(--color-base-100);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    outline: none;
    opacity: 1;
  }

  .transit-ps-field input:focus {
    border-color: var(--color-primary);
  }

  .transit-ps-toggle {
    padding: 1px 6px;
    font-size: 0.62rem;
    font-weight: 600;
    letter-spacing: 0.01em;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    opacity: 0.5;
    cursor: pointer;
    white-space: nowrap;
    transition: opacity 0.1s, background 0.1s, color 0.1s, border-color 0.1s;
  }

  .transit-ps-toggle:hover {
    opacity: 0.85;
    background: var(--color-base-300);
  }

  .transit-ps-toggle.transit-ps-toggle-on {
    opacity: 1;
    color: var(--color-primary);
    border-color: var(--color-primary);
    background: oklch(var(--p) / 0.1);
  }

  .transit-ps-toggle.transit-ps-toggle-on:hover {
    background: oklch(var(--p) / 0.18);
  }
</style>
