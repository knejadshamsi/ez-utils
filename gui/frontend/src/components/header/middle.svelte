<script>
  // Middle section of header
  // Demo props - can be uncommented to enable drawing toggle
  // export let isEditingEnabled = false;
  // export let onToggleEditing = () => {};
  
  import { commandArgs } from '../../store.svelte';
  import NetworkModeToggle from '../../services/edit/network/NetworkModeToggle.svelte';
</script>

<div class="flex items-center justify-center h-full">
  <!-- Demo: Drawing toggle button - uncomment to enable -->
  <!-- <button 
    class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
    onclick={onToggleEditing}
  >
    {isEditingEnabled ? 'Disable' : 'Enable'} Drawing
  </button> -->
  
  {#if commandArgs.fileEditMode === 'NETWORK' && commandArgs.validationStatus === 'VALIDATED_NETWORK'}
    <NetworkModeToggle />
  {/if}
</div>