import { mapState } from '../../map/mapState.svelte';

export interface ViewportBounds {
  minLat: number;
  maxLat: number;
  minLng: number;
  maxLng: number;
}

// Get current viewport bounds or default bounds
export function getCurrentViewportBounds(): ViewportBounds {
  if (mapState.map) {
    const bounds = mapState.map.getBounds();
    return {
      minLat: bounds.getSouth(),
      maxLat: bounds.getNorth(),
      minLng: bounds.getWest(),
      maxLng: bounds.getEast()
    };
  }
  
  // Fallback: Set to default view and get bounds
  return getDefaultViewportBounds();
}

export function getDefaultViewportBounds(): ViewportBounds {
  if (mapState.map) {
    // Temporarily set to default view to calculate bounds
    const currentView = mapState.map.getCenter();
    const currentZoom = mapState.map.getZoom();
    
    // Set to default
    mapState.map.setView([40.5, -73.5], 10);
    const bounds = mapState.map.getBounds();
    
    // Restore previous view if it was different
    if (currentView.lat !== 40.5 || currentView.lng !== -73.5 || currentZoom !== 10) {
      mapState.map.setView(currentView, currentZoom);
    }
    
    return {
      minLat: bounds.getSouth(),
      maxLat: bounds.getNorth(),
      minLng: bounds.getWest(),
      maxLng: bounds.getEast()
    };
  }
  
  // Ultimate fallback - approximate bounds for NYC zoom 10
  return {
    minLat: 40.45,
    maxLat: 40.55,
    minLng: -73.6,
    maxLng: -73.4
  };
}

// Reset map to default position
export function resetMapToDefault() {
  if (mapState.map) {
    mapState.map.setView([40.5, -73.5], 10);
  }
}