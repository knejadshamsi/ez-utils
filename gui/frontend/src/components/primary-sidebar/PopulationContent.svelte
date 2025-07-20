<script lang="ts">
  console.log('[PopulationContent] Component instantiated');
  import { Button, Checkbox } from 'flowbite-svelte';
  import { PlusOutline, TrashBinOutline, EditOutline } from 'flowbite-svelte-icons';
  import { 
    populationState, 
    toggleZone, 
    toggleVisibility, 
    selectPerson,
    getPersonsInSelectedZones,
    type Person
  } from '$lib/stores/population.svelte';
  import { appState, editingSession } from '$lib/stores/app.svelte.ts';
  import { onMount } from 'svelte';
  import { trackPersonChange } from '$lib/utils/populationChangeTracking';
  import { loadPopulationPage, handlePageChange, handleZoneFilterChange, initializeFilterSession, addZoneToFilter, removeZoneFromFilter } from '$lib/services/populationPagination';
  import { syncChanges } from '$lib/syncManager';
  import { mapState } from '../../map/mapState.svelte';
  import { startPolygonDrawing, stopPolygonDrawing, enableZoneEditing, disableZoneEditing, deleteZone as deleteMapZone } from '../../map/polygonDrawing';
  import { updateConnectedDots } from '../../map/updateConnectedDots';
  import { GetPersonsInPolygon } from '@wailsjs/go/gui/App';
  import type * as L from 'leaflet';
  
  let editingZoneId = $state<string | null>(null);
  
  console.log('[PopulationContent] After imports - commandArgs:', commandArgs);
  console.log('[PopulationContent] After imports - editingSession:', editingSession);

  // Load population data from backend
  onMount(async () => {
    console.log('PopulationContent onMount - editingSession.tableName:', editingSession.tableName);
    if (editingSession.tableName) {
      try {
        // Initialize filter session
        await initializeFilterSession(editingSession.tableName);
        
        // Load first page of population data (no zones initially)
        await loadPopulationPage(editingSession.tableName, 1);
        
      } catch (error) {
        console.error('Failed to load population data:', error);
      }
    } else {
      console.log('No editingSession.tableName available');
    }
  });

  async function handleAddNewPerson() {
    if (!mapState.map) return;
    
    // Set a flag to indicate we're waiting for a click
    populationState.isSelectingActivityLocation = true;
    appState.secondarySidebar = 'HIDDEN';
    
    // Create a temporary handler for the map click
    const clickHandler = (e: L.LeafletMouseEvent) => {
      // Reset the flag
      populationState.isSelectingActivityLocation = false;
      
      // Generate new person ID
      const timestamp = Date.now();
      const newId = `person_${timestamp}`;
      const activityId = `activity_${timestamp}_0`;
      
      // Get click coordinates
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
      // Create new person with Plan 1 and home activity at clicked location
      const newPerson: Person = {
        id: newId,
        zoneId: populationState.zones[0]?.id || 'default',
        plans: [{
          id: 1,
          activities: [{
            id: activityId,
            type: 'home',
            location: coords,
            startTime: '08:00',
            endTime: '08:30'
          }],
          legs: []
        }]
      };
      
      // Add to store
      populationState.persons.set(newId, newPerson);
      
      // Make visible by default
      populationState.visiblePersons.add(newId);
      
      // Track change
      trackPersonChange(newPerson, 'create');
      
      // Select the new person
      selectPerson(newId);
      
      // Open secondary sidebar
      appState.secondarySidebar = 'EXPANDED';
      
      // Update connected dots for new person
      updateConnectedDots();
      
      // Update total count
      populationState.totalPersons++;
      
      // Recalculate total pages if needed
      populationState.totalPages = Math.ceil(populationState.totalPersons / populationState.pageSize);
      
      // Remove the click handler
      mapState.map.off('click', clickHandler);
    };
    
    // Add click handler to map
    mapState.map.on('click', clickHandler);
  }

  function handlePersonClick(personId: string) {
    // Toggle behavior: if clicking the same person, deselect and hide sidebar
    if (populationState.selectedPersonId === personId) {
      selectPerson(null);
      appState.secondarySidebar = 'HIDDEN';
      // Clear connected dots when deselecting
      updateConnectedDots();
    } else {
      selectPerson(personId);
      appState.secondarySidebar = 'EXPANDED';
      // Update connected dots for selected person
      updateConnectedDots();
    }
  }

  function handleDeleteZone(zoneId: string) {
    // Remove zone from selected zones
    populationState.selectedZones.delete(zoneId);
    
    // Remove all persons in this zone
    const personsToDelete = Array.from(populationState.persons.values())
      .filter(person => person.zoneId === zoneId);
    
    personsToDelete.forEach(person => {
      populationState.persons.delete(person.id);
      trackPersonChange(person, 'delete');
    });
    
    // Remove zone from zones array
    populationState.zones = populationState.zones.filter(zone => zone.id !== zoneId);
    
    // Remove from map
    deleteMapZone(zoneId);
    
    // Stop editing if this zone was being edited
    if (editingZoneId === zoneId) {
      editingZoneId = null;
    }
  }

  async function handleDeletePerson(personId: string, event: Event) {
    event.stopPropagation(); // Prevent triggering person selection
    
    const person = populationState.persons.get(personId);
    if (person) {
      // Track deletion
      trackPersonChange(person, 'delete');
      
      // Remove from store
      populationState.persons.delete(personId);
      
      // Update zone count
      const zone = populationState.zones.find(z => z.id === person.zoneId);
      if (zone) {
        zone.personCount = Math.max(0, zone.personCount - 1);
      }
      
      // Clear selection if this person was selected
      if (populationState.selectedPersonId === personId) {
        populationState.selectedPersonId = null;
        appState.secondarySidebar = 'HIDDEN';
      }
      
      // Update total count
      populationState.totalPersons--;
      
      // Recalculate total pages
      populationState.totalPages = Math.ceil(populationState.totalPersons / populationState.pageSize);
      
      // If current page is now empty and not the first page, go to previous page
      if (populationState.persons.size === 0 && populationState.currentPage > 1) {
        await handlePageNavigation(populationState.currentPage - 1);
      }
    }
  }

  async function handleAddNewZone() {
    if (!mapState.map || !editingSession.tableName) return;
    
    populationState.isDrawingZone = true;
    
    startPolygonDrawing(mapState.map, async (zoneId: string, coordinates: [number, number][]) => {
      try {
        // Create zone object - remove the last point (closing point) for backend
        const polygonPoints = coordinates.slice(0, -1).map(coord => ({ 
          x: coord[0], 
          y: coord[1] 
        }));
        
        const zone = {
          id: zoneId,
          name: `Zone ${populationState.zones.length + 1}`,
          polygon: polygonPoints,
          boundingBox: null // Will be calculated by backend
        };
        
        console.log('Zone object:', zone);
        console.log('Polygon points count:', polygonPoints.length);
        
        console.log('Creating zone with coordinates:', coordinates);
        
        // Add zone to filter session
        await addZoneToFilter(
          zoneId,
          coordinates, // [lng, lat] pairs
          editingSession.tableName
        );
        
        // Add zone to frontend state (local only)
        const newZone = {
          id: zoneId,
          name: zone.name,
          polygon: coordinates,
          color: '#' + Math.floor(Math.random()*16777215).toString(16) // Random color
        };
        
        console.log('Adding zone to state:', newZone);
        populationState.zones.push(newZone);
        
        // Select the new zone
        populationState.selectedZones.add(zoneId);
        
        populationState.isDrawingZone = false;
      } catch (error) {
        console.error('Failed to create zone:', error);
        populationState.isDrawingZone = false;
        // Remove the drawn polygon
        deleteMapZone(zoneId);
        alert('Failed to create zone. Please try again.');
      }
    });
  }

  function handleCancelZoneDrawing() {
    if (mapState.map) {
      stopPolygonDrawing();
    }
    populationState.isDrawingZone = false;
  }
  
  function handleEditZone(zoneId: string) {
    if (editingZoneId === zoneId) {
      // Stop editing
      disableZoneEditing(zoneId);
      editingZoneId = null;
    } else {
      // Stop editing previous zone
      if (editingZoneId) {
        disableZoneEditing(editingZoneId);
      }
      // Start editing this zone
      enableZoneEditing(zoneId);
      editingZoneId = zoneId;
    }
  }

  function togglePersonVisibility(personId: string) {
    if (populationState.visiblePersons.has(personId)) {
      populationState.visiblePersons.delete(personId);
    } else {
      populationState.visiblePersons.add(personId);
    }
  }

  // Get persons to display based on selected zones
  const displayPersons = $derived(getPersonsInSelectedZones());
  
  // Handle zone toggle with pagination refresh
  async function handleToggleZone(zoneId: string) {
    const zone = populationState.zones.find(z => z.id === zoneId);
    if (!zone) return;
    
    if (populationState.selectedZones.has(zoneId)) {
      // Remove zone
      populationState.selectedZones.delete(zoneId);
      await removeZoneFromFilter(zoneId, editingSession.tableName);
    } else {
      // Add zone
      populationState.selectedZones.add(zoneId);
      await addZoneToFilter(zoneId, zone.polygon, editingSession.tableName);
    }
  }
  
  // Handle page navigation
  async function handlePageNavigation(page: number) {
    if (!editingSession.tableName) return;
    
    await handlePageChange(page, editingSession.tableName, async () => {
      const confirmed = confirm('You have unsaved changes. Save before switching pages?');
      if (confirmed) {
        await syncChanges();
      }
      return confirmed;
    });
  }
