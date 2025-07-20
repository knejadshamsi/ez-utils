<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from '$lib/stores/app.svelte.ts';
  import { commandArgs } from '$lib/stores/app.svelte.ts';
  import PopulationContent from './components/secondary-sidebar/PopulationContent.svelte';
  // MAP_TODO: Restore PT and Network content
  // import NetworkContent from './components/secondary-sidebar/NetworkContent.svelte';
  // import PTContent from './components/secondary-sidebar/PTContent.svelte';
  
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
  class="top-[72px] h-[calc(100vh-72px)] {state === 'COLLAPSED' ? 'w-16' : 'w-[420px]'} z-[1000]"
>
  <div class="h-full shadow-xl">
    {#if commandArgs.fileEditMode === 'POPULATION'}
      <PopulationContent />
    {:else if commandArgs.fileEditMode === 'NETWORK'}
      <!-- MAP_TODO: <NetworkContent /> -->
      <div class="p-4 text-gray-400">Network details temporarily disabled</div>
    {:else if commandArgs.fileEditMode === 'PT'}
      <!-- MAP_TODO: <PTContent /> -->
      <div class="p-4 text-gray-400">PT details temporarily disabled</div>
    {/if}
  </div>
</Drawer>