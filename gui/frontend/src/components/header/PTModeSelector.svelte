<script lang="ts">
  import { ptState, TRANSPORT_MODES } from '$lib/stores/pt.svelte';
  import type { TransportMode } from '$lib/stores/pt.svelte';
  import { PTService } from '$lib/api/pt';
  import { appState } from '$lib/stores/app.svelte';
  import EzRadioGroup from '../EzRadioGroup.svelte';
  import EzRadioOption from '../EzRadioOption.svelte';
  
  const modes = Object.values(TRANSPORT_MODES);
  
  async function handleModeChange(mode: string) {
    ptState.selected.mode = mode as TransportMode;
    
    // Clear selection when switching modes
    ptState.selected.lineId = null;
    ptState.selected.routeId = null;
    ptState.selected.stopId = null;
    
    // Close secondary sidebar when mode changes
    appState.secondarySidebar = 'HIDDEN';
    
    // Only fetch data if we have a valid processId
    if (appState.processId > 0) {
      try {
        await PTService.loadPTData(mode as TransportMode);
      } catch (error) {
        console.error('Failed to load PT data for mode:', mode, error);
      }
    }
  }
</script>

<EzRadioGroup 
  name="ptMode"
  bind:value={ptState.selected.mode}
  onchange={handleModeChange}
>
  {#each modes as mode}
    <EzRadioOption 
      name="ptMode" 
      value={mode.value} 
      checked={ptState.selected.mode === mode.value}
    >
      <div class="flex items-center gap-1.5">
        <span>{mode.icon}</span>
        {mode.label}
      </div>
    </EzRadioOption>
  {/each}
</EzRadioGroup>