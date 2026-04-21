<script lang="ts">
  import { t } from 'svelte-i18n';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import { confirmDeleteLine } from '$lib/stores/transit-actions';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
</script>

<ConfirmModal
  open={transitEdit.deleteTarget !== null}
  title={transitEdit.deleteTarget
    ? $t('transit.delete_line_title', { values: { id: transitEdit.deleteTarget.lineId } })
    : ''}
  confirmLabel={$t('transit.delete_confirm')}
  cancelLabel={$t('transit.cancel')}
  submitting={transitEdit.deleteTarget?.submitting ?? false}
  confirmDisabled={transitEdit.deleteTarget?.loadingPreview ?? false}
  onConfirm={() => void confirmDeleteLine()}
  onCancel={() => transitEdit.cancelDeleteLine()}
>
  {#if transitEdit.deleteTarget?.loadingPreview}
    <p>…</p>
  {:else if transitEdit.deleteTarget?.preview}
    {@const p = transitEdit.deleteTarget.preview}
    <p>{$t('transit.delete_line_cascade', {
      values: { routes: p.routes, profileStops: p.profileStops, pathLinks: p.pathLinks, departures: p.departures },
    })}</p>
    <p>{$t('transit.delete_line_cannot_undo')}</p>
  {/if}
  {#if transitEdit.deleteTarget?.error}
    <p class="confirm-modal-error">{transitEdit.deleteTarget.error}</p>
  {/if}
</ConfirmModal>

<style>
  :global(.confirm-modal-error) {
    color: #ef4444;
  }
</style>
