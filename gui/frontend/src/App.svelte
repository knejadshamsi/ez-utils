<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "./Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  // import WelcomeModal from "./WelcomeModal.svelte"; // TEMPORARY: Commented for development
  import ProcessingModal from "./components/ProcessingModal.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  import { appState } from "./store.svelte";
  
  // Demo: Editable layer state - uncomment to enable drawing functionality
  // let isEditingEnabled = $state(false);
  // const toggleEditing = () => {
  //   isEditingEnabled = !isEditingEnabled;
  // };
</script>

<Header />

<div class="h-[calc(100vh-72px)] bg-gray-50">
  <Map />
</div>

<PrimarySidebar state={appState.primarySidebar} />
<SecondarySidebar state={appState.secondarySidebar} />

<!-- TEMPORARY: Commented out for development - UNCOMMENT BEFORE COMMIT -->
<!-- <WelcomeModal /> -->
<ProcessingModal />
<ToastContainer />

