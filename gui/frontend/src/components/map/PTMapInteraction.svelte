<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import { ptState } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';

  let { map }: { map: MaplibreMap | null } = $props();

  export function startAddingStop() {
    // This method is not used anymore since we pre-create stops
    console.warn('[PTMapInteraction] startAddingStop is deprecated');
  }

  function stopAddingStop() {
    ptState.isAddingStop = false;
    if (map) {
      map.getCanvas().style.cursor = '';
    }
  }

  function handleMapClick(e: any) {
    // Old click-to-add logic is no longer used
    // Stop addition is now handled through pre-creation in PTContent.svelte
    // And location selection is handled in Map.svelte
  }

  function updateMapData() {
    // MapLibre GL rendering removed - now using deck.gl layers in PTMapLayer.svelte
  }

  $effect(() => {
    updateMapData();
  });

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && ptState.isAddingStop) {
      stopAddingStop();
    }
  }

  onMount(() => {
    // No longer needed - map clicks are handled in Map.svelte
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
  });
</script>