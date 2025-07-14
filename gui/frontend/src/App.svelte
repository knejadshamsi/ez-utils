<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "./Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  // import WelcomeModal from "./WelcomeModal.svelte"; // TEMPORARY: Commented for development
  import ProcessingModal from "./components/ProcessingModal.svelte";
  import { mapComponent } from './store.svelte';
  
  let mapInstance: any;
  
  $effect(() => {
    // Update global mapComponent reference when mapInstance is available
    if (mapInstance) {
      mapComponent.startAddingStop = mapInstance.startAddingStop;
    }
  });
  import { appState, commandArgs } from "./store.svelte";
  import { PTService } from "./services/edit/pt/ptService";
  
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
</script>

<Header />

<div class="h-[calc(100vh-72px)] bg-gray-50">
  <Map bind:this={mapInstance} />
</div>

<PrimarySidebar state={appState.primarySidebar} />
<SecondarySidebar state={appState.secondarySidebar} />

<!-- TEMPORARY: Commented out for development - UNCOMMENT BEFORE COMMIT -->
<!-- <WelcomeModal /> -->
<ProcessingModal />

