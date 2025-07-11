<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "./Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  import WelcomeModal from "./WelcomeModal.svelte";
  
  let primarySidebarHidden = $state(false);
  let secondarySidebarHidden = $state(false);
  
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

<PrimarySidebar bind:hidden={primarySidebarHidden} />
<SecondarySidebar bind:hidden={secondarySidebarHidden} />

<WelcomeModal />

