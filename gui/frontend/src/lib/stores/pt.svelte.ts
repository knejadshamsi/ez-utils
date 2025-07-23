import type { gui } from '@wailsjs/go/models';
import { mapPTLine, mapPTRoute, mapPTStop, mapPTRouteStop, mapPTDeparture } from '$lib/utils/ptModelMapping';
import type { PTLineExtended, PTRouteExtended, PTStopExtended, PTRouteStopExtended, PTDepartureExtended } from '$lib/utils/ptModelMapping';

export enum TransportMode {
  Bus = 'BUS',
  Metro = 'METRO',
  Tram = 'TRAM'
}

export interface StopTime {
  stopId: string;
  arrival: string;
  dwellMinutes: number;
  sequence: number;
}

export interface RouteWithTiming extends PTRouteExtended {
  firstDeparture: string;
  stopSequence: StopTime[];
  departures?: PTDepartureExtended[];
}

export interface LineWithRoutes extends PTLineExtended {
  routes: RouteWithTiming[];
}

export interface PTVisibility {
  stops: boolean;
  routes: boolean;
}

// Loading states for different data levels
export interface PTLoadingState {
  summaries: boolean;
  lines: Map<string, boolean>; // Loading state per line
  routes: Map<string, boolean>; // Loading state per route
  stops: boolean;
}

class PTState {
  modes = $state<Set<TransportMode>>(new Set([
    TransportMode.Bus,
    TransportMode.Metro,
    TransportMode.Tram
  ]));
  
  // Line summaries for lazy loading
  lineSummaries = $state<Map<string, gui.PTLineSummary>>(new Map());
  
  // Full data loaded on demand
  lines = $state<Map<string, LineWithRoutes>>(new Map());
  stops = $state<Map<string, PTStopExtended>>(new Map());
  selectedLineId = $state<string | null>(null);
  selectedRouteId = $state<string | null>(null);
  visibleModes = $state<Set<TransportMode>>(new Set([
    TransportMode.Bus,
    TransportMode.Metro,
    TransportMode.Tram
  ]));
  
  visibility = $state<PTVisibility>({
    stops: true,
    routes: true
  });
  
  // Loading states
  loading = $state<PTLoadingState>({
    summaries: false,
    lines: new Map(),
    routes: new Map(),
    stops: false
  });
  
  // Data loaded flags
  loadedLines = $state<Set<string>>(new Set());
  loadedRoutes = $state<Set<string>>(new Set());
  stopsLoaded = $state<boolean>(false);
  
  isAddingStop = $state<boolean>(false);
  isSelectingStopLocation = $state<boolean>(false);
  selectingStopId = $state<string | null>(null);
  updateVersion = $state<number>(0);
  isDraggingStop = $state<boolean>(false);
  isAddingMultipleStops = $state<boolean>(false);
  
  // Selected vehicle type for filtering
  selectedMode = $state<TransportMode>(TransportMode.Bus);

  selectedLine = $derived(
    this.selectedLineId ? this.lines.get(this.selectedLineId) : null
  );
  
  selectedRoute = $derived(() => {
    // Force recomputation when updateVersion changes
    const version = this.updateVersion;
    if (!this.selectedRouteId) return null;
    for (const line of this.lines.values()) {
      const route = line.routes.find(r => r.id === this.selectedRouteId);
      if (route) {
        return { line, route };
      }
    }
    return null;
  });

  visibleLines = $derived(
    Array.from(this.lines.values()).filter(line => 
      this.visibleModes.has(line.mode as TransportMode)
    )
  );

  linesByMode = $derived(() => {
    const grouped = new Map<TransportMode, LineWithRoutes[]>();
    
    for (const mode of this.visibleModes) {
      grouped.set(mode, []);
    }
    
    for (const line of this.visibleLines) {
      const mode = line.mode as TransportMode;
      if (grouped.has(mode)) {
        grouped.get(mode)!.push(line);
      }
    }
    
    return grouped;
  });
  
  // Get line summaries grouped by mode for lazy loading
  lineSummariesByMode = $derived(() => {
    const grouped = new Map<TransportMode, gui.PTLineSummary[]>();
    
    for (const mode of this.visibleModes) {
      grouped.set(mode, []);
    }
    
    for (const summary of this.lineSummaries.values()) {
      const mode = summary.mode as TransportMode;
      if (this.visibleModes.has(mode) && grouped.has(mode)) {
        grouped.get(mode)!.push(summary);
      }
    }
    
    return grouped;
  });
  
