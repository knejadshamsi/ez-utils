// PT Data Loading Functions - API calls with state updates
import { appState } from '$lib/stores/app.svelte';
import type { TransportMode, Line, Route, Stop, Departure, StopType, AccessibilityStatus } from './types';
import { lines, routes, stops, departures, sequence, selected } from './state.svelte';
import {
  fetchPTLinesByMode,
  fetchPTRoutesByLineID,
  fetchPTStopsByRouteID,
  fetchPTStops,
  fetchPTStopsInBounds,
  fetchPTDeparturesByRouteID,
  savePTLines,
  savePTRoutes,
  savePTStops,
  savePTRouteStops,
  savePTDepartures,
  deletePTLinesAPI,
  deletePTRoutesAPI,
  deletePTStopsAPI,
  deletePTDeparturesAPI
} from './api';
import { deleteLine, deleteRoute, deleteStop, deleteDeparture } from './crud.svelte';

// Helper functions
function markLineAsSaved(lineId: string): void {
  if (lines[lineId]) {
    lines[lineId].unsaved = false;
  }
}

function markRouteAsSaved(routeId: string): void {
  if (routes[routeId]) {
    routes[routeId].unsaved = false;
  }
}

function markStopsAsSaved(stopIds: string[]): void {
  stopIds.forEach(stopId => {
    if (stops[stopId]) {
      stops[stopId].unsaved = false;
    }
  });
}

function markDeparturesAsSaved(departureIds: string[]): void {
  departureIds.forEach(depId => {
    if (departures[depId]) {
      departures[depId].unsaved = false;
    }
  });
}

function getUnsavedItems(): {
  lines: Array<{ id: string; name: string; mode: TransportMode }>;
  routes: Array<{ id: string; name: string; lineId: string }>;
  stops: Stop[];
  departures: Array<Departure & { routeId: string }>;
} {
  const unsavedLines = Object.values(lines).filter(line => line.unsaved);
  const unsavedRoutes = Object.values(routes).filter(route => route.unsaved);
  const unsavedStops = Object.values(stops).filter(stop => stop.unsaved);
  const unsavedDepartures = Object.values(departures).filter(dep => dep.unsaved);
  
  return {
    lines: unsavedLines,
    routes: unsavedRoutes,
    stops: unsavedStops,
    departures: unsavedDepartures
  };
}

