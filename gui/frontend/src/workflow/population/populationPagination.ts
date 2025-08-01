import { 
  GetPopulationPaginated, 
  CreateFilterSession,
  AddZoneToSession,
  RemoveZoneFromSession,
  CloseFilterSession
} from '@wailsjs/go/gui/App';
import { populationState } from './state.svelte';
import { parsePersonXML } from './populationXmlParser';
import { changeTracker } from '$lib/changeTracker.svelte';
import type { Person } from './types';

// Cache for unsaved persons to maintain during zone filtering
const unsavedPersonsCache = new Map<string, Person>();

// Add person to unsaved cache (or remove if person is null)
export function addToUnsavedCache(personId: string, person: Person | null) {
  if (person === null) {
    unsavedPersonsCache.delete(personId);
  } else {
    unsavedPersonsCache.set(personId, person);
  }
}

// Store the current filter session ID
let currentSessionId: string | null = null;

// Load a specific page of population data
export async function loadPopulationPage(tableName: string, page: number) {
  console.log('Loading population page:', page, 'Table:', tableName, 'SessionID:', currentSessionId);
  
  if (!tableName) {
    console.error('No table name provided');
    return;
  }


  
  try {
    // Fetch paginated data - pass sessionId if we have one (from zones)
    const response = await GetPopulationPaginated(
      tableName, 
      page, 
      populationState.pageSize, 
      currentSessionId || '' // Empty string means no filtering
    );
    
    console.log(`Loaded page ${page}:`, response.persons?.length || 0, 'persons');
    console.log('Backend response details:', {
      currentPage: response.currentPage,
      totalPages: response.totalPages,
      totalCount: response.totalCount,
      personsReturned: response.persons?.length || 0
    });
    
    // Update pagination state
    populationState.currentPage = response.currentPage;
    populationState.totalPages = response.totalPages;
    populationState.totalPersons = response.totalCount;
    
    // Process and cache persons
    const processedPersons: Person[] = [];
    
    if (response.persons && response.persons.length > 0) {
      response.persons.forEach((person: any) => {
      const parsedData = parsePersonXML(person.raw_xml);
      
      const processedPerson: Person = {
        id: person.id,
        zoneId: '', // Zone is determined dynamically, not stored
        attributes: parsedData.attributes || [],
        plans: parsedData.plans || []
      };
      
      processedPersons.push(processedPerson);
    });
    }
    
    
    console.log('Processed persons count:', processedPersons.length);
    
    // Load into state
    loadPersonsIntoState(processedPersons);
    
  } catch (error) {
    console.error('Failed to load population page:', error);
  }
}

// Load persons into the active state
function loadPersonsIntoState(persons: Person[]) {
  console.log('Loading persons into state. Backend count:', persons.length, 'Cache count:', unsavedPersonsCache.size);
  
  // First, remove all backend persons (those NOT in cache)
  const personsToRemove: string[] = [];
  populationState.persons.forEach((person, id) => {
    if (!unsavedPersonsCache.has(id)) {
      personsToRemove.push(id);
    }
  });
  
  // Remove only backend persons
  personsToRemove.forEach(id => {
    populationState.persons.delete(id);
    populationState.visiblePersons.delete(id);
  });
  
  console.log('Removed', personsToRemove.length, 'backend persons');
  
  // Add new filtered persons from backend
  persons.forEach(person => {
    // Don't overwrite cached versions
    if (!unsavedPersonsCache.has(person.id)) {
      populationState.persons.set(person.id, person);
      
      // Restore visibility if it was previously set
      if (!populationState.visiblePersons.has(person.id)) {
        populationState.visiblePersons.add(person.id);
      }
    }
  });
  
  console.log('Final persons count in state:', populationState.persons.size);
}

// Clear unsaved cache after successful sync
export function clearUnsavedCache() {
  unsavedPersonsCache.clear();
}

// Initialize filter session for the table
export async function initializeFilterSession(tableName: string) {
  if (!tableName) {
    console.error('No table name provided');
    return;
  }

  try {
    // Close any existing session
    if (currentSessionId) {
      await CloseFilterSession(currentSessionId);
      currentSessionId = null;
    }
    
    // Create new filter session
    currentSessionId = await CreateFilterSession(tableName);
    console.log('Created filter session:', currentSessionId);
    
  } catch (error) {
    console.error('Failed to initialize filter session:', error);
    currentSessionId = null;
  }
}

// Add a zone to the filter session
export async function addZoneToFilter(
  zoneId: string, 
  polygon: [number, number][],
  tableName: string
) {
  console.log('Adding zone to filter:', zoneId, 'SessionID:', currentSessionId);
  
  if (!currentSessionId || !tableName) {
    console.error('No active session or table');
    return;
  }
  
  try {
    // Send polygon to backend - bbox calculation happens there
    await AddZoneToSession(
      currentSessionId,
      zoneId,
      polygon.map(p => ({ x: p[0], y: p[1] }))
    );
    
    console.log('Zone added to backend, reloading population data...');
    // Reload
    await loadPopulationPage(tableName, 1);
    
  } catch (error) {
    console.error('Failed to add zone to filter:', error);
  }
}

// Remove a zone from the filter session
export async function removeZoneFromFilter(zoneId: string, tableName: string) {
  if (!currentSessionId || !tableName) {
    console.error('No active session or table');
    return;
  }
  
  try {
    await RemoveZoneFromSession(currentSessionId, zoneId);
    
    // Reload
    await loadPopulationPage(tableName, 1);
    
  } catch (error) {
    console.error('Failed to remove zone from filter:', error);
  }
}

// Check if there are unsaved changes for current page
export function hasUnsavedChangesForCurrentPage(): boolean {
  const currentPagePersonIds = Array.from(populationState.persons.keys());
  
  return changeTracker.pendingChanges.some(change => {
    if (change.type === 'population' && change.elementType === 'person') {
      const personId = 'personId' in change ? change.personId : 
                      ('data' in change ? change.data.id : null);
      return personId && currentPagePersonIds.includes(personId);
    }
    return false;
  });
}

// Handle page change with sync check
export async function handlePageChange(
  newPage: number, 
  tableName: string,
  onSyncRequired?: () => Promise<boolean>
) {
  // Validate page number
  if (newPage < 1 || newPage > populationState.totalPages) {
    console.error('Invalid page number:', newPage);
    return;
  }
  
  // Check for unsaved changes
  if (hasUnsavedChangesForCurrentPage()) {
    if (onSyncRequired) {
      const shouldProceed = await onSyncRequired();
      if (!shouldProceed) {
        return; // User cancelled
      }
    }
  }
  
  // Load the new page
  await loadPopulationPage(tableName, newPage);
}

// Handle zone filter changes
export async function handleZoneFilterChange(tableName: string) {
  // Reset to first page and reload
  populationState.currentPage = 1;
  await loadPopulationPage(tableName, 1);
}

// Handle page size change
export async function handlePageSizeChange(newSize: number, tableName: string) {
  populationState.pageSize = newSize;
  
  // Reset to first page and reload
  populationState.currentPage = 1;
  await loadPopulationPage(tableName, 1);
}