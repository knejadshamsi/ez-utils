import L from 'leaflet';
import { mapState } from './mapState.svelte';
import { ptState } from '$lib/stores/pt.svelte';
import { PTService } from '$lib/api/pt';
import { updatePTVisualization } from './updatePTVisualization';
import type { gui } from '@wailsjs/go/models';

let currentStops: gui.Stop[] = [];
let stopMarkers: Map<string, L.Marker> = new Map();

const MIN_ZOOM_LEVEL = 14; // Don't load stops when zoomed out too far

export async function updateSharedStopsLayer() {
  if (!mapState.map || !mapState.sharedStopsLayer) return;
  
  // Only show shared stops when in ADDING_STOP mode
  if (ptState.editMode !== 'ADDING_STOP') {
    clearSharedStops();
    return;
  }
  
  // Check zoom level
  const zoom = mapState.map.getZoom();
  if (zoom < MIN_ZOOM_LEVEL) {
    clearSharedStops();
    return;
  }
  
  // Get current bounds
  const bounds = mapState.map.getBounds();
  const viewport = {
    minLat: bounds.getSouth(),
    maxLat: bounds.getNorth(),
    minLng: bounds.getWest(),
    maxLng: bounds.getEast()
  };
  
  // Load stops for current mode
  const mode = ptState.selected.mode;
  if (!mode) return;
  
  currentStops = await PTService.loadStopsInBounds(viewport, mode);
  
  // Clear existing markers
  clearSharedStops();
  
  // Add markers for each stop
  currentStops.forEach(stop => {
    const marker = L.marker([stop.lat, stop.lng], {
      icon: L.divIcon({
        className: 'shared-stop-marker',
        html: `<div class="shared-stop-icon" title="${stop.stopName}"></div>`,
        iconSize: [20, 20],
        iconAnchor: [10, 10]
      })
    });
    
    // Add click handler
    marker.on('click', () => handleSharedStopClick(stop));
    
    marker.addTo(mapState.sharedStopsLayer!);
    stopMarkers.set(stop.stopId, marker);
  });
}

function clearSharedStops() {
  if (mapState.sharedStopsLayer) {
    mapState.sharedStopsLayer.clearLayers();
  }
  stopMarkers.clear();
  currentStops = [];
}

export function handleSharedStopClick(stop: gui.Stop) {
  // Prevent event bubbling to map click handler
  L.DomEvent.stopPropagation(window.event as Event);
  
  // Add this stop to the current route
  const routeId = ptState.selected.routeId;
  if (!routeId) return;
  
  const ptData = ptState.getCurrentModeData();
  if (!ptData) return;
  
  // Get current stops for this route
  const routeStops = ptData.stops.filter(s => s.routeId === routeId);
  const nextSequence = Math.max(...routeStops.map(s => s.sequence), -1) + 1;
  
  // Get next stop name
  const nameCount = routeStops.filter(s => s.stopName.startsWith('Stop ')).length;
  const nextName = `Stop ${nameCount + 1}`;
  
  // Create new stop entry for this route
  const newStop = {
    routeId: routeId,
    stopId: stop.stopId,
    sequence: nextSequence,
    stopName: nextName, // Use sequential name, not the original stop name
    lat: stop.lat,
    lng: stop.lng,
    arrivalOffset: '00:00:00',
    departureOffset: '00:00:00',
    stopType: stop.stopType,
    wheelchairAccessible: stop.wheelchairAccessible,
    timingPoint: stop.timingPoint,
    sourceState: 'local' as const
  };
  
  // Add to state
  ptState.addStop(newStop);
  
  // Update visualization to show the new stop
  updatePTVisualization();
}

// Setup map event handlers
export function setupSharedStopsHandlers() {
  if (!mapState.map) return;
  
  // Update on moveend (drag end)
  mapState.map.on('moveend', updateSharedStopsLayer);
  
  // Update on zoomend
  mapState.map.on('zoomend', updateSharedStopsLayer);
}