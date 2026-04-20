<script lang="ts">
  import { t } from 'svelte-i18n';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import {
    confirmDeleteStopFacility,
    loadDeleteStopFacilityPreview,
  } from '$lib/stores/transit/profile-stops-actions';

  $effect(() => {
    const target = transitEdit.deleteStopFacilityTarget;
    if (target && target.loadingPreview) {
      void loadDeleteStopFacilityPreview();
    }
  });
</script>

{#if transitEdit.deleteStopFacilityTarget}
  {@const target = transitEdit.deleteStopFacilityTarget}
  <div class="fixed inset-0 z-[2600] flex items-center justify-center bg-black/70">
    <div class="w-[480px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      <h2 class="mb-2 text-lg font-semibold">
        {$t('transit.delete_stop_title', { values: { id: target.stopId } })}
      </h2>

      {#if target.loadingPreview}
        <p class="mb-6 text-sm text-base-content/60">…</p>
      {:else if target.preview}
        {@const p = target.preview}
        {#if p.profileStops === 0 && p.transfers === 0}
          <p class="mb-3 text-sm text-base-content/70">{$t('transit.delete_stop_no_refs')}</p>
        {:else}
          <p class="mb-3 text-sm text-base-content/70">
            {$t('transit.delete_stop_cascade', {
              values: { profileStops: p.profileStops, routes: p.routes, transfers: p.transfers },
            })}
          </p>
        {/if}
        <p class="mb-6 text-sm text-base-content/60">{$t('transit.delete_line_cannot_undo')}</p>
      {/if}

      {#if target.error}
        <p class="mb-4 text-sm text-error">{target.error}</p>
      {/if}

      <div class="flex justify-end gap-1">
        <button
          class="transit-modal-btn"
          disabled={target.submitting}
          onclick={() => transitEdit.cancelDeleteStopFacility()}
        >{$t('transit.cancel')}</button>
        <button
          class="transit-modal-btn transit-modal-btn-danger"
          disabled={target.submitting || target.loadingPreview}
          onclick={() => void confirmDeleteStopFacility()}
        >{$t('transit.delete_confirm')}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .transit-modal-btn {
    padding: 6px 14px;
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--color-base-content);
    background: none;
    border: 1px solid transparent;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.1s, color 0.1s, border-color 0.1s;
  }

  .transit-modal-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-modal-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .transit-modal-btn-danger {
    color: #ef4444;
  }

  .transit-modal-btn-danger:hover:not(:disabled) {
    background: none;
    border-color: #ef4444;
  }
</style>
