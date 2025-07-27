import { GetPTLinesByMode, GetPTStopsForRoute, GetPTDepartures } from '@wailsjs/go/gui/App';
import { ptState } from '$lib/stores/pt.svelte';
import { appState } from '$lib/stores/app.svelte';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(mode: string = 'BUS'): Promise<void> {
    try {
      const lines = await GetPTLinesByMode(appState.processId, mode);
      
      // Create or update PTData for the current mode
      const ptData = {
        mode: mode as any,
        lines: lines as gui.Line[],
        stops: [],
        departures: []
      };
      
      const existingIndex = ptState.ptData.findIndex(pd => pd.mode === mode);
      if (existingIndex >= 0) {
        ptState.ptData[existingIndex] = { ...ptState.ptData[existingIndex], lines: ptData.lines };
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