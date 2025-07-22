<script>
  // Middle section of header
  // Demo props - can be uncommented to enable drawing toggle
  // export let isEditingEnabled = false;
  // export let onToggleEditing = () => {};
  
  import { commandArgs } from '$lib/stores/app.svelte';
  import PTModeSelector from './PTModeSelector.svelte';
</script>

<div class="flex items-center justify-center h-full">
  <!-- Demo: Drawing toggle button - uncomment to enable -->
  <!-- <button 
    class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
    onclick={onToggleEditing}
  >
    {isEditingEnabled ? 'Disable' : 'Enable'} Drawing
  </button> -->
  
  <!-- PT Mode Selector when editing PT files -->
  {#if commandArgs.fileEditMode === 'PT'}
    <PTModeSelector />
  {/if}
  
  <!-- Network mode toggle moved to primary sidebar -->
</div>