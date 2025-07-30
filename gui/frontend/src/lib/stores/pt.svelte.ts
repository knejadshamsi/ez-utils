import { trackLineChange, trackRouteChange, trackStopChange, trackDepartureChange } from '$lib/utils/ptChangeTracking';
import { nanoid } from 'nanoid';

export type TransportMode = 'BUS' | 'METRO' | 'TRAM';

export type LineData = {
  name: string;
  unsaved: boolean;
  routes: {
    [routeId: string]: {
      name: string;
      unsaved: boolean;
    }
  }
};

export type PTData = {
  BUS: { [lineId: string]: LineData };
  METRO: { [lineId: string]: LineData };
  TRAM: { [lineId: string]: LineData };
};

export type StopType = 'REGULAR' | 'REQUEST' | 'BOARDING_ONLY' | 'ALIGHTING_ONLY';
export type AccessibilityStatus = 'YES' | 'NO' | 'UNKNOWN';

export type Stop = {
  stopId: string;
  stopName: string;
  lat: number;
  lng: number;
  arrivalOffset: string;
  departureOffset: string;
  stopType?: StopType;
  wheelchairAccessible?: AccessibilityStatus;
  timingPoint?: boolean;
  attributes?: Record<string, string | number | boolean>;
  unsaved: boolean;
};

export type Departure = {
  id: string;
  departureTime: string;
  vehicleRefId?: string;
  unsaved: boolean;
};

export type RouteStop = {
  linkId: string;
  routeId: string;
  stopId: string;
  sequence: number;
};

export const TRANSPORT_MODES = {
  BUS: { value: 'BUS', label: 'Bus', icon: '🚌', abbreviation: 'B' },
  METRO: { value: 'METRO', label: 'Metro', icon: '🚇', abbreviation: 'M' },
  TRAM: { value: 'TRAM', label: 'Tram', icon: '🚊', abbreviation: 'T' }
}

export type PTEditMode = 'NORMAL' | 'ADDING_STOP' | 'DRAGGING_STOP' | 'EDITING_STOP_ATTRIBUTES' | 'EDITING_DEPARTURES' | 'ADDING_MULTIPLE_STOPS' | 'SELECTING_STOP_LOCATION';

class PTState {
  // Nested object structure for lines and routes
  ptData = $state<PTData>({
    BUS: {},
    METRO: {},
    TRAM: {}
  });
  
  // Current route data with object-based storage
  currentRouteData = $state<{
    routeId: string | null;
    sequence: string[];  // Array of stopIds in order
    stops: { [stopId: string]: Stop };  // Keyed by stopId
    departures: { [departureId: string]: Departure };  // Keyed by departureId
  }>({
    routeId: null,
    sequence: [],
    stops: {},
    departures: {}
  });
  
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

  // LINE CRUD
  createLine(name: string): string {
    const lineId = `line_${nanoid(10)}`;
    this.ptData[this.selected.mode][lineId] = {
      name,
      unsaved: true,
      routes: {}
    };
    
    // Track change
    trackLineChange({ id: lineId, name, type: this.selected.mode }, 'save');
    
    return lineId;
  }

  updateLine(lineId: string, name: string): void {
    const line = this.ptData[this.selected.mode][lineId];
    if (line) {
      line.name = name;
      line.unsaved = true;
      
      // Track change
      trackLineChange({ id: lineId, name, type: this.selected.mode }, 'save');
    }
  }

  deleteLine(lineId: string): void {
    const mode = this.selected.mode;
    const line = this.ptData[mode][lineId];
    
    if (line) {
      // Clear selection if this line was selected
      if (this.selected.lineId === lineId) {
        this.selected.lineId = null;
        this.selected.routeId = null;
      }
      
      // Delete the line
      delete this.ptData[mode][lineId];
      
      // Track deletion if it was saved
      if (!line.unsaved) {
        trackLineChange({ id: lineId, name: line.name, type: mode }, 'delete');
      }
    }
  }

  // ROUTE CRUD
  createRoute(lineId: string, name: string): string {
    const routeId = `route_${nanoid(10)}`;
    const line = this.ptData[this.selected.mode][lineId];
    
    if (line) {
      line.routes[routeId] = {
        name,
        unsaved: true
      };
      
      // Mark line as unsaved too
      line.unsaved = true;
      
      // Track change
      trackRouteChange({ id: routeId, name, lineId }, 'save');
    }
    
    return routeId;
  }

  updateRoute(lineId: string, routeId: string, name: string): void {
    const line = this.ptData[this.selected.mode][lineId];
    if (line?.routes[routeId]) {
      line.routes[routeId].name = name;
      line.routes[routeId].unsaved = true;
      
      // Track change
      trackRouteChange({ id: routeId, name, lineId }, 'save');
    }
  }

