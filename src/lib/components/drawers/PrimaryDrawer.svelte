<script lang="ts">
  import Drawer from '$lib/components/Drawer.svelte';
  import PopulationPrimaryView from '$lib/components/population/PopulationPrimaryView.svelte';
  import NetworkPrimaryView from '$lib/components/network/NetworkPrimaryView.svelte';
  import TransitPrimaryView from '$lib/components/transit/TransitPrimaryView.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { mapViewport } from '$lib/stores/ui.svelte';
  import { NETWORK_ZOOM_THRESHOLD } from '$lib/stores/network.svelte';
</script>

{#if sources.activeKind}
  <Drawer id="primary" side="left" width={320} alwaysOpen={sources.activeKind === 'transit'}>
    {#if sources.activeKind === 'network'}
      {#if mapViewport.state.zoom >= NETWORK_ZOOM_THRESHOLD}
        <NetworkPrimaryView />
      {/if}
    {:else if sources.activeKind === 'population'}
      <PopulationPrimaryView />
    {:else if sources.activeKind === 'transit'}
      <TransitPrimaryView />
    {/if}
  </Drawer>
{/if}
