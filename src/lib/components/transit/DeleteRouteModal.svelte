<script lang="ts">
  import { t } from 'svelte-i18n';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import { confirmDeleteRoute } from '$lib/stores/transit-actions';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
</script>

<ConfirmModal
  open={transitEdit.deleteRouteTarget !== null}
  title={transitEdit.deleteRouteTarget
    ? $t('transit.delete_route_title', { values: { id: transitEdit.deleteRouteTarget.routeId } })
    : ''}
  confirmLabel={$t('transit.delete_confirm')}
  cancelLabel={$t('transit.cancel')}
  submitting={transitEdit.deleteRouteTarget?.submitting ?? false}
  confirmDisabled={transitEdit.deleteRouteTarget?.loadingPreview ?? false}
  onConfirm={() => void confirmDeleteRoute()}
  onCancel={() => transitEdit.cancelDeleteRoute()}
>
  {#if transitEdit.deleteRouteTarget?.loadingPreview}
    <p>…</p>
  {:else if transitEdit.deleteRouteTarget?.preview}
    {@const p = transitEdit.deleteRouteTarget.preview}
    <p>{$t('transit.delete_route_cascade', {
      values: { profileStops: p.profileStops, pathLinks: p.pathLinks, departures: p.departures },
    })}</p>
    <p>{$t('transit.delete_line_cannot_undo')}</p>
  {/if}
  {#if transitEdit.deleteRouteTarget?.error}
    <p class="confirm-modal-error">{transitEdit.deleteRouteTarget.error}</p>
  {/if}
</ConfirmModal>
