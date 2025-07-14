<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import { ptState } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';

  let { map }: { map: MaplibreMap | null } = $props();

  export function startAddingStop() {
    // Use the currently selected route from ptState
    if (!ptState.selectedRouteId) {
      console.warn('[PTMapInteraction] No route selected');
      return;
    }
    
    ptState.isAddingStop = true;
    if (map) {
      map.getCanvas().style.cursor = 'crosshair';
    }
  }
  
  function stopAddingStop() {
    ptState.isAddingStop = false;
    if (map) {
      map.getCanvas().style.cursor = '';
    }
  }

  function handleMapClick(e: any) {
    if (!ptState.isAddingStop || !map) return;

    const selectedRouteData = ptState.selectedRoute();
    if (!selectedRouteData) {
      console.error('[PTMapInteraction] No route selected');
      return;
    }

    const { line, route } = selectedRouteData;
    const { lng, lat } = e.lngLat;

    const stopId = generateId();
    const stopName = `Stop ${route.stopSequence.length + 1}`;
    
    const newStop = {
      id: stopId,
      name: stopName,
      location: [lng, lat] as [number, number],
      x: lng,
      y: lat,
      telemetryId: '',
      raw_xml: ''
    };

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'stop',
      action: 'add',
      processId: getCurrentProcessId(),
      stop: newStop
    });

    // Update stops map to trigger reactivity
    const newStops = new Map(ptState.stops);
    newStops.set(stopId, newStop);
    ptState.stops = newStops;

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

    // Calculate arrival offset from first departure
    const arrivalOffset = ptState.timeToMinutes(arrival) - ptState.timeToMinutes(route.firstDeparture);
    const departureOffset = arrivalOffset + dwellMinutes;

    const routeStop = {
      id: generateId(),
      route_id: route.id,
      stop_ref_id: stopId,
      stop_order: route.stopSequence.length + 1,
      arrival_offset: `PT${arrivalOffset}M`,
      departure_offset: `PT${departureOffset}M`,
      raw_xml: ''
    };

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'routeStop',
      action: 'add',
      processId: getCurrentProcessId(),
      routeStop: routeStop
    });

    // Create new route with updated stop sequence to trigger reactivity
    const updatedRoute = {
      ...route,
      stopSequence: [...route.stopSequence, newStopTime]
    };
    
    // Create new line with updated route
    const updatedLine = {
      ...line,
      routes: line.routes.map(r => r.id === route.id ? updatedRoute : r)
    };
    
    // Update state to trigger reactivity
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, updatedLine);
    ptState.lines = newLines;
    
    console.log('[PTMapInteraction] Added stop. New stop count:', updatedRoute.stopSequence.length);
    
    // Increment version to force derived property updates
    ptState.updateVersion++;

    stopAddingStop();
    updateMapData();
  }

  function updateMapData() {
    if (!map) return;

    const stopsGeoJSON = {
      type: 'FeatureCollection' as const,
      features: Array.from(ptState.stops.values())
        .filter(stop => ptState.visibility.stops)
        .map(stop => ({
          type: 'Feature' as const,
          properties: {
            id: stop.id,
            name: stop.name,
            type: 'stop'
          },
          geometry: {
            type: 'Point' as const,
            coordinates: stop.location
          }
        }))
    };

    const routesGeoJSON = {
      type: 'FeatureCollection' as const,
      features: ptState.visibility.routes ? Array.from(ptState.lines.values())
        .filter(line => ptState.visibleModes.has(line.mode as any))
        .flatMap(line => 
          line.routes.map(route => {
            const coordinates = route.stopSequence
              .map(st => ptState.stops.get(st.stopId))
              .filter(stop => stop !== undefined)
              .map(stop => stop!.location);

            return {
              type: 'Feature' as const,
              properties: {
                id: route.id,
                lineId: line.id,
                lineName: line.name,
                direction: route.direction,
                color: line.color,
                mode: line.mode,
                type: 'route'
              },
              geometry: {
                type: 'LineString' as const,
                coordinates: coordinates
              }
            };
          })
        ) : []
    };

    if (map.getSource('pt-stops')) {
      (map.getSource('pt-stops') as any).setData(stopsGeoJSON);
    } else {
      map.addSource('pt-stops', {
        type: 'geojson',
        data: stopsGeoJSON
      });
    }

    if (map.getSource('pt-routes')) {
      (map.getSource('pt-routes') as any).setData(routesGeoJSON);
    } else {
      map.addSource('pt-routes', {
        type: 'geojson',
        data: routesGeoJSON
      });
    }

    if (!map.getLayer('pt-routes-layer')) {
      map.addLayer({
        id: 'pt-routes-layer',
        type: 'line',
        source: 'pt-routes',
        paint: {
          'line-color': ['get', 'color'],
          'line-width': [
            'case',
            ['==', ['get', 'id'], ptState.selectedRouteId || ''],
            4,
            2
          ],
          'line-opacity': [
            'case',
            ['==', ['get', 'id'], ptState.selectedRouteId || ''],
            1,
            0.6
          ]
        }
      });
    }

    if (!map.getLayer('pt-stops-layer')) {
      map.addLayer({
        id: 'pt-stops-layer',
        type: 'circle',
        source: 'pt-stops',
        paint: {
          'circle-radius': 6,
          'circle-color': '#ffffff',
          'circle-stroke-color': '#333333',
          'circle-stroke-width': 2
        }
      });
    }

    if (!map.getLayer('pt-stops-selected')) {
      map.addLayer({
        id: 'pt-stops-selected',
        type: 'circle',
        source: 'pt-stops',
        filter: ['==', ['get', 'id'], ''],
        paint: {
          'circle-radius': 8,
          'circle-color': '#3B82F6',
          'circle-stroke-color': '#ffffff',
          'circle-stroke-width': 3
        }
      });
    }

    const selectedStopIds = new Set<string>();
    if (ptState.selectedRouteId) {
      const selectedRouteData = ptState.selectedRoute();
      if (selectedRouteData) {
        for (const stop of selectedRouteData.route.stopSequence) {
          selectedStopIds.add(stop.stopId);
        }
      }
    }

    map.setFilter('pt-stops-selected', ['in', ['get', 'id'], ['literal', Array.from(selectedStopIds)]]);
  }

  $effect(() => {
    updateMapData();
  });

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && ptState.isAddingStop) {
      stopAddingStop();
    }
  }

  onMount(() => {
    if (map) {
      map.on('click', handleMapClick);
      updateMapData();
    }
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    if (map) {
      map.off('click', handleMapClick);
      
      if (map.getLayer('pt-stops-selected')) map.removeLayer('pt-stops-selected');
      if (map.getLayer('pt-stops-layer')) map.removeLayer('pt-stops-layer');
      if (map.getLayer('pt-routes-layer')) map.removeLayer('pt-routes-layer');
      
      if (map.getSource('pt-stops')) map.removeSource('pt-stops');
      if (map.getSource('pt-routes')) map.removeSource('pt-routes');
    }
    window.removeEventListener('keydown', handleKeyDown);
  });
</script>