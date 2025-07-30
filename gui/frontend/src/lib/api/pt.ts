import { GetPTLinesByMode, GetPTRoutesByLineID, GetPTStopsByRouteID, GetPTDeparturesByRouteID, GetPTStopsInBounds, GetPTStops, InsertPTLines, InsertPTRoutes, UpdatePTLines, UpdatePTRoutes, DeletePTLines, DeletePTRoutes, InsertPTStops, InsertPTRouteStops, UpdatePTStops, UpdatePTRouteStops, DeletePTStops, DeletePTRouteStops, InsertPTDepartures, UpdatePTDepartures, DeletePTDepartures, GetMaxSequenceForRoute, InsertSingleRouteStop, DeleteSingleRouteStop } from '@wailsjs/go/gui/App';
import { clearRouteChanges, clearAllChanges } from '$lib/utils/ptChangeTracking';
import { ptState, type TransportMode } from '$lib/stores/pt.svelte';
import { appState } from '$lib/stores/app.svelte';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(mode: TransportMode): Promise<void> {
    try {
      // Load lines and routes together
      const lines = await GetPTLinesByMode(appState.processId, mode);
      
      // Build the data structure
      const data: Array<{ id: string; name: string; type: 'LINE' | 'ROUTE'; lineId?: string }> = [];
      
      // Add all lines
      for (const line of lines) {
        data.push({
          id: line.id,
          name: line.name,
          type: 'LINE'
        });
        
        // Load routes for each line
        try {
          const routes = await GetPTRoutesByLineID(appState.processId, line.id);
          for (const route of routes) {
            data.push({
              id: route.id,
              name: route.name,
              type: 'ROUTE',
              lineId: line.id
            });
          }
        } catch (error) {
          // Continue even if one line's routes fail to load
          console.error(`Failed to load routes for line ${line.id}:`, error);
        }
      }
      
      // Load into state
      ptState.loadLinesAndRoutes(mode, data);
      
    } catch (error) {
      throw new Error(`Failed to load ${mode} data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async loadRouteStops(routeId: string): Promise<void> {
    try {
      // 1. Always fetch fresh sequence from route_stops table
      const routeStops = await GetPTStopsByRouteID(appState.processId, routeId);
      
      // 2. Extract sequence and unique stop IDs
      const sequence = routeStops
        .sort((a, b) => a.sequence - b.sequence)
        .map(rs => rs.stopId);
      
      const uniqueStopIds = [...new Set(sequence)];
      
      if (uniqueStopIds.length === 0) {
        ptState.loadRouteData(routeId, [], [], Object.values(ptState.currentRouteData.departures));
        return;
      }
      
      // 3. Check what stops we already have locally
      const localStops: gui.Stop[] = [];
      const missingStopIds: string[] = [];
      
      uniqueStopIds.forEach(stopId => {
        const localStop = ptState.currentRouteData.stops[stopId];
        if (localStop && localStop.unsaved) {
          // Keep unsaved local version
          localStops.push(localStop as any);
        } else {
          missingStopIds.push(stopId);
        }
      });
      
      // 4. Fetch only missing stops from database
      let fetchedStops: gui.Stop[] = [];
      if (missingStopIds.length > 0) {
        fetchedStops = await GetPTStops(appState.processId, missingStopIds);
      }
      
      // 5. Combine local and fetched stops
      const allStops = [...localStops, ...fetchedStops];
      
      // Load into current route data with sequence
      ptState.loadRouteData(routeId, sequence, allStops, Object.values(ptState.currentRouteData.departures));
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
      
      // Load into current route data
      ptState.loadRouteData(routeId, ptState.currentRouteData.sequence, Object.values(ptState.currentRouteData.stops), departures);
    } catch (error) {
      throw new Error(`Failed to load departures for route: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async saveCurrentRoute(): Promise<void> {
    if (!ptState.currentRouteData.routeId) return;
    
    try {
      const unsaved = ptState.getUnsavedItems();
      
      // Save unsaved stops
      if (unsaved.stops.length > 0) {
        const newStops = unsaved.stops.filter(s => s.stopId.startsWith('stop_'));
        const existingStops = unsaved.stops.filter(s => !s.stopId.startsWith('stop_'));
        
        // Insert new stops
        if (newStops.length > 0) {
          await InsertPTStops(appState.processId, newStops);
        }
        
        // Update existing stops
        if (existingStops.length > 0) {
          await UpdatePTStops(appState.processId, existingStops);
        }
        
        // Mark as saved
        ptState.markStopsAsSaved(unsaved.stops.map(s => s.stopId));
      }
      
      // Save unsaved departures
      if (unsaved.departures.length > 0) {
        const newDepartures = unsaved.departures.filter(d => d.id.startsWith('dep_'));
        const existingDepartures = unsaved.departures.filter(d => !d.id.startsWith('dep_'));
        
        // Insert new departures
        if (newDepartures.length > 0) {
          await InsertPTDepartures(appState.processId, newDepartures);
        }
        
        // Update existing departures
        if (existingDepartures.length > 0) {
          await UpdatePTDepartures(appState.processId, existingDepartures);
        }
        
        // Mark as saved
        ptState.markDeparturesAsSaved(unsaved.departures.map(d => d.id));
      }
      
    } catch (error) {
      throw new Error(`Failed to save route data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async saveUnsavedItems(): Promise<void> {
    try {
      const unsaved = ptState.getUnsavedItems();
      
      // Save unsaved lines
      if (unsaved.lines.length > 0) {
        const linesToInsert = unsaved.lines.map(l => ({
          id: l.id,
          name: l.name,
          type: l.mode
        }));
        await InsertPTLines(appState.processId, linesToInsert);
        
        // Mark as saved
        unsaved.lines.forEach(l => ptState.markLineAsSaved(l.id));
      }
      
      // Save unsaved routes
      if (unsaved.routes.length > 0) {
        const routesToInsert = unsaved.routes.map(r => ({
          id: r.id,
          name: r.name,
          lineId: r.lineId
        }));
        await InsertPTRoutes(appState.processId, routesToInsert);
        
        // Mark as saved
        unsaved.routes.forEach(r => ptState.markRouteAsSaved(r.lineId, r.id));
      }
      
    } catch (error) {
      throw new Error(`Failed to save unsaved items: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  // Direct route-stop operations (no local state)
  static async addStopToRoute(routeId: string, stopId: string): Promise<void> {
    try {
      // Get next sequence number
      const maxSequence = await GetMaxSequenceForRoute(appState.processId, routeId);
      const nextSequence = (maxSequence || 0) + 1;
      
      // Insert directly to route_stops table
      await InsertSingleRouteStop(appState.processId, {
        linkId: `rs_${routeId}_${stopId}`,
        routeId,
        stopId,
        sequence: nextSequence
      });
      
      // Reload route data to get fresh sequence
      await this.loadRouteStops(routeId);
    } catch (error) {
      throw new Error(`Failed to add stop to route: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  static async removeStopFromRoute(routeId: string, stopId: string): Promise<void> {
    try {
      // Delete directly from route_stops table
      await DeleteSingleRouteStop(appState.processId, routeId, stopId);
      
      // Reload route data to get fresh sequence
      await this.loadRouteStops(routeId);
    } catch (error) {
      throw new Error(`Failed to remove stop from route: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

}