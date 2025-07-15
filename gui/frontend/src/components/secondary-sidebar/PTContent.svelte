<script lang="ts">
  import { Button, Input } from 'flowbite-svelte';
  import { CloseOutline, PlusOutline, TrashBinOutline, MapPinOutline, EditOutline } from 'flowbite-svelte-icons';
  import { ptState } from '$lib/stores/pt.svelte';
  import { appState, mapComponent, editingSession } from '$lib/stores/app.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { trackStopChange, trackRouteChange, trackRouteStopChange } from '$lib/utils/ptChangeTracking';

  const selectedRouteData = $derived(ptState.selectedRoute());
  const line = $derived(selectedRouteData?.line);
  const route = $derived(selectedRouteData?.route);
  
  // Force re-render when stops change
  const stopCount = $derived(route?.stopSequence.length || 0);
  let editingCoords = $state<{ stopId: string; x: string; y: string } | null>(null);
  
  // Debug logging
  $effect(() => {
    console.log('[SecondaryPTSidebar] selectedRouteData:', selectedRouteData);
    console.log('[SecondaryPTSidebar] line:', line);
    console.log('[SecondaryPTSidebar] route:', route);
    console.log('[SecondaryPTSidebar] ptState.selectedRouteId:', ptState.selectedRouteId);
    console.log('[SecondaryPTSidebar] stopCount:', stopCount);
    if (route) {
      console.log('[SecondaryPTSidebar] Route stops:', route.stopSequence.length);
      console.log('[SecondaryPTSidebar] Stop IDs:', route.stopSequence.map((s: any) => s.stopId));
    }
  });

  function handleClose() {
    ptState.setSelectedRoute(null);
    ptState.setSelectedLine(null);
    appState.secondarySidebar = 'HIDDEN';
  }

  function handleAddStopToRoute() {
    if (!line || !route) return;
    
    console.log('[SecondaryPTSidebar] Adding new stop to route:', route.id);
    
    // Create a new stop with default values (similar to population activity)
    const stopId = `stop_${Date.now()}`;
    const newStop = {
      id: stopId,
      name: `New Stop`,
      location: [0, 0] as [number, number], // Will be set by map click
      x: 0,
      y: 0,
      telemetryId: '',
      raw_xml: ''
    };
    
    // Add stop to the state
    const newStops = new Map(ptState.stops);
    newStops.set(stopId, newStop);
    ptState.stops = newStops;
    
    // Track the new stop
    trackStopChange(newStop, 'add');
    
    // Calculate arrival time for the new stop
    let arrival: string;
    let dwellMinutes = 2;
    
    if (route.stopSequence.length === 0) {
      arrival = route.firstDeparture || '06:00';
    } else {
      const lastStop = route.stopSequence[route.stopSequence.length - 1];
      const lastDepartureMinutes = ptState.timeToMinutes(lastStop.arrival) + lastStop.dwellMinutes;
      arrival = ptState.minutesToTime(lastDepartureMinutes + 5);
    }
    
    // Add to route sequence
    const newStopTime = {
      stopId: stopId,
      arrival: arrival,
      dwellMinutes: dwellMinutes,
      sequence: route.stopSequence.length + 1
    };
    
    // Update route with new stop
    const updatedRoute = {
      ...route,
      stopSequence: [...route.stopSequence, newStopTime]
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    // Track the new route stop
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: arrival,
      dwellMinutes: dwellMinutes,
      sequence: route.stopSequence.length,
      routeIndex: 0
    }, 'add');
    
    // Also track the route update
    trackRouteChange(updatedRoute, 'update');
    
    // Immediately prompt for location selection (like population)
    ptState.isSelectingStopLocation = true;
    ptState.selectingStopId = stopId;
    
    console.log('[SecondaryPTSidebar] Created stop, entering location selection mode');
  }

  function selectStopLocation(stopId: string) {
    ptState.isSelectingStopLocation = true;
    ptState.selectingStopId = stopId;
  }
  
  function updateStopName(stopId: string, newName: string) {
    const stop = ptState.stops.get(stopId);
    if (stop) {
      const updatedStop = { ...stop, name: newName };
      const newStops = new Map(ptState.stops);
      newStops.set(stopId, updatedStop);
      ptState.stops = newStops;
      
      // Track the stop update
      trackStopChange(updatedStop, 'update');
    }
  }
  
  function updateArrivalTime(stopId: string, arrivalTime: string) {
    if (!route) return;
    
    // Find the stop in the route sequence
    const stopInRoute = route.stopSequence.find(s => s.stopId === stopId);
    if (!stopInRoute) return;
    
    // Update the stop in the route sequence
    const updatedRoute = {
      ...route,
      stopSequence: route.stopSequence.map(stop =>
        stop.stopId === stopId
          ? { ...stop, arrival: arrivalTime }
          : stop
      )
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    // Track the route stop update
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: arrivalTime,
      dwellMinutes: stopInRoute.dwellMinutes,
      sequence: stopInRoute.sequence,
      routeIndex: 0
    }, 'update');
    
    // Also track the route update
    trackRouteChange(updatedRoute, 'update');
  }
  
  function updateDwellTime(stopId: string, dwellMinutes: number) {
    if (!route) return;
    
    // Ensure minimum 1 minute dwell time
    const validDwellTime = Math.max(1, dwellMinutes);
    
    // Find the stop in the route sequence
    const stopInRoute = route.stopSequence.find(s => s.stopId === stopId);
    if (!stopInRoute) return;
    
    // Find and update the stop in the route sequence
    const updatedRoute = {
      ...route,
      stopSequence: route.stopSequence.map(stop =>
        stop.stopId === stopId
          ? { ...stop, dwellMinutes: validDwellTime }
          : stop
      )
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    // Track the route stop update
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: stopInRoute.arrival,
      dwellMinutes: validDwellTime,
      sequence: stopInRoute.sequence,
      routeIndex: 0
    }, 'update');
    
    // Also track the route update
    trackRouteChange(updatedRoute, 'update');
  }
  
  function updateRouteName(newName: string) {
    if (!route || !line) return;
    
    const updatedRoute = {
      ...route,
      direction: newName
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    // Track the route update
    trackRouteChange(updatedRoute, 'update');
  }

  function handleDeleteStopFromRoute(stopId: string) {
    if (!route || !line) return;
    
    // Find the stop in the route sequence
    const stopIndex = route.stopSequence.findIndex(s => s.stopId === stopId);
    if (stopIndex === -1) return;
    
    const stopToDelete = route.stopSequence[stopIndex];
    
    // Update the route by removing the stop
    const updatedRoute = {
      ...route,
      stopSequence: route.stopSequence.filter(s => s.stopId !== stopId)
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    // Track the route stop deletion
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: stopToDelete.arrival,
      dwellMinutes: stopToDelete.dwellMinutes,
      sequence: stopToDelete.sequence,
      routeIndex: 0
    }, 'delete');
    
    // Also track the route update
    trackRouteChange(updatedRoute, 'update');
  }
  
  function handleDeleteRoute() {
    if (!line || !route) return;
    
    // Track the route deletion
    trackRouteChange(route, 'delete');
    
    // Delete from local state
    ptState.deleteRoute(line.id, route.id);
    
    // Close the sidebar
    appState.secondarySidebar = 'HIDDEN';
  }
  
  function startEditingCoords(stopId: string) {
    const stop = ptState.stops.get(stopId);
    if (stop) {
      editingCoords = {
        stopId,
        x: stop.x?.toString() || '0',
        y: stop.y?.toString() || '0'
      };
    }
  }
  
  function saveCoords() {
    if (!editingCoords) return;
    
    const x = parseFloat(editingCoords.x);
    const y = parseFloat(editingCoords.y);
    
    if (isNaN(x) || isNaN(y)) {
      alert('Invalid coordinates. Please enter valid numbers.');
      return;
    }
    
    const stop = ptState.stops.get(editingCoords.stopId);
    if (stop) {
      const updatedStop = { ...stop, x, y, location: [x, y] as [number, number] };
      const newStops = new Map(ptState.stops);
      newStops.set(editingCoords.stopId, updatedStop);
      ptState.stops = newStops;
      
      // Track the stop update
      trackStopChange(updatedStop, 'update');
    }
    
    editingCoords = null;
  }
  
  function cancelEditingCoords() {
    editingCoords = null;
  }

