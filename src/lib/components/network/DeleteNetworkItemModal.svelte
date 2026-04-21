<script lang="ts">
  import { t } from 'svelte-i18n';
  import { network } from '$lib/stores/network.svelte';
  import { confirmDelete } from '$lib/stores/network-actions';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
</script>

<ConfirmModal
  open={network.deleteTarget !== null}
  title={network.deleteTarget?.kind === 'link' ? $t('network.delete_link') : $t('network.delete_node')}
  confirmLabel={$t('network.delete_confirm')}
  cancelLabel={$t('network.cancel')}
  onConfirm={() => void confirmDelete()}
  onCancel={() => network.cancelDelete()}
>
  {#if network.deleteTarget?.kind === 'link'}
    <p>{$t('network.delete_link_message', { values: { id: network.deleteTarget.id } })}</p>
    <p>{$t('network.delete_cannot_undo')}</p>
  {:else if network.deleteTarget}
    <p>{$t('network.delete_node_message', { values: { id: network.deleteTarget.id } })}</p>
    {#if network.deleteTarget.connectedLinks === 1}
      <p>{$t('network.delete_node_cascade_one', { values: { count: network.deleteTarget.connectedLinks } })}</p>
    {:else if network.deleteTarget.connectedLinks > 1}
      <p>{$t('network.delete_node_cascade_many', { values: { count: network.deleteTarget.connectedLinks } })}</p>
    {/if}
    <p>{$t('network.delete_cannot_undo')}</p>
  {/if}
</ConfirmModal>
