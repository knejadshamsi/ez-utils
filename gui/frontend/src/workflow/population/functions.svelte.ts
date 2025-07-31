// Population functions
import type { ActivityType, TravelMode } from './types';
import { populationState } from './state.svelte';

// Selection functions
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

// Visibility functions
export function toggleVisibility(type: 'persons' | 'plans') {
  populationState.visibility[type] = !populationState.visibility[type];
}

// Pagination functions
export function setCurrentPage(page: number) {
  populationState.currentPage = page;
}

// Activity and travel mode configurations
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