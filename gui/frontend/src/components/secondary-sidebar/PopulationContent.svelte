<script lang="ts">
  import { Button, Select, Input, CloseButton } from 'flowbite-svelte';
  import { TrashBinOutline, PlusOutline, MapPinOutline } from 'flowbite-svelte-icons';
  import CompactSelect from '../CompactSelect.svelte';
  import { 
    populationState,
    activityTypeConfig,
    travelModeConfig,
    type Activity,
    type Leg,
    type ActivityType,
    type TravelMode
  } from '$lib/stores/population.svelte';
  import { appState } from '$lib/stores/app.svelte.ts';
  import { GetPerson } from '@wailsjs/go/gui/App';
  import { editingSession } from '$lib/stores/app.svelte.ts';
  import { trackActivityChange } from '$lib/utils/populationChangeTracking';
  import { parsePersonXML } from '$lib/utils/populationXmlParser';
  
  // Get selected person
  const selectedPerson = $derived(
    populationState.selectedPersonId 
      ? populationState.persons.get(populationState.selectedPersonId)
      : null
  );
  
  // Current plan - using shared state
  const currentPlan = $derived(selectedPerson?.plans[populationState.currentPlanIndex]);
  
  // Reset plan index when person changes
  $effect(() => {
    if (selectedPerson && populationState.currentPlanIndex >= selectedPerson.plans.length) {
      populationState.currentPlanIndex = 0;
    }
  });
  
  // Load person details when selection changes
  $effect(() => {
    if (populationState.selectedPersonId && editingSession.tableName) {
      loadPersonDetails(populationState.selectedPersonId);
    }
  });
  
  async function loadPersonDetails(personId: string) {
    try {
      const personData = await GetPerson(editingSession.tableName, personId);
      
      // Parse the raw XML to get updated plans
      const parsedData = parsePersonXML(personData.raw_xml);
      const person = populationState.persons.get(personId);
      
      if (person) {
        // Create a new person object with parsed plans to trigger reactivity
        const updatedPerson = {
          ...person,
          plans: parsedData.plans || []
        };
        
        // Ensure at least one plan exists
        if (updatedPerson.plans.length === 0) {
          updatedPerson.plans.push({
            type: 'weekday' as const,
            activities: [],
            legs: []
          });
        }
        
        // Update the Map to trigger reactivity
        populationState.persons.set(personId, updatedPerson);
      }
    } catch (error) {
      console.error('Failed to load person details:', error);
    }
  }
  
  function handleClose() {
    populationState.selectedPersonId = null;
    appState.secondarySidebar = 'HIDDEN';
  }
  
  function addActivity() {
    if (!selectedPerson || !currentPlan) return;
    
    const newActivity: Activity = {
      id: `activity_${Date.now()}`,
      type: 'home',
      location: [0, 0], // Will be set by map click
      startTime: '09:00',
      endTime: '17:00'
    };
    
    // Create a new person object with updated activities
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? { ...plan, activities: [...plan.activities, newActivity] }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Immediately prompt for location selection
    populationState.isSelectingActivityLocation = true;
    populationState.selectingActivityId = newActivity.id;
    
    autoSortActivities();
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function deleteActivity(activityId: string) {
    if (!selectedPerson || !currentPlan) return;
    
    // Create a new person object with filtered activities
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? { ...plan, activities: plan.activities.filter(a => a.id !== activityId) }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    autoSortActivities();
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateActivityTime(activityId: string, field: 'startTime' | 'endTime', value: string) {
    if (!selectedPerson || !currentPlan) return;
    
    // Create a new person object with updated activity time
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? {
              ...plan,
              activities: plan.activities.map(a => 
                a.id === activityId 
                  ? { ...a, [field]: value }
                  : a
              )
            }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    autoSortActivities();
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateActivityType(activityId: string, type: ActivityType) {
    if (!selectedPerson || !currentPlan) return;
    
    // Create a new person object with updated activity type
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? {
              ...plan,
              activities: plan.activities.map(a => 
                a.id === activityId 
                  ? { ...a, type }
                  : a
              )
            }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateLegMode(leg: Leg, mode: TravelMode) {
    if (!selectedPerson || !currentPlan) return;
    
    // Create a new person object with updated leg mode
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? {
              ...plan,
              legs: plan.legs.map(l => 
                l.fromActivityId === leg.fromActivityId && l.toActivityId === leg.toActivityId
                  ? { ...l, mode }
                  : l
              )
            }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateLegDuration(leg: Leg, duration: number) {
    if (!selectedPerson || !currentPlan) return;
    
    // Create a new person object with updated leg duration
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? {
              ...plan,
              legs: plan.legs.map(l => 
                l.fromActivityId === leg.fromActivityId && l.toActivityId === leg.toActivityId
                  ? { ...l, duration }
                  : l
              )
            }
          : plan
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Track change
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function autoSortActivities() {
    if (!selectedPerson || !currentPlan) return;
    
    // Get current person from the Map to ensure we have the latest data
    const currentPerson = populationState.persons.get(selectedPerson.id);
    if (!currentPerson) return;
    
    const plan = currentPerson.plans[populationState.currentPlanIndex];
    if (!plan) return;
    
    // Sort activities by start time
    const sortedActivities = [...plan.activities].sort((a, b) => {
      const [hoursA = 0, minutesA = 0] = a.startTime.split(':').map(Number);
      const [hoursB = 0, minutesB = 0] = b.startTime.split(':').map(Number);
      return (hoursA * 60 + minutesA) - (hoursB * 60 + minutesB);
    });
    
    // Preserve existing leg data where possible
    const existingLegs = new Map(plan.legs.map(leg => 
      [`${leg.fromActivityId}-${leg.toActivityId}`, leg]
    ));
    
    // Regenerate legs
    const newLegs: Leg[] = [];
    for (let i = 0; i < sortedActivities.length - 1; i++) {
      const fromId = sortedActivities[i].id;
      const toId = sortedActivities[i + 1].id;
      const existingLeg = existingLegs.get(`${fromId}-${toId}`);
      
      newLegs.push({
        fromActivityId: fromId,
        toActivityId: toId,
        mode: existingLeg?.mode || "person's choice",
        duration: existingLeg?.duration || 30
      });
    }
    
    // Create a new person object with sorted activities and regenerated legs
    const updatedPerson = {
      ...currentPerson,
      plans: currentPerson.plans.map((p, index) => 
        index === populationState.currentPlanIndex 
          ? { ...p, activities: sortedActivities, legs: newLegs }
          : p
      )
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(currentPerson.id, updatedPerson);
  }
  
  function addPlan() {
    if (!selectedPerson) return;
    
    // Create a new person object with added plan
    const updatedPerson = {
      ...selectedPerson,
      plans: [...selectedPerson.plans, {
        type: 'weekend',
        activities: [],
        legs: []
      }]
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Update plan index
    populationState.currentPlanIndex = updatedPerson.plans.length - 1;
  }
  
  function deletePlan() {
    if (!selectedPerson || selectedPerson.plans.length <= 1) return;
    
    // Create a new person object with deleted plan
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.filter((_, index) => index !== populationState.currentPlanIndex)
    };
    
    // Update the Map to trigger reactivity
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    // Update plan index
    populationState.currentPlanIndex = Math.max(0, populationState.currentPlanIndex - 1);
  }
  
  function selectActivityLocation(activityId: string) {
    populationState.isSelectingActivityLocation = true;
    populationState.selectingActivityId = activityId;
  }
  
  function formatLocation(location: [number, number]): string {
    if (location[0] === 0 && location[1] === 0) {
      return 'Not set';
    }
    return `${location[1].toFixed(4)}, ${location[0].toFixed(4)}`;
  }
  
  // Activity type options
  const activityTypeOptions = Object.entries(activityTypeConfig).map(([value, config]) => ({
    value,
    name: `${config.icon} ${config.label}`
  }));
  
  // Travel mode options
  const travelModeOptions = Object.entries(travelModeConfig).map(([value, config]) => ({
    value,
    name: `${config.icon} ${config.label}`
  }));
  
</script>

{#if selectedPerson}
  <div class="h-full flex flex-col bg-gray-800">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b border-gray-600">
      <h2 class="text-lg font-semibold text-white">{selectedPerson.id}</h2>
      <CloseButton onclick={handleClose} class="text-gray-400 hover:text-white" />
    </div>
    
    <!-- Plan selector -->
    <div class="p-4 border-b border-gray-600">
      <div class="flex items-center gap-2 flex-nowrap">
        <span class="text-sm text-gray-300 whitespace-nowrap">Currently viewing</span>
        <Select 
          size="sm"
          bind:value={populationState.currentPlanIndex}
          items={selectedPerson.plans.map((p, i) => ({ value: i, name: `Plan ${i + 1}` }))}
          class="bg-gray-700 text-white border-gray-600 min-w-0"
        />
        <Button size="xs" color="primary" onclick={addPlan} class="whitespace-nowrap">
          Add new plan
        </Button>
      </div>
    </div>
    
    <!-- Activities and legs -->
    <div class="flex-1 overflow-y-auto px-5 py-5">
      {#if currentPlan}
        <div class="space-y-3">
          {#each currentPlan.activities as activity, index}
            <!-- Activity -->
            <div class="bg-gray-700 rounded-lg px-3 pt-3 pb-7">
              <div class="flex items-baseline mb-2">
                <span class="text-lg font-semibold text-white">{index + 1}.</span>
                <h4 class="text-sm font-medium text-gray-300 ml-2">Activity</h4>
              </div>
              <div class="flex items-center gap-2 mb-2">
                    <span class="text-sm text-gray-300 whitespace-nowrap">Time Period:</span>
                    <div class="flex items-center gap-2 flex-1 justify-end">
                      <Input 
                        type="time" 
                        size="sm" 
                        value={activity.startTime}
                        onchange={(e) => updateActivityTime(activity.id, 'startTime', (e.target as HTMLInputElement).value)}
                        class="bg-gray-600 text-white border-gray-500 w-24 h-7 flex-1"
                      />
                      <span class="text-white">-</span>
                      <Input 
                        type="time" 
                        size="sm" 
                        value={activity.endTime}
                        onchange={(e) => updateActivityTime(activity.id, 'endTime', (e.target as HTMLInputElement).value)}
                        class="bg-gray-600 text-white border-gray-500 w-24 h-7 flex-1"
                      />
                    </div>
              </div>
              <div class="flex items-center justify-center gap-2 w-full">
                <div class="flex items-center gap-2 w-full max-w-sm">
                  <span class="text-sm text-gray-300">Type:</span>
                  <CompactSelect 
                    items={activityTypeOptions}
                    value={activity.type}
                    onchange={(e) => updateActivityType(activity.id, (e.target as HTMLSelectElement).value as ActivityType)}
                  />
                  <Button size="sm" color="red" onclick={() => deleteActivity(activity.id)} class="h-6 px-2">
                    <TrashBinOutline class="w-3 h-3" />
                  </Button>
                </div>
              </div>
              <div class="flex items-center gap-2 mt-2">
                <span class="text-sm text-gray-300">Location:</span>
                <span class="text-xs text-gray-400 flex-1">{formatLocation(activity.location)}</span>
                <Button 
                  size="xs" 
                  color={populationState.selectingActivityId === activity.id ? "yellow" : "alternative"}
                  onclick={() => selectActivityLocation(activity.id)} 
                  class="h-6 px-2"
                >
                  <MapPinOutline class="w-3 h-3 mr-1" />
                  {populationState.selectingActivityId === activity.id ? 'Click map' : 'Set'}
                </Button>
              </div>
            </div>
            
            <!-- Leg (if not last activity) -->
            {#if index < currentPlan.activities.length - 1}
              {@const leg = currentPlan.legs.find(l => l.fromActivityId === activity.id)}
              {#if leg}
                <div class="mx-4 my-3 relative">
                  <div class="absolute left-4 top-0 w-0.5 h-full bg-gray-500"></div>
                  <div class="bg-gray-700 rounded-lg p-3 ml-8 border border-gray-600">
                    <div class="flex items-center gap-2">
                      <span class="text-xs text-gray-400">Travel by:</span>
                      <CompactSelect 
                        items={travelModeOptions}
                        value={leg.mode}
                        onchange={(e) => updateLegMode(leg, (e.target as HTMLSelectElement).value as TravelMode)}
                      />
                    </div>
                    <div class="flex items-center gap-2 mt-2">
                      <span class="text-xs text-gray-400">Duration:</span>
                      <Input 
                        type="number" 
                        size="sm" 
                        value={leg.duration}
                        disabled={leg.mode === "person's choice"}
                        onchange={(e) => updateLegDuration(leg, parseInt((e.target as HTMLInputElement).value) || 0)}
                        class="bg-gray-600 text-white border-gray-500 h-6 w-20 text-xs {leg.mode === "person's choice" ? 'opacity-50 cursor-not-allowed' : ''}"
                      />
                      <span class="text-xs text-gray-400">minutes</span>
                    </div>
                  </div>
                </div>
              {/if}
            {/if}
          {/each}
        </div>
      {/if}
    </div>
    
    <!-- Footer -->
    <div class="p-4 border-t border-gray-600 space-y-2">
      <Button color="primary" size="sm" class="w-full" onclick={addActivity}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Activity
      </Button>
      <Button color="red" size="sm" class="w-full" onclick={deletePlan} disabled={selectedPerson.plans.length <= 1}>
        <TrashBinOutline class="w-4 h-4 mr-2" />
        Delete Plan
      </Button>
    </div>
  </div>
{:else}
  <div class="h-full flex items-center justify-center text-gray-400 bg-gray-800">
    <p>Select a person to view details</p>
  </div>
{/if}