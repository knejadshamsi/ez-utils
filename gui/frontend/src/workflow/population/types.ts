// Population workflow type definitions

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

export interface PersonData {
  id: string;
  lng: number;
  lat: number;
  rawXML: string;
}

export interface PersonUpdate {
  id: string;
  lng: number;
  lat: number;
  rawXML: string;
}

export interface Zone {
  id: string;
  name: string;
  personCount: number;
  geometry?: any;
}

export type PopulationMode = 'NORMAL' | 'DRAWING_ZONE' | 'ADDING_PERSON' | 'VIEWING_SIDEBAR';
export type SidebarInteractionMode = 'NORMAL' | 'DRAGGING_ACTIVITIES' | 'EDITING_ATTRIBUTES';

// Note: PopulationState uses Svelte's reactive collections in the actual implementation
// This is the abstract interface - the concrete implementation uses SvelteMap/SvelteSet
export interface PopulationState {
  mode: PopulationMode;
  sidebarInteraction: SidebarInteractionMode;
  zones: Zone[];
  persons: Map<string, Person>;
  selectedPersonId: string | null;
  selectedZones: Set<string>;
  visiblePersons: Set<string>;
  visibility: {
    persons: boolean;
    plans: boolean;
  };
  currentPlanIndex: number;
  currentPage: number;
  pageSize: number;
  totalPages: number;
}

export interface PopulationFilters {
  ageRange?: [number, number];
  gender?: 'male' | 'female' | 'all';
  activityTypes?: string[];
  spatialBounds?: BoundingBox;
}

export interface PopulationStats {
  totalPersons: number;
  averageAge: number;
  genderDistribution: {
    male: number;
    female: number;
    other: number;
  };
  activityCounts: Record<string, number>;
}

export interface PopulationEditState {
  selectedPersons: Set<string>;
  editMode: 'select' | 'move' | 'delete';
  bulkEditActive: boolean;
}

export interface BoundingBox {
  north: number;
  south: number;
  east: number;
  west: number;
}

export interface PopulationDensity {
  gridCell: string;
  count: number;
  density: number;
  bounds: BoundingBox;
}

export interface PopulationQuery {
  bounds?: BoundingBox;
  limit?: number;
  offset?: number;
  filters?: PopulationFilters;
}