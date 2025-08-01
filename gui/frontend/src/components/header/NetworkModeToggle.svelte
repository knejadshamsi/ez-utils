<script lang="ts">
  import { appState, type NetworkMode } from '$lib/stores/app.svelte';
  import EzRadioGroup from '../EzRadioGroup.svelte';
  import EzRadioOption from '../EzRadioOption.svelte';
  import { EyeOutline, CogOutline, PlusOutline } from 'flowbite-svelte-icons';

  function handleModeChange(mode: string) {
    appState.networkMode = mode as NetworkMode;
    
    // Update sidebar visibility based on mode
    if (mode === 'VIEW') {
      appState.primarySidebar = 'EXPANDED';
      appState.secondarySidebar = 'HIDDEN';
    } else if (mode === 'EDIT' || mode === 'CREATE') {
      // Both Edit and Create modes show primary sidebar
      appState.primarySidebar = 'EXPANDED';
      // Secondary sidebar shown when entity selected
      appState.secondarySidebar = 'HIDDEN';
    }
  }
</script>

<EzRadioGroup 
  name="networkMode"
  bind:value={appState.networkMode}
  onchange={handleModeChange}
>
  <EzRadioOption 
    name="networkMode" 
    value="VIEW" 
    checked={appState.networkMode === 'VIEW'}
  >
    <div class="flex items-center gap-1.5">
      <EyeOutline class="w-4 h-4" />
      View
    </div>
  </EzRadioOption>
  <EzRadioOption 
    name="networkMode" 
    value="EDIT" 
    checked={appState.networkMode === 'EDIT'}
  >
    <div class="flex items-center gap-1.5">
      <CogOutline class="w-4 h-4" />
      Edit
    </div>
  </EzRadioOption>
  <EzRadioOption 
    name="networkMode" 
    value="CREATE" 
    checked={appState.networkMode === 'CREATE'}
  >
    <div class="flex items-center gap-1.5">
      <PlusOutline class="w-4 h-4" />
      Create
    </div>
  </EzRadioOption>
</EzRadioGroup>