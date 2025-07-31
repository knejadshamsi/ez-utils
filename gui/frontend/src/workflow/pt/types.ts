// PT workflow type definitions

export type TransportMode = 'BUS' | 'METRO' | 'TRAM';

export const TRANSPORT_MODES = {
  BUS: { name: 'Bus', icon: '🚌' },
  METRO: { name: 'Metro', icon: '🚇' },
  TRAM: { name: 'Tram', icon: '🚋' },
};

export type Line = {
  id: string;
  name: string;
  mode: TransportMode;
  unsaved: boolean;
};

export type Route = {
  id: string;
  name: string;
  lineId: string;
  unsaved: boolean;
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
  stopType: StopType;
  wheelchairAccessible: AccessibilityStatus;
  timingPoint: boolean;
  attributes?: Record<string, string | number | boolean>;
  unsaved: boolean;
};

export type Departure = {
  id: string;
  routeId: string;
  departureTime: string;
  vehicleRefId?: string;
  unsaved: boolean;
};

export interface RouteStop {
  linkId: string;
  routeId: string;
  stopId: string;
  sequence: number;
}

export interface PTStopUpdate {
  stopName?: string;
  lat?: number;
  lng?: number;
  arrivalOffset?: string;
  departureOffset?: string;
  stopType?: StopType;
  wheelchairAccessible?: AccessibilityStatus;
  timingPoint?: boolean;
}

export interface CreateStopParams {
  routeId: string;
  stopId: string;
  stopName: string;
  lat: number;
  lng: number;
  arrivalOffset: string;
  departureOffset: string;
}

export interface PTData {
  [mode: string]: {
    [lineId: string]: {
      name: string;
      routes: {
        [routeId: string]: {
          name: string;
        };
      };
    };
  };
}

export interface CurrentRouteData {
  stops: Stop[];
  departures: Departure[];
}

export interface PTSelection {
  mode: TransportMode;
  lineId: string | null;
  routeId: string | null;
  stopId: string | null;
}

export type PTEditMode = 'NORMAL' | 'ADDING_STOP' | 'DRAGGING_STOP' | 'EDITING_STOP_ATTRIBUTES' | 'EDITING_DEPARTURES' | 'ADDING_MULTIPLE_STOPS' | 'SELECTING_STOP_LOCATION';

export interface PTSelection {
  mode: TransportMode;
  lineId: string | null;
  routeId: string | null;
  stopId: string | null;
  editMode: PTEditMode;
}