  deleteRoute(lineId: string, routeId: string): void {
    const line = this.ptData[this.selected.mode][lineId];
    if (line?.routes[routeId]) {
      const route = line.routes[routeId];
      
      // Clear selection if this route was selected
      if (this.selected.routeId === routeId) {
        this.selected.routeId = null;
      }
      
      // Clear current route data if it was loaded
      if (this.currentRouteData.routeId === routeId) {
        this.currentRouteData = {
          routeId: null,
          sequence: [],
          stops: {},
          departures: {}
        };
      }
      
      // Delete the route
      delete line.routes[routeId];
      
      // Track deletion if it was saved
      if (!route.unsaved) {
        trackRouteChange({ id: routeId, name: route.name, lineId }, 'delete');
      }
    }
  }

  // STOP CRUD (operates on currentRouteData)
  createStop(stopData: Omit<Stop, 'unsaved'>): string {
    if (!this.currentRouteData.routeId) return '';
    
    const stopId = stopData.stopId || `stop_${nanoid(10)}`;
    
    const newStop: Stop = {
      ...stopData,
      stopId,
      stopType: stopData.stopType || 'REGULAR',
      wheelchairAccessible: stopData.wheelchairAccessible || 'UNKNOWN',
      timingPoint: stopData.timingPoint || false,
      unsaved: true
    };
    
    // Add to stops map
    this.currentRouteData.stops[stopId] = newStop;
    
    // Track change
    trackStopChange(newStop, 'save');
    
    return stopId;
  }

  updateStop(stopId: string, updates: Partial<Stop>): void {
    const stop = this.currentRouteData.stops[stopId];
    if (stop) {
      Object.assign(stop, updates);
      stop.unsaved = true;
      
      // Track change
      trackStopChange(stop, 'save');
    }
  }

  deleteStop(stopId: string): void {
    const stop = this.currentRouteData.stops[stopId];
    if (stop) {
      // Remove from sequence
      this.currentRouteData.sequence = this.currentRouteData.sequence.filter(id => id !== stopId);
      
      // Remove from stops map
      delete this.currentRouteData.stops[stopId];
      
      // Track deletion
      trackStopChange(stop, 'delete');
    }
  }

  // DEPARTURE CRUD
  createDeparture(time: string, vehicleId?: string): string {
    if (!this.currentRouteData.routeId) return '';
    
    const departureId = `dep_${nanoid(10)}`;
    
    const newDeparture: Departure = {
      id: departureId,
      departureTime: time,
      vehicleRefId: vehicleId,
      unsaved: true
    };
    
    // Add to departures map
    this.currentRouteData.departures[departureId] = newDeparture;
    
    // Track change
    trackDepartureChange({ ...newDeparture, routeId: this.currentRouteData.routeId }, 'save');
    
    return departureId;
  }

  updateDeparture(departureId: string, updates: Partial<Departure>): void {
    const departure = this.currentRouteData.departures[departureId];
    if (departure) {
      Object.assign(departure, updates);
      departure.unsaved = true;
      
      // Track change
      trackDepartureChange({ ...departure, routeId: this.currentRouteData.routeId! }, 'save');
    }
  }

  deleteDeparture(departureId: string): void {
    const departure = this.currentRouteData.departures[departureId];
    if (departure) {
      // Remove from departures map
      delete this.currentRouteData.departures[departureId];
      
      // Track deletion
      trackDepartureChange({ ...departure, routeId: this.currentRouteData.routeId! }, 'delete');
    }
  }

  // BATCH OPERATIONS
  batchCreateStops(stops: Array<Omit<Stop, 'unsaved'>>): string[] {
    return stops.map(stopData => this.createStop(stopData));
  }

  // MODE SWITCHING
  switchMode(newMode: TransportMode): void {
    this.selected.mode = newMode;
    this.selected.lineId = null;
    this.selected.routeId = null;
    this.selected.stopId = null;
    
    // Clear current route data
    this.currentRouteData = {
      routeId: null,
      sequence: [],
      stops: {},
      departures: {}
    };
  }

  // DATA LOADING
  loadLinesAndRoutes(mode: TransportMode, data: Array<{
    id: string;
    name: string;
    type: 'LINE' | 'ROUTE';
    lineId?: string;
  }>): void {
    // Clear existing saved data for this mode
    Object.keys(this.ptData[mode]).forEach(lineId => {
      const line = this.ptData[mode][lineId];
      if (!line.unsaved) {
        delete this.ptData[mode][lineId];
      }
    });
    
    // Process the data
    data.forEach(item => {
      if (item.type === 'LINE') {
        // Add line if it doesn't exist as unsaved
        if (!this.ptData[mode][item.id]?.unsaved) {
          this.ptData[mode][item.id] = {
            name: item.name,
            unsaved: false,
            routes: this.ptData[mode][item.id]?.routes || {}
          };
        }
      } else if (item.type === 'ROUTE' && item.lineId) {
        // Ensure line exists
        if (!this.ptData[mode][item.lineId]) {
          // This shouldn't happen, but handle gracefully
          this.ptData[mode][item.lineId] = {
            name: 'Unknown Line',
            unsaved: false,
            routes: {}
          };
        }
        
        // Add route if it doesn't exist as unsaved
        if (!this.ptData[mode][item.lineId].routes[item.id]?.unsaved) {
          this.ptData[mode][item.lineId].routes[item.id] = {
            name: item.name,
            unsaved: false
          };
        }
      }
    });
  }

