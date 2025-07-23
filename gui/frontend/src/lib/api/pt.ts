import { GetPTLines, GetPTRoutes, GetPTStops, GetPTRouteStops, GetPTDepartures, GetPTLineSummaries, GetPTLine } from '@wailsjs/go/gui/App';
import { ptState } from '$lib/stores/pt.svelte';
import { appState } from '$lib/stores/app.svelte';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(mode: string = 'BUS'): Promise<void> {
    try {
      ptState.loading.summaries = true;
      const summaries = await GetPTLineSummaries(appState.processId, mode);
      
      ptState.loadLineSummaries(summaries as gui.PTLineSummary[]);
      ptState.loading.summaries = false;
      
    } catch (error) {
      ptState.loading.summaries = false;
      throw error;
    }
  }
  
  static async loadLineData(lineId: string): Promise<void> {
    if (ptState.isLineLoaded(lineId)) {
      return;
    }
    
    try {
      ptState.loading.lines.set(lineId, true);
      
      const [line, routes] = await Promise.all([
        GetPTLine(appState.processId, lineId),
        GetPTRoutes(appState.processId, lineId)
      ]);
      
      const routeStopPromises = (routes as gui.PTRoute[]).map(route => 
        GetPTRouteStops(appState.processId, route.id)
      );
      const routeStopArrays = await Promise.all(routeStopPromises);
      const routeStops = routeStopArrays.flat();
      
      // Load stops if not already loaded
      if (!ptState.stopsLoaded) {
        ptState.loading.stops = true;
        const stops = await GetPTStops(appState.processId);
        ptState.loadPTData(
          [line] as gui.PTLine[],
          stops as gui.PTStop[],
          routes as gui.PTRoute[],
          routeStops as gui.PTRouteStop[],
          [], // No departures yet
          false // Don't merge, replace all data
        );
        ptState.stopsLoaded = true;
        ptState.loading.stops = false;
      } else {
        // Just add the line data
        ptState.loadPTData(
          [line] as gui.PTLine[],
          [], // Stops already loaded
          routes as gui.PTRoute[],
          routeStops as gui.PTRouteStop[],
          [], // No departures yet
          true // Merge with existing data
        );
      }
      
      ptState.loadedLines.add(lineId);
      ptState.loading.lines.set(lineId, false);
      
    } catch (error) {
      ptState.loading.lines.set(lineId, false);
      throw error;
    }
  }
  
  // Load departures for a specific route
  static async loadRouteDepartures(routeId: string): Promise<void> {
    if (ptState.isRouteLoaded(routeId)) {
      return;
    }
    
    try {
      ptState.loading.routes.set(routeId, true);
      
      const departures = await GetPTDepartures(appState.processId, routeId);
      
      // Update existing route with departures
      ptState.updateRouteDepartures(routeId, departures as gui.PTDeparture[]);
      
      ptState.loadedRoutes.add(routeId);
      ptState.loading.routes.set(routeId, false);
      
    } catch (error) {
      ptState.loading.routes.set(routeId, false);
      throw error;
    }
  }

}