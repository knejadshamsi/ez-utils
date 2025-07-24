import { SvelteMap, SvelteSet } from 'svelte/reactivity';

export type ActivityType = 'home' | 'work' | 'school' | 'shop' | 'eat' | 'recreation' | 'other';
export type TravelMode = "person's choice" | 'car' | 'walk' | 'bus' | 'metro' | 'tram' | 'bike';
export interface Activity {
  id: string;
  type: ActivityType;
  location: [number, number];
  startTime: string;
  endTime: string;
}

export interface Leg {
  fromActivityId: string;
  toActivityId: string;
  mode: TravelMode;
  duration: number;
}

export interface Plan {
  id: number;
  activities: Activity[];
  legs: Leg[];
}

export interface PersonAttribute {
  name: string;
  type: 'java.lang.Integer' | 'java.lang.Boolean' | 'java.lang.String' | 'java.lang.Double';
  value: string | number | boolean;
  included: boolean;
}

export interface Person {
  id: string;
  zoneId: string;
  attributes: PersonAttribute[];
  plans: Plan[];
}

export interface Zone {
  id: string;
  name: string;
  personCount: number;
  geometry?: any;
}
export type PopulationMode = 'NORMAL' | 'DRAWING_ZONE' | 'ADDING_PERSON' | 'VIEWING_SIDEBAR';
export type SidebarInteractionMode = 'NORMAL' | 'DRAGGING_ACTIVITIES' | 'EDITING_ATTRIBUTES';

export interface PopulationState {
  mode: PopulationMode;
  sidebarInteraction: SidebarInteractionMode;
  zones: Zone[];
  persons: SvelteMap<string, Person>;
  selectedPersonId: string | null;
  selectedZones: SvelteSet<string>;
  visiblePersons: SvelteSet<string>;
  visibility: {
    persons: boolean;
    plans: boolean;
  };
  currentPlanIndex: number;
  currentPage: number;
  pageSize: number;
  totalPages: number;
  totalPersons: number | null;
}
export const populationState = $state<PopulationState>({
  mode: 'NORMAL',
  sidebarInteraction: 'NORMAL',
  zones: [],
  persons: new SvelteMap(),
  selectedPersonId: null,
  selectedZones: new SvelteSet(),
  visiblePersons: new SvelteSet(),
  visibility: {
    persons: true,
    plans: true
  },
  currentPlanIndex: 0,
  currentPage: 1,
  pageSize: 50,
  totalPages: 1,
  totalPersons: null
});
export function selectPerson(personId: string) {
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

export function setCurrentPage(page: number) {
  populationState.currentPage = page;
}


export const activityTypeConfig = {
  home: { icon: '🏠', label: 'Home', color: '#10b981' },
  work: { icon: '🏢', label: 'Work', color: '#3b82f6' },
  school: { icon: '🏫', label: 'School', color: '#8b5cf6' },
  shop: { icon: '🛒', label: 'Shop', color: '#f59e0b' },
  eat: { icon: '🍽️', label: 'Eat', color: '#ef4444' },
  recreation: { icon: '🏃', label: 'Recreation', color: '#ec4899' },
  other: { icon: '🚗', label: 'Other', color: '#6b7280' }
};
export const travelModeConfig = {
  "person's choice": { icon: '🧭', label: "Person's Choice" },
  car: { icon: '🚗', label: 'Car' },
  walk: { icon: '🚶', label: 'Walk' },
  bus: { icon: '🚌', label: 'Bus' },
  metro: { icon: '🚇', label: 'Metro' },
  tram: { icon: '🚊', label: 'Tram' },
  bike: { icon: '🚲', label: 'Bike' }
};