<script lang="ts">
  import { Button, Checkbox } from 'flowbite-svelte';
  import { PlusOutline, TrashBinOutline, EditOutline } from 'flowbite-svelte-icons';
  import { populationState } from '@workflow/population/state.svelte';
  import { toggleZone, toggleVisibility, selectPerson } from '@workflow/population/functions.svelte';
  import type { Person } from '@workflow/population/types';
  import { appState } from '$lib/stores/app.svelte';
  import { trackPersonChange } from '@workflow/population/populationChangeTracking';
  import { loadPopulationPage, handlePageChange, handleZoneFilterChange, initializeFilterSession, addZoneToFilter, removeZoneFromFilter } from '@workflow/population/populationPagination';
  import { syncChanges } from '$lib/syncManager';
  import { mapState } from '../../map/mapState.svelte';
  import { startPolygonDrawing, stopPolygonDrawing, enableZoneEditing, disableZoneEditing, deleteZone as deleteMapZone } from '../../map/polygonDrawing';
  import { updateConnectedDots } from '../../map/updateConnectedDots';
  import { ValidateZoneBeforeAdd } from '@wailsjs/go/gui/App';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import type * as L from 'leaflet';
  
  let editingZoneId = $state<string | null>(null);

    $effect(() => {
    if (appState.processId && appState.processId !== 0 && appState.display === 'LOADING') {
      loadPopulationData();
    }
  });
  
  async function loadPopulationData() {
    try {
            const tableName = `population_data_${appState.processId}`;
            await loadPopulationPage(tableName, 1);
      
    } catch (error) {
      console.error('Failed to load population data:', error);
    }
  }

  async function handleAddNewPerson() {
    if (!mapState.map) return;
    
        populationState.mode = 'ADDING_PERSON';
    appState.secondarySidebar = 'HIDDEN';
    
        const clickHandler = (e: L.LeafletMouseEvent) => {
            populationState.mode = 'NORMAL';
      
            const timestamp = Date.now();
      const newId = `person_${timestamp}`;
      const activityId = `activity_${timestamp}_0`;
      
            const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
            const newPerson: Person = {
        id: newId,
        zoneId: populationState.zones[0]?.id || 'default',
        attributes: [],         plans: [{
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
      
            populationState.persons.set(newId, newPerson);
      
            populationState.visiblePersons.add(newId);
      
            trackPersonChange(newPerson, 'create');
      
            selectPerson(newId);
      
            appState.secondarySidebar = 'EXPANDED';
      
            updateConnectedDots();
      
            populationState.totalPersons++;
      
            populationState.totalPages = Math.ceil(populationState.totalPersons / populationState.pageSize);
      
            mapState.map.off('click', clickHandler);
    };
    
        mapState.map.on('click', clickHandler);
        
        // Store handler for cleanup
        (window as any).__addPersonHandler = clickHandler;
  }
  
  // Clean up event handlers on component destroy
  $effect(() => {
    return () => {
      // Clean up any existing click handler when component unmounts
      if (mapState.map && (window as any).__addPersonHandler) {
        mapState.map.off('click', (window as any).__addPersonHandler);
        delete (window as any).__addPersonHandler;
      }
    };
  });

  function handlePersonClick(personId: string) {
        if (populationState.selectedPersonId === personId) {
      selectPerson(null);
      appState.secondarySidebar = 'HIDDEN';
            updateConnectedDots();
    } else {
      selectPerson(personId);
      appState.secondarySidebar = 'EXPANDED';
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
    if (!mapState.map || !appState.processId) return;
    
    populationState.mode = 'DRAWING_ZONE';
    
    startPolygonDrawing(mapState.map, async (zoneId: string, coordinates: [number, number][]) => {
      try {
        // Create zone object - remove the last point (closing point) for backend
        const polygonPoints = coordinates.slice(0, -1).map(coord => ({ 
          x: coord[0], 
          y: coord[1] 
        }));
        
        const tableName = `population_data_${appState.processId}`;
        
        // Validate zone before adding (max 1000 persons per zone)
        const maxThreshold = 1000;
        const validation = await ValidateZoneBeforeAdd(tableName, polygonPoints, maxThreshold);
        
        if (!validation.valid) {
          populationState.mode = 'NORMAL';
          // Remove the drawn polygon
          deleteMapZone(zoneId);
          alert(validation.error || 'Zone validation failed');
          return;
        }
        
        const zone = {
          id: zoneId,
          name: `Zone ${populationState.zones.length + 1}`,
          polygon: polygonPoints,
          boundingBox: null // Will be calculated by backend
        };
        
        // Initialize filter session if this is the first zone
        if (populationState.zones.length === 0) {
          await initializeFilterSession(tableName);
        }
        
        // Add zone to filter session
        await addZoneToFilter(
          zoneId,
          coordinates, // [lng, lat] pairs
          tableName
        );
        
        // Add zone to frontend state (local only)
        const newZone: Zone = {
          id: zoneId,
          name: zone.name,
          personCount: validation.count || 0,
          geometry: coordinates
        };
        
        populationState.zones.push(newZone);
        
        // Select the new zone
        populationState.selectedZones.add(zoneId);
        
        // Force transition to edit mode for the newly created zone
        populationState.mode = 'NORMAL';
        enableZoneEditing(zoneId, handleZoneUpdate);
        editingZoneId = zoneId;
      } catch (error) {
        console.error('Failed to create zone:', error);
        populationState.mode = 'NORMAL';
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
    populationState.mode = 'NORMAL';
  }
  
  async function handleZoneUpdate(zoneId: string, newCoordinates: [number, number][]) {
    if (!appState.processId) return;
    
    const zone = populationState.zones.find(z => z.id === zoneId);
    if (!zone) return;
    
    const tableName = `population_data_${appState.processId}`;
    
    try {
      // Validate the new zone boundaries
      const polygonPoints = newCoordinates.slice(0, -1).map(coord => ({ 
        x: coord[0], 
        y: coord[1] 
      }));
      
      const validation = await ValidateZoneBeforeAdd(tableName, polygonPoints, 1000);
      
      if (!validation.valid) {
        alert(validation.error || 'Zone validation failed');
        // TODO: Revert polygon to original shape
        return;
      }
      
      // Remove the old zone from filter
      await removeZoneFromFilter(zoneId, tableName);
      
      // Add it back with new coordinates
      await addZoneToFilter(zoneId, newCoordinates, tableName);
      
      // Update local zone data
      zone.geometry = newCoordinates;
      zone.personCount = validation.count || 0;
      
    } catch (error) {
      console.error('Failed to update zone:', error);
      alert('Failed to update zone. Please try again.');
    }
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
      enableZoneEditing(zoneId, handleZoneUpdate);
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

  // Get persons to display - backend handles all filtering
  const displayPersons = $derived(Array.from(populationState.persons.values()));
  
  // Check if a person has unsaved changes
  function isPersonUnsaved(personId: string): boolean {
    return changeTracker.pendingChanges.some(change => 
      change.type === 'population' && 
      change.elementType === 'person' &&
      ((change.action === 'add' && 'data' in change && change.data.id === personId) ||
       (change.action === 'update' && 'personId' in change && change.personId === personId))
    );
  }
  
  // Count unsaved new persons
  const unsavedNewPersonsCount = $derived(
    changeTracker.pendingChanges.filter(change => 
      change.type === 'population' && 
      change.elementType === 'person' &&
      change.action === 'add'
    ).length
  );
  
  // Handle zone toggle with pagination refresh
  async function handleToggleZone(zoneId: string) {
    const zone = populationState.zones.find(z => z.id === zoneId);
    if (!zone) return;
    
    const tableName = `population_data_${appState.processId}`;
    if (populationState.selectedZones.has(zoneId)) {
      // Remove zone
      populationState.selectedZones.delete(zoneId);
      await removeZoneFromFilter(zoneId, tableName);
    } else {
      // Add zone
      populationState.selectedZones.add(zoneId);
      await addZoneToFilter(zoneId, zone.geometry, tableName);
    }
  }
  
  // Handle page navigation
  async function handlePageNavigation(page: number) {
    if (!appState.processId) return;
    
    const tableName = `population_data_${appState.processId}`;
    await handlePageChange(page, tableName, async () => {
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
      disabled={populationState.mode === 'ADDING_PERSON'}
    >
      <PlusOutline class="w-4 h-4 mr-2" />
      {populationState.mode === 'ADDING_PERSON' ? 'Click map to place person...' : 'Add New Person'}
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
        Zones {populationState.mode === 'DRAWING_ZONE' ? '(Drawing...)' : ''}
      </h3>
      {#if populationState.mode === 'DRAWING_ZONE'}
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
        {#if unsavedNewPersonsCount > 0}
          <span class="text-yellow-500">({unsavedNewPersonsCount} unsaved)</span>
        {/if}
      </span>
    </div>
    
    <div class="flex-1 overflow-y-auto pr-2">
      <div class="space-y-1">
        {#each displayPersons as person}
            <div class="group flex items-center gap-2 rounded p-1 hover:bg-gray-700
                        {populationState.selectedPersonId === person.id ? 'bg-blue-900 hover:bg-blue-900' : ''}
                        {isPersonUnsaved(person.id) ? 'border-l-2 border-yellow-500' : ''}">
              <Checkbox 
                checked={populationState.visiblePersons.has(person.id)}
                onchange={() => togglePersonVisibility(person.id)}
                onclick={(e: Event) => e.stopPropagation()}
              />
              <button
                class="flex-1 text-left px-2 py-1 text-sm flex items-center gap-2
                       {populationState.selectedPersonId === person.id ? 'text-blue-200' : 'text-white'}"
                onclick={() => handlePersonClick(person.id)}
              >
                {person.id}
                {#if isPersonUnsaved(person.id)}
                  <span class="text-xs text-yellow-500" title="Unsaved changes">⚠️</span>
                {/if}
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
      </div>
    </div>
    
    <!-- Pagination Controls -->
    {#if populationState.totalPersons && populationState.totalPersons > populationState.pageSize}
      <div class="mt-4 pt-4 border-t border-gray-600">
        <div class="flex items-center justify-between mb-2">
          <Button 
            size="xs" 
            color="alternative"
            disabled={populationState.currentPage === 1}
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
            disabled={populationState.currentPage === populationState.totalPages}
            onclick={() => handlePageNavigation(populationState.currentPage + 1)}
          >
            Next
          </Button>
        </div>
        
        <div class="text-center">
          <span class="text-xs text-gray-400">
            Showing {displayPersons.length} of {populationState.totalPersons + unsavedNewPersonsCount} persons
          </span>
        </div>
      </div>
    {/if}
  </section>
</div>