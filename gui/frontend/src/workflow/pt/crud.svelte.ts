// PT CRUD Operations - Create, Read, Update, Delete Functions
import { trackLineChange, trackRouteChange, trackStopChange, trackDepartureChange } from './ptChangeTracking';
import { nanoid } from 'nanoid';
import type { TransportMode, Line, Route, Stop, Departure } from './types';
import { lines, routes, stops, departures, sequence, selected } from './state.svelte';

// LINES CRUD
export function createLine(name: string, mode: TransportMode): string {
  const lineId = `line_${nanoid(10)}`;
  lines[lineId] = {
    id: lineId,
    name,
    mode,
    unsaved: true
  };
  
  trackLineChange({ id: lineId, name, type: mode }, 'save');
  return lineId;
}

export function updateLine(lineId: string, updates: Partial<Line>): void {
  if (lines[lineId]) {
    Object.assign(lines[lineId], updates);
    lines[lineId].unsaved = true;
    trackLineChange({ id: lineId, name: lines[lineId].name, type: lines[lineId].mode }, 'save');
  }
}

export function deleteLine(lineId: string): void {
  if (lines[lineId]) {
    const line = lines[lineId];
    
    // Delete all routes associated with this line
    Object.keys(routes).forEach(routeId => {
      if (routes[routeId].lineId === lineId) {
        delete routes[routeId];
      }
    });
    
    // Clear selection if this line was selected
    if (selected.lineId === lineId) {
      selected.lineId = null;
      selected.routeId = null;
    }
    
    // Track change
    trackLineChange({ id: lineId, name: line.name, type: line.mode }, 'delete');
    delete lines[lineId];
  }
}

// ROUTES CRUD
export function createRoute(name: string, lineId: string): string {
  const routeId = `route_${nanoid(10)}`;
  routes[routeId] = {
    id: routeId,
    name,
    lineId,
    unsaved: true
  };
  
  // Mark parent line as unsaved
  if (lines[lineId]) {
    lines[lineId].unsaved = true;
  }
  
  trackRouteChange({ id: routeId, name, lineId }, 'save');
  return routeId;
}

export function updateRoute(routeId: string, updates: Partial<Route>): void {
  if (routes[routeId]) {
    const oldLineId = routes[routeId].lineId;
    Object.assign(routes[routeId], updates);
    routes[routeId].unsaved = true;
    
    // Mark line as unsaved
    if (routes[routeId].lineId) {
      lines[routes[routeId].lineId].unsaved = true;
    }
    
    // If lineId changed, mark old line as unsaved too
    if (oldLineId !== routes[routeId].lineId && lines[oldLineId]) {
      lines[oldLineId].unsaved = true;
    }
    
    trackRouteChange({ id: routeId, name: routes[routeId].name, lineId: routes[routeId].lineId }, 'save');
  }
}

export function deleteRoute(routeId: string): void {
  if (routes[routeId]) {
    const lineId = routes[routeId].lineId;
    
    // Clear selection if this route was selected
    if (selected.routeId === routeId) {
      selected.routeId = null;
    }
    
    // Mark parent line as unsaved
    if (lines[lineId]) {
      lines[lineId].unsaved = true;
    }
    
    trackRouteChange({ id: routeId, name: routes[routeId].name, lineId }, 'delete');
    delete routes[routeId];
  }
}

// STOPS CRUD (for current route)
export function createStop(stopData: Omit<Stop, 'unsaved'>): string {
  const stopId = stopData.stopId || `stop_${nanoid(10)}`;
  stops[stopId] = {
    ...stopData,
    stopId,
    stopType: stopData.stopType || 'REGULAR',
    wheelchairAccessible: stopData.wheelchairAccessible || 'UNKNOWN',
    timingPoint: stopData.timingPoint || false,
    unsaved: true
  };
  
  trackStopChange(stops[stopId], 'save');
  return stopId;
}

export function updateStop(stopId: string, updates: Partial<Stop>): void {
  if (stops[stopId]) {
    Object.assign(stops[stopId], updates);
    stops[stopId].unsaved = true;
    trackStopChange(stops[stopId], 'save');
  }
}

export function deleteStop(stopId: string): void {
  if (stops[stopId]) {
    trackStopChange(stops[stopId], 'delete');
    delete stops[stopId];
    
    // Remove from sequence
    sequence.splice(sequence.indexOf(stopId), 1);
    
    // Clear selection if this stop was selected
    if (selected.stopId === stopId) {
      selected.stopId = null;
    }
  }
}

// DEPARTURES CRUD (for current route)
export function createDeparture(time: string, vehicleId?: string, routeId?: string): string {
  const departureId = `dep_${nanoid(10)}`;
  const finalRouteId = routeId || selected.routeId;
  
  if (!finalRouteId) return '';
  
  departures[departureId] = {
    id: departureId,
    routeId: finalRouteId,
    departureTime: time,
    vehicleRefId: vehicleId,
    unsaved: true
  };
  
  trackDepartureChange({ ...departures[departureId], routeId: finalRouteId! }, 'save');
  return departureId;
}

export function updateDeparture(departureId: string, updates: Partial<Departure>): void {
  if (departures[departureId]) {
    Object.assign(departures[departureId], updates);
    departures[departureId].unsaved = true;
    trackDepartureChange({ ...departures[departureId], routeId: departures[departureId].routeId }, 'save');
  }
}

export function deleteDeparture(departureId: string): void {
  if (departures[departureId]) {
    trackDepartureChange({ ...departures[departureId], routeId: departures[departureId].routeId }, 'delete');
    delete departures[departureId];
  }
}