</script>

{#if route && line}
  <div class="h-full flex flex-col bg-gray-800">
    <!-- Header with route info and close button -->
    <div class="flex items-center justify-between p-4 border-b border-gray-600">
      <div class="flex-1">
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-300">Route:</span>
          <Input 
            type="text"
            size="sm"
            value={route.direction || ''}
            placeholder="Unnamed Route"
            onchange={(e) => updateRouteName((e.target as HTMLInputElement).value)}
            class="bg-gray-700 text-white border-gray-600 h-7 text-sm font-semibold w-48"
          />
        </div>
        <p class="text-sm text-gray-400 mt-1">
          {ptState.getModeIcon(line.mode)} {line.name} • {stopCount} stops
        </p>
      </div>
      <Button
        size="xs"
        color="none"
        onclick={handleClose}
        class="text-gray-400 hover:text-white p-1"
      >
        <CloseOutline size="sm" />
      </Button>
    </div>


    <!-- Main scrollable content -->
    <div class="flex-1 overflow-y-auto px-5 py-5">
      {#if route.stopSequence.length === 0}
        <div class="text-sm text-gray-400 italic text-center py-8">No stops defined</div>
      {:else}
        <div class="space-y-4">
          {#each route.stopSequence as stop, stopIndex (`${stop.stopId}-${stopIndex}`)}
            <!-- Stop card -->
            <div class="relative">
              <div class="bg-gray-700 rounded-lg p-4 border border-gray-600 relative">
                <!-- Stop number indicator -->
                <div class="absolute -left-3 top-4 w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center text-xs text-white font-medium">
                  {stopIndex + 1}
                </div>
                
                <!-- Stop info -->
                <div class="ml-4">
                  <div class="flex items-center justify-between mb-2">
                    <Input 
                      type="text"
                      size="sm"
                      value={ptState.getStopName(stop.stopId)}
                      onchange={(e) => updateStopName(stop.stopId, (e.target as HTMLInputElement).value)}
                      class="bg-gray-600 text-white border-gray-500 h-7 text-sm font-medium flex-1 mr-2"
                      placeholder="Stop name"
                    />
                    <Button size="xs" color="red" class="h-6 px-2" onclick={() => handleDeleteStopFromRoute(stop.stopId)}>
                      <TrashBinOutline class="w-3 h-3" />
                    </Button>
                  </div>
                  
                  <div class="grid grid-cols-2 gap-3 text-xs">
                    <div>
                      <span class="text-gray-300">Arrival:</span>
                      <Input 
                        type="time" 
                        size="sm" 
                        value={stop.arrival || ''}
                        onchange={(e) => updateArrivalTime(stop.stopId, (e.target as HTMLInputElement).value)}
                        class="bg-gray-600 text-white border-gray-500 h-6 text-xs mt-1"
                      />
                    </div>
                    <div>
                      <span class="text-gray-300">Dwell (mins):</span>
                      <Input 
                        type="number" 
                        size="sm" 
                        value={stop.dwellMinutes || 2}
                        min="1"
                        onchange={(e) => updateDwellTime(stop.stopId, parseInt((e.target as HTMLInputElement).value) || 1)}
                        class="bg-gray-600 text-white border-gray-500 h-6 text-xs mt-1"
                      />
                    </div>
                  </div>
                  
                  <div class="flex items-center gap-2 mt-2">
                    <span class="text-xs text-gray-300">Location:</span>
                    {#if editingCoords?.stopId === stop.stopId}
                      <div class="flex items-center gap-1 flex-1">
                        <Input 
                          type="text"
                          size="sm"
                          bind:value={editingCoords.x}
                          placeholder="X"
                          class="bg-gray-600 text-white border-gray-500 h-6 text-xs w-20"
                        />
                        <span class="text-xs text-gray-400">,</span>
                        <Input 
                          type="text"
                          size="sm"
                          bind:value={editingCoords.y}
                          placeholder="Y"
                          class="bg-gray-600 text-white border-gray-500 h-6 text-xs w-20"
                        />
                        <Button size="xs" color="green" onclick={saveCoords} class="h-6 px-2">
                          ✓
                        </Button>
                        <Button size="xs" color="red" onclick={cancelEditingCoords} class="h-6 px-2">
                          ✗
                        </Button>
                      </div>
                    {:else}
                      <span class="text-xs text-gray-400 flex-1">{ptState.getStopCoords(stop.stopId)}</span>
                      <Button 
                        size="xs" 
                        color="alternative"
                        onclick={() => startEditingCoords(stop.stopId)}
                        class="h-6 px-2 bg-gray-600 hover:bg-gray-500 text-white border-gray-500"
                      >
                        <EditOutline class="w-3 h-3" />
                      </Button>
                      <Button 
                        size="xs" 
                        color={ptState.selectingStopId === stop.stopId ? "yellow" : "alternative"}
                        onclick={() => selectStopLocation(stop.stopId)}
                        class="h-6 px-2 {ptState.selectingStopId === stop.stopId ? '' : 'bg-gray-600 hover:bg-gray-500 text-white border-gray-500'}"
                      >
                        <MapPinOutline class="w-3 h-3 mr-1" />
                        {ptState.selectingStopId === stop.stopId ? 'Click map' : 'Set'}
                      </Button>
                    {/if}
                  </div>
                </div>
              </div>
              
              <!-- Visual connector to next stop -->
              {#if stopIndex < route.stopSequence.length - 1}
                <div class="absolute left-0 top-16 w-0.5 h-8 bg-gray-500 ml-2"></div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
      
      <!-- Location selection feedback -->
      {#if ptState.isSelectingStopLocation}
        <div class="mt-4 p-3 bg-yellow-900/20 rounded-lg border border-yellow-800">
          <p class="text-xs text-yellow-300 text-center">
            Click on the map to set stop location
          </p>
          <p class="text-xs text-yellow-400 text-center mt-1">
            The stop with yellow highlight is being positioned
          </p>
        </div>
      {/if}
    </div>
    
    <!-- Footer -->
    <div class="p-4 border-t border-gray-600 space-y-2">
      <Button
        size="sm"
        color="primary"
        class="w-full"
        onclick={handleAddStopToRoute}
      >
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Stop
      </Button>
      <Button
        size="sm"
        color="red"
        class="w-full"
        onclick={handleDeleteRoute}
      >
        <TrashBinOutline class="w-4 h-4 mr-2" />
        Delete Route
      </Button>
    </div>
  </div>
{:else}
  <div class="h-full flex items-center justify-center text-gray-400 bg-gray-800">
    <div class="text-center">
      <p class="mb-2">No route selected</p>
      <p class="text-sm">Select a route from the primary sidebar to view its details</p>
    </div>
  </div>
{/if}