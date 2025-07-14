import { GetPTLines, GetPTRoutes, GetPTStops, GetPTRouteStops, GetPTDepartures } from '@wailsjs/go/gui/App';
import { ptState } from '$lib/stores/pt.svelte';
import { editingSession } from '$lib/stores/app.svelte.ts';
import type { gui } from '@wailsjs/go/models';

export class PTService {
  static async loadPTData(): Promise<void> {
    console.log('[PTService] Loading PT data for processId:', editingSession.processId);
    try {
      // Get all lines and stops first
      const [lines, stops] = await Promise.all([
        GetPTLines(editingSession.processId),
        GetPTStops(editingSession.processId)
      ]);
      
      console.log('[PTService] Loaded lines:', lines?.length || 0, 'stops:', stops?.length || 0);

      // Get routes for all lines
      const routePromises = (lines as gui.PTLine[]).map(line => 
        GetPTRoutes(editingSession.processId, line.id)
      );
      const routeArrays = await Promise.all(routePromises);
      const routes = routeArrays.flat();

      // Get route stops and departures for all routes
      const routeStopPromises = (routes as gui.PTRoute[]).map(route => 
        GetPTRouteStops(editingSession.processId, route.id)
      );
      const departurePromises = (routes as gui.PTRoute[]).map(route => 
        GetPTDepartures(editingSession.processId, route.id)
      );
      
      const [routeStopArrays, departureArrays] = await Promise.all([
        Promise.all(routeStopPromises),
        Promise.all(departurePromises)
      ]);
      
      const routeStops = routeStopArrays.flat();
      const departures = departureArrays.flat();

      ptState.loadPTData(
        lines as gui.PTLine[],
        stops as gui.PTStop[],
        routes as gui.PTRoute[],
        routeStops as gui.PTRouteStop[],
        departures as gui.PTDeparture[]
      );
      
      console.log('[PTService] PT data loaded successfully');
      console.log('[PTService] ptState lines:', ptState.lines.size, 'stops:', ptState.stops.size);
    } catch (error) {
      console.error('[PTService] Failed to load PT data:', error);
      throw error;
    }
  }

}