  // Get line summaries for the selected mode only
  selectedModeSummaries = $derived(() => {
    const summaries: gui.PTLineSummary[] = [];
    for (const summary of this.lineSummaries.values()) {
      if (summary.mode === this.selectedMode) {
        summaries.push(summary);
      }
    }
    return summaries;
  });

  toggleMode(mode: TransportMode) {
    const newSet = new Set(this.visibleModes);
    if (newSet.has(mode)) {
      newSet.delete(mode);
    } else {
      newSet.add(mode);
    }
    this.visibleModes = newSet;
  }

  setSelectedMode(mode: TransportMode) {
    this.selectedMode = mode;
    // Clear selected line and route when mode changes
    this.selectedLineId = null;
    this.selectedRouteId = null;
    
    // Clear all line data for the previous mode
    this.lineSummaries = new Map();
    this.lines = new Map();
    this.loadedLines = new Set();
    
    // Keep stops as they can be shared across modes
    // but clear loading states
    this.loading.lines = new Map();
    this.loading.routes = new Map();
  }

  setSelectedLine(lineId: string | null) {
    this.selectedLineId = lineId;
    // Clear route selection when line changes
    if (this.selectedRouteId) {
      const selectedRouteData = this.selectedRoute;
      if (!selectedRouteData || selectedRouteData.line.id !== lineId) {
        this.selectedRouteId = null;
      }
    }
  }
  
  setSelectedRoute(routeId: string | null) {
    this.selectedRouteId = routeId;
    // Auto-select the line when route is selected
    if (routeId) {
      for (const line of this.lines.values()) {
        if (line.routes.some(r => r.id === routeId)) {
          this.selectedLineId = line.id;
          break;
        }
      }
    }
  }

  toggleVisibility(type: 'stops' | 'routes') {
    this.visibility[type] = !this.visibility[type];
  }

  // Load only line summaries initially
  loadLineSummaries(summaries: gui.PTLineSummary[]) {
    
    // Always create a new map, even if empty
    const newSummaries = new Map<string, gui.PTLineSummary>();
    
    if (summaries && summaries.length > 0) {
      for (const summary of summaries) {
        newSummaries.set(summary.id, summary);
      }
    }
    
    this.lineSummaries = newSummaries;
  }
  
  // Check if a line's full data is loaded
  isLineLoaded(lineId: string): boolean {
    return this.loadedLines.has(lineId);
  }
  
  // Check if a route's departures are loaded
  isRouteLoaded(routeId: string): boolean {
    return this.loadedRoutes.has(routeId);
  }
  
  loadPTData(lines: gui.PTLine[], stops: gui.PTStop[], routes: gui.PTRoute[], routeStops: gui.PTRouteStop[], departures: gui.PTDeparture[], merge: boolean = false) {
    // Handle stops - merge or replace
    if (stops.length > 0) {
      if (merge) {
        // Merge new stops into existing
        for (const stop of stops) {
          const mappedStop = mapPTStop(stop);
          this.stops.set(mappedStop.id, mappedStop);
        }
      } else {
        // Replace all stops
        const newStops = new Map<string, PTStopExtended>();
        for (const stop of stops) {
          const mappedStop = mapPTStop(stop);
          newStops.set(mappedStop.id, mappedStop);
        }
        this.stops = newStops;
      }
    }

    const routeMap = new Map<string, RouteWithTiming>();
    const routeStopsByRoute = new Map<string, PTRouteStopExtended[]>();
    const departuresByRoute = new Map<string, PTDepartureExtended[]>();

    for (const route of routes) {
      const mappedRoute = mapPTRoute(route);
      routeMap.set(mappedRoute.id, {
        ...mappedRoute,
        firstDeparture: '',
        stopSequence: []
      });
    }

    for (const routeStop of routeStops) {
      const mappedRouteStop = mapPTRouteStop(routeStop);
      if (!routeStopsByRoute.has(mappedRouteStop.routeId)) {
        routeStopsByRoute.set(mappedRouteStop.routeId, []);
      }
      routeStopsByRoute.get(mappedRouteStop.routeId)!.push(mappedRouteStop);
    }

    for (const departure of departures) {
      const mappedDeparture = mapPTDeparture(departure);
      if (!departuresByRoute.has(mappedDeparture.routeId)) {
        departuresByRoute.set(mappedDeparture.routeId, []);
      }
      departuresByRoute.get(mappedDeparture.routeId)!.push(mappedDeparture);
    }

    for (const [routeId, route] of routeMap) {
      const routeStops = routeStopsByRoute.get(routeId) || [];
      const departures = departuresByRoute.get(routeId) || [];
      
      routeStops.sort((a, b) => a.sequence - b.sequence);
      
      if (departures.length > 0) {
        departures.sort((a, b) => a.departureTime.localeCompare(b.departureTime));
        route.firstDeparture = departures[0].departureTime.substring(11, 16);
      }
      
      // Store departures in the route
      route.departures = departures;

      route.stopSequence = routeStops.map((rs, index) => ({
        stopId: rs.stopId,
        arrival: rs.arrivalOffset !== undefined ? this.calculateArrivalTime(route.firstDeparture, rs.arrivalOffset) : route.firstDeparture || '00:00',
        dwellMinutes: rs.dwellTime || 0,
        sequence: rs.sequence
      }));
    }

    const newLines = new Map<string, LineWithRoutes>();
    for (const line of lines) {
      const mappedLine = mapPTLine(line);
      const lineRoutes = routes
        .filter(r => r.line_id === line.id)
        .map(r => {
          const mappedRouteId = mapPTRoute(r).id;
          return routeMap.get(mappedRouteId);
        })
        .filter(r => r !== undefined) as RouteWithTiming[];
      
      newLines.set(mappedLine.id, {
        ...mappedLine,
        routes: lineRoutes
      });
    }
    // Merge or replace lines
    if (merge) {
      // Merge new lines into existing
      for (const [lineId, lineData] of newLines) {
        this.lines.set(lineId, lineData);
      }
    } else {
      // Replace all lines
      this.lines = newLines;
    }
  }

