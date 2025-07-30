import { trackLineChange, trackRouteChange, trackStopChange, trackDepartureChange } from '$lib/utils/ptChangeTracking';
import { nanoid } from 'nanoid';

export type TransportMode = 'BUS' | 'METRO' | 'TRAM';

export type SourceState = 'persisted' | 'local' | 'modified';

export type Line = {
  id: string;
  type: TransportMode;
  name: string;
  sourceState: SourceState;
};

export type Route = {
  id: string;
  lineId: string;
  name: string;
  sourceState: SourceState;
};

export type StopType = 'REGULAR' | 'REQUEST' | 'BOARDING_ONLY' | 'ALIGHTING_ONLY';
export type AccessibilityStatus = 'YES' | 'NO' | 'UNKNOWN';

export type Stop = {
  routeId: string;
  stopId: string;
  arrivalOffset: string;
  departureOffset: string;
  stopName: string;
  lat: number;
  lng: number;
  sequence: number;
  stopType?: StopType;
  wheelchairAccessible?: AccessibilityStatus;
  timingPoint?: boolean;
  attributes?: Record<string, string | number | boolean>;
  sourceState: SourceState;
};

export type Departure = {
  id: string;
  routeId: string;
  departureTime: string;
  vehicleRefId?: string;
  sourceState: SourceState;
};

export type PTData = {
  mode: TransportMode;
  lines: Line[];
  routes: Route[];
  stops: Stop[];
  departures: Departure[];
};

export const TRANSPORT_MODES = {
  BUS: { value: 'BUS', label: 'Bus', icon: '🚌', abbreviation: 'B' },
  METRO: { value: 'METRO', label: 'Metro', icon: '🚇', abbreviation: 'M' },
  TRAM: { value: 'TRAM', label: 'Tram', icon: '🚊', abbreviation: 'T' }
}

export type PTEditMode = 'NORMAL' | 'ADDING_STOP' | 'DRAGGING_STOP' | 'EDITING_STOP_ATTRIBUTES' | 'EDITING_DEPARTURES' | 'ADDING_MULTIPLE_STOPS' | 'SELECTING_STOP_LOCATION';

class PTState {
  // Single data source with sourceState
  ptData = $state<PTData[]>([
    { mode: 'BUS', lines: [], routes: [], stops: [], departures: [] },
    { mode: 'METRO', lines: [], routes: [], stops: [], departures: [] },
    { mode: 'TRAM', lines: [], routes: [], stops: [], departures: [] }
  ]);
  
  selected = $state<{
    mode: TransportMode;
    lineId: string | null;
    routeId: string | null;
    stopId: string | null;
  }>({
    mode: 'BUS',
    lineId: null,
    routeId: null,
    stopId: null
  });
  
  visibility = $state<{ stops: boolean; routes: boolean }>({
    stops: true,
    routes: true
  });
  
  editMode = $state<PTEditMode>('NORMAL');
  selectingStopId = $state<string | null>(null);

  // Helper to get current mode data
  getCurrentModeData() {
    return this.ptData.find(pd => pd.mode === this.selected.mode)!;
  }

  // LINE CRUD
  createLine(name: string): void {
    const ptData = this.getCurrentModeData();
    const newLine: Line = {
      id: `line_${nanoid(10)}`,
      name,
      type: this.selected.mode,
      sourceState: 'local'
    };
    ptData.lines.push(newLine);
    
    // Track change as save
    trackLineChange(newLine, 'save');
  }

  updateLine(lineId: string, updates: {name?: string}): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.lines.findIndex(l => l.id === lineId);
    
