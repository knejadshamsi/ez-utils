import { GetPTLines, GetPTRoutes, GetPTStops, GetPTRouteStops, GetPTDepartures, GetPTLineSummaries, GetPTLine } from '@wailsjs/go/gui/App';
import { ptState } from '$lib/stores/pt.svelte';
import { editingSession } from '$lib/stores/app.svelte.ts';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  // Load line summaries for a specific mode
  static async loadPTData(mode: string = 'BUS'): Promise<void> {
    console.log('[PTService] Loading PT line summaries for mode:', mode, 'processId:', editingSession.processId);
    try {
      // Only load line summaries for the specified mode
      ptState.loading.summaries = true;
      const summaries = await GetPTLineSummaries(editingSession.processId, mode);
      
      console.log('[PTService] Loaded line summaries for', mode, ':', summaries?.length || 0);
      console.log('[PTService] Raw summaries:', summaries);
      ptState.loadLineSummaries(summaries as gui.PTLineSummary[]);
      ptState.loading.summaries = false;
      
      console.log('[PTService] PT line summaries loaded successfully');
    } catch (error) {
      console.error('[PTService] Failed to load PT line summaries:', error);
      ptState.loading.summaries = false;
      throw error;
    }
  }
  
  // Load full data for a specific line
  static async loadLineData(lineId: string): Promise<void> {
    if (ptState.isLineLoaded(lineId)) {
      console.log('[PTService] Line already loaded:', lineId);
      return;
    }
    
    console.log('[PTService] Loading line data for:', lineId);
    try {
      ptState.loading.lines.set(lineId, true);
      
      // Get line details and its routes
      const [line, routes] = await Promise.all([
        GetPTLine(editingSession.processId, lineId),
        GetPTRoutes(editingSession.processId, lineId)
      ]);
      
      // Get route stops for all routes
      const routeStopPromises = (routes as gui.PTRoute[]).map(route => 
        GetPTRouteStops(editingSession.processId, route.id)
      );
      const routeStopArrays = await Promise.all(routeStopPromises);
      const routeStops = routeStopArrays.flat();
      
      // Load stops if not already loaded
      if (!ptState.stopsLoaded) {
        ptState.loading.stops = true;
        const stops = await GetPTStops(editingSession.processId);
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
      
      console.log('[PTService] Line data loaded successfully:', lineId);
    } catch (error) {
      console.error('[PTService] Failed to load line data:', error);
      ptState.loading.lines.set(lineId, false);
      throw error;
    }
  }
  
  // Load departures for a specific route
  static async loadRouteDepartures(routeId: string): Promise<void> {
    if (ptState.isRouteLoaded(routeId)) {
      console.log('[PTService] Route departures already loaded:', routeId);
      return;
    }
    
    console.log('[PTService] Loading route departures for:', routeId);
    try {
      ptState.loading.routes.set(routeId, true);
      
      const departures = await GetPTDepartures(editingSession.processId, routeId);
      
      // Update existing route with departures
      ptState.updateRouteDepartures(routeId, departures as gui.PTDeparture[]);
      
      ptState.loadedRoutes.add(routeId);
      ptState.loading.routes.set(routeId, false);
      
      console.log('[PTService] Route departures loaded successfully:', routeId);
    } catch (error) {
      console.error('[PTService] Failed to load route departures:', error);
      ptState.loading.routes.set(routeId, false);
      throw error;
    }
  }

}