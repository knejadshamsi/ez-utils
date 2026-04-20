<script lang="ts">
  import { t } from 'svelte-i18n';
  import { network } from '$lib/stores/network.svelte';
  import { confirmDelete } from '$lib/stores/network-actions';
</script>

{#if network.deleteTarget}
  <div class="fixed inset-0 z-[2600] flex items-center justify-center bg-black/70">
    <div class="w-[440px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      {#if network.deleteTarget.kind === 'link'}
        <h2 class="mb-2 text-lg font-semibold">{$t('network.delete_link')}</h2>
        <p class="mb-6 text-sm text-base-content/70">
          {$t('network.delete_link_message', { values: { id: network.deleteTarget.id } })} {$t('network.delete_cannot_undo')}
        </p>
      {:else}
        <h2 class="mb-2 text-lg font-semibold">{$t('network.delete_node')}</h2>
        <p class="mb-6 text-sm text-base-content/70">
          {$t('network.delete_node_message', { values: { id: network.deleteTarget.id } })}
          {#if network.deleteTarget.connectedLinks === 1}
            {$t('network.delete_node_cascade_one', { values: { count: network.deleteTarget.connectedLinks } })}
          {:else if network.deleteTarget.connectedLinks > 1}
            {$t('network.delete_node_cascade_many', { values: { count: network.deleteTarget.connectedLinks } })}
          {/if}
          {$t('network.delete_cannot_undo')}
        </p>
      {/if}

      <div class="flex justify-end gap-2">
        <button class="btn btn-ghost" onclick={() => network.cancelDelete()}>{$t('network.cancel')}</button>
        <button class="btn btn-error" onclick={() => void confirmDelete()}>{$t('network.delete_confirm')}</button>
      </div>
    </div>
  </div>
{/if}
