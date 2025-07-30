<script lang="ts">
  import { Button, Input, Toggle, Label } from 'flowbite-svelte';
  import { CloseOutline, PlusOutline, TrashBinOutline, CogOutline, MapPinOutline, CheckOutline, CloseCircleOutline, EditOutline } from 'flowbite-svelte-icons';
  import { ptState, TRANSPORT_MODES, type StopType, type AccessibilityStatus } from '$lib/stores/pt.svelte';
  import { appState } from '$lib/stores/app.svelte';
  import DepartureManagementModal from '../modals/pt/DepartureManagementModal.svelte';
  import { mapState } from '../../map/mapState.svelte';
  import type * as L from 'leaflet';
  import { PTService } from '$lib/api/pt';
  import { updatePTVisualization } from '../../map/updatePTVisualization';
  import { updateSharedStopsLayer } from '../../map/sharedStops';
  import CompactSelect from '../CompactSelect.svelte';
  import { nanoid } from 'nanoid';
  
  let editingStopName = $state<{ stopId: string; name: string } | null>(null);
  let editingRouteName = $state<string | null>(null);
  
  // Get current selected route data
  const selectedLine = $derived(() => {
    if (!ptState.selected.lineId) return null;
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.lines.find(l => l.id === ptState.selected.lineId);
  });
  
  const selectedRoute = $derived(() => {
    if (!ptState.selected.routeId) return null;
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.routes.find(r => r.id === ptState.selected.routeId);
  });
  
  const routeStops = $derived(() => {
    if (!ptState.selected.routeId) return [];
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.stops.filter(s => s.routeId === ptState.selected.routeId).sort((a, b) => a.sequence - b.sequence) || [];
  });
  
  const departureCount = $derived(() => {
    if (!ptState.selected.routeId) return 0;
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.departures.filter(d => d.routeId === ptState.selected.routeId).length || 0;
  });

  const routeDepartures = $derived(() => {
    if (!ptState.selected.routeId) return [];
    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    return ptData?.departures.filter(d => d.routeId === ptState.selected.routeId) || [];
  });
  
  const departureItems = $derived(() => {
    const items = [{ value: 'offset', name: 'Show Offsets' }];
    routeDepartures().forEach(dep => {
      items.push({ value: dep.id, name: dep.departureTime });
    });
    return items;
  });
  
  let selectedDepartureId = $state<string>('offset');
  
  function addOffsetToTime(departureTime: string, offset: string): string {
    // Parse departure time (HH:MM)
    const [depHours, depMinutes] = departureTime.split(':').map(Number);
    const [offsetHours, offsetMinutes] = offset.split(':').map(Number);
    
    // Calculate total minutes
    let totalMinutes = depHours * 60 + depMinutes + offsetHours * 60 + offsetMinutes;
    
    // Handle day overflow
    const hours = Math.floor(totalMinutes / 60) % 24;
    const minutes = totalMinutes % 60;
    
    // Format as HH:MM
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
  }

  function handleClose() {
    ptState.selected.routeId = null;
    ptState.selected.stopId = null;
    appState.secondarySidebar = 'HIDDEN';
  }

  function handleDeleteRoute() {
    if (!selectedRoute()) return;
    
    if (confirm(`Delete route "${selectedRoute()?.name}"? This will also delete all stops and departures.`)) {
      ptState.deleteRoute(selectedRoute()!.id);
      handleClose();
    }
  }

  function handleDeleteStop(stopId: string) {
    ptState.deleteStop(stopId);
    // Update map visualization after deletion
    updatePTVisualization();
  }
  
  function handleManageDepartures() {
    ptState.editMode = 'EDITING_DEPARTURES';
  }
  
  function handleStopClick(stopId: string) {
    // Toggle selection
    if (ptState.selected.stopId === stopId) {
      ptState.selected.stopId = null;
    } else {
      ptState.selected.stopId = stopId;
    }
  }
  
  function startAddingStops() {
    if (!mapState.map || !ptState.selected.routeId) return;
    
    // Remove any existing handler first
    const existingHandler = (window as any).__addStopHandler;
    if (existingHandler) {
      mapState.map.off('click', existingHandler);
    }
    
    ptState.editMode = 'ADDING_STOP';
    
    // Update shared stops layer
    updateSharedStopsLayer();
    
    const clickHandler = async (e: L.LeafletMouseEvent) => {
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
      // Get next stop name
      const stopName = 'New Stop';
      
      // Create the stop
      ptState.createStop({
        routeId: ptState.selected.routeId!,
        stopId: `stop_${nanoid(10)}`,
        stopName,
        lat: coords[1],
        lng: coords[0],
        arrivalOffset: '00:00:00',
        departureOffset: '00:00:00'
      });
      
      // Update map visualization
      updatePTVisualization();
    };
    
    mapState.map.on('click', clickHandler);
    (window as any).__addStopHandler = clickHandler;
  }
  
  function stopAddingStops() {
    if (!mapState.map) return;
    
    ptState.editMode = 'NORMAL';
    
    const handler = (window as any).__addStopHandler;
    if (handler) {
      mapState.map.off('click', handler);
      delete (window as any).__addStopHandler;
    }
    
    // Clear shared stops layer
    updateSharedStopsLayer();
  }
  
  // Clean up event handlers on component destroy
  $effect(() => {
    return () => {
      // Clean up any existing click handler when component unmounts
      if (mapState.map && (window as any).__addStopHandler) {
        mapState.map.off('click', (window as any).__addStopHandler);
        delete (window as any).__addStopHandler;
      }
      
      // Reset edit mode if we're in the middle of adding stops
      if (ptState.editMode === 'ADDING_STOP') {
        ptState.editMode = 'NORMAL';
      }
    };
  });
  
  function toggleStopDragging() {
    ptState.editMode = ptState.editMode === 'DRAGGING_STOP' ? 'NORMAL' : 'DRAGGING_STOP';
    updatePTVisualization();
  }
  
  function startEditingStopName(stop: any) {
    editingStopName = { stopId: stop.stopId, name: stop.stopName };
  }
  
  function saveStopName() {
    if (editingStopName && editingStopName.name.trim()) {
      ptState.updateStop(editingStopName.stopId, { stopName: editingStopName.name.trim() });
      editingStopName = null;
    }
  }
  
  function cancelEditingStopName() {
    editingStopName = null;
  }
  
  function updateStopType(stopId: string, type: StopType) {
    ptState.updateStop(stopId, { stopType: type });
  }
  
  function updateAccessibility(stopId: string, status: AccessibilityStatus) {
    ptState.updateStop(stopId, { wheelchairAccessible: status });
  }
  
  function toggleTimingPoint(stopId: string, currentValue: boolean) {
    ptState.updateStop(stopId, { timingPoint: !currentValue });
  }
  
  function updateOffset(stopId: string, field: 'arrivalOffset' | 'departureOffset', value: string) {
    // Accept both HH:MM and HH:MM:SS formats
    if (/^[0-9]{2}:[0-9]{2}(:[0-9]{2})?$/.test(value)) {
      // Normalize to HH:MM:SS format if needed
      const normalizedValue = value.length === 5 ? `${value}:00` : value;
      ptState.updateStop(stopId, { [field]: normalizedValue });
    }
  }
  
  function startEditingRouteName() {
    if (selectedRoute()) {
      editingRouteName = selectedRoute()!.name;
    }
  }
  
  function saveRouteName() {
    if (editingRouteName && editingRouteName.trim() && selectedRoute()) {
      ptState.updateRoute(selectedRoute()!.id, { name: editingRouteName.trim() });
      editingRouteName = null;
    }
  }
  
  function cancelEditingRouteName() {
    editingRouteName = null;
  }
