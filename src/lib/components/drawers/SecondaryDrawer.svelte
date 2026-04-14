<script lang="ts">
  import Drawer from '$lib/components/Drawer.svelte';
  import PopulationSecondaryView from '$lib/components/population/PopulationSecondaryView.svelte';
  import NetworkSecondaryView from '$lib/components/network/NetworkSecondaryView.svelte';
  import TransitSecondaryView from '$lib/components/transit/TransitSecondaryView.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { transit } from '$lib/stores/transit.svelte';

  let transitHasRoute = $derived(sources.activeKind === 'transit' && transit.selectedRouteId !== null);
  let show = $derived(
    sources.activeKind === 'network' || sources.activeKind === 'population' || transitHasRoute
  );
</script>

{#if show}
  <Drawer id="secondary" side="right" width={420} alwaysOpen={transitHasRoute}>
    {#if sources.activeKind === 'network'}
      <NetworkSecondaryView />
    {:else if sources.activeKind === 'population'}
      <PopulationSecondaryView />
    {:else if sources.activeKind === 'transit'}
      <TransitSecondaryView />
    {/if}
  </Drawer>
{/if}
