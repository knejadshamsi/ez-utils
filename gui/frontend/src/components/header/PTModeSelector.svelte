<script lang="ts">
  import { selected } from '@workflow/pt/state.svelte';
  import { saveCurrentRoute, loadPTData } from '@workflow/pt/loading.svelte';
  import { selectMode } from '@workflow/pt/functions.svelte';
  import type { TransportMode } from '@workflow/pt/types';
  import { appState } from '$lib/stores/app.svelte';
  import { updatePTVisualization } from '../../map/updatePTVisualization';
  import EzRadioGroup from '../EzRadioGroup.svelte';
  import EzRadioOption from '../EzRadioOption.svelte';
  
  const TRANSPORT_MODES = {
    BUS: { value: 'BUS', label: 'Bus', icon: '🚌' },
    METRO: { value: 'METRO', label: 'Metro', icon: '🚇' },
    TRAM: { value: 'TRAM', label: 'Tram', icon: '🚋' },
  };
  const modes = Object.values(TRANSPORT_MODES);
  
  async function handleModeChange(mode: string) {
    // Auto-save current route if needed
    if (selected.routeId) {
      try {
        await saveCurrentRoute();
      } catch (error) {
        console.error('Failed to auto-save route:', error);
      }
    }
    
    // Switch mode (clears selections and current route data)
    selectMode(mode as TransportMode);
    
    // Clear the map visualization
    updatePTVisualization();
    
    // Close secondary sidebar when mode changes
    appState.secondarySidebar = 'HIDDEN';
    
    // Only fetch data if we have a valid processId
    if (appState.processId > 0) {
      try {
        await loadPTData(mode as TransportMode);
      } catch (error) {
        console.error('Failed to load PT data for mode:', mode, error);
      }
    }
  }
</script>

<EzRadioGroup 
  name="ptMode"
  bind:value={selected.mode}
  onchange={handleModeChange}
>
  {#each modes as mode}
    <EzRadioOption 
      name="ptMode" 
      value={mode.value} 
      checked={selected.mode === mode.value}
    >
      <div class="flex items-center gap-1.5">
        <span>{mode.icon}</span>
        {mode.label}
      </div>
    </EzRadioOption>
  {/each}
</EzRadioGroup>