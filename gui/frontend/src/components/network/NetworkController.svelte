<script lang="ts">
  import { onMount } from 'svelte';
  import { commandArgs, appState } from '$lib/stores/app.svelte.ts';
  import { loadAllNetworkData } from '$lib/api/network';
  import { networkState } from '$lib/stores/network.svelte';
  
  let isLoading = false;
  let error: string | null = null;
  
  onMount(async () => {
    if (commandArgs.fileEditMode === 'NETWORK' && commandArgs.validationStatus === 'VALIDATED_NETWORK') {
      await loadNetworkData();
    }
  });
  
  async function loadNetworkData() {
    isLoading = true;
    error = null;
    
    try {
      await loadAllNetworkData();
      console.log(`Loaded ${networkState.nodes.length} nodes and ${networkState.links.length} links`);
    } catch (err) {
      error = err.message || 'Failed to load network data';
      console.error('Failed to load network data:', err);
    } finally {
      isLoading = false;
    }
  }
  
  // Reload data when switching back to network mode
  $: if (commandArgs.fileEditMode === 'NETWORK' && appState.display === 'EDITING' && networkState.nodes.length === 0) {
    loadNetworkData();
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