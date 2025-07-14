<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from './store.svelte';
  import { commandArgs } from './store.svelte';
  import PopulationContent from './components/secondary-sidebar/PopulationContent.svelte';
  import NetworkContent from './components/secondary-sidebar/NetworkContent.svelte';
  import PTContent from './components/secondary-sidebar/PTContent.svelte';
  import { SecondaryPTSidebar } from './services/edit/pt';
  import NetworkSecondarySidebar from './services/edit/network/NetworkSecondarySidebar.svelte';
  
  export let state: SidebarState = 'HIDDEN';
  
  const transitionParams = {
    x: 420,
    duration: 200,
    easing: sineIn
  };
</script>

<Drawer 
  hidden={state === 'HIDDEN'}
  position="fixed"
  placement="right"
  transitionParams={transitionParams}
  backdrop={false}
  activateClickOutside={false}
  bodyScrolling={true}
  class="top-[72px] h-[calc(100vh-72px)] {state === 'COLLAPSED' ? 'w-16' : 'w-[420px]'}"
>
  <div class="h-full border-l border-gray-200 shadow-xl">
    {#if commandArgs.fileEditMode === 'POPULATION'}
      <PopulationContent />
    {:else if commandArgs.fileEditMode === 'NETWORK'}
      <NetworkSecondarySidebar />
    {:else if commandArgs.fileEditMode === 'PT'}
      <SecondaryPTSidebar />
    {/if}
  </div>
</Drawer>