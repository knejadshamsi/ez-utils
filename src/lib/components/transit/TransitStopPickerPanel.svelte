<script lang="ts">
  import { t } from 'svelte-i18n';

  import { transitEdit } from '$lib/stores/transit.svelte';
  import { setProfileStopPickerQuery } from '$lib/stores/transit/profile-stops-actions';
</script>

{#if transitEdit.profileStopsDraft?.pickerOpen}
  {@const d = transitEdit.profileStopsDraft}
  <div class="transit-ps-picker">
    <input
      class="transit-ps-picker-input"
      type="text"
      placeholder={$t('transit.search_stop_placeholder')}
      value={d.pickerQuery}
      oninput={(e) => setProfileStopPickerQuery((e.target as HTMLInputElement).value)}
    />
    <div class="transit-ps-picker-list">
      {#if d.pickerLoading && d.pickerResults.length === 0}
        <div class="transit-ps-picker-empty">…</div>
      {:else if d.pickerResults.length === 0}
        <div class="transit-ps-picker-empty">{$t('transit.no_stops_match')}</div>
      {:else}
        {#each d.pickerResults as facility (facility.id)}
          <button
            class="transit-ps-picker-row"
            onclick={() => transitEdit.appendProfileStopFromFacility(facility)}
          >
            <span class="transit-ps-picker-name">{facility.name ?? facility.id}</span>
            <span class="transit-ps-picker-id">{facility.id}</span>
          </button>
        {/each}
      {/if}
    </div>
  </div>
{/if}

<style>
  .transit-ps-picker {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 12px 12px;
    border-top: 1px solid var(--color-base-300);
    background: var(--color-base-100);
    max-height: 260px;
    overflow: hidden;
  }

  .transit-ps-picker-input {
    padding: 5px 8px;
    font-size: 0.78rem;
    background: var(--color-base-200);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    outline: none;
  }

  .transit-ps-picker-input:focus {
    border-color: var(--color-primary);
  }

  .transit-ps-picker-list {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .transit-ps-picker-empty {
    padding: 10px;
    text-align: center;
    font-size: 0.72rem;
    opacity: 0.5;
  }

  .transit-ps-picker-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 6px 8px;
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
    border-radius: 4px;
    color: var(--color-base-content);
  }

  .transit-ps-picker-row:hover {
    background: var(--color-base-300);
  }

  .transit-ps-picker-name {
    flex: 1;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.8rem;
  }

  .transit-ps-picker-id {
    font-family: monospace;
    font-size: 0.68rem;
    opacity: 0.65;
  }
</style>