</script>

<div class="h-full flex flex-col">
  <!-- Operations Section -->
  <section class="p-4 border-b border-gray-600">
    
    <Button 
      color="primary" 
      size="sm" 
      class="w-full mb-3"
      onclick={handleAddNewPerson}
      disabled={populationState.isSelectingActivityLocation}
    >
      <PlusOutline class="w-4 h-4 mr-2" />
      {populationState.isSelectingActivityLocation ? 'Click map to place person...' : 'Add New Person'}
    </Button>
    
    <div class="flex items-center gap-4 text-sm">
      <label class="flex items-center gap-2 text-white cursor-pointer">
        <Checkbox 
          checked={populationState.visibility.persons}
          onchange={() => toggleVisibility('persons')}
        />
        Show Persons
      </label>
      <label class="flex items-center gap-2 text-white cursor-pointer">
        <Checkbox 
          checked={populationState.visibility.plans}
          onchange={() => toggleVisibility('plans')}
        />
        Plans
      </label>
    </div>
  </section>

  <!-- Zones Section -->
  <section class="p-4 border-b border-gray-600">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wider">
        Zones {populationState.isDrawingZone ? '(Drawing...)' : ''}
      </h3>
      {#if populationState.isDrawingZone}
        <Button 
          size="xs" 
          color="red" 
          onclick={handleCancelZoneDrawing}
        >
          Cancel
        </Button>
      {:else}
        <Button 
          size="xs" 
          color="primary" 
          onclick={handleAddNewZone}
        >
          <PlusOutline class="w-3 h-3" />
        </Button>
      {/if}
    </div>
    
    <div class="space-y-2">
      {#each populationState.zones as zone}
        <div class="group flex items-center gap-2 text-sm hover:bg-gray-700 p-1 rounded">
          <label class="flex items-center gap-2 flex-1 cursor-pointer">
            <Checkbox 
              checked={populationState.selectedZones.has(zone.id)}
              onchange={() => handleToggleZone(zone.id)}
            />
            <span class="flex-1 text-white">{zone.name}</span>
          </label>
          <Button 
            size="xs" 
            color={editingZoneId === zone.id ? 'primary' : 'alternative'}
            onclick={() => handleEditZone(zone.id)}
            class="p-1 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <EditOutline class="w-3 h-3" />
          </Button>
          <Button 
            size="xs" 
            color="red" 
            onclick={() => handleDeleteZone(zone.id)}
            class="p-1 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <TrashBinOutline class="w-3 h-3" />
          </Button>
        </div>
      {/each}
    </div>
  </section>

  <!-- People Section -->
  <section class="flex-1 overflow-hidden flex flex-col p-4">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wider">
        People
      </h3>
      <span class="text-xs text-gray-400">
        {populationState.totalPersons} total
      </span>
    </div>
    
    <div class="flex-1 overflow-y-auto pr-2">
      <div class="space-y-1">
        {#if populationState.isLoadingPage}
          <div class="flex justify-center py-4">
            <div class="text-gray-400">Loading...</div>
          </div>
        {:else}
          {#each displayPersons as person}
            <div class="group flex items-center gap-2 rounded p-1 hover:bg-gray-700
                        {populationState.selectedPersonId === person.id ? 'bg-blue-900 hover:bg-blue-900' : ''}">
              <Checkbox 
                checked={populationState.visiblePersons.has(person.id)}
                onchange={() => togglePersonVisibility(person.id)}
                onclick={(e: Event) => e.stopPropagation()}
              />
              <button
                class="flex-1 text-left px-2 py-1 text-sm
                       {populationState.selectedPersonId === person.id ? 'text-blue-200' : 'text-white'}"
                onclick={() => handlePersonClick(person.id)}
              >
                {person.id}
              </button>
              <Button 
                size="xs" 
                color="red" 
                onclick={(e: Event) => handleDeletePerson(person.id, e)}
                class="p-1 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <TrashBinOutline class="w-3 h-3" />
              </Button>
            </div>
          {/each}
        {/if}
      </div>
    </div>
    
    <!-- Pagination Controls -->
    <div class="mt-4 pt-4 border-t border-gray-600">
      <div class="flex items-center justify-between mb-2">
        <Button 
          size="xs" 
          color="alternative"
          disabled={populationState.currentPage === 1 || populationState.isLoadingPage}
          onclick={() => handlePageNavigation(populationState.currentPage - 1)}
        >
          Previous
        </Button>
        
        <span class="text-sm text-gray-300">
          Page {populationState.currentPage} of {populationState.totalPages}
        </span>
        
        <Button 
          size="xs" 
          color="alternative"
          disabled={populationState.currentPage === populationState.totalPages || populationState.isLoadingPage}
          onclick={() => handlePageNavigation(populationState.currentPage + 1)}
        >
          Next
        </Button>
      </div>
      
      <div class="text-center">
        <span class="text-xs text-gray-400">
          Showing {displayPersons.length} of {populationState.totalPersons} persons
        </span>
      </div>
    </div>
  </section>
</div>