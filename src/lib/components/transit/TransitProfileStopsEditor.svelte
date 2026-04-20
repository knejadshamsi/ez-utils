<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { Check, Plus, Search, Undo2, X } from 'lucide-svelte';

  import { transit, transitEdit } from '$lib/stores/transit.svelte';
  import {
    discardProfileStops,
    runProfileStopPickerSearch,
    saveProfileStops,
  } from '$lib/stores/transit/profile-stops-actions';
  import TransitPathLinksSegmentEditor from './TransitPathLinksSegmentEditor.svelte';
  import TransitProfileStopCard from './TransitProfileStopCard.svelte';
  import TransitStopCreatePanel from './TransitStopCreatePanel.svelte';
  import TransitStopPickerPanel from './TransitStopPickerPanel.svelte';

  $effect(() => {
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    const loading = transit.loadingProfileStops;
    const items = transit.profileStops;
    untrack(() => {
      if (!lineId || !routeId) {
        transitEdit.clearProfileStopsDraft();
        return;
      }
      if (loading) return;
      const current = transitEdit.profileStopsDraft;
      const sameRoute =
        current && current.lineId === lineId && current.routeId === routeId;
      if (!sameRoute) {
        transitEdit.loadProfileStopsDraft(lineId, routeId, items);
      }
    });
  });

  $effect(() => {
    const draft = transitEdit.profileStopsDraft;
    if (!draft || !draft.pickerOpen) return;
    if (draft.pickerQuery.length === 0 && draft.pickerResults.length === 0 && !draft.pickerLoading) {
      void runProfileStopPickerSearch();
    }
  });
</script>

{#if transitEdit.profileStopsDraft}
  {@const d = transitEdit.profileStopsDraft}
  {@const dirty = transitEdit.profileStopsDirty}
  <div class="transit-ps-wrap">
    <div class="transit-ps-header">
      {#if d.saveState === 'error'}
        <span class="transit-ps-status transit-ps-status-err" title={d.saveError ?? ''}>
          {$t('transit.save_failed')}
        </span>
      {:else if d.saveState === 'saving'}
        <span class="transit-ps-status">{$t('transit.saving')}</span>
      {/if}

      {#if dirty || d.saveState === 'saving' || d.saveState === 'error'}
        <button
          class="transit-ps-icon-btn"
          onclick={() => discardProfileStops()}
          disabled={d.saveState === 'saving'}
          title={$t('transit.discard')}
          aria-label={$t('transit.discard')}
        ><Undo2 size={13} /></button>
        <button
          class="transit-ps-icon-btn transit-ps-icon-btn-primary"
          onclick={() => void saveProfileStops()}
          disabled={d.saveState === 'saving'}
          title={$t('transit.save')}
          aria-label={$t('transit.save')}
        ><Check size={14} /></button>
      {/if}

      <span class="transit-ps-spacer"></span>

      <button
        class="transit-ps-btn"
        class:transit-ps-btn-active={d.pickerOpen}
        onclick={() => transitEdit.togglePicker()}
      >
        {#if d.pickerOpen}<X size={13} />{:else}<Search size={13} />{/if}
        {$t('transit.from_existing')}
      </button>
      <button
        class="transit-ps-btn"
        class:transit-ps-btn-active={d.createOpen}
        onclick={() => transitEdit.toggleCreate()}
      >
        {#if d.createOpen}<X size={13} />{:else}<Plus size={13} />{/if}
        {$t('transit.add_new_stop')}
      </button>
    </div>

    <TransitStopCreatePanel />
    <TransitStopPickerPanel />

    <div class="transit-ps-list">
      {#each d.stops as stop, index (stop.key)}
        <TransitProfileStopCard
          {stop}
          {index}
          total={d.stops.length}
          onMoveUp={() => transitEdit.moveProfileStop(index, -1)}
          onMoveDown={() => transitEdit.moveProfileStop(index, 1)}
          onRemove={() => transitEdit.removeProfileStop(index)}
          onFieldChange={(field, value) => transitEdit.updateProfileStopField(index, field, value)}
        />
        {#if index < d.stops.length - 1 && transit.selectedLineId && transit.selectedRouteId}
          <TransitPathLinksSegmentEditor
            fromLinkRef={stop.stopLinkRefId}
            toLinkRef={d.stops[index + 1].stopLinkRefId}
            lineId={transit.selectedLineId}
            routeId={transit.selectedRouteId}
          />
        {/if}
      {/each}
      {#if d.stops.length === 0 && !d.pickerOpen && !d.createOpen}
        <div class="transit-ps-empty">{$t('transit.no_profile_stops')}</div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .transit-ps-wrap {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-height: 0;
  }

  .transit-ps-header {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-base-300);
    background: var(--color-base-100);
  }

  .transit-ps-spacer { flex: 1; }

  .transit-ps-status {
    font-size: 0.7rem;
    font-weight: 600;
    opacity: 0.75;
    white-space: nowrap;
  }

  .transit-ps-status-err {
    color: #ef4444;
    opacity: 1;
  }

  .transit-ps-btn {
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
    transition: background 0.1s, border-color 0.1s, color 0.1s;
  }

  .transit-ps-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-ps-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .transit-ps-btn-active {
    background: var(--color-base-300);
  }

  .transit-ps-icon-btn {
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
    transition: background 0.1s, opacity 0.1s, color 0.1s, border-color 0.1s;
  }

  .transit-ps-icon-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-300);
  }

  .transit-ps-icon-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .transit-ps-icon-btn-primary {
    color: var(--color-primary);
    border-color: var(--color-primary);
  }

  .transit-ps-icon-btn-primary:hover:not(:disabled) {
    background: var(--color-primary);
    color: var(--color-base-100);
  }

  .transit-ps-list {
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 8px 12px 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .transit-ps-empty {
    padding: 20px;
    text-align: center;
    font-size: 0.76rem;
    color: oklch(var(--bc) / 0.5);
  }
</style>
