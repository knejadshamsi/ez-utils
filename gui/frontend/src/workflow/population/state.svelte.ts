// Population state management
import { SvelteMap, SvelteSet } from 'svelte/reactivity';
import type { 
  ActivityType, TravelMode, Activity, Leg, Plan, PersonAttribute, Person, Zone,
  PopulationMode, SidebarInteractionMode
} from './types';

export const populationState = $state<{
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
}>({
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