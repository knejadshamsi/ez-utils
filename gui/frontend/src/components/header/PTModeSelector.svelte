<script lang="ts">
  import { ButtonGroup, Button } from 'flowbite-svelte';
  import { ptState, TransportMode } from '$lib/stores/pt.svelte';
  import { PTService } from '$lib/api/pt';
  
  const modes = [
    { value: TransportMode.Bus, label: 'Bus', icon: '🚌' },
    { value: TransportMode.Metro, label: 'Metro', icon: '🚇' },
    { value: TransportMode.Tram, label: 'Tram', icon: '🚊' }
  ];
  
  async function handleModeChange(mode: TransportMode) {
    // Clear existing data and set new mode
    ptState.setSelectedMode(mode);
    
    // Fetch new mode data
    try {
      await PTService.loadPTData(mode);
    } catch (error) {
      console.error('Failed to load PT data for mode:', mode, error);
    }
  }
</script>

<div class="flex items-center justify-center h-full">
  <ButtonGroup class="shadow-sm">
    {#each modes as mode}
      <Button
        size="sm"
        color={ptState.selectedMode === mode.value ? 'primary' : 'alternative'}
        onclick={() => handleModeChange(mode.value)}
        class="px-4 py-2 text-sm font-medium transition-all duration-200 {ptState.selectedMode === mode.value ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300 hover:bg-gray-600'}"
      >
        <span class="mr-2">{mode.icon}</span>
        {mode.label}
      </Button>
    {/each}
  </ButtonGroup>
</div>