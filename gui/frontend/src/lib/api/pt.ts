import { GetPTLinesByMode, GetPTRoutesByLineID, GetPTStopsByRouteID, GetPTDeparturesByRouteID, GetPTStopsInBounds, GetPTStops } from '@wailsjs/go/gui/App';
import { ptState, type TransportMode } from '$lib/stores/pt.svelte';
import { appState } from '$lib/stores/app.svelte';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(mode: TransportMode): Promise<void> {
    try {
      // Clear persisted data first
      ptState.clearPersistedData();
      
      // Load only lines for the mode initially
      const lines = await GetPTLinesByMode(appState.processId, mode);
      
      // Initialize empty arrays for routes, stops, and departures
      // These will be loaded lazily when user expands lines
      const routes = [];
      const stops = [];
      const departures = [];
      
      // Load into persisted state with just lines
      ptState.loadPTData({ lines, routes, stops, departures });
      
    } catch (error) {
      throw new Error(`Failed to load ${mode} data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async loadLineRoutes(lineId: string): Promise<void> {
    try {
      // Load routes for this line
      const routes = await GetPTRoutesByLineID(appState.processId, lineId);
      
      const ptData = ptState.getCurrentModeData();
      
      // Remove existing routes for this line (preserving local/modified)
      ptData.routes = ptData.routes.filter(r => 
        r.lineId !== lineId || r.sourceState !== 'persisted'
      );
      
      // Add new routes with persisted sourceState
      const persistedRoutes = routes.map(r => ({ 
        ...r, 
        sourceState: 'persisted' as const 
      }));
      ptData.routes.push(...persistedRoutes);
      
      // Load stops and departures for each route
      for (const route of routes) {
        await PTService.loadRouteStops(route.id);
        await PTService.loadRouteDepartures(route.id);
      }
    } catch (error) {
      throw new Error(`Failed to load routes for line: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async loadRouteStops(routeId: string): Promise<void> {
    try {
      // Get route-stop junctions
      const routeStops = await GetPTStopsByRouteID(appState.processId, routeId);
      
      // Get stop IDs
      const stopIds = routeStops.map(rs => rs.stopId);
      
      // Get actual stop data
      const stops = await GetPTStops(appState.processId, stopIds);
      
      // Create a map for quick lookup
      const stopMap = new Map(stops.map(s => [s.stopId, s]));
      
      const ptData = ptState.getCurrentModeData();
      
      // Remove existing stops for this route (preserving local/modified)
      ptData.stops = ptData.stops.filter(s => 
        s.routeId !== routeId || s.sourceState !== 'persisted'
      );
      
      // Combine route-stop and stop data
      const combinedStops = routeStops.map(rs => {
        const stop = stopMap.get(rs.stopId);
        if (!stop) return null;
        
        return {
          routeId: rs.routeId,
          stopId: rs.stopId,
          sequence: rs.sequence,
          stopName: stop.stopName,
          lat: stop.lat,
          lng: stop.lng,
          arrivalOffset: stop.arrivalOffset,
          departureOffset: stop.departureOffset,
          stopType: stop.stopType,
          wheelchairAccessible: stop.wheelchairAccessible,
          timingPoint: stop.timingPoint,
          sourceState: 'persisted' as const
        };
      }).filter(s => s !== null);
      
      ptData.stops.push(...combinedStops);
    } catch (error) {
      throw new Error(`Failed to load stops for route: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async loadStopsInBounds(bounds: { minLat: number; maxLat: number; minLng: number; maxLng: number }, mode: TransportMode): Promise<gui.Stop[]> {
    try {
      const stops = await GetPTStopsInBounds(appState.processId, mode, bounds);
      return stops;
    } catch (error) {
      return [];
    }
  }
  
  static async loadRouteDepartures(routeId: string): Promise<void> {
    try {
      const departures = await GetPTDeparturesByRouteID(appState.processId, routeId);
      const ptData = ptState.getCurrentModeData();
      
      // Remove existing departures for this route (preserving local/modified)
      ptData.departures = ptData.departures.filter(d => 
        d.routeId !== routeId || d.sourceState !== 'persisted'
      );
      
      // Add new departures with persisted sourceState
      const persistedDepartures = departures.map(d => ({ 
        ...d, 
        sourceState: 'persisted' as const 
      }));
      ptData.departures.push(...persistedDepartures);
    } catch (error) {
      throw new Error(`Failed to load departures for route: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

}