  calculateArrivalTime(firstDeparture: string, offsetMinutes: number): string {
    if (!firstDeparture || !firstDeparture.includes(':')) {
      // If no first departure, assume 00:00 as base
      const totalMinutes = offsetMinutes;
      const newHours = Math.floor(totalMinutes / 60) % 24;
      const newMinutes = totalMinutes % 60;
      return `${newHours.toString().padStart(2, '0')}:${newMinutes.toString().padStart(2, '0')}`;
    }
    const [hours, minutes] = firstDeparture.split(':').map(Number);
    const totalMinutes = hours * 60 + minutes + offsetMinutes;
    const newHours = Math.floor(totalMinutes / 60) % 24;
    const newMinutes = totalMinutes % 60;
    return `${newHours.toString().padStart(2, '0')}:${newMinutes.toString().padStart(2, '0')}`;
  }

  recalculateSubsequentTimes(lineId: string, routeIndex: number, fromStopIndex: number) {
    const line = this.lines.get(lineId);
    if (!line || !line.routes[routeIndex]) return;

    const route = line.routes[routeIndex];
    const stops = route.stopSequence;

    for (let i = fromStopIndex + 1; i < stops.length; i++) {
      const prevStop = stops[i - 1];
      const currentStop = stops[i];
      
      const prevDepartureMinutes = this.timeToMinutes(prevStop.arrival) + prevStop.dwellMinutes;
      const travelTime = this.timeToMinutes(currentStop.arrival) - prevDepartureMinutes;
      
      if (travelTime < 0) {
        const newArrivalMinutes = prevDepartureMinutes + 5;
        currentStop.arrival = this.minutesToTime(newArrivalMinutes);
      }
    }
  }

  timeToMinutes(time: string): number {
    const [hours, minutes] = time.split(':').map(Number);
    return hours * 60 + minutes;
  }

  minutesToTime(minutes: number): string {
    const hours = Math.floor(minutes / 60) % 24;
    const mins = minutes % 60;
    return `${hours.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}`;
  }

  getTotalTravelTime(route: RouteWithTiming): number {
    if (route.stopSequence.length < 2) return 0;
    
    const firstStop = route.stopSequence[0];
    const lastStop = route.stopSequence[route.stopSequence.length - 1];
    
    return this.timeToMinutes(lastStop.arrival) - this.timeToMinutes(firstStop.arrival);
  }


  getModeIcon(mode: TransportMode): string {
    const icons = {
      [TransportMode.Bus]: '🚌',
      [TransportMode.Metro]: '🚇',
      [TransportMode.Tram]: '🚊'
    };
    return icons[mode] || '🚌';
  }

