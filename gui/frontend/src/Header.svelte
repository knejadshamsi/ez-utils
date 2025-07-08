<script>
  import Left from './components/header/left.svelte';
  import Middle from './components/header/middle.svelte';
  import Right from './components/header/right.svelte';
</script>

<header class="w-full h-[72px] bg-[#0a1628] grid grid-cols-3">
  <div class="px-6">
    <Left />
  </div>
  <div>
    <Middle />
  </div>
  <div class="px-6">
    <Right />
  </div>
</header>