<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "@map/Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  import WelcomeModal from "./components/modals/WelcomeModal.svelte";
  import ProcessingModal from "./components/modals/ProcessingModal.svelte";
  import LoadingModal from "./components/modals/LoadingModal.svelte";
  import ExportProgressModal from "./components/modals/ExportProgressModal.svelte";
  import AttributesModal from "./components/modals/AttributesModal.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { appState, commandArgs, mapComponent } from "$lib/stores/app.svelte.ts";
  import { PTService } from "$lib/api/pt";
  import NetworkController from "./components/network/NetworkController.svelte";
  import { networkState } from "$lib/stores/network.svelte";
  
  // Load PT data when editing PT files - using onMount instead of $effect
  import { onMount } from 'svelte';
  
  onMount(async () => {
    if (appState.display === 'EDITING' && commandArgs.fileEditMode === 'PT') {
      console.log('[App] Loading PT data for default mode (bus)...');
      try {
        await PTService.loadPTData('BUS'); // Load only BUS mode initially
      } catch (error) {
        console.error('[App] Failed to load PT data on app start:', error);
      }
    }
  });
  
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
<LoadingModal />
<ExportProgressModal />
<AttributesModal />
<ToastContainer />

{#if commandArgs.fileEditMode === 'NETWORK'}
  <NetworkController />
{/if}

