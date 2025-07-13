import { SvelteMap, SvelteSet } from 'svelte/reactivity';

export type ActivityType = 'home' | 'work' | 'school' | 'shop' | 'eat' | 'recreation' | 'other';
export type TravelMode = 'car' | 'walk' | 'transit' | 'bike';
export type PlanType = 'weekday' | 'weekend' | 'holiday';

export interface Activity {
  id: string;
  type: ActivityType;
  location: [number, number];
  startTime: string; // "HH:MM"
  endTime: string;   // "HH:MM"
}

export interface Leg {
  fromActivityId: string;
  toActivityId: string;
  mode: TravelMode;
  duration: number; // minutes
}

export interface Plan {
  type: PlanType;
  activities: Activity[];
  legs: Leg[];
}

export interface Person {
  id: string;
  zoneId: string;
  plans: Plan[];
}

export interface Zone {
  id: string;
  name: string;
  personCount: number;
  geometry?: any; // GeoJSON geometry
}

export interface PopulationState {
  zones: Zone[];
  persons: SvelteMap<string, Person>;
  selectedPersonId: string | null;
  selectedZones: SvelteSet<string>;
  visiblePersons: SvelteSet<string>;
  pendingChanges: SvelteMap<string, any>;
  visibility: {
    persons: boolean;
    plans: boolean;
  };
  isDrawingZone: boolean;
  isSelectingActivityLocation: boolean;
  selectingActivityId: string | null;
  currentPlanIndex: number;
}

// Create the reactive population state
export const populationState = $state<PopulationState>({
  zones: [],
  persons: new SvelteMap(),
  selectedPersonId: null,
  selectedZones: new SvelteSet(),
  visiblePersons: new SvelteSet(),
  pendingChanges: new SvelteMap(),
  visibility: {
    persons: true,
    plans: true
  },
  isDrawingZone: false,
  isSelectingActivityLocation: false,
  selectingActivityId: null,
  currentPlanIndex: 0
});

// Helper functions
export function selectPerson(personId: string | null) {
  populationState.selectedPersonId = personId;
}

export function toggleZone(zoneId: string) {
  if (populationState.selectedZones.has(zoneId)) {
    populationState.selectedZones.delete(zoneId);
  } else {
    populationState.selectedZones.add(zoneId);
  }
}

export function toggleVisibility(type: 'persons' | 'plans') {
  populationState.visibility[type] = !populationState.visibility[type];
}

export function getPersonsInSelectedZones(): Person[] {
  if (populationState.selectedZones.size === 0) {
    return Array.from(populationState.persons.values());
  }
  
  return Array.from(populationState.persons.values()).filter(
    person => populationState.selectedZones.has(person.zoneId)
  );
}

// Activity type metadata
export const activityTypeConfig = {
  home: { icon: '🏠', label: 'Home', color: '#10b981' },
  work: { icon: '🏢', label: 'Work', color: '#3b82f6' },
  school: { icon: '🏫', label: 'School', color: '#8b5cf6' },
  shop: { icon: '🛒', label: 'Shop', color: '#f59e0b' },
  eat: { icon: '🍽️', label: 'Eat', color: '#ef4444' },
  recreation: { icon: '🏃', label: 'Recreation', color: '#ec4899' },
  other: { icon: '🚗', label: 'Other', color: '#6b7280' }
};

// Travel mode metadata
export const travelModeConfig = {
  car: { icon: '🚗', label: 'Car' },
  walk: { icon: '🚶', label: 'Walk' },
  transit: { icon: '🚌', label: 'Transit' },
  bike: { icon: '🚲', label: 'Bike' }
};