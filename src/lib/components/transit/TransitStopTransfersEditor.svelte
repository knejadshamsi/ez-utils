<script lang="ts">
  import { invoke } from '@tauri-apps/api/core';
  import { t } from 'svelte-i18n';
  import { Check, Plus, Trash2, X } from 'lucide-svelte';

  import { sources } from '$lib/stores/data.svelte';
  import { transitStops } from '$lib/stores/transit.svelte';
  import type { ListedStop } from '$lib/stores/transit/types';

  interface Props {
    stopId: string;
  }

  let { stopId }: Props = $props();

  // In-progress per-row edits.
  let editingKey = $state<string | null>(null);
  let editingValue = $state('');
  let busy = $state(false);
  let error = $state<string | null>(null);

  // Add-transfer form state.
  let addOpen = $state(false);
  let addQuery = $state('');
  let addResults = $state<ListedStop[]>([]);
  let addPickedStop = $state<ListedStop | null>(null);
  let addTime = $state('');
  let addSearching = $state(false);
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  function activeSource(): string | null {
    const a = sources.active;
    return a && a.kind === 'transit' ? a.name : null;
  }

  async function refresh() {
    transitStops.toggleExpansion(stopId);
    transitStops.toggleExpansion(stopId);
  }

  async function saveEditedTime(fromStop: string, toStop: string) {
    const sourceName = activeSource();
    if (!sourceName) return;
    const seconds = Number(editingValue);
    if (!Number.isFinite(seconds) || seconds < 0) {
      error = $t('transit.transfer_time_invalid');
      return;
    }
    busy = true;
    error = null;
    try {
      await invoke('upsert_transit_transfer_cmd', {
        sourceName, fromStop, toStop, transferTime: seconds,
      });
      editingKey = null;
      editingValue = '';
      await refresh();
    } catch (e) {
      const err = e as { message?: string };
      error = err?.message ?? 'Save failed';
    } finally {
      busy = false;
    }
  }

  async function deleteTransfer(fromStop: string, toStop: string) {
    const sourceName = activeSource();
    if (!sourceName) return;
    busy = true;
    error = null;
    try {
      await invoke('delete_transit_transfer_cmd', { sourceName, fromStop, toStop });
      await refresh();
    } catch (e) {
      const err = e as { message?: string };
      error = err?.message ?? 'Delete failed';
    } finally {
      busy = false;
    }
  }

  function setAddQuery(q: string) {
    addQuery = q;
    addPickedStop = null;
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => { void runAddSearch(); }, 200);
  }

  async function runAddSearch() {
    const sourceName = activeSource();
    if (!sourceName) return;
    addSearching = true;
    try {
      const results = await invoke<ListedStop[]>('search_transit_stops_cmd', {
        sourceName, query: addQuery.trim().length > 0 ? addQuery.trim() : null, limit: 20,
      });
      addResults = results.filter((s) => s.id !== stopId);
    } catch {
      addResults = [];
    } finally {
      addSearching = false;
    }
  }

  async function saveNewTransfer() {
    const sourceName = activeSource();
    if (!sourceName || !addPickedStop) return;
    const seconds = Number(addTime);
    if (!Number.isFinite(seconds) || seconds < 0) {
      error = $t('transit.transfer_time_invalid');
      return;
    }
    busy = true;
    error = null;
    try {
      await invoke('upsert_transit_transfer_cmd', {
        sourceName, fromStop: stopId, toStop: addPickedStop.id, transferTime: seconds,
      });
      addOpen = false;
      addQuery = '';
      addPickedStop = null;
      addTime = '';
      addResults = [];
      await refresh();
    } catch (e) {
      const err = e as { message?: string };
      error = err?.message ?? 'Save failed';
    } finally {
      busy = false;
    }
  }

  function keyFor(direction: string, otherStopId: string): string {
    return `${direction}|${otherStopId}`;
  }

  function beginEdit(direction: string, otherStopId: string, currentTime: number) {
    editingKey = keyFor(direction, otherStopId);
    editingValue = String(currentTime);
    error = null;
  }