// Data loading functions
export async function loadPTData(mode: TransportMode): Promise<void> {
  try {
    // Check if we already have complete data for this mode
    const existingLines = Object.values(lines).filter(line => line.mode === mode);
    const existingRoutes = Object.values(routes).filter(route => 
      existingLines.some(line => line.id === route.lineId)
    );
    
    if (existingLines.length > 0 && existingRoutes.length > 0) {
      return; // Already loaded with routes
    }
    
    const linesData = await fetchPTLinesByMode(appState.processId, mode);
    
    // Clear existing data for this mode
    Object.keys(lines).forEach(key => {
      if (key in lines && lines[key].mode === mode) {
        delete lines[key];
      }
    });
    
    // Add lines
    linesData.forEach(line => {
      lines[line.id] = {
        id: line.id,
        name: line.name,
        mode: mode,
        unsaved: false
      };
    });
    
    // Load routes for each line
    for (const line of linesData) {
      const routesData = await fetchPTRoutesByLineID(appState.processId, line.id);
      routesData.forEach(route => {
        routes[route.id] = {
          id: route.id,
          name: route.name,
          lineId: line.id,
          unsaved: false
        };
      });
    }
    
  } catch (error) {
    throw new Error(`Failed to load ${mode} data: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function loadRouteStops(routeId: string): Promise<void> {
  try {
    // Check if we already have data for this route
    if (sequence.length > 0 && selected.routeId === routeId) {
      return; // Already loaded for this route
    }
    
    // Clear existing stops and sequence
    Object.keys(stops).forEach(key => delete stops[key]);
    sequence.splice(0, sequence.length);
    
    const routeStops = await fetchPTStopsByRouteID(appState.processId, routeId);
    
    if (routeStops.length === 0) {
      return;
    }
    
    // Sort by sequence
    const sortedStops = routeStops.sort((a, b) => a.sequence - b.sequence);
    
    // Build sequence
    sequence.splice(0, sequence.length, ...sortedStops.map(rs => rs.stopId));
    
    // Load stop details
    const stopIds = [...new Set(sortedStops.map(rs => rs.stopId))];
    if (stopIds.length > 0) {
      const stopDetails = await fetchPTStops(appState.processId, stopIds);
      
      stopDetails.forEach(stop => {
        stops[stop.stopId] = {
          stopId: stop.stopId,
          stopName: stop.stopName,
          lat: stop.lat,
          lng: stop.lng,
          arrivalOffset: stop.arrivalOffset,
          departureOffset: stop.departureOffset,
          stopType: (stop.stopType as StopType) || 'REGULAR',
          wheelchairAccessible: (stop.wheelchairAccessible as AccessibilityStatus) || 'UNKNOWN',
          timingPoint: stop.timingPoint || false,
          attributes: stop.attributes,
          unsaved: false
        };
      });
    }
    
  } catch (error) {
    throw new Error(`Failed to load stops for route: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function loadStopsInBounds(bounds: { minLat: number; maxLat: number; minLng: number; maxLng: number }, mode: TransportMode) {
  try {
    const stops = await fetchPTStopsInBounds(appState.processId, mode, bounds);
    return stops;
  } catch (error) {
    return [];
  }
}

export async function loadRouteDepartures(routeId: string): Promise<void> {
  try {
    // Clear existing departures
    Object.keys(departures).forEach(key => delete departures[key]);
    
    const fetchedDepartures = await fetchPTDeparturesByRouteID(appState.processId, routeId);
    
    fetchedDepartures.forEach(dep => {
      departures[dep.id] = {
        id: dep.id,
        routeId: routeId,
        departureTime: dep.departureTime,
        vehicleRefId: dep.vehicleRefId,
        unsaved: false
      };
    });
    
  } catch (error) {
    throw new Error(`Failed to load departures for route: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

// Save functions
export async function saveCurrentRoute(): Promise<void> {
  if (!selected.routeId) return;
  
  try {
    const unsaved = getUnsavedItems();
    
    // Save lines
    if (unsaved.lines.length > 0) {
      await savePTLines(appState.processId, unsaved.lines.map(line => ({
        id: line.id,
        name: line.name,
        mode: line.mode as string
      })));
      unsaved.lines.forEach(line => markLineAsSaved(line.id));
    }
    
    // Save routes
    if (unsaved.routes.length > 0) {
      await savePTRoutes(appState.processId, unsaved.routes.map(route => ({
        id: route.id,
        name: route.name,
        lineId: route.lineId
      })));
      unsaved.routes.forEach(route => markRouteAsSaved(route.id));
    }
    
    // Save stops
    if (unsaved.stops.length > 0) {
      await savePTStops(appState.processId, unsaved.stops);
      markStopsAsSaved(unsaved.stops.map(s => s.stopId));
    }
    
    // Save departures
    if (unsaved.departures.length > 0) {
      const departuresToSave = unsaved.departures.filter(dep => dep.routeId === selected.routeId);
      await savePTDepartures(appState.processId, departuresToSave);
      markDeparturesAsSaved(departuresToSave.map(d => d.id));
    }
    
    // Save route-sequence relationships
    if (sequence.length > 0 && selected.routeId) {
      const routeStops = sequence.map((stopId, index) => ({
        linkId: `${selected.routeId}_${stopId}`,
        routeId: selected.routeId!,
        stopId: stopId,
        sequence: index + 1
      }));
      await savePTRouteStops(appState.processId, routeStops);
    }
    
  } catch (error) {
    throw new Error(`Failed to save route data: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

// Route-stop operations
export async function addStopToRoute(stopId: string): Promise<void> {
  if (!selected.routeId) return;
  
  try {
    // Add to sequence first (import from functions.svelte.ts)
    const index = sequence.indexOf(stopId);
    if (index === -1) {
      sequence.push(stopId);
    }
    
    const routeStops = sequence.map((id, index) => ({
      linkId: `${selected.routeId}_${id}`,
      routeId: selected.routeId!,
      stopId: id,
      sequence: index + 1
    }));
    
    await savePTRouteStops(appState.processId, routeStops);
    
  } catch (error) {
    throw new Error(`Failed to add stop to route: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function removeStopFromRoute(stopId: string): Promise<void> {  
  if (!selected.routeId) return;
  
  try {
    // Remove from sequence first (import from functions.svelte.ts)
    const index = sequence.indexOf(stopId);
    if (index > -1) {
      sequence.splice(index, 1);
    }
    
    const routeStops = sequence.map((id, index) => ({
      linkId: `${selected.routeId}_${id}`,
      routeId: selected.routeId!,
      stopId: id,
      sequence: index + 1
    }));
    
    await savePTRouteStops(appState.processId, routeStops);
    
  } catch (error) {
    throw new Error(`Failed to remove stop from route: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

// Delete operations (API calls)
export async function deleteLineFromAPI(lineId: string): Promise<void> {
  try {
    await deletePTLinesAPI(appState.processId, [lineId]);
    deleteLine(lineId); // Update local state
  } catch (error) {
    throw new Error(`Failed to delete line: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function deleteRouteFromAPI(routeId: string): Promise<void> {
  try {
    await deletePTRoutesAPI(appState.processId, [routeId]);
    deleteRoute(routeId); // Update local state
  } catch (error) {
    throw new Error(`Failed to delete route: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function deleteStopFromAPI(stopId: string): Promise<void> {
  try {
    await deletePTStopsAPI(appState.processId, [stopId]);
    deleteStop(stopId); // Update local state
  } catch (error) {
    throw new Error(`Failed to delete stop: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

export async function deleteDepartureFromAPI(departureId: string): Promise<void> {
  try {
    await deletePTDeparturesAPI(appState.processId, [departureId]);
    deleteDeparture(departureId); // Update local state
  } catch (error) {
    throw new Error(`Failed to delete departure: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}