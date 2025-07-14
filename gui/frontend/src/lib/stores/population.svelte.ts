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
  // Pagination state
  currentPage: number;
  pageSize: number;
  totalPages: number;
  totalPersons: number;
  isLoadingPage: boolean;
  pageCache: Map<number, Person[]>;
  cacheTimestamps: Map<number, number>;
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
  currentPlanIndex: 0,
  // Pagination state
  currentPage: 1,
  pageSize: 50,
  totalPages: 1,
  totalPersons: 0,
  isLoadingPage: false,
  pageCache: new Map(),
  cacheTimestamps: new Map()
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

// Pagination helpers
export function setCurrentPage(page: number) {
  populationState.currentPage = page;
}

export function setPageSize(size: number) {
  populationState.pageSize = size;
  populationState.currentPage = 1; // Reset to first page when changing page size
}

export function clearPageCache() {
  populationState.pageCache.clear();
  populationState.cacheTimestamps.clear();
}

// Check if a page is cached and fresh (less than 30 seconds old)
export function isPageCached(page: number): boolean {
  const timestamp = populationState.cacheTimestamps.get(page);
  if (!timestamp) return false;
  
  const age = Date.now() - timestamp;
  return age < 30000; // 30 seconds
}

// LRU cache management - keep only 5 most recent pages
export function updatePageCache(page: number, persons: Person[]) {
  populationState.pageCache.set(page, persons);
  populationState.cacheTimestamps.set(page, Date.now());
  
  // If cache exceeds 5 pages, remove oldest
  if (populationState.pageCache.size > 5) {
    let oldestPage = -1;
    let oldestTime = Date.now();
    
    populationState.cacheTimestamps.forEach((time, pageNum) => {
      if (time < oldestTime && pageNum !== page) {
        oldestTime = time;
        oldestPage = pageNum;
      }
    });
    
    if (oldestPage !== -1) {
      populationState.pageCache.delete(oldestPage);
      populationState.cacheTimestamps.delete(oldestPage);
    }
  }
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