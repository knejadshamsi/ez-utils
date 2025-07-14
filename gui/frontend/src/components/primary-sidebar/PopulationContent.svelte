<script lang="ts">
  import { Button, Checkbox } from 'flowbite-svelte';
  import { PlusOutline, TrashBinOutline } from 'flowbite-svelte-icons';
  import { 
    populationState, 
    toggleZone, 
    toggleVisibility, 
    selectPerson,
    getPersonsInSelectedZones 
  } from '$lib/stores/population.svelte';
  import { appState, editingSession } from '$lib/stores/app.svelte.ts';
  import { onMount } from 'svelte';
  import { trackPersonChange } from '$lib/utils/populationChangeTracking';
  import { loadPopulationPage, loadZones, handlePageChange, handleZoneFilterChange } from '$lib/services/populationPagination';
  import { syncChanges } from '$lib/syncManager';

  // Load population data from backend
  onMount(async () => {
    console.log('PopulationContent onMount - editingSession.tableName:', editingSession.tableName);
    if (editingSession.tableName) {
      try {
        // Load zones first
        await loadZones(editingSession.tableName);
        
        // Load first page of population data
        await loadPopulationPage(editingSession.tableName, 1);
        
      } catch (error) {
        console.error('Failed to load population data:', error);
      }
    } else {
      console.log('No editingSession.tableName available');
    }
  });

  async function handleAddNewPerson() {
    // Generate new person ID
    const timestamp = Date.now();
    const newId = `person_${timestamp}`;
    
    // Create new person
    const newPerson = {
      id: newId,
      zoneId: populationState.zones[0]?.id || 'default',
      plans: [{
        type: 'weekday' as const,
        activities: [],
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
    
    // Update total count
    populationState.totalPersons++;
    
    // Recalculate total pages if needed
    populationState.totalPages = Math.ceil(populationState.totalPersons / populationState.pageSize);
  }

  function handlePersonClick(personId: string) {
    // Toggle behavior: if clicking the same person, deselect and hide sidebar
    if (populationState.selectedPersonId === personId) {
      selectPerson(null);
      appState.secondarySidebar = 'HIDDEN';
    } else {
      selectPerson(personId);
      appState.secondarySidebar = 'EXPANDED';
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

  function handleAddNewZone() {
    // Trigger map drawing mode
    populationState.isDrawingZone = true;
    console.log('Activating map drawing mode for new zone');
  }

  function handleCancelZoneDrawing() {
    // Cancel drawing mode
    populationState.isDrawingZone = false;
    console.log('Cancelled zone drawing');
    
    // Clear any partially drawn features
    if (typeof window !== 'undefined') {
      // Dispatch event to clear editable layer
      window.dispatchEvent(new CustomEvent('clearEditableLayer'));
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
    toggleZone(zoneId);
    // Reload data with new zone filter
    if (editingSession.tableName) {
      await handleZoneFilterChange(editingSession.tableName);
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
    >
      <PlusOutline class="w-4 h-4 mr-2" />
      Add New Person
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
            <span class="text-xs text-gray-300">({zone.personCount} persons)</span>
          </label>
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