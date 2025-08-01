<script>
  import Left from './components/header/left.svelte';
  import Middle from './components/header/middle.svelte';
  import Right from './components/header/right.svelte';
  
  // Demo: Uncomment to enable drawing toggle functionality
  // let { isEditingEnabled = false, toggleEditing = () => {} } = $props();
</script>

<header class="w-full h-[72px] bg-[#0a1628] grid grid-cols-3">
  <div class="px-6">
    <Left />
  </div>
  <div>
    <Middle />
    <!-- Demo: Uncomment to enable drawing toggle -->
    <!-- <Middle {isEditingEnabled} onToggleEditing={toggleEditing} /> -->
  </div>
  <div class="px-6">
    <Right />
  </div>
</header>