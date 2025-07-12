<script lang="ts">
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  import type { SidebarState } from './store.svelte';
  
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
    <!-- Primary sidebar content -->
  </div>
</Drawer>