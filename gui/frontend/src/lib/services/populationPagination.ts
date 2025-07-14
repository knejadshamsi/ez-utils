import { GetPopulationPaginated, GetZonesWithCounts, GetAllZones, GetZoneStats } from '@wailsjs/go/gui/App';
import { populationState, updatePageCache, isPageCached, clearPageCache } from '$lib/stores/population.svelte';
import { parsePersonXML } from '$lib/utils/populationXmlParser';
import { changeTracker } from '$lib/changeTracker.svelte';
import type { Person } from '$lib/stores/population.svelte';

// Load a specific page of population data
export async function loadPopulationPage(tableName: string, page: number, forceReload = false) {
  if (!tableName) {
    console.error('No table name provided');
    return;
  }

  // Check cache first unless force reload
  if (!forceReload && isPageCached(page)) {
    const cachedPersons = populationState.pageCache.get(page);
    if (cachedPersons) {
      console.log(`Loading page ${page} from cache`);
      loadPersonsIntoState(cachedPersons);
      populationState.currentPage = page;
      return;
    }
  }

  populationState.isLoadingPage = true;
  
  try {
    // Get selected zones for filtering
    // If no zones are selected, we still pass empty array to get empty result
    const zoneFilter = Array.from(populationState.selectedZones);
    
    // Fetch paginated data
    const response = await GetPopulationPaginated(
      tableName, 
      page, 
      populationState.pageSize, 
      zoneFilter
    );
    
    console.log(`Loaded page ${page}:`, response.persons?.length || 0, 'persons');
    
    // Update pagination state
    populationState.currentPage = response.currentPage;
    populationState.totalPages = response.totalPages;
    populationState.totalPersons = response.totalCount;
    
    // Process and cache persons
    const processedPersons: Person[] = [];
    
    response.persons.forEach((person: any) => {
      const parsedData = parsePersonXML(person.raw_xml);
      
      const processedPerson: Person = {
        id: person.id,
        zoneId: '', // Zone is determined dynamically, not stored
        plans: parsedData.plans || []
      };
      
      processedPersons.push(processedPerson);
    });
    
    // Update cache
    updatePageCache(page, processedPersons);
    
    // Load into state
    loadPersonsIntoState(processedPersons);
    
  } catch (error) {
    console.error('Failed to load population page:', error);
  } finally {
    populationState.isLoadingPage = false;
  }
}

// Load persons into the active state
function loadPersonsIntoState(persons: Person[]) {
  // Clear current persons
  populationState.persons.clear();
  
  // Add new persons
  persons.forEach(person => {
    populationState.persons.set(person.id, person);
    
    // Restore visibility if it was previously set
    if (!populationState.visiblePersons.has(person.id)) {
      populationState.visiblePersons.add(person.id);
    }
  });
}

// Load zones independently of pagination
export async function loadZones(tableName: string) {
  if (!tableName) {
    console.error('No table name provided');
    return;
  }

  try {
    // Get all zones
    const zones = await GetAllZones();
    
    // Get zone statistics
    const zoneStats = await GetZoneStats(tableName);
    
    // Convert to Zone format with person counts
    populationState.zones = (zones || []).map(zone => ({
      id: zone.id,
      name: zone.name,
      personCount: zoneStats[zone.id]?.personCount || 0
    }));
    
    // Select all zones by default
    populationState.zones.forEach(zone => {
      populationState.selectedZones.add(zone.id);
    });
    
    console.log('Loaded zones:', populationState.zones.length);
  } catch (error) {
    console.error('Failed to load zones:', error);
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
  // Clear cache as filtering changes data
  clearPageCache();
  
  // Reset to first page and reload
  populationState.currentPage = 1;
  await loadPopulationPage(tableName, 1, true);
}

// Handle page size change
export async function handlePageSizeChange(newSize: number, tableName: string) {
  populationState.pageSize = newSize;
  
  // Clear cache as page boundaries change
  clearPageCache();
  
  // Reset to first page and reload
  populationState.currentPage = 1;
  await loadPopulationPage(tableName, 1, true);
}