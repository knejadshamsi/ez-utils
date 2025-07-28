import { GetPTLinesByMode, GetPTStopsForRoute, GetPTDepartures } from '@wailsjs/go/gui/App';
import { ptState, type Line, type TransportMode } from '$lib/stores/pt.svelte';
import { appState } from '$lib/stores/app.svelte';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(mode: string = 'BUS'): Promise<void> {
    try {
      const backendLines = await GetPTLinesByMode(appState.processId, mode);
      
      // Convert backend lines to frontend Line type
      const lines: Line[] = backendLines.map(line => ({
        id: line.id,
        type: line.type as TransportMode,
        name: line.name,
        routes: line.routes
        // local is undefined for backend lines (not local)
      }));
      
      // Create or update PTData for the current mode
      const ptData = {
        mode: mode as TransportMode,
        lines: lines,
        stops: [],
        departures: []
      };
      
      const existingIndex = ptState.ptData.findIndex(pd => pd.mode === mode);
      if (existingIndex >= 0) {
        // Preserve local lines before overwriting
        const localLines = ptState.ptData[existingIndex].lines.filter(line => line.local === true);
        
        // Set backend lines
        ptState.ptData[existingIndex].lines = ptData.lines;
        
        // Add back local lines
        ptState.ptData[existingIndex].lines.push(...localLines);
      } else {
        ptState.ptData.push(ptData);
      }
      
    } catch (error) {
      throw error;
    }
  }
  
  static async loadRouteStops(routeId: string): Promise<void> {
    try {
      
      const stops = await GetPTStopsForRoute(appState.processId, routeId);
      // Add stops to current mode's PTData
      const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
      if (ptData) {
        // Remove existing stops for this route and add new ones
        ptData.stops = ptData.stops.filter(s => s.routeId !== routeId);
        ptData.stops.push(...(stops as gui.Stop[]));
      }
    } catch (error) {
      throw error;
    }
  }
  
  // Load departures for a specific route
  static async loadRouteDepartures(routeId: string): Promise<void> {
    try {
      
      const departures = await GetPTDepartures(appState.processId, routeId);
      // Add departures to current mode's PTData
      const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
      if (ptData) {
        // Remove existing departures for this route and add new ones
        ptData.departures = ptData.departures.filter(d => d.routeId !== routeId);
        ptData.departures.push(...(departures as gui.Departure[]));
      }
    } catch (error) {
      throw error;
    }
  }

}