</script>

<div class="flex flex-col h-full bg-gray-800">
  {#if selectedRoute()}
    <!-- Header -->
    <div class="p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between">
        <div class="group">
          <div class="flex items-center gap-2">
            {#if editingRouteName !== null}
              <div class="flex items-center gap-1">
                <Input
                  type="text"
                  bind:value={editingRouteName}
                  size="sm"
                  class="text-lg font-semibold"
                  onkeydown={(e) => {
                    if (e.key === 'Enter') saveRouteName();
                    if (e.key === 'Escape') cancelEditingRouteName();
                  }}
                  autofocus
                />
                <Button size="xs" color="primary" onclick={saveRouteName}>
                  <CheckOutline size="xs" />
                </Button>
                <Button size="xs" color="alternative" onclick={cancelEditingRouteName}>
                  <CloseCircleOutline size="xs" />
                </Button>
              </div>
            {:else}
              <span class="text-lg font-semibold text-white">{selectedRoute()?.name || 'Route'}</span>
            {/if}
            <span class="text-sm text-gray-400">
              {TRANSPORT_MODES[ptState.selected.mode]?.icon}
            </span>
            {#if editingRouteName === null}
              <button
                class="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-gray-700 rounded"
                onclick={startEditingRouteName}
              >
                <EditOutline size="sm" class="text-gray-400" />
              </button>
            {/if}
          </div>
          <p class="text-sm text-gray-400">
            {selectedLine()?.name} • {routeStops().length} stops • {departureCount()} departures
          </p>
        </div>
        <div class="flex items-center gap-2">
          <Button 
            size="xs" 
            color="alternative" 
            onclick={handleManageDepartures}
            title="Manage Departures"
          >
            <CogOutline size="xs" />
          </Button>
          <Button 
            size="xs" 
            color="alternative" 
            onclick={handleClose}
          >
            <CloseOutline size="xs" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto">
      <!-- Stops Section -->
      <div class="p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-medium text-white">Stops</h3>
          {#if routeDepartures().length > 0}
            <div class="w-32">
              <CompactSelect
                bind:value={selectedDepartureId}
                items={departureItems()}
                size="sm"
              />
            </div>
          {/if}
        </div>
        
        {#if routeStops().length === 0}
          <p class="text-sm text-gray-400 text-center py-4">No stops</p>
        {:else}
          <div class="space-y-2">
            {#each routeStops().sort((a, b) => a.sequence - b.sequence) as stop, index}
              <div 
                class="rounded border {ptState.selected.stopId === stop.stopId ? 'ring-2 ring-blue-500 border-blue-500' : 'border-gray-200 dark:border-gray-700'}"
              >
                <button 
                  type="button"
                  class="w-full text-left p-2 bg-gray-800 cursor-pointer group"
                  onclick={() => handleStopClick(stop.stopId)}
                >
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2 flex-1">
                      <span class="text-xs font-bold text-gray-500 bg-gray-700 rounded-full w-6 h-6 flex items-center justify-center">
                        {index + 1}
                      </span>
                      <div class="flex-1">
                        {#if editingStopName?.stopId === stop.stopId}
                          <div class="flex items-center gap-1" onclick={(e) => e.stopPropagation()}>
                            <Input
                              type="text"
                              bind:value={editingStopName.name}
                              size="sm"
                              class="flex-1"
                              onkeydown={(e) => {
                                if (e.key === 'Enter') saveStopName();
                                if (e.key === 'Escape') cancelEditingStopName();
                              }}
                              autofocus
                            />
                            <Button size="xs" color="primary" onclick={saveStopName}>
                              <CheckOutline size="xs" />
                            </Button>
                            <Button size="xs" color="alternative" onclick={cancelEditingStopName}>
                              <CloseCircleOutline size="xs" />
                            </Button>
                          </div>
                        {:else}
                          <div class="flex items-center gap-1">
                            <p class="font-medium text-sm text-white">
                              {stop.stopName}
                            </p>
                            <button
                              class="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-gray-700 rounded"
                              onclick={(e) => { e.stopPropagation(); startEditingStopName(stop); }}
                            >
                              <EditOutline size="xs" class="text-gray-400" />
                            </button>
                          </div>
                        {/if}
                        {#if selectedDepartureId !== 'offset'}
                          {@const selectedDeparture = routeDepartures().find(d => d.id === selectedDepartureId)}
                          {#if selectedDeparture}
                            <p class="text-xs text-gray-300">
                              Arr: {addOffsetToTime(selectedDeparture.departureTime, stop.arrivalOffset)} • 
                              Dep: {addOffsetToTime(selectedDeparture.departureTime, stop.departureOffset)}
                            </p>
                          {/if}
                        {:else}
                          <p class="text-xs text-gray-400">
                            Arrival: {stop.arrivalOffset} • Departure: {stop.departureOffset}
                          </p>
                        {/if}
                      </div>
                    </div>
                    <Button
                      size="xs"
                      color="red"
                      onclick={(e: MouseEvent) => {
                        e.stopPropagation();
                        handleDeleteStop(stop.stopId);
                      }}
                    >
                      <TrashBinOutline size="xs" />
                    </Button>
                  </div>
                </button>
                
                {#if ptState.selected.stopId === stop.stopId}
                  <div class="border-t border-gray-700 p-3 space-y-3 bg-gray-800">
                    <!-- Offset Times -->
                    <div class="flex gap-2">
                      <div class="flex-1">
                        <label class="text-xs text-gray-500">Arrival Offset</label>
                        <Input
                          type="text"
                          value={stop.arrivalOffset}
                          size="sm"
                          pattern="[0-9]{2}:[0-9]{2}(:[0-9]{2})?"
                          placeholder="HH:MM:SS"
                          onchange={(e) => updateOffset(stop.stopId, 'arrivalOffset', e.target.value)}
                        />
                      </div>
                      <div class="flex-1">
                        <label class="text-xs text-gray-500">Departure Offset</label>
                        <Input
                          type="text"
                          value={stop.departureOffset}
                          size="sm"
                          pattern="[0-9]{2}:[0-9]{2}(:[0-9]{2})?"
                          placeholder="HH:MM:SS"
                          onchange={(e) => updateOffset(stop.stopId, 'departureOffset', e.target.value)}
                        />
                      </div>
                    </div>
                    
                    <!-- Accessibility -->
                    <div>
                      <label class="text-xs text-gray-400 block mb-1">Wheelchair Accessible</label>
                      <div class="flex gap-1">
                        <button
                          class="px-2 py-1 text-xs rounded border {stop.wheelchairAccessible === 'YES' ? 'bg-green-600 text-white border-green-600' : 'bg-gray-700 text-gray-300 border-gray-600 hover:bg-gray-600'}"
                          onclick={() => updateAccessibility(stop.stopId, 'YES')}
                        >
                          ♿ Yes
                        </button>
                        <button
                          class="px-2 py-1 text-xs rounded border {stop.wheelchairAccessible === 'NO' ? 'bg-red-600 text-white border-red-600' : 'bg-gray-700 text-gray-300 border-gray-600 hover:bg-gray-600'}"
                          onclick={() => updateAccessibility(stop.stopId, 'NO')}
                        >
                          No
                        </button>
                        <button
                          class="px-2 py-1 text-xs rounded border {stop.wheelchairAccessible === 'UNKNOWN' || !stop.wheelchairAccessible ? 'bg-gray-600 text-white border-gray-600' : 'bg-gray-700 text-gray-300 border-gray-600 hover:bg-gray-600'}"
                          onclick={() => updateAccessibility(stop.stopId, 'UNKNOWN')}
                        >
                          ?
                        </button>
                      </div>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Footer -->
    <div class="p-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
      <Button 
        size="sm" 
        color={ptState.editMode === 'DRAGGING_STOP' ? "yellow" : "alternative"}
        class="w-full"
        onclick={toggleStopDragging}
      >
        <MapPinOutline size="sm" class="mr-1" />
        {ptState.editMode === 'DRAGGING_STOP' ? 'Done Editing Locations' : 'Edit Stop Locations'}
      </Button>
      <Button 
        size="sm" 
        color={ptState.editMode === 'ADDING_STOP' ? "yellow" : "primary"}
        class="w-full"
        onclick={ptState.editMode === 'ADDING_STOP' ? stopAddingStops : startAddingStops}
      >
        {#if ptState.editMode === 'ADDING_STOP'}
          Done Adding Stops
        {:else}
          <PlusOutline size="sm" class="mr-1" />
          Add Stop
        {/if}
      </Button>
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
  {/if}
</div>

{#if ptState.selected.routeId}
  <DepartureManagementModal routeId={ptState.selected.routeId} />
{/if}