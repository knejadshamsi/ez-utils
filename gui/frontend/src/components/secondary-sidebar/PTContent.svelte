<script lang="ts">
  import { Button, Input } from 'flowbite-svelte';
  import { CloseOutline, PlusOutline, TrashBinOutline, MapPinOutline, EditOutline } from 'flowbite-svelte-icons';
  import { ptState } from '$lib/stores/pt.svelte';
  import { appState, mapComponent } from '$lib/stores/app.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { trackStopChange, trackRouteChange, trackRouteStopChange } from '$lib/utils/ptChangeTracking';
  import { updatePTVisualization } from '../../map/updatePTVisualization';
  import { mapState } from '../../map/mapState.svelte';
  import AddDepartureModal from '../modals/pt/AddDepartureModal.svelte';
  import CompactSelect from '../CompactSelect.svelte';
  import { generateId } from '$lib/utils/generateId';
  import type * as L from 'leaflet';

  const selectedRouteData = $derived(ptState.selectedRoute());
  const line = $derived(selectedRouteData?.line);
  const route = $derived(selectedRouteData?.route);
  
  const stopCount = $derived(route?.stopSequence.length || 0);
  let editingCoords = $state<{ stopId: string; lng: string; lat: string } | null>(null);
  let showAddDepartureModal = $state(false);
  let selectedDepartureId = $state<string>('');

  function handleClose() {
    ptState.setSelectedRoute(null);
    ptState.setSelectedLine(null);
    appState.secondarySidebar = 'HIDDEN';
    updatePTVisualization();
  }

  function startAddingStops() {
    if (!line || !route || !mapState.map) return;
    
    ptState.isAddingMultipleStops = true;
    
    const clickHandler = (e: L.LeafletMouseEvent) => {
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
      const stopId = `stop_${Date.now()}`;
      const newStop = {
        id: stopId,
        name: `Stop ${route.stopSequence.length + 1}`,
        location: coords,
        x: coords[0],
        y: coords[1],
        telemetryId: '',
        raw_xml: ''
      };
      
      const newStops = new Map(ptState.stops);
      newStops.set(stopId, newStop);
      ptState.stops = newStops;
      
      trackStopChange(newStop, 'add');
      
      let arrival: string;
      let dwellMinutes = 2;
      
      if (route.stopSequence.length === 0) {
        arrival = route.firstDeparture || '06:00';
      } else {
        const lastStop = route.stopSequence[route.stopSequence.length - 1];
        const lastDepartureMinutes = ptState.timeToMinutes(lastStop.arrival) + lastStop.dwellMinutes;
        arrival = ptState.minutesToTime(lastDepartureMinutes + 5);
      }
      
      const newStopTime = {
        stopId: stopId,
        arrival: arrival,
        dwellMinutes: dwellMinutes,
        sequence: route.stopSequence.length + 1
      };
      
      const currentRouteData = ptState.selectedRoute();
      if (!currentRouteData) return;
      
      const updatedRoute = {
        ...currentRouteData.route,
        stopSequence: [...currentRouteData.route.stopSequence, newStopTime]
      };
      
      const updatedLine = {
        ...currentRouteData.line,
        routes: currentRouteData.line.routes.map(r => r.id === updatedRoute.id ? updatedRoute : r)
      };
      
      const newLines = new Map(ptState.lines);
      newLines.set(updatedLine.id, updatedLine);
      ptState.lines = newLines;
      
      trackRouteStopChange({
        routeId: updatedRoute.id,
        stopId: stopId,
        arrival: arrival,
        dwellMinutes: dwellMinutes,
        sequence: updatedRoute.stopSequence.length - 1,
        routeIndex: 0
      }, 'add');
      
      // Also track the route update
      trackRouteChange(updatedRoute, 'update');
      
      updatePTVisualization();
    };
    
    mapState.map.on('click', clickHandler);
    
    (window as any).__addPTStopHandler = clickHandler;
  }
  
  function stopAddingStops() {
    if (!mapState.map) return;
    
    ptState.isAddingMultipleStops = false;
    
    const handler = (window as any).__addPTStopHandler;
    if (handler) {
      mapState.map.off('click', handler);
      delete (window as any).__addPTStopHandler;
    }
  }

  function selectStopLocation(stopId: string) {
    if (!mapState.map) return;
    
    if (ptState.selectingStopId === stopId) {
      ptState.isSelectingStopLocation = false;
      ptState.selectingStopId = null;
      return;
    }
    
    ptState.isSelectingStopLocation = true;
    ptState.selectingStopId = stopId;
    
    const clickHandler = (e: L.LeafletMouseEvent) => {
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
      const stop = ptState.stops.get(stopId);
      if (stop) {
        const updatedStop = {
          ...stop,
          x: coords[0],
          y: coords[1],
          location: coords
        };
        
        const newStops = new Map(ptState.stops);
        newStops.set(stopId, updatedStop);
        ptState.stops = newStops;
        
        trackStopChange(updatedStop, 'update');
      }
      
      ptState.isSelectingStopLocation = false;
      ptState.selectingStopId = null;
      
      updatePTVisualization();
      
      mapState.map.off('click', clickHandler);
    };
    
    mapState.map.on('click', clickHandler);
  }
  
  function updateStopName(stopId: string, newName: string) {
    const stop = ptState.stops.get(stopId);
    if (stop) {
      const updatedStop = { ...stop, name: newName };
      const newStops = new Map(ptState.stops);
      newStops.set(stopId, updatedStop);
      ptState.stops = newStops;
      
      trackStopChange(updatedStop, 'update');
      
      if (route && route.stopSequence.some(s => s.stopId === stopId)) {
        updatePTVisualization();
      }
    }
  }
  
  function updateArrivalTime(stopId: string, arrivalOffset: string) {
    if (!route) return;
    
    const stopInRoute = route.stopSequence.find(s => s.stopId === stopId);
    if (!stopInRoute) return;
    
    let formattedOffset = arrivalOffset;
    if (arrivalOffset.split(':').length === 2) {
      formattedOffset = arrivalOffset + ':00';
    }
    
    const updatedRoute = {
      ...route,
      stopSequence: route.stopSequence.map(stop =>
        stop.stopId === stopId
          ? { ...stop, arrival: formattedOffset }
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
    
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: formattedOffset,
      dwellMinutes: stopInRoute.dwellMinutes,
      sequence: stopInRoute.sequence,
      routeIndex: 0
    }, 'update');
    
    trackRouteChange(updatedRoute, 'update');
  }
  
  function formatOffset(offset: string): string {
    const parts = offset.split(':');
    if (parts.length === 2) {
      return offset + ':00';
    }
    return offset;
  }

  function calculateDepartureOffset(arrivalOffset: string, dwellMinutes: number): string {
    const parts = arrivalOffset.split(':');
    let seconds = 0;
    if (parts.length >= 2) {
      seconds = parseInt(parts[0]) * 3600 + parseInt(parts[1]) * 60;
      if (parts.length === 3) {
        seconds += parseInt(parts[2]);
      }
    }
    
    const departureSeconds = seconds + (dwellMinutes * 60);
    
    const hours = Math.floor(departureSeconds / 3600);
    const minutes = Math.floor((departureSeconds % 3600) / 60);
    const secs = departureSeconds % 60;
    
    return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  }
  
  function updateDepartureOffset(stopId: string, departureOffset: string) {
    if (!route) return;
    
    const stopInRoute = route.stopSequence.find(s => s.stopId === stopId);
    if (!stopInRoute) return;
    
    const arrivalParts = stopInRoute.arrival.split(':');
    let arrivalSeconds = 0;
    if (arrivalParts.length >= 2) {
      arrivalSeconds = parseInt(arrivalParts[0]) * 3600 + parseInt(arrivalParts[1]) * 60;
      if (arrivalParts.length === 3) {
        arrivalSeconds += parseInt(arrivalParts[2]);
      }
    }
    
    const departureParts = departureOffset.split(':');
    let departureSeconds = 0;
    if (departureParts.length >= 2) {
      departureSeconds = parseInt(departureParts[0]) * 3600 + parseInt(departureParts[1]) * 60;
      if (departureParts.length === 3) {
        departureSeconds += parseInt(departureParts[2]);
      }
    }
    
    const dwellMinutes = Math.max(0, Math.floor((departureSeconds - arrivalSeconds) / 60));
    
    const updatedRoute = {
      ...route,
      stopSequence: route.stopSequence.map(stop =>
        stop.stopId === stopId
          ? { ...stop, dwellMinutes: dwellMinutes }
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
    
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: stopInRoute.arrival,
      dwellMinutes: dwellMinutes,
      sequence: stopInRoute.sequence,
      routeIndex: 0
    }, 'update');
    
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
    
    trackRouteChange(updatedRoute, 'update');
  }

  function handleDeleteStopFromRoute(stopId: string) {
    if (!route || !line) return;
    
    const stopIndex = route.stopSequence.findIndex(s => s.stopId === stopId);
    if (stopIndex === -1) return;
    
    const stopToDelete = route.stopSequence[stopIndex];
    
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
    
    trackRouteStopChange({
      routeId: route.id,
      stopId: stopId,
      arrival: stopToDelete.arrival,
      dwellMinutes: stopToDelete.dwellMinutes,
      sequence: stopToDelete.sequence,
      routeIndex: 0
    }, 'delete');
    
    trackRouteChange(updatedRoute, 'update');
    
    updatePTVisualization();
  }
  
  function handleDeleteRoute() {
    if (!line || !route) return;
    
    trackRouteChange(route, 'delete');
    
    ptState.deleteRoute(line.id, route.id);
    
    appState.secondarySidebar = 'HIDDEN';
  }
  
  function startEditingCoords(stopId: string) {
    const stop = ptState.stops.get(stopId);
    if (stop) {
      editingCoords = {
        stopId,
        lng: stop.lng?.toString() || '0',
        lat: stop.lat?.toString() || '0'
      };
    }
  }
  
  function saveCoords() {
    if (!editingCoords) return;
    
    const lng = parseFloat(editingCoords.lng);
    const lat = parseFloat(editingCoords.lat);
    
    if (isNaN(lng) || isNaN(lat)) {
      alert('Invalid coordinates. Please enter valid numbers.');
      return;
    }
    
    const stop = ptState.stops.get(editingCoords.stopId);
    if (stop) {
      const updatedStop = { ...stop, lng, lat, location: [lng, lat] as [number, number] };
      const newStops = new Map(ptState.stops);
      newStops.set(editingCoords.stopId, updatedStop);
      ptState.stops = newStops;
      
      trackStopChange(updatedStop, 'update');
      
      updatePTVisualization();
    }
    
    editingCoords = null;
  }
  
  function cancelEditingCoords() {
    editingCoords = null;
  }
  
  function toggleStopDragging() {
    ptState.isDraggingStop = !ptState.isDraggingStop;
    updatePTVisualization();
  }
  
  function handleDepartureAdd(departure: any) {
    if (!line || !route) return;
    
    ptState.addDeparture(line.id, route.id, departure);
    
  }
  
  function handleDeleteDeparture() {
    if (!line || !route || !selectedDepartureId || !route.departures) return;
    
    const departureIndex = route.departures.findIndex(d => d.id === selectedDepartureId);
    if (departureIndex === -1) return;
    
    const updatedDepartures = route.departures.filter(d => d.id !== selectedDepartureId);
    
    const updatedRoute = {
      ...route,
      departures: updatedDepartures
    };
    
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    if (updatedDepartures.length > 0) {
      const nextIndex = Math.min(departureIndex, updatedDepartures.length - 1);
      selectedDepartureId = updatedDepartures[nextIndex].id;
    } else {
      selectedDepartureId = '';
    }
    
  }
  
  $effect(() => {
    if (route && (!route.departures || route.departures.length === 0)) {
      const defaultDeparture = {
        id: generateId(),
        route_id: route.id,
        routeId: route.id,
        departure_time: '06:00:00',
        departureTime: '06:00:00',
        vehicle_id: `${line?.mode.toLowerCase()}_1`,
        raw_xml: ''
      };
      handleDepartureAdd(defaultDeparture);
    }
  });

