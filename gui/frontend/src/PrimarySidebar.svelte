<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from '$lib/stores/app.svelte.ts';
  import { commandArgs } from '$lib/stores/app.svelte.ts';
  import PopulationContent from './components/primary-sidebar/PopulationContent.svelte';
  import NetworkContent from './components/primary-sidebar/NetworkContent.svelte';
  import PTContent from './components/primary-sidebar/PTContent.svelte';
  
  export let state: SidebarState = 'EXPANDED';
  
  const transitionParams = {
    x: -320,
    duration: 200,
    easing: sineIn
  };
</script>

<Drawer 
  hidden={state === 'HIDDEN'}
  position="fixed"
  placement="left"
  transitionParams={transitionParams}
  backdrop={false}
  activateClickOutside={false}
  bodyScrolling={true}
  class="top-[72px] h-[calc(100vh-72px)] {state === 'COLLAPSED' ? 'w-16' : 'w-80'}"
>
  <div class="h-full shadow-xl">
    {#if commandArgs.fileEditMode === 'POPULATION'}
      <PopulationContent />
    {:else if commandArgs.fileEditMode === 'NETWORK'}
      <NetworkContent />
    {:else if commandArgs.fileEditMode === 'PT'}
      <PTContent />
    {/if}
  </div>
</Drawer>