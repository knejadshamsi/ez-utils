import { trackLineChange, trackRouteChange, trackStopChange, trackDepartureChange } from '$lib/utils/ptChangeTracking';
import { nanoid } from 'nanoid';

export type TransportMode = 'BUS' | 'METRO' | 'TRAM';

export type Line = {
  id: string;
  type: TransportMode;
  name: string;
  routes: Array<{ id: string; name: string; stops: number }>;
  local?: boolean;
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
};

export type Departure = {
  id: string;
  routeId: string;
  departureTime: string;
  vehicleRefId?: string;
};

export type PTData = {
  mode: TransportMode;
  lines: Line[];
  stops: Stop[];
  departures: Departure[];
};

export const TRANSPORT_MODES = {
  BUS: { value: 'BUS', label: 'Bus', icon: '🚌', abbreviation: 'B' },
  METRO: { value: 'METRO', label: 'Metro', icon: '🚇', abbreviation: 'M' },
  TRAM: { value: 'TRAM', label: 'Tram', icon: '🚊', abbreviation: 'T' }
}

export type PTEditMode = 'NORMAL' | 'ADDING_STOP' | 'DRAGGING_STOP' | 'EDITING_STOP_ATTRIBUTES' | 'EDITING_DEPARTURES';

class PTState {
  ptData = $state<PTData[]>([]);
  
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
  
  // Map interaction states
  isDraggingStop = $state<boolean>(false);
  isAddingMultipleStops = $state<boolean>(false);
  isSelectingStopLocation = $state<boolean>(false);
  selectingStopId = $state<string | null>(null);

  // Helper to get or create PTData for current mode
  private getOrCreatePTData(): PTData {
    let ptData = this.ptData.find(pd => pd.mode === this.selected.mode);
    if (!ptData) {
      ptData = {
        mode: this.selected.mode,
        lines: [],
        stops: [],
        departures: []
      };
      this.ptData.push(ptData);
    }
    return ptData;
  }

  // LINE CRUD
  createLine(name: string): void {
    const ptData = this.getOrCreatePTData();
    const newLine: Line = {
      id: `line_${nanoid(10)}`,
      name,
      type: this.selected.mode,
      routes: [],
      local: true
    };
    ptData.lines.push(newLine);
    
    
    // Track change
    trackLineChange(newLine, 'add');
  }

  updateLine(lineId: string, updates: {name?: string}): void {
    const ptData = this.getOrCreatePTData();
    const lineIndex = ptData.lines.findIndex(l => l.id === lineId);
    if (lineIndex !== -1) {
      Object.assign(ptData.lines[lineIndex], updates);
      // Mark as local when edited
      ptData.lines[lineIndex].local = true;
      
      
      // Track change
      trackLineChange(ptData.lines[lineIndex], 'update');
    }
  }

  deleteLine(lineId: string): void {
    const ptData = this.getOrCreatePTData();
    const lineIndex = ptData.lines.findIndex(l => l.id === lineId);
    if (lineIndex !== -1) {
      const line = ptData.lines[lineIndex];
      
      // Delete all stops and departures for all routes in this line
      line.routes.forEach(route => {
        // Remove stops for this route
        ptData.stops = ptData.stops.filter(s => s.routeId !== route.id);
        // Remove departures for this route
        ptData.departures = ptData.departures.filter(d => d.routeId !== route.id);
      });
      
      // Remove the line
      ptData.lines.splice(lineIndex, 1);
      
      
      // Clear selection if this line was selected
      if (this.selected.lineId === lineId) {
        this.selected.lineId = null;
        this.selected.routeId = null;
      }
      
      // Track change
      trackLineChange(line, 'delete');
    }
  }

  // ROUTE CRUD
  createRoute(lineId: string, name: string): void {
    const ptData = this.getOrCreatePTData();
    const line = ptData.lines.find(l => l.id === lineId);
    if (line) {
      const newRoute = {
        id: `route_${nanoid(10)}`,
        name,
        stops: 0
      };
      line.routes.push(newRoute);
      
      // Track change
      trackRouteChange({...newRoute, lineId}, 'add');
    }
  }

  updateRoute(lineId: string, routeId: string, updates: {name?: string}): void {
    const ptData = this.getOrCreatePTData();
    const line = ptData.lines.find(l => l.id === lineId);
    if (line) {
      const route = line.routes.find(r => r.id === routeId);
      if (route) {
        Object.assign(route, updates);
        // Track change
        trackRouteChange({...route, lineId}, 'update');
      }
    }
  }

  deleteRoute(lineId: string, routeId: string): void {
    const ptData = this.getOrCreatePTData();
    const line = ptData.lines.find(l => l.id === lineId);
    if (line) {
      const routeIndex = line.routes.findIndex(r => r.id === routeId);
      if (routeIndex !== -1) {
        // Remove the route
        line.routes.splice(routeIndex, 1);
        
        // Remove all stops for this route
        ptData.stops = ptData.stops.filter(s => s.routeId !== routeId);
        
        // Remove all departures for this route
        ptData.departures = ptData.departures.filter(d => d.routeId !== routeId);
        
        // Clear selection if this route was selected
        if (this.selected.routeId === routeId) {
          this.selected.routeId = null;
        }
        
        // Track change
        trackRouteChange({id: routeId, name: '', lineId}, 'delete');
      }
    }
  }