  getModeAbbreviation(mode: TransportMode): string {
    const abbreviations = {
      [TransportMode.Bus]: 'B',
      [TransportMode.Metro]: 'M',
      [TransportMode.Tram]: 'T'
    };
    return abbreviations[mode] || 'B';
  }

  getStopName(stopId: string): string {
    const stop = this.stops.get(stopId);
    return stop?.name || `Stop ${stopId}`;
  }

  getStopCoords(stopId: string): string {
    const stop = this.stops.get(stopId);
    if (stop?.lng !== undefined && stop?.lat !== undefined) {
      return `${stop.lng.toFixed(4)}, ${stop.lat.toFixed(4)}`;
    }
    return 'No coordinates';
  }

  getLinesForStop(stopId: string): LineWithRoutes[] {
    const linesSet = new Set<string>();
    const result: LineWithRoutes[] = [];
    
    // Check all lines and their routes
    for (const line of this.lines.values()) {
      for (const route of line.routes) {
        // Check if this route contains the stop
        if (route.stopSequence.some(stop => stop.stopId === stopId)) {
          if (!linesSet.has(line.id)) {
            linesSet.add(line.id);
            result.push(line);
          }
        }
      }
    }
    
    return result;
  }

  // Add a single departure to a route
  addDeparture(lineId: string, routeId: string, departure: PTDepartureExtended): void {
    const line = this.lines.get(lineId);
    if (!line) return;
    
    const routeIndex = line.routes.findIndex(r => r.id === routeId);
    if (routeIndex === -1) return;
    
    const route = line.routes[routeIndex];
    
    // Initialize departures array if not exists
    const currentDepartures = route.departures || [];
    
    // Add and sort departures - create new array for reactivity
    const newDepartures = [...currentDepartures, departure].sort((a, b) => a.departureTime.localeCompare(b.departureTime));
    
    // Create updated route with new departures
    const updatedRoute = {
      ...route,
      departures: newDepartures
    };
    
    // Update first departure if this is the earliest
    if (newDepartures.length === 1 || departure.departureTime < updatedRoute.firstDeparture) {
      updatedRoute.firstDeparture = departure.departureTime.substring(0, 5);
      
      // Update stop times if they were previously calculated
      if (updatedRoute.stopSequence.length > 0) {
        updatedRoute.stopSequence = [
          { ...updatedRoute.stopSequence[0], arrival: updatedRoute.firstDeparture },
          ...updatedRoute.stopSequence.slice(1)
        ];
      }
    }
    
    // Create new line with updated route
    const updatedLine = {
      ...line,
      routes: line.routes.map((r, i) => i === routeIndex ? updatedRoute : r)
    };
    
    // Trigger reactivity
    this.lines.set(lineId, updatedLine);
    this.updateVersion++;
  }

  // Update departures for a specific route
  updateRouteDepartures(routeId: string, departures: gui.PTDeparture[]): void {
    // Find the line containing this route
    for (const [lineId, line] of this.lines) {
      const routeIndex = line.routes.findIndex(r => r.id === routeId);
      if (routeIndex !== -1) {
        const route = line.routes[routeIndex];
        
        // Map and sort departures
        const mappedDepartures = departures.map(d => mapPTDeparture(d));
        mappedDepartures.sort((a, b) => a.departureTime.localeCompare(b.departureTime));
        
        // Update first departure time
        if (mappedDepartures.length > 0) {
          route.firstDeparture = mappedDepartures[0].departureTime.substring(11, 16);
          
          // Update stop times if they were previously calculated from an offset
          // We don't recalculate them here as we don't have the original offset data
          // Just update the first stop's arrival to match the first departure
          if (route.stopSequence.length > 0) {
            route.stopSequence[0].arrival = route.firstDeparture;
          }
        }
        
        // Trigger reactivity by updating the line
        this.lines.set(lineId, { ...line });
        this.updateVersion++;
        break;
      }
    }
  }
  
  deleteRoute(lineId: string, routeId: string): void {
    const line = this.lines.get(lineId);
    if (!line) return;
    
    // Remove the route from the line's routes array
    line.routes = line.routes.filter(r => r.id !== routeId);
    
    // If this was the selected route, clear selection
    if (this.selectedRouteId === routeId) {
      this.selectedRouteId = null;
      this.selectedLineId = null;
    }
    
    // Trigger reactivity by updating the Map
    this.lines.set(lineId, { ...line });
  }
}

export const ptState = new PTState();