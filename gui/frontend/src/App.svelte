<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "./Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  // import WelcomeModal from "./components/modals/WelcomeModal.svelte"; // TEMPORARY: Commented for development
  import ProcessingModal from "./components/modals/ProcessingModal.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { appState, commandArgs, mapComponent } from "$lib/stores/app.svelte.ts";
  import { PTService } from "$lib/api/pt";
  import NetworkController from "./components/network/NetworkController.svelte";
  import { networkState } from "$lib/stores/network.svelte";
  
  let mapInstance: any;
  
  // Simple boolean to control map editing for network mode
  let mapEditingEnabled = $state(false);
  
  $effect(() => {
    // Update global mapComponent reference when mapInstance is available
    if (mapInstance) {
      mapComponent.startAddingStop = mapInstance.startAddingStop;
    }
  });
  
  // Demo: Editable layer state - uncomment to enable drawing functionality
  // let isEditingEnabled = $state(false);
  // const toggleEditing = () => {
  //   isEditingEnabled = !isEditingEnabled;
  // };

  // Load PT data when app starts in editing mode with PT file mode
  $effect(() => {
    console.log('[App] Effect triggered - display:', appState.display, 'fileEditMode:', commandArgs.fileEditMode);
    if (appState.display === 'EDITING' && commandArgs.fileEditMode === 'PT') {
      console.log('[App] Loading PT data...');
      PTService.loadPTData().catch(error => {
        console.error('[App] Failed to load PT data on app start:', error);
      });
    }
  });
  
  // Watch for network drawing state changes and clear polygon when starting new drawing
  $effect(() => {
    if (networkState.isDrawingPolygon) {
      console.log('Enabling map editing for polygon drawing');
      mapEditingEnabled = true;
      // Clear any existing selection polygon when starting new drawing
      networkState.selection.selectedPolygon = null;
    }
  });
  
  // Handle polygon completion
  function handlePolygonComplete(polygon: number[][]) {
    console.log('Polygon completed:', polygon);
    networkState.selection.selectedPolygon = polygon;
    networkState.isDrawingPolygon = false;
    mapEditingEnabled = false;
  }
</script>

<Header />

<div class="h-[calc(100vh-72px)] bg-gray-50">
  <Map 
    bind:this={mapInstance}
    isEditingEnabled={mapEditingEnabled} 
    onPolygonComplete={handlePolygonComplete}
    selectionPolygon={networkState.selection.selectedPolygon} 
  />
</div>

<PrimarySidebar state={appState.primarySidebar} />
<SecondarySidebar state={appState.secondarySidebar} />

<!-- TEMPORARY: Commented out for development - UNCOMMENT BEFORE COMMIT -->
<!-- <WelcomeModal /> -->
<ProcessingModal />
<ToastContainer />

{#if commandArgs.fileEditMode === 'NETWORK'}
  <NetworkController />
{/if}

