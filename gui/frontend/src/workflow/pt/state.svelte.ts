// PT State Management - Reactive Stores Only
import type { 
  TransportMode, Line, Route, Stop, Departure, PTEditMode
} from './types';

// Flat state stores as specified in PRD
export const lines = $state<Record<string, Line>>({});
export const routes = $state<Record<string, Route>>({});
export const stops = $state<Record<string, Stop>>({});
export const departures = $state<Record<string, Departure>>({});
export const sequence = $state<string[]>([]);

// UI state for the PT workflow
export const selected = $state<{
  mode: TransportMode;
  lineId: string | null;
  routeId: string | null;
  stopId: string | null;
  editMode: PTEditMode;
}>({
  mode: 'BUS',
  lineId: null,
  routeId: null,
  stopId: null,
  editMode: 'NORMAL'
});