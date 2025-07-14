<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from './store.svelte';
  import { commandArgs } from './store.svelte';
  import { SecondaryPTSidebar } from './services/edit/pt';
  
  export let state: SidebarState = 'HIDDEN';
  
  const transitionParams = {
    x: 320,
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
  class="top-[72px] h-[calc(100vh-72px)] {state === 'COLLAPSED' ? 'w-16' : 'w-96'}"
>
  <div class="h-full">
    {#if commandArgs.fileEditMode === 'PT'}
      <SecondaryPTSidebar />
    {:else if commandArgs.fileEditMode === 'POPULATION'}
      <!-- Population mode secondary sidebar -->
      <div class="p-4">Population mode not implemented</div>
    {:else if commandArgs.fileEditMode === 'NETWORK'}
      <!-- Network mode secondary sidebar -->
      <div class="p-4">Network mode not implemented</div>
    {/if}
  </div>
</Drawer>