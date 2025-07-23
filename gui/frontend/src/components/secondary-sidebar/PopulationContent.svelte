<script lang="ts">
  import { Button, Select, Input, CloseButton } from 'flowbite-svelte';
  import { TrashBinOutline, PlusOutline, MapPinOutline, CogOutline } from 'flowbite-svelte-icons';
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
  import { appState } from '$lib/stores/app.svelte';
  import { GetPerson } from '@wailsjs/go/gui/App';
  import { trackActivityChange } from '$lib/utils/populationChangeTracking';
  import { parsePersonXML } from '$lib/utils/populationXmlParser';
  import { updateConnectedDots } from '../../map/updateConnectedDots';
  import { mapState } from '../../map/mapState.svelte';
  import type * as L from 'leaflet';
  
  // Get selected person
  const selectedPerson = $derived(
    populationState.selectedPersonId 
      ? populationState.persons.get(populationState.selectedPersonId)
      : null
  );
  
  // Current plan - using shared state
  const currentPlan = $derived(selectedPerson?.plans[populationState.currentPlanIndex]);
  
  // Load person details when selection changes
  $effect(() => {
    if (populationState.selectedPersonId && appState.processId) {
      loadPersonDetails(populationState.selectedPersonId);
    }
  });
  
  async function loadPersonDetails(personId: string) {
    try {
      const tableName = `population_data_${appState.processId}`;
      const personData = await GetPerson(tableName, personId);
      
      const parsedData = parsePersonXML(personData.raw_xml);
      const person = populationState.persons.get(personId);
      
      if (person) {
        const updatedPerson = {
          ...person,
          plans: parsedData.plans || []
        };
        
        if (updatedPerson.plans.length === 0) {
          updatedPerson.plans.push({
            id: 1,
            activities: [],
            legs: []
          });
        }
        
        populationState.persons.set(personId, updatedPerson);
      }
    } catch (error) {
      console.error('Failed to load person details:', error);
    }
  }
  
  function handleClose() {
    populationState.selectedPersonId = null;
    appState.secondarySidebar = 'HIDDEN';
    updateConnectedDots();
  }
  
  function addActivity() {
    if (!selectedPerson || !currentPlan || !mapState.map) return;
    
    populationState.isAddingActivity = true;
    
    const clickHandler = (e: L.LeafletMouseEvent) => {
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      const newActivity: Activity = {
        id: `activity_${Date.now()}`,
        type: 'other',
        location: coords,
        startTime: getNextStartTime(),
        endTime: getNextEndTime()
      };
      
      const updatedPerson = {
        ...selectedPerson,
        plans: selectedPerson.plans.map((plan, index) => 
          index === populationState.currentPlanIndex 
            ? { ...plan, activities: [...plan.activities, newActivity] }
            : plan
        )
      };
      
      populationState.persons.set(selectedPerson.id, updatedPerson);
      
      autoSortActivities();
      
      trackActivityChange(selectedPerson.id, updatedPerson);
      
      updateConnectedDots();
    };
    
    mapState.map.on('click', clickHandler);
    
    (window as any).__addActivityHandler = clickHandler;
  }
  
  function stopAddingActivities() {
    if (!mapState.map) return;
    
    populationState.isAddingActivity = false;
    
    const handler = (window as any).__addActivityHandler;
    if (handler) {
      mapState.map.off('click', handler);
      delete (window as any).__addActivityHandler;
    }
  }
  
  function getNextStartTime(): string {
    if (!currentPlan || currentPlan.activities.length === 0) return '08:00';
    
    const lastActivity = currentPlan.activities[currentPlan.activities.length - 1];
    const [hours, minutes] = lastActivity.endTime.split(':').map(Number);
    
    let newMinutes = minutes + 30;
    let newHours = hours;
    
    if (newMinutes >= 60) {
      newMinutes -= 60;
      newHours += 1;
    }
    
    return `${String(newHours).padStart(2, '0')}:${String(newMinutes).padStart(2, '0')}`;
  }
  
  function getNextEndTime(): string {
    const startTime = getNextStartTime();
    const [hours, minutes] = startTime.split(':').map(Number);
    
    let newMinutes = minutes + 30;
    let newHours = hours;
    
    if (newMinutes >= 60) {
      newMinutes -= 60;
      newHours += 1;
    }
    
    return `${String(newHours).padStart(2, '0')}:${String(newMinutes).padStart(2, '0')}`;
  }
  
  function deleteActivity(activityId: string) {
    if (!selectedPerson || !currentPlan) return;
    
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.map((plan, index) => 
        index === populationState.currentPlanIndex 
          ? { ...plan, activities: plan.activities.filter(a => a.id !== activityId) }
          : plan
      )
    };
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    autoSortActivities();
    
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateActivityTime(activityId: string, field: 'startTime' | 'endTime', value: string) {
    if (!selectedPerson || !currentPlan) return;
    
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
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    autoSortActivities();
    
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateActivityType(activityId: string, type: ActivityType) {
    if (!selectedPerson || !currentPlan) return;
    
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
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateLegMode(leg: Leg, mode: TravelMode) {
    if (!selectedPerson || !currentPlan) return;
    
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
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function updateLegDuration(leg: Leg, duration: number) {
    if (!selectedPerson || !currentPlan) return;
    
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
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    trackActivityChange(selectedPerson.id, updatedPerson);
  }
  
  function autoSortActivities() {
    if (!selectedPerson || !currentPlan) return;
    
    const currentPerson = populationState.persons.get(selectedPerson.id);
    if (!currentPerson) return;
    
    const plan = currentPerson.plans[populationState.currentPlanIndex];
    if (!plan) return;
    
    const sortedActivities = [...plan.activities].sort((a, b) => {
      const [hoursA = 0, minutesA = 0] = a.startTime.split(':').map(Number);
      const [hoursB = 0, minutesB = 0] = b.startTime.split(':').map(Number);
      return (hoursA * 60 + minutesA) - (hoursB * 60 + minutesB);
    });
    
    const existingLegs = new Map(plan.legs.map(leg => 
      [`${leg.fromActivityId}-${leg.toActivityId}`, leg]
    ));
    
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
    
    const updatedPerson = {
      ...currentPerson,
      plans: currentPerson.plans.map((p, index) => 
        index === populationState.currentPlanIndex 
          ? { ...p, activities: sortedActivities, legs: newLegs }
          : p
      )
    };
    
    populationState.persons.set(currentPerson.id, updatedPerson);
    
    updateConnectedDots();
  }
  
  function addPlan() {
    if (!selectedPerson) return;
    
    const updatedPerson = {
      ...selectedPerson,
      plans: [...selectedPerson.plans, {
        id: selectedPerson.plans.length + 1,
        activities: [],
        legs: []
      }]
    };
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    populationState.currentPlanIndex = updatedPerson.plans.length - 1;
  }
  
  function deletePlan() {
    if (!selectedPerson || selectedPerson.plans.length <= 1) return;
    
    const updatedPerson = {
      ...selectedPerson,
      plans: selectedPerson.plans.filter((_, index) => index !== populationState.currentPlanIndex)
    };
    
    populationState.persons.set(selectedPerson.id, updatedPerson);
    
    populationState.currentPlanIndex = Math.max(0, populationState.currentPlanIndex - 1);
  }
  
  function selectActivityLocation(activityId: string) {
    if (!mapState.map || !selectedPerson || !currentPlan) return;
    
    if (populationState.selectingActivityId === activityId) {
      populationState.isSelectingActivityLocation = false;
      populationState.selectingActivityId = null;
      return;
    }
    
    populationState.isSelectingActivityLocation = true;
    populationState.selectingActivityId = activityId;
    
    const clickHandler = (e: L.LeafletMouseEvent) => {
      const coords: [number, number] = [e.latlng.lng, e.latlng.lat];
      
      const updatedPerson = {
        ...selectedPerson,
        plans: selectedPerson.plans.map((plan, index) => 
          index === populationState.currentPlanIndex 
            ? {
                ...plan,
                activities: plan.activities.map(a => 
                  a.id === activityId 
                    ? { ...a, location: coords }
                    : a
                )
              }
            : plan
        )
      };
      
      populationState.persons.set(selectedPerson.id, updatedPerson);
      
      populationState.isSelectingActivityLocation = false;
      populationState.selectingActivityId = null;
      
      trackActivityChange(selectedPerson.id, updatedPerson);
      
      updateConnectedDots();
      
      mapState.map.off('click', clickHandler);
    };
    
    mapState.map.on('click', clickHandler);
  }
  
  function formatLocation(location: [number, number]): string {
    if (location[0] === 0 && location[1] === 0) {
      return 'Not set';
    }
    return `${location[1].toFixed(4)}, ${location[0].toFixed(4)}`;
  }
  
  function toggleActivityDragging() {
    populationState.isDraggingActivity = !populationState.isDraggingActivity;
    updateConnectedDots();
  }
  
  const activityTypeOptions = Object.entries(activityTypeConfig).map(([value, config]) => ({
    value,
    name: `${config.icon} ${config.label}`
  }));
  
  const travelModeOptions = Object.entries(travelModeConfig).map(([value, config]) => ({
    value,
    name: `${config.icon} ${config.label}`
  }));
  
</script>

{#if selectedPerson}
  <div class="h-full flex flex-col bg-gray-800">
    <div class="flex items-center justify-between p-4 border-b border-gray-600">
      <div class="flex items-center gap-2">
        <h2 class="text-lg font-semibold text-white">{selectedPerson.id}</h2>
        <Button
          size="xs"
          class="!p-1.5"
          color="alternative"
          onclick={() => populationState.showAttributesModal = true}
          title="Edit Attributes"
        >
          <CogOutline class="w-4 h-4" />
        </Button>
      </div>
      <CloseButton onclick={handleClose} class="text-gray-400 hover:text-white" />
    </div>
    
    <div class="p-4 border-b border-gray-600">
      <div class="flex items-center gap-2 flex-nowrap">
        <span class="text-sm text-gray-300 whitespace-nowrap">Currently viewing</span>
        <Select 
          size="sm"
          bind:value={populationState.currentPlanIndex}
          onchange={() => updateConnectedDots()}
          items={selectedPerson.plans.map((p, i) => ({ value: i, name: `Plan ${i + 1}` }))}
          class="bg-gray-700 text-white border-gray-600 min-w-0"
        />
        <Button size="xs" color="primary" onclick={addPlan} class="whitespace-nowrap">
          Add new plan
        </Button>
      </div>
    </div>
    
    <div class="flex-1 overflow-y-auto px-5 py-5">
      {#if currentPlan}
        <div class="space-y-3">
          {#each currentPlan.activities as activity, index}
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
                      <span class="text-xs text-gray-400">mins</span>
                    </div>
                  </div>
                </div>
              {/if}
            {/if}
          {/each}
        </div>
      {/if}
    </div>
    
    <div class="p-4 border-t border-gray-600 space-y-2">
      <Button 
        color={populationState.isDraggingActivity ? "yellow" : "alternative"} 
        size="sm" 
        class="w-full" 
        onclick={toggleActivityDragging}
      >
        <MapPinOutline class="w-4 h-4 mr-2" />
        {populationState.isDraggingActivity ? 'Done Editing Locations' : 'Edit Locations'}
      </Button>
      <Button 
        color={populationState.isAddingActivity ? "yellow" : "primary"} 
        size="sm" 
        class="w-full" 
        onclick={populationState.isAddingActivity ? stopAddingActivities : addActivity}
      >
        {#if populationState.isAddingActivity}
          Done Adding Activities
        {:else}
          <PlusOutline class="w-4 h-4 mr-2" />
          Add Activity
        {/if}
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