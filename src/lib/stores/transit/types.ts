export type TransitMode = 'bus' | 'metro' | 'tram';

export interface TransitModeFlags {
  bus: boolean;
  metro: boolean;
  tram: boolean;
}

export type SecondaryTab = 'stops' | 'departures';
export type PrimaryTab = 'lines' | 'stops';

export interface ListedLine {
  id: string;
  name: string | null;
  routeCount: number;
  modes: string[];
}

export interface ListedRoute {
  id: string;
  lineId: string;
  transportMode: string;
  description: string | null;
  stopCount: number;
}

export interface ListedProfileStop {
  sequence: number;
  stopRefId: string;
  stopName: string | null;
  stopLng: number | null;
  stopLat: number | null;
  stopLinkRefId: string | null;
  stopAreaId: string | null;
  stopIsBlocking: boolean | null;
  arrivalOffset: string | null;
  departureOffset: string | null;
  allowBoarding: boolean;
  allowAlighting: boolean;
  awaitDeparture: boolean;
}

export interface ListedPathLink {
  sequence: number;
  linkId: string;
}

export interface ListedDeparture {
  id: string;
  departureTime: string;
  vehicleRefId: string | null;
}

export interface ListedStop {
  id: string;
  name: string | null;
  lng: number;
  lat: number;
  linkRefId: string | null;
  stopAreaId: string | null;
  isBlocking: boolean;
  modes: string[];
}

export interface NearbyNetworkLink {
  id: string;
  fromLng: number;
  fromLat: number;
  toLng: number;
  toLat: number;
}

export interface StopLineUsage {
  lineId: string;
  lineName: string | null;
  routeCount: number;
  modes: string[];
}

export interface StopRouteUsage {
  lineId: string;
  routeId: string;
  transportMode: string;
  description: string | null;
}

export interface StopTransfer {
  otherStopId: string;
  otherStopName: string | null;
  transferTime: number;
  direction: 'from' | 'to';
}

export interface StopDetail {
  linesLoading: boolean;
  linesError: string | null;
  lines: StopLineUsage[];
  routesLoading: boolean;
  routesError: string | null;
  routes: StopRouteUsage[];
  transfersLoading: boolean;
  transfersError: string | null;
  transfers: StopTransfer[];
}

