import type { gui } from '../../../../wailsjs/go/models';

// Map backend models to frontend models

export interface PTLineExtended extends gui.PTLine {
  name: string;
  number: string;
  agencyId: string;
  telemetryId: string;
}

export interface PTRouteExtended extends gui.PTRoute {
  direction: string;
  telemetryId: string;
  lineId: string; // Map from line_id
}

export interface PTStopExtended extends gui.PTStop {
  location: [number, number];
  telemetryId: string;
}

export interface PTRouteStopExtended extends gui.PTRouteStop {
  routeId: string; // Map from route_id
  stopId: string; // Map from stop_ref_id
  sequence: number; // Map from stop_order
  arrivalOffset: number; // Parse from arrival_offset string
  dwellTime: number; // Calculate from arrival_offset and departure_offset
}

export interface PTDepartureExtended extends gui.PTDeparture {
  routeId: string; // Map from route_id
  departureTime: string; // Map from departure_time
  vehicle_id?: string;
}

// Helper functions to convert between backend and frontend models

export function mapPTLine(line: gui.PTLine): PTLineExtended {
  // Parse name, number from raw_xml if available
  // For now, use defaults
  return {
    ...line,
    name: `Line ${line.id}`,
    number: '',
    agencyId: '',
    telemetryId: ''
  };
}

export function mapPTRoute(route: gui.PTRoute): PTRouteExtended {
  return {
    ...route,
    direction: 'Unknown Direction',
    telemetryId: '',
    lineId: route.line_id
  };
}

export function mapPTStop(stop: gui.PTStop): PTStopExtended {
  return {
    ...stop,
    location: [stop.x, stop.y],
    telemetryId: ''
  };
}

export function mapPTRouteStop(routeStop: gui.PTRouteStop): PTRouteStopExtended {
  const arrivalMinutes = parseTimeOffset(routeStop.arrival_offset);
  const departureMinutes = parseTimeOffset(routeStop.departure_offset);
  
  return {
    ...routeStop,
    routeId: routeStop.route_id,
    stopId: routeStop.stop_ref_id,
    sequence: routeStop.stop_order,
    arrivalOffset: arrivalMinutes,
    dwellTime: departureMinutes - arrivalMinutes
  };
}

export function mapPTDeparture(departure: gui.PTDeparture): PTDepartureExtended {
  return {
    ...departure,
    routeId: departure.route_id,
    departureTime: departure.departure_time
  };
}

function parseTimeOffset(offset: string): number {
  // Parse PT offset format (e.g., "PT5M" -> 5 minutes)
  const match = offset.match(/PT(\d+)M/);
  return match ? parseInt(match[1]) : 0;
}