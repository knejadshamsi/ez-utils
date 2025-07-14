<script lang="ts">
  import { Button, Input, Label } from 'flowbite-svelte';
  import { TrashBinOutline, PlusOutline } from 'flowbite-svelte-icons';
  import { ptState, type LineWithRoutes, type RouteWithTiming } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import StopItem from './StopItem.svelte';
  import AddStopModal from '../modals/pt/AddStopModal.svelte';

  let { line, route, routeIndex }: { line: LineWithRoutes, route: RouteWithTiming, routeIndex: number } = $props();

  let showAddStopModal = $state(false);
  let editingDeparture = $state(false);
  let tempDeparture = $state(route.firstDeparture);

  $effect(() => {
    tempDeparture = route.firstDeparture;
  });

  function handleFirstDepartureChange() {
    if (!tempDeparture.match(/^\d{2}:\d{2}$/)) {
      tempDeparture = route.firstDeparture;
      editingDeparture = false;
      return;
    }

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'route',
      action: 'update',
      processId: getCurrentProcessId(),
      routeId: route.id,
      update: {
        firstDeparture: tempDeparture
      }
    });

    route.firstDeparture = tempDeparture;
    
    route.stopSequence.forEach((stop, index) => {
      if (index === 0) {
        stop.arrival = tempDeparture;
      } else {
        const offset = ptState.timeToMinutes(stop.arrival) - ptState.timeToMinutes(route.stopSequence[0].arrival);
        stop.arrival = ptState.calculateArrivalTime(tempDeparture, offset);
      }
    });

    editingDeparture = false;
  }

  function handleAddStop() {
    showAddStopModal = true;
  }

  function handleDeleteRoute() {
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'route',
      action: 'delete',
      processId: getCurrentProcessId(),
      routeId: route.id
    });

    const lineRoutes = line.routes.filter(r => r.id !== route.id);
    line.routes = lineRoutes;
  }

  const totalTime = $derived(ptState.getTotalTravelTime(route));
</script>

<div class="border-b border-gray-200 dark:border-gray-700">
  <div class="px-4 py-3 bg-gray-50 dark:bg-gray-800">
    <div class="flex items-center justify-between mb-2">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {route.direction || `Route ${routeIndex + 1}`}
      </h3>
      <Button
        size="xs"
        color="red"
        onclick={handleDeleteRoute}
      >
        <TrashBinOutline size="xs" />
      </Button>
    </div>
    
    <div class="flex items-center gap-2">
      <Label class="text-xs">First departure:</Label>
      {#if editingDeparture}
        <Input
          type="text"
          size="sm"
          bind:value={tempDeparture}
          placeholder="HH:MM"
          pattern="^\d{2}:\d{2}$"
          class="w-20"
          onblur={handleFirstDepartureChange}
          onkeydown={(e) => {
            if (e.key === 'Enter') handleFirstDepartureChange();
            if (e.key === 'Escape') {
              tempDeparture = route.firstDeparture;
              editingDeparture = false;
            }
          }}
          autofocus
        />
      {:else}
        <button
          class="px-2 py-1 text-sm font-mono bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-600"
          onclick={() => editingDeparture = true}
        >
          {route.firstDeparture || '--:--'}
        </button>
      {/if}
    </div>
  </div>

  <div class="max-h-96 overflow-y-auto">
    {#each route.stopSequence as stop, index (stop.stopId + index)}
      <StopItem
        {line}
        {route}
        {routeIndex}
        {stop}
        stopIndex={index}
      />
    {/each}
  </div>

  {#if route.stopSequence.length >= 2}
    <div class="px-4 py-2 bg-gray-50 dark:bg-gray-800 text-sm text-gray-600 dark:text-gray-400">
      Total time: {totalTime} minutes
    </div>
  {/if}

  <div class="p-2">
    <Button
      size="xs"
      color="light"
      class="w-full"
      onclick={handleAddStop}
    >
      <PlusOutline size="xs" class="mr-1" />
      Add Stop
    </Button>
  </div>
</div>

<AddStopModal
  bind:open={showAddStopModal}
  {line}
  {route}
  {routeIndex}
/>