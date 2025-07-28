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
    location: [stop.lng, stop.lat],
    telemetryId: ''
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