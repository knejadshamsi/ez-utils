<script lang="ts">
  import { t } from 'svelte-i18n';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import {
    confirmDeleteStopFacility,
    loadDeleteStopFacilityPreview,
  } from '$lib/stores/transit/profile-stops-actions';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';

  $effect(() => {
    const target = transitEdit.deleteStopFacilityTarget;
    if (target && target.loadingPreview) {
      void loadDeleteStopFacilityPreview();
    }
  });
</script>

<ConfirmModal
  open={transitEdit.deleteStopFacilityTarget !== null}
  title={transitEdit.deleteStopFacilityTarget
    ? $t('transit.delete_stop_title', { values: { id: transitEdit.deleteStopFacilityTarget.stopId } })
    : ''}
  confirmLabel={$t('transit.delete_confirm')}
  cancelLabel={$t('transit.cancel')}
  submitting={transitEdit.deleteStopFacilityTarget?.submitting ?? false}
  confirmDisabled={transitEdit.deleteStopFacilityTarget?.loadingPreview ?? false}
  onConfirm={() => void confirmDeleteStopFacility()}
  onCancel={() => transitEdit.cancelDeleteStopFacility()}
>
  {#if transitEdit.deleteStopFacilityTarget?.loadingPreview}
    <p>…</p>
  {:else if transitEdit.deleteStopFacilityTarget?.preview}
    {@const p = transitEdit.deleteStopFacilityTarget.preview}
    {#if p.profileStops === 0 && p.transfers === 0}
      <p>{$t('transit.delete_stop_no_refs')}</p>
    {:else}
      <p>{$t('transit.delete_stop_cascade', {
        values: { profileStops: p.profileStops, routes: p.routes, transfers: p.transfers },
      })}</p>
    {/if}
    <p>{$t('transit.delete_line_cannot_undo')}</p>
  {/if}
  {#if transitEdit.deleteStopFacilityTarget?.error}
    <p class="confirm-modal-error">{transitEdit.deleteStopFacilityTarget.error}</p>
  {/if}
</ConfirmModal>
