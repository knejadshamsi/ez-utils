import type { gui } from '../../../../wailsjs/go/models';
import { mapPTLine, mapPTRoute, mapPTStop, mapPTRouteStop, mapPTDeparture } from './modelMapping';
import type { PTLineExtended, PTRouteExtended, PTStopExtended, PTRouteStopExtended, PTDepartureExtended } from './modelMapping';

export enum TransportMode {
  Bus = 'bus',
  Rail = 'rail',
  Tram = 'tram',
  Ferry = 'ferry'
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
}

export interface LineWithRoutes extends PTLineExtended {
  routes: RouteWithTiming[];
}

export interface PTVisibility {
  stops: boolean;
  routes: boolean;
}

class PTState {
  modes = $state<Set<TransportMode>>(new Set([
    TransportMode.Bus,
    TransportMode.Rail,
    TransportMode.Tram
  ]));
  
  lines = $state<Map<string, LineWithRoutes>>(new Map());
  stops = $state<Map<string, PTStopExtended>>(new Map());
  selectedLineId = $state<string | null>(null);
  selectedRouteId = $state<string | null>(null);
  visibleModes = $state<Set<TransportMode>>(new Set([
    TransportMode.Bus,
    TransportMode.Rail,
    TransportMode.Tram
  ]));
  
  visibility = $state<PTVisibility>({
    stops: true,
    routes: true
  });
  
  isAddingStop = $state<boolean>(false);
  updateVersion = $state<number>(0);

  selectedLine = $derived(
    this.selectedLineId ? this.lines.get(this.selectedLineId) : null
  );
  
  selectedRoute = $derived(() => {
    // Force recomputation when updateVersion changes
    const version = this.updateVersion;
    console.log('[ptState] Computing selectedRoute with selectedRouteId:', this.selectedRouteId, 'version:', version);
    if (!this.selectedRouteId) return null;
    for (const line of this.lines.values()) {
      const route = line.routes.find(r => r.id === this.selectedRouteId);
      if (route) {
        console.log('[ptState] Found selected route:', route.id, 'in line:', line.id);
        return { line, route };
      }
    }
    console.log('[ptState] Selected route not found!');
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

  toggleMode(mode: TransportMode) {
    const newSet = new Set(this.visibleModes);
    if (newSet.has(mode)) {
      newSet.delete(mode);
      console.log(`[ptState] Mode ${mode} hidden, visible modes:`, Array.from(newSet));
    } else {
      newSet.add(mode);
      console.log(`[ptState] Mode ${mode} shown, visible modes:`, Array.from(newSet));
    }
    this.visibleModes = newSet;
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
    console.log('[ptState] setSelectedRoute called with:', routeId);
    this.selectedRouteId = routeId;
    // Auto-select the line when route is selected
    if (routeId) {
      for (const line of this.lines.values()) {
        if (line.routes.some(r => r.id === routeId)) {
          this.selectedLineId = line.id;
          console.log('[ptState] Found route in line:', line.id);
          break;
        }
      }
    }
    console.log('[ptState] After setSelectedRoute - selectedRouteId:', this.selectedRouteId);
  }

  toggleVisibility(type: 'stops' | 'routes') {
    this.visibility[type] = !this.visibility[type];
  }

  loadPTData(lines: gui.PTLine[], stops: gui.PTStop[], routes: gui.PTRoute[], routeStops: gui.PTRouteStop[], departures: gui.PTDeparture[]) {
    const newStops = new Map<string, PTStopExtended>();
    for (const stop of stops) {
      const mappedStop = mapPTStop(stop);
      newStops.set(mappedStop.id, mappedStop);
    }
    this.stops = newStops;

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

      route.stopSequence = routeStops.map((rs, index) => ({
        stopId: rs.stopId,
        arrival: rs.arrivalOffset ? this.calculateArrivalTime(route.firstDeparture, rs.arrivalOffset) : route.firstDeparture,
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
    this.lines = newLines;
  }

  calculateArrivalTime(firstDeparture: string, offsetMinutes: number): string {
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
      [TransportMode.Rail]: '🚇',
      [TransportMode.Tram]: '🚊',
      [TransportMode.Ferry]: '⛴️'
    };
    return icons[mode] || '🚌';
  }
}

export const ptState = new PTState();