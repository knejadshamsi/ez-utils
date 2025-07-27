<script lang="ts">
  import { Button } from 'flowbite-svelte';
  import { CloseOutline, PlusOutline, TrashBinOutline } from 'flowbite-svelte-icons';
  import { ptState, TRANSPORT_MODES } from '$lib/stores/pt.svelte';
  import { appState } from '$lib/stores/app.svelte';
  import AddStopModal from '../modals/pt/AddStopModal.svelte';
  import AddDepartureModal from '../modals/pt/AddDepartureModal.svelte';

  let showAddStopModal = $state(false);
  let showAddDepartureModal = $state(false);
  
  // Get current selected route data
  const selectedLine = $derived(() => {
    if (!ptState.selected.lineId) return null;
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.lines.find(l => l.id === ptState.selected.lineId);
  });
  
  const selectedRoute = $derived(() => {
    if (!ptState.selected.routeId || !selectedLine()) return null;
    return selectedLine()?.routes.find(r => r.id === ptState.selected.routeId);
  });
  
  const routeStops = $derived(() => {
    if (!ptState.selected.routeId) return [];
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.stops.filter(s => s.routeId === ptState.selected.routeId) || [];
  });
  
  const routeDepartures = $derived(() => {
    if (!ptState.selected.routeId) return [];
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.departures.filter(d => d.routeId === ptState.selected.routeId) || [];
  });

  function handleClose() {
    ptState.selected.routeId = null;
    ptState.selected.stopId = null;
    appState.secondarySidebar = 'HIDDEN';
  }

  function handleDeleteRoute() {
    if (!selectedLine() || !selectedRoute()) return;
    
    if (confirm(`Delete route "${selectedRoute()?.name}"? This will also delete all stops and departures.`)) {
      ptState.deleteRoute(selectedLine()!.id, selectedRoute()!.id);
      handleClose();
    }
  }

  function handleDeleteStop(stopId: string) {
    ptState.deleteStop(stopId);
  }
  
  function handleDeleteDeparture(departureId: string) {
    ptState.deleteDeparture(departureId);
  }
  
  function handleStopClick(stopId: string) {
    ptState.selected.stopId = stopId;
  }
</script>

<div class="flex flex-col h-full">
  {#if selectedRoute()}
    <!-- Header -->
    <div class="p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between">
        <div>
          <div class="flex items-center gap-2">
            <span class="text-lg font-semibold">{selectedRoute()?.name || 'Route'}</span>
            <span class="text-sm text-gray-500">
              {TRANSPORT_MODES[ptState.selected.mode]?.icon}
            </span>
          </div>
          <p class="text-sm text-gray-500">
            {selectedLine()?.name} • {routeStops().length} stops • {routeDepartures().length} departures
          </p>
        </div>
        <Button 
          size="xs" 
          color="alternative" 
          onclick={handleClose}
        >
          <CloseOutline size="xs" />
        </Button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto">
      <!-- Stops Section -->
      <div class="p-4 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-medium">Stops</h3>
          <Button 
            size="xs" 
            color="primary"
            onclick={() => showAddStopModal = true}
          >
            <PlusOutline size="xs" class="mr-1" />
            Add Stop
          </Button>
        </div>
        
        {#if routeStops().length === 0}
          <p class="text-sm text-gray-500 text-center py-4">No stops</p>
        {:else}
          <div class="space-y-2">
            {#each routeStops().sort((a, b) => a.sequence - b.sequence) as stop}
              <div 
                class="p-2 rounded border hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer {ptState.selected.stopId === stop.stopId ? 'ring-2 ring-blue-500 border-blue-500' : 'border-gray-200 dark:border-gray-700'}"
                onclick={() => handleStopClick(stop.stopId)}
              >
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-medium text-sm">{stop.stopName}</p>
                    <p class="text-xs text-gray-500">
                      Arr: {stop.arrivalOffset} • Dep: {stop.departureOffset}
                    </p>
                  </div>
                  <Button
                    size="xs"
                    color="red"
                    onclick={(e) => {
                      e.stopPropagation();
                      handleDeleteStop(stop.stopId);
                    }}
                  >
                    <TrashBinOutline size="xs" />
                  </Button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Departures Section -->
      <div class="p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-medium">Departures</h3>
          <Button 
            size="xs" 
            color="primary"
            onclick={() => showAddDepartureModal = true}
          >
            <PlusOutline size="xs" class="mr-1" />
            Add Departure
          </Button>
        </div>
        
        {#if routeDepartures().length === 0}
          <p class="text-sm text-gray-500 text-center py-4">No departures</p>
        {:else}
          <div class="space-y-2">
            {#each routeDepartures().sort((a, b) => a.departureTime.localeCompare(b.departureTime)) as departure}
              <div class="p-2 rounded border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-medium text-sm">{departure.departureTime}</p>
                    {#if departure.vehicleRefId}
                      <p class="text-xs text-gray-500">Vehicle: {departure.vehicleRefId}</p>
                    {/if}
                  </div>
                  <Button
                    size="xs"
                    color="red"
                    onclick={() => handleDeleteDeparture(departure.id)}
                  >
                    <TrashBinOutline size="xs" />
                  </Button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-gray-200 dark:border-gray-700">
      <Button 
        color="red" 
        size="sm"
        class="w-full"
        onclick={handleDeleteRoute}
      >
        <TrashBinOutline size="sm" class="mr-1" />
        Delete Route
      </Button>
    </div>
  {:else}
    <div class="flex items-center justify-center h-full text-gray-500">
      Select a route to view details
    </div>
  {/if}
</div>

{#if ptState.selected.routeId}
  <AddStopModal bind:open={showAddStopModal} routeId={ptState.selected.routeId} />
  <AddDepartureModal bind:open={showAddDepartureModal} routeId={ptState.selected.routeId} />
{/if}