    if (index !== -1) {
      const line = ptData.lines[index];
      Object.assign(line, updates);
      
      // Update sourceState if it was persisted
      if (line.sourceState === 'persisted') {
        line.sourceState = 'modified';
      }
      
      trackLineChange(line, 'save');
    }
  }

  deleteLine(lineId: string): void {
    // Clear selection if this line was selected
    if (this.selected.lineId === lineId) {
      this.selected.lineId = null;
      this.selected.routeId = null;
    }

    const ptData = this.getCurrentModeData();
    const index = ptData.lines.findIndex(l => l.id === lineId);
    
    if (index !== -1) {
      const line = ptData.lines[index];
      
      if (line.sourceState === 'local') {
        // Remove from data entirely
        ptData.lines.splice(index, 1);
        
        // Also delete local routes, stops, departures for this line
        const routeIds = ptData.routes.filter(r => r.lineId === lineId).map(r => r.id);
        ptData.routes = ptData.routes.filter(r => r.lineId !== lineId);
        ptData.stops = ptData.stops.filter(s => !routeIds.includes(s.routeId));
        ptData.departures = ptData.departures.filter(d => !routeIds.includes(d.routeId));
      } else {
        // For persisted/modified, track delete
        ptData.lines.splice(index, 1);
        trackLineChange(line, 'delete');
      }
    }
  }

  // ROUTE CRUD
  createRoute(lineId: string, name: string): void {
    const ptData = this.getCurrentModeData();
    const newRoute: Route = {
      id: `route_${nanoid(10)}`,
      lineId,
      name,
      sourceState: 'local'
    };
    ptData.routes.push(newRoute);
    
    // Track change as save
    trackRouteChange(newRoute, 'save');
  }

  updateRoute(routeId: string, updates: {name?: string}): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.routes.findIndex(r => r.id === routeId);
    
    if (index !== -1) {
      const route = ptData.routes[index];
      Object.assign(route, updates);
      
      // Update sourceState if it was persisted
      if (route.sourceState === 'persisted') {
        route.sourceState = 'modified';
      }
      
      trackRouteChange(route, 'save');
    }
  }

  deleteRoute(routeId: string): void {
    // Clear selection if this route was selected
    if (this.selected.routeId === routeId) {
      this.selected.routeId = null;
    }

    const ptData = this.getCurrentModeData();
    const index = ptData.routes.findIndex(r => r.id === routeId);
    
    if (index !== -1) {
      const route = ptData.routes[index];
      
      if (route.sourceState === 'local') {
        // Remove from data entirely
        ptData.routes.splice(index, 1);
        
        // Also delete local stops and departures for this route
        ptData.stops = ptData.stops.filter(s => s.routeId !== routeId);
        ptData.departures = ptData.departures.filter(d => d.routeId !== routeId);
      } else {
        // For persisted/modified, track delete
        ptData.routes.splice(index, 1);
        trackRouteChange(route, 'delete');
      }
    }
  }

  // STOP CRUD
  createStop(stopData: Omit<Stop, 'sequence' | 'sourceState'>): void {
    const ptData = this.getCurrentModeData();
    
    // Calculate next sequence number for this route
    const routeStops = ptData.stops.filter(s => s.routeId === stopData.routeId);
    const maxSequence = routeStops.reduce((max, s) => Math.max(max, s.sequence), 0);
    
    const newStop: Stop = {
      ...stopData,
      sequence: maxSequence + 1,
      stopType: stopData.stopType || 'REGULAR',
      wheelchairAccessible: stopData.wheelchairAccessible || 'UNKNOWN',
      timingPoint: stopData.timingPoint || false,
      sourceState: 'local'
    };
    
    ptData.stops.push(newStop);
    
    // Track change
    trackStopChange(newStop, 'save');
  }

  updateStop(stopId: string, updates: Partial<Stop>): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.stops.findIndex(s => s.stopId === stopId);
    
    if (index !== -1) {
      const stop = ptData.stops[index];
      Object.assign(stop, updates);
      
      // Update sourceState if it was persisted
      if (stop.sourceState === 'persisted') {
        stop.sourceState = 'modified';
      }
      
      trackStopChange(stop, 'save');
    }
  }

  deleteStop(stopId: string): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.stops.findIndex(s => s.stopId === stopId);
    
    if (index !== -1) {
      const stop = ptData.stops[index];
      const routeId = stop.routeId;
      
      if (stop.sourceState === 'local') {
        // Remove from data entirely
        ptData.stops.splice(index, 1);
      } else {
        // For persisted/modified, track delete
        ptData.stops.splice(index, 1);
        trackStopChange(stop, 'delete');
      }
      
      // Resequence remaining stops for this route
      const remainingStops = ptData.stops
        .filter(s => s.routeId === routeId)
        .sort((a, b) => a.sequence - b.sequence);
      
      remainingStops.forEach((s, idx) => {
        s.sequence = idx + 1;
      });
    }
  }


  // DEPARTURE CRUD
  createDeparture(routeId: string, time: string, vehicleId?: string): void {
    const ptData = this.getCurrentModeData();
    const newDeparture: Departure = {
      id: `dep_${nanoid(10)}`,
      routeId,
      departureTime: time,
      vehicleRefId: vehicleId,
      sourceState: 'local'
    };
    
    ptData.departures.push(newDeparture);
    
    // Track change
    trackDepartureChange(newDeparture, 'save');
  }

  updateDeparture(departureId: string, updates: Partial<Departure>): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.departures.findIndex(d => d.id === departureId);
    
    if (index !== -1) {
      const departure = ptData.departures[index];
      Object.assign(departure, updates);
      
      // Update sourceState if it was persisted
      if (departure.sourceState === 'persisted') {
        departure.sourceState = 'modified';
      }
      
      trackDepartureChange(departure, 'save');
    }
  }

  deleteDeparture(departureId: string): void {
    const ptData = this.getCurrentModeData();
    const index = ptData.departures.findIndex(d => d.id === departureId);
    
    if (index !== -1) {
      const departure = ptData.departures[index];
      
      if (departure.sourceState === 'local') {
        // Remove from data entirely
        ptData.departures.splice(index, 1);
      } else {
        // For persisted/modified, track delete
        ptData.departures.splice(index, 1);
        trackDepartureChange(departure, 'delete');
      }
    }
  }

  // BATCH OPERATIONS
  batchCreateStops(routeId: string, stops: Omit<Stop, 'sequence' | 'sourceState'>[]): void {
    const ptData = this.getCurrentModeData();
    
    // Get current max sequence for this route
    const routeStops = ptData.stops.filter(s => s.routeId === routeId);
    let sequence = routeStops.reduce((max, s) => Math.max(max, s.sequence), 0);
    
    const newStops = stops.map(stopData => {
      sequence++;
      const newStop: Stop = {
        ...stopData,
        routeId,
        sequence,
        sourceState: 'local'
      };
      // Track each stop
      trackStopChange(newStop, 'save');
      return newStop;
    });
    
    ptData.stops.push(...newStops);
  }

  // MODE SWITCHING
  switchMode(newMode: TransportMode): void {
    this.selected.mode = newMode;
    this.selected.lineId = null;
    this.selected.routeId = null;
    this.selected.stopId = null;
  }
  
  clearPersistedData(): void {
    // Clear only persisted data for the current mode
    const ptData = this.getCurrentModeData();
    // Remove all persisted items
    ptData.lines = ptData.lines.filter(l => l.sourceState !== 'persisted');
    ptData.routes = ptData.routes.filter(r => r.sourceState !== 'persisted');
    ptData.stops = ptData.stops.filter(s => s.sourceState !== 'persisted');
    ptData.departures = ptData.departures.filter(d => d.sourceState !== 'persisted');
  }
  
  // SAVE AND LOAD
  loadPTData(data: { lines: Line[], routes: Route[], stops: Stop[], departures: Departure[] }): void {
    const ptData = this.getCurrentModeData();
    
    // Add sourceState to loaded data and merge with existing local/modified
    const persistedLines = data.lines.map(l => ({ ...l, sourceState: 'persisted' as SourceState }));
    const persistedRoutes = data.routes.map(r => ({ ...r, sourceState: 'persisted' as SourceState }));
    const persistedStops = data.stops.map(s => ({ ...s, sourceState: 'persisted' as SourceState }));
    const persistedDepartures = data.departures.map(d => ({ ...d, sourceState: 'persisted' as SourceState }));
    
    // Keep local and modified items, add new persisted items
    ptData.lines = [...ptData.lines.filter(l => l.sourceState !== 'persisted'), ...persistedLines];
    ptData.routes = [...ptData.routes.filter(r => r.sourceState !== 'persisted'), ...persistedRoutes];
    ptData.stops = [...ptData.stops.filter(s => s.sourceState !== 'persisted'), ...persistedStops];
    ptData.departures = [...ptData.departures.filter(d => d.sourceState !== 'persisted'), ...persistedDepartures];
  }
  
  getDataForSave(): { 
    localLines: Line[], 
    modifiedLines: Line[],
    localRoutes: Route[],
    modifiedRoutes: Route[],
    localStops: Stop[],
    modifiedStops: Stop[],
    localDepartures: Departure[],
    modifiedDepartures: Departure[]
  } {
    const ptData = this.getCurrentModeData();
    return {
      localLines: ptData.lines.filter(l => l.sourceState === 'local'),
      modifiedLines: ptData.lines.filter(l => l.sourceState === 'modified'),
      localRoutes: ptData.routes.filter(r => r.sourceState === 'local'),
      modifiedRoutes: ptData.routes.filter(r => r.sourceState === 'modified'),
      localStops: ptData.stops.filter(s => s.sourceState === 'local'),
      modifiedStops: ptData.stops.filter(s => s.sourceState === 'modified'),
      localDepartures: ptData.departures.filter(d => d.sourceState === 'local'),
      modifiedDepartures: ptData.departures.filter(d => d.sourceState === 'modified')
    };
  }
  
  clearLocalAndModified(): void {
    const ptData = this.getCurrentModeData();
    
    // Remove all local and modified items
    ptData.lines = ptData.lines.filter(l => l.sourceState === 'persisted');
    ptData.routes = ptData.routes.filter(r => r.sourceState === 'persisted');
    ptData.stops = ptData.stops.filter(s => s.sourceState === 'persisted');
    ptData.departures = ptData.departures.filter(d => d.sourceState === 'persisted');
    
    // Update sourceState of all remaining items to persisted
    ptData.lines.forEach(l => l.sourceState = 'persisted');
    ptData.routes.forEach(r => r.sourceState = 'persisted');
    ptData.stops.forEach(s => s.sourceState = 'persisted');
    ptData.departures.forEach(d => d.sourceState = 'persisted');
  }
}

export const ptState = new PTState();