export interface StopsPage {
  items: ListedStop[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ViewportBounds {
  west: number;
  south: number;
  east: number;
  north: number;
}

export interface LinePathPoint {
  routeId: string;
  sequence: number;
  lng: number;
  lat: number;
}

export interface GeoPoint {
  lng: number;
  lat: number;
}

export interface LinkedRoutePath {
  lineId: string;
  routeId: string;
  networkSource: string;
  linkCount: number;
  points: GeoPoint[];
}

export interface LineDeletePreview {
  routes: number;
  profileStops: number;
  pathLinks: number;
  departures: number;
}

export interface RouteDeletePreview {
  profileStops: number;
  pathLinks: number;
  departures: number;
}

export interface StopFacilityDeletePreview {
  stopId: string;
  profileStops: number;
  routes: number;
  transfers: number;
}

export interface EditableProfileStop {
  key: string;
  stopRefId: string;
  stopName: string | null;
  stopLng: number | null;
  stopLat: number | null;
  stopLinkRefId: string | null;
  stopAreaId: string | null;
  stopIsBlocking: boolean | null;
  arrivalOffset: string;
  departureOffset: string;
  allowBoarding: boolean;
  allowAlighting: boolean;
  awaitDeparture: boolean;
}

export interface ProfileStopsDraft {
  lineId: string;
  routeId: string;
  stops: EditableProfileStop[];
  originalSnapshot: string;
  saveState: 'idle' | 'saving' | 'saved' | 'error';
  saveError: string | null;
  pickerOpen: boolean;
  pickerQuery: string;
  pickerResults: ListedStop[];
  pickerLoading: boolean;
  createOpen: boolean;
  createId: string;
  createName: string;
  createLng: string;
  createLat: string;
  createLinkRefId: string;
  createStopAreaId: string;
  createIsBlocking: boolean;
  createNearbyLinks: NearbyNetworkLink[];
  createNearbyLinksLoading: boolean;
  createSubmitting: boolean;
  createError: string | null;
}

export interface DeleteStopFacilityTarget {
  stopId: string;
  preview: StopFacilityDeletePreview | null;
  loadingPreview: boolean;
  submitting: boolean;
  error: string | null;
}

export interface EditStopFacilityTarget {
  originalId: string;
  id: string;
  name: string;
  lng: string;
  lat: string;
  linkRefId: string;
  stopAreaId: string;
  isBlocking: boolean;
  nearbyLinks: NearbyNetworkLink[];
  nearbyLinksLoading: boolean;
  submitting: boolean;
  error: string | null;
}

export interface ListedVehicle {
  id: string;
  vehicleType: string | null;
  referenced: boolean;
}

export interface EditableDeparture {
  key: string;
  id: string;
  departureTime: string;
  vehicleRefId: string;
}

export interface DeparturesDraft {
  lineId: string;
  routeId: string;
  departures: EditableDeparture[];
  originalSnapshot: string;
  saveState: 'idle' | 'saving' | 'saved' | 'error';
  saveError: string | null;
  vehicles: ListedVehicle[];
  vehiclesLoaded: boolean;
}

export interface CreateLineForm {
  open: boolean;
  id: string;
  name: string;
  transportMode: string;
  submitting: boolean;
  error: string | null;
}

export interface CreateRouteForm {
  open: boolean;
  id: string;
  description: string;
  submitting: boolean;
  error: string | null;
}

export interface RenameLineTarget {
  originalId: string;
  id: string;
  name: string;
  transportMode: string;
  submitting: boolean;
  error: string | null;
}

export interface DeleteLineTarget {
  lineId: string;
  preview: LineDeletePreview | null;
  loadingPreview: boolean;
  submitting: boolean;
  error: string | null;
}

export interface RenameRouteTarget {
  lineId: string;
  originalId: string;
  id: string;
  description: string;
  submitting: boolean;
  error: string | null;
}

export interface DeleteRouteTarget {
  lineId: string;
  routeId: string;
  preview: RouteDeletePreview | null;
  loadingPreview: boolean;
  submitting: boolean;
  error: string | null;
}

/** Maps the user-facing mode toggle set to the raw transport_mode strings we expect in data. */
export const MODE_ALIASES: Record<TransitMode, string[]> = {
  bus: ['bus'],
  metro: ['metro', 'subway', 'rail'],
  tram: ['tram', 'light_rail', 'lightrail', 'streetcar'],
};

export function lineMatchesMode(line: ListedLine, flags: TransitModeFlags): boolean {
  const enabled = (Object.keys(flags) as TransitMode[]).filter((m) => flags[m]);
  if (enabled.length === 0) return false;
  if (line.modes.length === 0) return true;
  const lineModesLower = line.modes.map((m) => m.toLowerCase());
  return enabled.some((m) => MODE_ALIASES[m].some((alias) => lineModesLower.includes(alias)));
}

/** Flat list of lowercase transport_mode strings the backend should match against. */
export function resolveEnabledModeAliases(flags: TransitModeFlags): string[] {
  const enabled = (Object.keys(flags) as TransitMode[]).filter((m) => flags[m]);
  const aliases = new Set<string>();
  for (const m of enabled) {
    for (const a of MODE_ALIASES[m]) aliases.add(a);
  }
  return [...aliases];
}

export function extractMessage(err: unknown): string {
  if (typeof err === 'object' && err !== null && 'message' in err) {
    const maybe = (err as { message?: unknown }).message;
    if (typeof maybe === 'string') return maybe;
  }
  return String(err);
}
