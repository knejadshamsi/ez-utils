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
  import ConfirmationModal from "./components/modals/ConfirmationModal.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { appState, commandArgs, mapComponent } from "$lib/stores/app.svelte";
  import { confirmationState } from "$lib/stores/confirmationModal.svelte";
  import NetworkController from "./components/network/NetworkController.svelte";
  import { networkState } from "$lib/stores/network.svelte";
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
<ConfirmationModal 
  bind:open={confirmationState.isOpen}
  title={confirmationState.options?.title || ''}
  message={confirmationState.options?.message || ''}
  confirmText={confirmationState.options?.confirmText}
  cancelText={confirmationState.options?.cancelText}
  onConfirm={confirmationState.options?.onConfirm || (() => {})}
  onCancel={confirmationState.options?.onCancel || (() => {})}
/>
<ToastContainer />

{#if commandArgs.fileEditMode === 'NETWORK'}
  <NetworkController />
{/if}