  loadRouteData(routeId: string, sequence: string[], stops: Stop[], departures: Departure[]): void {
    // Keep existing unsaved stops and departures
    const existingUnsavedStops: { [id: string]: Stop } = {};
    const existingUnsavedDepartures: { [id: string]: Departure } = {};
    
    if (this.currentRouteData.routeId === routeId) {
      // Keep unsaved data from current route
      Object.entries(this.currentRouteData.stops).forEach(([id, stop]) => {
        if (stop.unsaved) {
          existingUnsavedStops[id] = stop;
        }
      });
      
      Object.entries(this.currentRouteData.departures).forEach(([id, dep]) => {
        if (dep.unsaved) {
          existingUnsavedDepartures[id] = dep;
        }
      });
    }
    
    // Convert arrays to objects
    const stopsMap: { [stopId: string]: Stop } = {};
    stops.forEach(stop => {
      // Use existing unsaved version if available
      stopsMap[stop.stopId] = existingUnsavedStops[stop.stopId] || { ...stop, unsaved: false };
    });
    
    const departuresMap: { [departureId: string]: Departure } = {};
    departures.forEach(dep => {
      // Use existing unsaved version if available
      departuresMap[dep.id] = existingUnsavedDepartures[dep.id] || { ...dep, unsaved: false };
    });
    
    // Add any unsaved items that weren't in the fetched data
    Object.assign(stopsMap, existingUnsavedStops);
    Object.assign(departuresMap, existingUnsavedDepartures);
    
    this.currentRouteData = {
      routeId,
      sequence,
      stops: stopsMap,
      departures: departuresMap
    };
  }

  // SAVE HELPERS
  markLineAsSaved(lineId: string): void {
    const line = this.ptData[this.selected.mode][lineId];
    if (line) {
      line.unsaved = false;
      // Mark all routes as saved too
      Object.keys(line.routes).forEach(routeId => {
        line.routes[routeId].unsaved = false;
      });
    }
  }

  markRouteAsSaved(lineId: string, routeId: string): void {
    const line = this.ptData[this.selected.mode][lineId];
    if (line?.routes[routeId]) {
      line.routes[routeId].unsaved = false;
    }
  }

  markStopsAsSaved(stopIds: string[]): void {
    stopIds.forEach(stopId => {
      if (this.currentRouteData.stops[stopId]) {
        this.currentRouteData.stops[stopId].unsaved = false;
      }
    });
  }

  markDeparturesAsSaved(departureIds: string[]): void {
    departureIds.forEach(depId => {
      if (this.currentRouteData.departures[depId]) {
        this.currentRouteData.departures[depId].unsaved = false;
      }
    });
  }

  // Get all unsaved items
  getUnsavedItems(): {
    lines: Array<{ id: string; name: string; mode: TransportMode }>;
    routes: Array<{ id: string; name: string; lineId: string; mode: TransportMode }>;
    stops: Stop[];
    departures: Array<Departure & { routeId: string }>;
  } {
    const lines: Array<{ id: string; name: string; mode: TransportMode }> = [];
    const routes: Array<{ id: string; name: string; lineId: string; mode: TransportMode }> = [];
    
    Object.entries(this.ptData).forEach(([mode, modeData]) => {
      Object.entries(modeData).forEach(([lineId, line]) => {
        if (line.unsaved) {
          lines.push({ id: lineId, name: line.name, mode: mode as TransportMode });
        }
        
        Object.entries(line.routes).forEach(([routeId, route]) => {
          if (route.unsaved) {
            routes.push({ 
              id: routeId, 
              name: route.name, 
              lineId, 
              mode: mode as TransportMode 
            });
          }
        });
      });
    });
    
    // Get unsaved stops and departures from current route
    const stops = Object.values(this.currentRouteData.stops).filter(s => s.unsaved);
    const departures = Object.values(this.currentRouteData.departures)
      .filter(d => d.unsaved)
      .map(d => ({ ...d, routeId: this.currentRouteData.routeId! }));
    
    return { lines, routes, stops, departures };
  }
}

export const ptState = new PTState();