<script lang="ts">
  import { Button, Timeline, TimelineItem } from 'flowbite-svelte';
  import { CloseOutline, PlusOutline } from 'flowbite-svelte-icons';
  import { ptState } from './ptState.svelte';
  import { appState, mapComponent } from '../../../store.svelte';
  import StopItem from './StopItem.svelte';

  const selectedRouteData = $derived(ptState.selectedRoute());
  const line = $derived(selectedRouteData?.line);
  const route = $derived(selectedRouteData?.route);
  
  // Force re-render when stops change
  const stopCount = $derived(route?.stopSequence.length || 0);
  
  // Debug logging
  $effect(() => {
    console.log('[SecondaryPTSidebar] selectedRouteData:', selectedRouteData);
    console.log('[SecondaryPTSidebar] line:', line);
    console.log('[SecondaryPTSidebar] route:', route);
    console.log('[SecondaryPTSidebar] ptState.selectedRouteId:', ptState.selectedRouteId);
    console.log('[SecondaryPTSidebar] stopCount:', stopCount);
    if (route) {
      console.log('[SecondaryPTSidebar] Route stops:', route.stopSequence.length);
      console.log('[SecondaryPTSidebar] Stop IDs:', route.stopSequence.map(s => s.stopId));
    }
  });

  function handleClose() {
    ptState.setSelectedRoute(null);
    ptState.setSelectedLine(null);
    appState.secondarySidebar = 'HIDDEN';
  }

  function handleAddStopToRoute() {
    if (!line || !route) return;
    
    console.log('[SecondaryPTSidebar] Starting add stop mode for route:', route.id);
    
    // Call map component to enable click-to-add mode
    if (mapComponent?.startAddingStop) {
      mapComponent.startAddingStop();
    } else {
      console.warn('[SecondaryPTSidebar] Map component not available');
    }
  }

</script>

{#if route && line}
  <div class="h-full flex flex-col">
    <div class="p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between mb-2">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          Route: {route.direction || 'Unnamed Route'}
        </h2>
        <Button
          size="xs"
          color="light"
          onclick={handleClose}
        >
          <CloseOutline size="sm" />
        </Button>
      </div>
      <p class="text-sm text-gray-500 dark:text-gray-400">
        {ptState.getModeIcon(line.mode)} {line.name} • {stopCount} stops
      </p>
    </div>

    <div class="flex-1 overflow-y-auto p-4">
      <Timeline>
        <TimelineItem title="Route Timeline" date={`First departure: ${route.firstDeparture || 'Not set'}`}>
          <div class="space-y-2">
            {#if route.stopSequence.length === 0}
              <div class="text-sm text-gray-500 italic mb-3">No stops defined</div>
            {:else}
              {#each route.stopSequence as stop, stopIndex (`${stop.stopId}-${stopIndex}`)}
                <StopItem
                  {line}
                  {route}
                  routeIndex={0}
                  {stop}
                  {stopIndex}
                />
              {/each}
            {/if}
            
            <div class="pt-2 border-t border-gray-100 dark:border-gray-600">
              {#if ptState.isAddingStop}
                <div class="mb-2 p-2 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800">
                  <p class="text-xs text-blue-700 dark:text-blue-300 text-center">
                    Click on the map to place a stop
                  </p>
                  <p class="text-xs text-blue-600 dark:text-blue-400 text-center mt-1">
                    Press ESC to cancel
                  </p>
                </div>
                <Button
                  size="xs"
                  color="red"
                  class="w-full"
                  onclick={() => ptState.isAddingStop = false}
                >
                  <CloseOutline size="xs" class="mr-1" />
                  Cancel Adding Stop
                </Button>
              {:else}
                <Button
                  size="xs"
                  color="primary"
                  class="w-full"
                  onclick={handleAddStopToRoute}
                >
                  <PlusOutline size="xs" class="mr-1" />
                  Add Stop
                </Button>
              {/if}
            </div>
          </div>
        </TimelineItem>
      </Timeline>
    </div>
  </div>
{:else}
  <div class="h-full flex items-center justify-center p-4">
    <div class="text-center text-gray-500 dark:text-gray-400">
      <p class="mb-2">No route selected</p>
      <p class="text-sm">Select a route from the primary sidebar to view its timeline</p>
    </div>
  </div>
{/if}