</script>

{#if route && line}
  <div class="h-full flex flex-col bg-gray-800">
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

    <div class="px-4 py-3 border-b border-gray-600">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-300">Departures</h3>
        <div class="flex items-center gap-2">
          <CompactSelect
            size="sm"
            bind:value={selectedDepartureId}
            items={[
              { value: '', name: 'Select departure' },
              ...(route?.departures?.map(dep => ({
                value: dep.id,
                name: `${dep.departureTime} - ${dep.vehicle_id || 'No vehicle'}`
              })) || [{ value: '', name: 'No departures yet' }])
            ]}
          />
          {#if selectedDepartureId && route?.departures && route.departures.length > 1}
            <Button
              size="xs"
              color="red"
              onclick={handleDeleteDeparture}
              class="text-xs"
            >
              <TrashBinOutline class="w-3 h-3" />
            </Button>
          {/if}
          <Button
            size="xs"
            color="primary"
            onclick={() => showAddDepartureModal = true}
            class="text-xs"
          >
            <PlusOutline class="w-3 h-3 mr-1" />
            Add Departure
          </Button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto px-5 py-5">
      {#if route.stopSequence.length === 0}
        <div class="text-sm text-gray-400 italic text-center py-8">No stops defined</div>
      {:else}
        <div class="space-y-4">
          {#each route.stopSequence as stop, stopIndex (`${stop.stopId}-${stopIndex}`)}
            <div class="relative">
              <div class="bg-gray-700 rounded-lg p-4 border border-gray-600 relative">
                <div 
                  class="absolute -left-3 top-4 w-6 h-6 rounded-full flex items-center justify-center text-xs text-white font-medium bg-blue-500"
                >
                  {stopIndex + 1}
                </div>
                
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
                      <span class="text-gray-300">Arrival Offset:</span>
                      <Input 
                        type="text" 
                        size="sm" 
                        value={formatOffset(stop.arrival || '00:00:00')}
                        placeholder="HH:MM:SS"
                        pattern="[0-9]{2}:[0-9]{2}:[0-9]{2}"
                        onchange={(e) => updateArrivalTime(stop.stopId, (e.target as HTMLInputElement).value)}
                        class="bg-gray-600 text-white border-gray-500 h-6 text-xs mt-1"
                      />
                    </div>
                    <div>
                      <span class="text-gray-300">Departure Offset:</span>
                      <Input 
                        type="text" 
                        size="sm" 
                        value={calculateDepartureOffset(stop.arrival, stop.dwellMinutes)}
                        placeholder="HH:MM:SS"
                        pattern="[0-9]{2}:[0-9]{2}:[0-9]{2}"
                        onchange={(e) => updateDepartureOffset(stop.stopId, (e.target as HTMLInputElement).value)}
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
                          bind:value={editingCoords.lng}
                          placeholder="Lng"
                          class="bg-gray-600 text-white border-gray-500 h-6 text-xs w-20"
                        />
                        <span class="text-xs text-gray-400">,</span>
                        <Input 
                          type="text"
                          size="sm"
                          bind:value={editingCoords.lat}
                          placeholder="Lat"
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
              
              {#if stopIndex < route.stopSequence.length - 1}
                <div class="absolute left-0 top-16 w-0.5 h-8 bg-gray-500 ml-2"></div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
      
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
    
    <div class="p-4 border-t border-gray-600 space-y-2">
      <Button 
        color={ptState.isDraggingStop ? "yellow" : "alternative"} 
        size="sm" 
        class="w-full" 
        onclick={toggleStopDragging}
      >
        <MapPinOutline class="w-4 h-4 mr-2" />
        {ptState.isDraggingStop ? 'Done Editing Stop Locations' : 'Edit Stop Locations'}
      </Button>
      <Button
        size="sm"
        color={ptState.isAddingMultipleStops ? "yellow" : "primary"}
        class="w-full"
        onclick={ptState.isAddingMultipleStops ? stopAddingStops : startAddingStops}
      >
        {#if ptState.isAddingMultipleStops}
          Done Adding Stops
        {:else}
          <PlusOutline class="w-4 h-4 mr-2" />
          Add Stop
        {/if}
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

{#if route}
  <AddDepartureModal 
    bind:open={showAddDepartureModal} 
    routeId={route.id}
    onDepartureAdd={handleDepartureAdd}
  />
{/if}