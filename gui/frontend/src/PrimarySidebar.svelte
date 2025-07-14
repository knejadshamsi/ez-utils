<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from './store.svelte';
  import { commandArgs } from './store.svelte';
  import { PrimaryPTSidebar } from './services/edit/pt';
  
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
    {#if commandArgs.fileEditMode === 'PT'}
      <PrimaryPTSidebar />
    {:else if commandArgs.fileEditMode === 'POPULATION'}
      <!-- Population mode sidebar -->
      <div class="p-4">Population mode not implemented</div>
    {:else if commandArgs.fileEditMode === 'NETWORK'}
      <!-- Network mode sidebar -->
      <div class="p-4">Network mode not implemented</div>
    {/if}
  </div>
</Drawer>