<script lang="ts">
  import { Button, Input, Label } from 'flowbite-svelte';
  import { TrashBinOutline } from 'flowbite-svelte-icons';
  import { ptState, type LineWithRoutes, type RouteWithTiming, type StopTime } from './ptState.svelte';
  import { changeTracker } from '../../../lib/changeTracker.svelte';
  import { getCurrentProcessId } from './getCurrentProcessId';

  let { line, route, routeIndex, stop, stopIndex }: { 
    line: LineWithRoutes, 
    route: RouteWithTiming, 
    routeIndex: number,
    stop: StopTime,
    stopIndex: number 
  } = $props();

  let editingArrival = $state(false);
  let editingDwell = $state(false);
  let tempArrival = $state(stop.arrival);
  let tempDwell = $state(stop.dwellMinutes.toString());

  const stopInfo = $derived(ptState.stops.get(stop.stopId));

  $effect(() => {
    tempArrival = stop.arrival;
    tempDwell = stop.dwellMinutes.toString();
  });

  function handleArrivalChange() {
    if (!tempArrival.match(/^\d{2}:\d{2}$/)) {
      tempArrival = stop.arrival;
      editingArrival = false;
      return;
    }

    const oldArrival = stop.arrival;
    stop.arrival = tempArrival;

    if (stopIndex === 0 && oldArrival !== tempArrival) {
      const timeDiff = ptState.timeToMinutes(tempArrival) - ptState.timeToMinutes(oldArrival);
      route.firstDeparture = tempArrival;
      
      for (let i = 1; i < route.stopSequence.length; i++) {
        const nextStop = route.stopSequence[i];
        const newMinutes = ptState.timeToMinutes(nextStop.arrival) + timeDiff;
        nextStop.arrival = ptState.minutesToTime(newMinutes);
      }
    } else if (stopIndex > 0) {
      ptState.recalculateSubsequentTimes(line.id, routeIndex, stopIndex);
    }

    // Calculate new arrival offset from first departure
    const arrivalOffset = ptState.timeToMinutes(stop.arrival) - ptState.timeToMinutes(route.firstDeparture);
    const departureOffset = arrivalOffset + stop.dwellMinutes;

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'routeStop',
      action: 'update',
      processId: getCurrentProcessId(),
      routeId: route.id,
      stopOrder: stop.sequence,
      update: {
        arrival_offset: `PT${arrivalOffset}M`,
        departure_offset: `PT${departureOffset}M`
      }
    });

    editingArrival = false;
  }

  function handleDwellChange() {
    const dwellMinutes = parseInt(tempDwell);
    if (isNaN(dwellMinutes) || dwellMinutes < 0) {
      tempDwell = stop.dwellMinutes.toString();
      editingDwell = false;
      return;
    }

    stop.dwellMinutes = dwellMinutes;
    
    if (stopIndex < route.stopSequence.length - 1) {
      ptState.recalculateSubsequentTimes(line.id, routeIndex, stopIndex);
    }

    // Calculate new offsets
    const arrivalOffset = ptState.timeToMinutes(stop.arrival) - ptState.timeToMinutes(route.firstDeparture);
    const departureOffset = arrivalOffset + dwellMinutes;

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'routeStop',
      action: 'update',
      processId: getCurrentProcessId(),
      routeId: route.id,
      stopOrder: stop.sequence,
      update: {
        arrival_offset: `PT${arrivalOffset}M`,
        departure_offset: `PT${departureOffset}M`
      }
    });

    editingDwell = false;
  }

  function handleDelete() {
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'routeStop',
      action: 'delete',
      processId: getCurrentProcessId(),
      routeId: route.id,
      stopOrder: stop.sequence
    });

    route.stopSequence = route.stopSequence.filter((_, index) => index !== stopIndex);
    
    route.stopSequence.forEach((s, index) => {
      s.sequence = index + 1;
    });

    if (route.stopSequence.length > stopIndex) {
      ptState.recalculateSubsequentTimes(line.id, routeIndex, stopIndex - 1);
    }
  }

  const isValidTime = $derived(() => {
    if (stopIndex === 0) return true;
    const prevStop = route.stopSequence[stopIndex - 1];
    const prevDepartureMinutes = ptState.timeToMinutes(prevStop.arrival) + prevStop.dwellMinutes;
    const currentArrivalMinutes = ptState.timeToMinutes(stop.arrival);
    return currentArrivalMinutes >= prevDepartureMinutes;
  });
