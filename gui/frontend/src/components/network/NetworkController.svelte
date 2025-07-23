<script lang="ts">
  import { onMount } from 'svelte';
  import { commandArgs, appState } from '$lib/stores/app.svelte';
  import { loadInitialViewportData, loadNetworkInPolygon } from '$lib/api/network';
  import { networkState } from '$lib/stores/network.svelte';
  import { mapState } from '../../map/mapState.svelte';
  import { updateNetworkVisualization } from '../../map/updateNetworkVisualization';
  
  let isLoading = false;
  let error: string | null = null;
  let hasInitialLoad = false;
  
  onMount(async () => {
    if (commandArgs.fileEditMode === 'NETWORK' && commandArgs.validationStatus === 'VALIDATED_NETWORK') {
      await loadInitialNetworkData();
    }
  });
  
  async function loadInitialNetworkData() {
    if (!mapState.map || hasInitialLoad) return;
    
    isLoading = true;
    error = null;
    hasInitialLoad = true;
    
    try {
      await loadInitialViewportData(mapState.map);
      console.log(`Loaded initial viewport: ${networkState.nodes.length} nodes and ${networkState.links.length} links`);
      // Update map visualization
      updateNetworkVisualization();
    } catch (err) {
      error = err.message || 'Failed to load network data';
      console.error('Failed to load network data:', err);
      hasInitialLoad = false;
    } finally {
      isLoading = false;
    }
  }
  
  // Load data for drawn polygons
  async function loadPolygonData(polygonCoords: number[][]) {
    isLoading = true;
    error = null;
    
    try {
      const beforeCount = networkState.nodes.length;
      await loadNetworkInPolygon(polygonCoords);
      const newCount = networkState.nodes.length - beforeCount;
      console.log(`Loaded polygon area: ${newCount} new nodes`);
      updateNetworkVisualization();
    } catch (err) {
      error = err.message || 'Failed to load polygon data';
      console.error('Failed to load polygon data:', err);
    } finally {
      isLoading = false;
    }
  }
  
  // React to polygon selection from networkState
  $: if (networkState.selection.selectedPolygon && networkState.selection.selectedPolygon.length > 0) {
    loadPolygonData(networkState.selection.selectedPolygon);
  }
  
  // Reload initial data when switching back to network mode
  $: if (commandArgs.fileEditMode === 'NETWORK' && appState.display === 'EDITING' && !hasInitialLoad && mapState.map) {
    loadInitialNetworkData();
  }
</script>

{#if isLoading}
  <div class="fixed top-20 right-4 bg-blue-100 text-blue-800 px-4 py-2 rounded shadow">
    Loading network data...
  </div>
{/if}

{#if error}
  <div class="fixed top-20 right-4 bg-red-100 text-red-800 px-4 py-2 rounded shadow">
    Error: {error}
  </div>
{/if}