  // STOP CRUD
  createStop(stopData: Omit<Stop, 'sequence'>): void {
    const ptData = this.getOrCreatePTData();
    
    // Calculate next sequence number for this route
    const routeStops = ptData.stops.filter(s => s.routeId === stopData.routeId);
    const maxSequence = routeStops.reduce((max, s) => Math.max(max, s.sequence), 0);
    
    const newStop: Stop = {
      ...stopData,
      sequence: maxSequence + 1,
      stopType: stopData.stopType || 'REGULAR',
      wheelchairAccessible: stopData.wheelchairAccessible || 'UNKNOWN',
      timingPoint: stopData.timingPoint || false
    };
    
    ptData.stops.push(newStop);
    
    // Update route stop count
    this.updateRouteStopCount(stopData.routeId);
    
    // Track change
    trackStopChange(newStop, 'add');
  }

  updateStop(stopId: string, updates: Partial<Stop>): void {
    const ptData = this.getOrCreatePTData();
    const stopIndex = ptData.stops.findIndex(s => s.stopId === stopId);
    if (stopIndex !== -1) {
      Object.assign(ptData.stops[stopIndex], updates);
      // Track change
      trackStopChange(ptData.stops[stopIndex], 'update');
    }
  }

  deleteStop(stopId: string): void {
    const ptData = this.getOrCreatePTData();
    const stopIndex = ptData.stops.findIndex(s => s.stopId === stopId);
    if (stopIndex !== -1) {
      const stop = ptData.stops[stopIndex];
      const routeId = stop.routeId;
      
      // Remove the stop
      ptData.stops.splice(stopIndex, 1);
      
      // Resequence remaining stops for this route
      const remainingStops = ptData.stops
        .filter(s => s.routeId === routeId)
        .sort((a, b) => a.sequence - b.sequence);
      
      remainingStops.forEach((stop, index) => {
        stop.sequence = index + 1;
      });
      
      // Update route stop count
      this.updateRouteStopCount(routeId);
      
      // Track change
      trackStopChange(stop, 'delete');
    }
  }

  // COMMENTED OUT - Stop reordering not implemented yet
  // reorderStops(routeId: string, stopIds: string[]): void {
  //   const ptData = this.getOrCreatePTData();
  //   const routeStops = ptData.stops.filter(s => s.routeId === routeId);
    
  //   stopIds.forEach((stopId, index) => {
  //     const stop = routeStops.find(s => s.stopId === stopId);
  //     if (stop) {
  //       stop.sequence = index + 1;
  //       // Track change
  //       trackStopChange(stop, 'update');
  //     }
  //   });
  // }

  // DEPARTURE CRUD
  createDeparture(routeId: string, time: string, vehicleId?: string): void {
    const ptData = this.getOrCreatePTData();
    const newDeparture: Departure = {
      id: `dep_${nanoid(10)}`,
      routeId,
      departureTime: time,
      vehicleRefId: vehicleId
    };
    
    ptData.departures.push(newDeparture);
    
    // Track change
    trackDepartureChange(newDeparture, 'add');
  }

  updateDeparture(departureId: string, updates: Partial<Departure>): void {
    const ptData = this.getOrCreatePTData();
    const depIndex = ptData.departures.findIndex(d => d.id === departureId);
    if (depIndex !== -1) {
      Object.assign(ptData.departures[depIndex], updates);
      // Track change
      trackDepartureChange(ptData.departures[depIndex], 'update');
    }
  }

  deleteDeparture(departureId: string): void {
    const ptData = this.getOrCreatePTData();
    const depIndex = ptData.departures.findIndex(d => d.id === departureId);
    if (depIndex !== -1) {
      const departure = ptData.departures[depIndex];
      ptData.departures.splice(depIndex, 1);
      // Track change
      trackDepartureChange(departure, 'delete');
    }
  }

  // BATCH OPERATIONS
  batchCreateStops(routeId: string, stops: Omit<Stop, 'sequence'>[]): void {
    const ptData = this.getOrCreatePTData();
    
    // Get current max sequence for this route
    const routeStops = ptData.stops.filter(s => s.routeId === routeId);
    let sequence = routeStops.reduce((max, s) => Math.max(max, s.sequence), 0);
    
    const newStops = stops.map(stopData => {
      sequence++;
      const newStop: Stop = {
        ...stopData,
        routeId,
        sequence
      };
      // Track each stop
      trackStopChange(newStop, 'add');
      return newStop;
    });
    
    ptData.stops.push(...newStops);
    this.updateRouteStopCount(routeId);
  }

  deleteRouteWithAllData(lineId: string, routeId: string): void {
    // This combines deleteRoute + cascading delete of stops/departures
    this.deleteRoute(lineId, routeId);
  }

  // HELPER METHODS
  private updateRouteStopCount(routeId: string): void {
    const ptData = this.getOrCreatePTData();
    const stopCount = ptData.stops.filter(s => s.routeId === routeId).length;
    
    // Find the line containing this route and update stop count
    for (let i = 0; i < ptData.lines.length; i++) {
      const routeIndex = ptData.lines[i].routes.findIndex(r => r.id === routeId);
      if (routeIndex !== -1) {
        // Update the stop count
        ptData.lines[i].routes[routeIndex].stops = stopCount;
        // Force reactivity by reassigning the array
        ptData.lines = [...ptData.lines];
        break;
      }
    }
  }
}

export const ptState = new PTState();