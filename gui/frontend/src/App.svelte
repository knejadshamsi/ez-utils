<script lang="ts">
  import "./app.css";
  import Header from "./Header.svelte";
  import Map from "./Map.svelte";
  import PrimarySidebar from "./PrimarySidebar.svelte";
  import SecondarySidebar from "./SecondarySidebar.svelte";
  
  let primarySidebarHidden = $state(false);
  let secondarySidebarHidden = $state(false);
</script>

<Header />

<div class="h-[calc(100vh-72px)] bg-gray-50">
  <Map />
</div>

<PrimarySidebar bind:hidden={primarySidebarHidden} />
<SecondarySidebar bind:hidden={secondarySidebarHidden} />

