<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from './store.svelte';
  
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
  <div class="h-full border-l border-gray-200 shadow-xl">
    <!-- Secondary sidebar content -->
  </div>
</Drawer>