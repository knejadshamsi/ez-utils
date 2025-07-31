// PT API Layer - Pure API Functions
import { 
  GetPTLinesByMode, 
  GetPTRoutesByLineID, 
  GetPTStopsByRouteID, 
  GetPTDeparturesByRouteID, 
  GetPTStopsInBounds, 
  GetPTStops, 
  SavePTLines, 
  SavePTRoutes, 
  SavePTStops, 
  SavePTRouteStops, 
  SavePTDepartures, 
  DeletePTLines, 
  DeletePTRoutes, 
  DeletePTStops, 
  DeletePTDepartures 
} from '@wailsjs/go/gui/App';
import type { gui } from '@wailsjs/go/models';
import type { TransportMode, Line, Route, Stop, Departure } from './types';

// Data Loading API Functions
export async function fetchPTLinesByMode(processId: number, mode: TransportMode): Promise<Array<{ id: string; name: string }>> {
  return await GetPTLinesByMode(processId, mode);
}

export async function fetchPTRoutesByLineID(processId: number, lineId: string): Promise<Array<{ id: string; name: string }>> {
  return await GetPTRoutesByLineID(processId, lineId);
}

export async function fetchPTStopsByRouteID(processId: number, routeId: string): Promise<Array<{ stopId: string; sequence: number }>> {
  return await GetPTStopsByRouteID(processId, routeId);
}

export async function fetchPTStops(processId: number, stopIds: string[]): Promise<gui.Stop[]> {
  return await GetPTStops(processId, stopIds);
}

export async function fetchPTStopsInBounds(
  processId: number, 
  mode: TransportMode, 
  bounds: { minLat: number; maxLat: number; minLng: number; maxLng: number }
): Promise<gui.Stop[]> {
  return await GetPTStopsInBounds(processId, mode, bounds);
}

export async function fetchPTDeparturesByRouteID(processId: number, routeId: string): Promise<Array<{ id: string; departureTime: string; vehicleRefId?: string }>> {
  return await GetPTDeparturesByRouteID(processId, routeId);
}

// Save API Functions
export async function savePTLines(processId: number, lines: Array<{ id: string; name: string; type: string }>): Promise<void> {
  await SavePTLines(processId, lines);
}

export async function savePTRoutes(processId: number, routes: Array<{ id: string; name: string; lineId: string }>): Promise<void> {
  await SavePTRoutes(processId, routes);
}

export async function savePTStops(processId: number, stops: Stop[]): Promise<void> {
  await SavePTStops(processId, stops);
}

export async function savePTRouteStops(processId: number, routeStops: Array<{ linkId: string; routeId: string; stopId: string; sequence: number }>): Promise<void> {
  await SavePTRouteStops(processId, routeStops);
}

export async function savePTDepartures(processId: number, departures: Array<Departure & { routeId: string }>): Promise<void> {
  await SavePTDepartures(processId, departures);
}

// Delete API Functions
export async function deletePTLinesAPI(processId: number, lineIds: string[]): Promise<void> {
  await DeletePTLines(processId, lineIds);
}

export async function deletePTRoutesAPI(processId: number, routeIds: string[]): Promise<void> {
  await DeletePTRoutes(processId, routeIds);
}

export async function deletePTStopsAPI(processId: number, stopIds: string[]): Promise<void> {
  await DeletePTStops(processId, stopIds);
}

export async function deletePTDeparturesAPI(processId: number, departureIds: string[]): Promise<void> {
  await DeletePTDepartures(processId, departureIds);
}