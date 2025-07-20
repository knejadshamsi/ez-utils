<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "@map/Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  import WelcomeModal from "./components/modals/WelcomeModal.svelte";
  import ProcessingModal from "./components/modals/ProcessingModal.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { appState, commandArgs, mapComponent } from "$lib/stores/app.svelte.ts";
  // MAP_TODO: Restore PT and Network imports
  // import { PTService } from "$lib/api/pt";
  // import NetworkController from "./components/network/NetworkController.svelte";
  // import { networkState } from "$lib/stores/network.svelte";
  
  // Map-related code removed - to be replaced with Leaflet implementation
  
  // Demo: Editable layer state - uncomment to enable drawing functionality
  // let isEditingEnabled = $state(false);
  // const toggleEditing = () => {
  //   isEditingEnabled = !isEditingEnabled;
  // };

  // MAP_TODO: Restore PT data loading
  // $effect(() => {
  //   console.log('[App] Effect triggered - display:', appState.display, 'fileEditMode:', commandArgs.fileEditMode);
  //   if (appState.display === 'EDITING' && commandArgs.fileEditMode === 'PT') {
  //     console.log('[App] Loading PT data...');
  //     PTService.loadPTData().catch(error => {
  //       console.error('[App] Failed to load PT data on app start:', error);
  //     });
  //   }
  // });
  
  // Network drawing handlers - to be reimplemented with Leaflet
</script>

<Header />

<div class="h-[calc(100vh-72px)] bg-gray-50">
  <Map />
</div>

<PrimarySidebar state={appState.primarySidebar} />
<SecondarySidebar state={appState.secondarySidebar} />

<WelcomeModal />
<ProcessingModal />
<ToastContainer />

<!-- MAP_TODO: Restore network controller
{#if commandArgs.fileEditMode === 'NETWORK'}
  <NetworkController />
{/if}
-->