</script>

<div class="transit-stop-detail-section">
  <div class="transit-stop-detail-section-title">{$t('transit.stop_transfers')}</div>
  {#if transitStops.detail.transfersLoading}
    <div class="transit-stop-detail-muted">…</div>
  {:else if transitStops.detail.transfersError}
    <div class="transit-stop-detail-muted transit-error">{transitStops.detail.transfersError}</div>
  {:else if transitStops.detail.transfers.length === 0}
    <div class="transit-stop-detail-muted">{$t('transit.stop_none_transfers')}</div>
  {:else}
    {#each transitStops.detail.transfers as tr (keyFor(tr.direction, tr.otherStopId))}
      {@const fromStop = tr.direction === 'from' ? stopId : tr.otherStopId}
      {@const toStop = tr.direction === 'from' ? tr.otherStopId : stopId}
      {@const isEditing = editingKey === keyFor(tr.direction, tr.otherStopId)}
      <div class="transit-stop-transfer-row">
        <span class="transit-stop-transfer-dir">
          {tr.direction === 'from' ? $t('transit.transfer_from') : $t('transit.transfer_to')}
        </span>
        <span class="transit-line-name">{tr.otherStopName ?? tr.otherStopId}</span>
        {#if isEditing}
          <input
            class="transit-transfer-time-input"
            type="number"
            min="0"
            value={editingValue}
            oninput={(e) => (editingValue = (e.target as HTMLInputElement).value)}
            disabled={busy}
          />
          <button class="transit-transfer-icon-btn" onclick={() => void saveEditedTime(fromStop, toStop)} disabled={busy} title={$t('transit.save')}>
            <Check size={11} />
          </button>
          <button class="transit-transfer-icon-btn" onclick={() => { editingKey = null; }} disabled={busy} title={$t('transit.cancel')}>
            <X size={11} />
          </button>
        {:else}
          <button class="transit-transfer-time" onclick={() => beginEdit(tr.direction, tr.otherStopId, tr.transferTime)}>
            {tr.transferTime}{$t('transit.seconds_short')}
          </button>
          <button class="transit-transfer-icon-btn transit-transfer-del" onclick={() => void deleteTransfer(fromStop, toStop)} disabled={busy} title={$t('transit.delete_transfer')}>
            <Trash2 size={11} />
          </button>
        {/if}
      </div>
    {/each}
  {/if}

  {#if error}<div class="transit-transfer-error">{error}</div>{/if}

  {#if !addOpen}
    <button class="transit-transfer-add-btn" onclick={() => { addOpen = true; void runAddSearch(); }}>
      <Plus size={11} /> {$t('transit.add_transfer')}
    </button>
  {:else}
    <div class="transit-transfer-add-panel">
      <input
        class="transit-transfer-add-search"
        type="text"
        placeholder={$t('transit.search_stop_placeholder')}
        value={addQuery}
        oninput={(e) => setAddQuery((e.target as HTMLInputElement).value)}
      />
      {#if addPickedStop}
        <div class="transit-transfer-add-picked">
          → {addPickedStop.name ?? addPickedStop.id}
          <span class="transit-transfer-add-picked-id">{addPickedStop.id}</span>
        </div>
      {:else if addResults.length > 0}
        <div class="transit-transfer-add-results">
          {#each addResults.slice(0, 10) as s (s.id)}
            <button class="transit-transfer-add-result" onclick={() => { addPickedStop = s; }}>
              {s.name ?? s.id} <span class="transit-transfer-add-picked-id">{s.id}</span>
            </button>
          {/each}
        </div>
      {:else if addSearching}
        <div class="transit-stop-detail-muted">…</div>
      {/if}
      <div class="transit-transfer-add-row">
        <input
          class="transit-transfer-time-input"
          type="number"
          min="0"
          placeholder={$t('transit.seconds_placeholder')}
          value={addTime}
          oninput={(e) => (addTime = (e.target as HTMLInputElement).value)}
        />
        <button class="transit-transfer-icon-btn" onclick={() => { addOpen = false; addPickedStop = null; addQuery = ''; addResults = []; }} disabled={busy}>
          <X size={11} />
        </button>
        <button class="transit-transfer-icon-btn transit-transfer-btn-primary" onclick={() => void saveNewTransfer()} disabled={busy || !addPickedStop || addTime.length === 0} title={$t('transit.save')}>
          <Check size={11} />
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
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

  .transit-stop-detail-muted {
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.5);
    font-style: italic;
  }

  .transit-stop-transfer-row {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 0;
    font-size: 0.78rem;
  }

  .transit-stop-transfer-dir {
    flex-shrink: 0;
    font-size: 0.66rem;
    color: oklch(var(--bc) / 0.6);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
    width: 36px;
  }

  .transit-line-name {
    flex: 1;
    min-width: 0;
    font-size: 0.8rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-transfer-time {
    padding: 1px 7px;
    border-radius: 10px;
    background: var(--color-base-200);
    font-size: 0.72rem;
    font-weight: 600;
    color: oklch(var(--bc) / 0.75);
    border: 1px solid transparent;
    cursor: pointer;
  }
  .transit-transfer-time:hover { border-color: var(--color-base-300); }

  .transit-transfer-time-input {
    width: 70px;
    padding: 2px 6px;
    font-size: 0.74rem;
    background: var(--color-base-100);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 3px;
    outline: none;
  }
  .transit-transfer-time-input:focus { border-color: var(--color-primary); }

  .transit-transfer-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    background: none;
    color: var(--color-base-content);
    border: 1px solid transparent;
    border-radius: 3px;
    opacity: 0.65;
    cursor: pointer;
  }
  .transit-transfer-icon-btn:hover:not(:disabled) { opacity: 1; border-color: var(--color-base-300); }
  .transit-transfer-icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
  .transit-transfer-del:hover:not(:disabled) { color: #ef4444; border-color: #ef4444; }
  .transit-transfer-btn-primary { color: var(--color-primary); border-color: var(--color-primary); opacity: 1; }

  .transit-transfer-error {
    font-size: 0.7rem;
    color: #ef4444;
    padding: 2px 0;
  }

  .transit-transfer-add-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    font-size: 0.7rem;
    font-weight: 600;
    background: none;
    color: var(--color-base-content);
    border: 1px dashed var(--color-base-300);
    border-radius: 4px;
    cursor: pointer;
    opacity: 0.7;
    margin-top: 4px;
    align-self: flex-start;
  }
  .transit-transfer-add-btn:hover { opacity: 1; background: var(--color-base-200); }

  .transit-transfer-add-panel {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding: 6px;
    background: var(--color-base-100);
    border: 1px solid var(--color-primary);
    border-radius: 4px;
  }

  .transit-transfer-add-search {
    padding: 3px 7px;
    font-size: 0.74rem;
    background: var(--color-base-200);
    border: 1px solid var(--color-base-300);
    border-radius: 3px;
    outline: none;
  }

  .transit-transfer-add-picked {
    font-size: 0.74rem;
    padding: 3px 6px;
    background: var(--color-base-200);
    border-radius: 3px;
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .transit-transfer-add-picked-id {
    font-family: monospace;
    font-size: 0.62rem;
    opacity: 0.7;
  }

  .transit-transfer-add-results {
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 140px;
    overflow-y: auto;
  }

  .transit-transfer-add-result {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 6px;
    background: none;
    border: none;
    text-align: left;
    font-size: 0.74rem;
    color: var(--color-base-content);
    border-radius: 3px;
    cursor: pointer;
  }
  .transit-transfer-add-result:hover { background: var(--color-base-200); }

  .transit-transfer-add-row {
    display: flex;
    gap: 4px;
    align-items: center;
    justify-content: flex-end;
  }
</style>
