<script>
  import { Drawer } from 'flowbite-svelte';
  import { sineIn } from 'svelte/easing';
  
  export let hidden = false;
  
  const transitionParams = {
    x: 320,
    duration: 200,
    easing: sineIn
  };
</script>

<Drawer 
  bind:hidden
  position="fixed"
  placement="right"
  transitionParams={transitionParams}
  backdrop={false}
  activateClickOutside={false}
  bodyScrolling={true}
  class="top-[72px] h-[calc(100vh-72px)]"
  width="w-96"
>
  <div class="h-full border-l border-gray-200 shadow-xl">
    <!-- Secondary sidebar content -->
  </div>
</Drawer>