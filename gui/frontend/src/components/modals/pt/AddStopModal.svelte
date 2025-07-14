<script lang="ts">
  import { Modal, Button, Label, Input, Select, Helper } from 'flowbite-svelte';
  import { ptState, type LineWithRoutes, type RouteWithTiming, type StopTime } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';

  let { open = $bindable(), line, route, routeIndex }: { open: boolean; line: LineWithRoutes; route: RouteWithTiming; routeIndex: number } = $props();
  let stopName = $state('');
  let existingStopId = $state('');
  let arrival = $state('');
  let dwellMinutes = $state('2');
  let error = $state('');
  let useExisting = $state(false);

  const availableStops = $derived(
    Array.from(ptState.stops.values())
      .filter(stop => !route.stopSequence.some(s => s.stopId === stop.id))
      .map(stop => ({ value: stop.id, name: stop.name }))
  );

  const suggestedArrival = $derived(() => {
    if (route.stopSequence.length === 0) {
      return route.firstDeparture || '06:00';
    }
    
    const lastStop = route.stopSequence[route.stopSequence.length - 1];
    const lastDepartureMinutes = ptState.timeToMinutes(lastStop.arrival) + lastStop.dwellMinutes;
    return ptState.minutesToTime(lastDepartureMinutes + 5);
  });

  $effect(() => {
    if (open) {
      arrival = suggestedArrival();
    }
  });

  function resetForm() {
    stopName = '';
    existingStopId = '';
    arrival = '';
    dwellMinutes = '2';
    error = '';
    useExisting = false;
  }

  function validateForm(): boolean {
    if (!useExisting && !stopName.trim()) {
      error = 'Stop name is required';
      return false;
    }

    if (useExisting && !existingStopId) {
      error = 'Please select an existing stop';
      return false;
    }

    if (!arrival.match(/^\d{2}:\d{2}$/)) {
      error = 'Invalid time format. Use HH:MM';
      return false;
    }

    const dwell = parseInt(dwellMinutes);
    if (isNaN(dwell) || dwell < 0 || dwell > 60) {
      error = 'Dwell time must be between 0 and 60 minutes';
      return false;
    }

    if (route.stopSequence.length > 0) {
      const lastStop = route.stopSequence[route.stopSequence.length - 1];
      const lastDepartureMinutes = ptState.timeToMinutes(lastStop.arrival) + lastStop.dwellMinutes;
      const newArrivalMinutes = ptState.timeToMinutes(arrival);
      
      if (newArrivalMinutes < lastDepartureMinutes) {
        error = 'Arrival time must be after the previous stop\'s departure';
        return false;
      }
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    let stopId: string;
    
    if (useExisting) {
      stopId = existingStopId;
    } else {
      stopId = generateId();
      const newStop = {
        id: stopId,
        name: stopName.trim(),
        location: [0, 0] as [number, number],
        x: 0,
        y: 0,
        telemetryId: '',
        raw_xml: ''
      };
      
      changeTracker.pendingChanges.push({
        type: 'pt',
        elementType: 'stop',
        action: 'add',
        processId: getCurrentProcessId(),
        stop: newStop
      });
      
      ptState.stops.set(stopId, newStop);
    }

    const newStopTime: StopTime = {
      stopId: stopId,
      arrival: arrival,
      dwellMinutes: parseInt(dwellMinutes),
      sequence: route.stopSequence.length + 1
    };

    // Calculate arrival offset from first departure
    const arrivalOffset = ptState.timeToMinutes(arrival) - ptState.timeToMinutes(route.firstDeparture);
    const departureOffset = arrivalOffset + parseInt(dwellMinutes);

    const routeStop = {
      id: generateId(),
      route_id: route.id,
      stop_ref_id: stopId,
      stop_order: route.stopSequence.length + 1,
      arrival_offset: `PT${arrivalOffset}M`,
      departure_offset: `PT${departureOffset}M`,
      raw_xml: ''
    };

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'routeStop',
      action: 'add',
      processId: getCurrentProcessId(),
      routeStop: routeStop
    });

    route.stopSequence.push(newStopTime);
    
    open = false;
    resetForm();
  }

  function handleCancel() {
    open = false;
    resetForm();
  }
</script>

<Modal bind:open title="Add Stop" size="sm">
  <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }}>
    <div class="space-y-4">
      <div>
        <Label class="mb-2">
          <input
            type="checkbox"
            bind:checked={useExisting}
            class="mr-2"
          />
          Use existing stop
        </Label>
      </div>

      {#if useExisting}
        <div>
          <Label for="existingStop" class="mb-2">Select Stop *</Label>
          <Select
            id="existingStop"
            bind:value={existingStopId}
            items={availableStops}
            placeholder="Choose a stop"
            required
          />
        </div>
      {:else}
        <div>
          <Label for="stopName" class="mb-2">Stop Name *</Label>
          <Input
            id="stopName"
            bind:value={stopName}
            placeholder="Enter stop name"
            required
          />
        </div>
      {/if}

      <div>
        <Label for="arrival" class="mb-2">Arrival Time *</Label>
        <Input
          id="arrival"
          bind:value={arrival}
          placeholder="HH:MM"
          pattern="^\d{2}:\d{2}$"
          required
        />
        <Helper class="mt-1">Suggested: {suggestedArrival}</Helper>
      </div>

      <div>
        <Label for="dwell" class="mb-2">Dwell Time (minutes) *</Label>
        <Input
          id="dwell"
          type="number"
          bind:value={dwellMinutes}
          min="0"
          max="60"
          required
        />
      </div>

      {#if error}
        <Helper color="red">{error}</Helper>
      {/if}
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      <Button type="submit" color="primary">Add Stop</Button>
    </div>
  </form>
</Modal>