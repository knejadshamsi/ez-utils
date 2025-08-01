// PT Functions - Utility Functions and State Management
import type { TransportMode, Line, Route, Stop, Departure } from './types';
import { lines, routes, stops, departures, sequence, selected } from './state.svelte';

// SEQUENCE MANAGEMENT
export function updateSequence(newSequence: string[]): void {
  sequence.splice(0, sequence.length, ...newSequence);
}

export function addStopToSequence(stopId: string, index?: number): void {
  if (index !== undefined) {
    sequence.splice(index, 0, stopId);
  } else {
    sequence.push(stopId);
  }
}

export function removeStopFromSequence(stopId: string): void {
  const index = sequence.indexOf(stopId);
  if (index > -1) {
    sequence.splice(index, 1);
  }
}

// MODE SWITCHING
export function switchMode(newMode: TransportMode): void {
  selected.mode = newMode;
  selected.lineId = null;
  selected.routeId = null;
  selected.stopId = null;
  
  // Clear stops and departures for the new mode
  Object.keys(stops).forEach(key => delete stops[key]);
  Object.keys(departures).forEach(key => delete departures[key]);
  sequence.splice(0, sequence.length);
}

// DATA LOADING AND STATE MANAGEMENT HELPERS
export function loadLinesAndRoutes(mode: TransportMode, data: Array<{
  id: string;
  name: string;
  type: 'LINE' | 'ROUTE';
  lineId?: string;
}>): void {
  // Clear existing data for this mode
  Object.keys(lines).forEach(key => {
    if (lines[key].mode === mode) {
      delete lines[key];
    }
  });
  Object.keys(routes).forEach(key => {
    if (routes[key].lineId && lines[routes[key].lineId]?.mode === mode) {
      delete routes[key];
    }
  });
  
  // Process the data
  data.forEach(item => {
    if (item.type === 'LINE') {
      if (!lines[item.id]) {
        lines[item.id] = {
          id: item.id,
          name: item.name,
          mode,
          unsaved: false
        };
      }
    } else if (item.type === 'ROUTE' && item.lineId) {
      if (!routes[item.id]) {
        routes[item.id] = {
          id: item.id,
          name: item.name,
          lineId: item.lineId,
          unsaved: false
        };
      }
    }
  });
}

export function loadRouteData(routeId: string, newSequence: string[], newStops: Stop[], newDepartures: Departure[]): void {
  // Clear existing stops and departures
  Object.keys(stops).forEach(key => delete (stops as any)[key]);
  Object.keys(departures).forEach(key => delete (departures as any)[key]);
  sequence.splice(0, sequence.length);
  
  // Load new data
  sequence.splice(0, sequence.length, ...newSequence);
  
  newStops.forEach(stop => {
    (stops as any)[stop.stopId] = stop;
  });
  
  newDepartures.forEach(departure => {
    (departures as any)[departure.id] = departure;
  });
}

// CLEAR STATE
export function clearRouteData(): void {
  Object.keys(stops).forEach(key => delete stops[key]);
  Object.keys(departures).forEach(key => delete departures[key]);
  sequence.splice(0, sequence.length);
  selected.routeId = null;
  selected.stopId = null;
}

// UTILITY METHODS
export function getRoutesForLine(lineId: string) {
  return Object.values(routes).filter(route => route.lineId === lineId);
}

export function getStopsInSequence() {
  return sequence.map(stopId => stops[stopId]).filter(Boolean);
}

export function getDeparturesForRoute(routeId: string) {
  return Object.values(departures).filter(dep => dep.routeId === routeId);
}

// UI STATE MANAGEMENT METHODS
export function selectMode(mode: TransportMode) {
  switchMode(mode);
}

export function selectLine(lineId: string) {
  selected.lineId = lineId;
  selected.routeId = null;
  selected.stopId = null;
  
  // Clear route-specific data
  clearRouteData();
}

export function selectRoute(routeId: string | null) {
  selected.routeId = routeId;
  selected.stopId = null;
}

export function selectStop(stopId: string | null) {
  selected.stopId = selected.stopId === stopId ? null : stopId;
}

export function setEditMode(mode: typeof selected.editMode) {
  selected.editMode = mode;
}