</script>

<div class="px-4 py-3 border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
  <div class="flex items-start justify-between mb-2">
    <div class="flex-1">
      <div class="font-medium text-sm text-gray-900 dark:text-white">
        {stopIndex + 1}. {stopInfo?.name || 'Unknown Stop'}
      </div>
      
      <div class="mt-2 space-y-1">
        <div class="flex items-center gap-2">
          <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Arrival:</Label>
          {#if editingArrival}
            <Input
              type="text"
              size="sm"
              bind:value={tempArrival}
              placeholder="HH:MM"
              pattern="^\d{2}:\d{2}$"
              class="w-20 {!isValidTime ? 'border-red-500' : ''}"
              onblur={handleArrivalChange}
              onkeydown={(e) => {
                if (e.key === 'Enter') handleArrivalChange();
                if (e.key === 'Escape') {
                  tempArrival = stop.arrival;
                  editingArrival = false;
                }
              }}
              autofocus
            />
          {:else}
            <button
              class="px-2 py-1 text-xs font-mono bg-white dark:bg-gray-700 border rounded hover:bg-gray-100 dark:hover:bg-gray-600 {!isValidTime ? 'border-red-500 text-red-600 dark:text-red-400' : 'border-gray-300 dark:border-gray-600'}"
              onclick={() => editingArrival = true}
            >
              {stop.arrival}
            </button>
          {/if}
        </div>
        
        <div class="flex items-center gap-2">
          <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Departure:</Label>
          <span class="px-2 py-1 text-xs font-mono bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded">
            {ptState.minutesToTime(ptState.timeToMinutes(stop.arrival) + stop.dwellMinutes)}
          </span>
        </div>
        
        <div class="flex items-center gap-2">
          <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Dwell:</Label>
          {#if editingDwell}
            <div class="flex items-center gap-1">
              <Input
                type="number"
                size="sm"
                bind:value={tempDwell}
                min="0"
                max="60"
                class="w-16"
                onblur={handleDwellChange}
                onkeydown={(e) => {
                  if (e.key === 'Enter') handleDwellChange();
                  if (e.key === 'Escape') {
                    tempDwell = stop.dwellMinutes.toString();
                    editingDwell = false;
                  }
                }}
                autofocus
              />
              <span class="text-xs text-gray-500">min</span>
            </div>
          {:else}
            <button
              class="px-2 py-1 text-xs bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-600"
              onclick={() => editingDwell = true}
            >
              {stop.dwellMinutes} min
            </button>
          {/if}
        </div>
        
        <div class="flex items-center gap-2">
          <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Stop ID:</Label>
          <span class="text-xs font-mono text-gray-600 dark:text-gray-400">{stop.stopId}</span>
        </div>
        
        {#if stopInfo}
          <div class="flex items-center gap-2">
            <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Location:</Label>
            <span class="text-xs font-mono text-gray-600 dark:text-gray-400">
              {stopInfo.x.toFixed(6)}, {stopInfo.y.toFixed(6)}
            </span>
          </div>
        {/if}
        
        <div class="flex items-center gap-2">
          <Label class="text-xs text-gray-500 dark:text-gray-400 w-16">Sequence:</Label>
          <span class="text-xs text-gray-600 dark:text-gray-400">{stop.sequence}</span>
        </div>
      </div>
    </div>
    
    <Button
      size="xs"
      color="red"
      onclick={handleDelete}
      disabled={route.stopSequence.length <= 2}
    >
      <TrashBinOutline size="xs" />
    </Button>
  </div>
</div>