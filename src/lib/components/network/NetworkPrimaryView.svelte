<script lang="ts">
  import { t } from 'svelte-i18n';
  import NetworkLinksList from '$lib/components/network/NetworkLinksList.svelte';
  import NetworkNodesList from '$lib/components/network/NetworkNodesList.svelte';
  import { network } from '$lib/stores/network.svelte';
  import { sources } from '$lib/stores/data.svelte';
</script>

<div class="network-panel">
  <header class="network-panel-header">
    <h2>{sources.activeName ?? $t('network.no_active_source')}</h2>
  </header>

  {#if network.error}
    <p class="network-error">{network.error}</p>
  {/if}

  <div class="network-panel-body">
    <NetworkLinksList />
    <NetworkNodesList />
  </div>
</div>

<style>
  .network-panel {
    height: 100%;
    display: flex;
    flex-direction: column;
    color: var(--color-base-content);
  }
  .network-panel-header {
    padding: 16px;
    border-bottom: 1px solid var(--color-base-300);
  }
  .network-panel-header h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 700;
  }
  .network-panel-header p {
    margin: 4px 0 0;
    font-size: 0.75rem;
    opacity: 0.7;
  }
  .network-panel-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }
  .network-error {
    padding: 12px 16px;
    margin: 0;
    font-size: 0.8rem;
    color: #b91c1c;
  }
</style>
