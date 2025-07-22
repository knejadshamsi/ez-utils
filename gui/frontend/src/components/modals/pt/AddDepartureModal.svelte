<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { generateId } from '$lib/utils/generateId';
  import type { PTDepartureExtended } from '$lib/utils/ptModelMapping';

  let { 
    open = $bindable(), 
    routeId,
    onDepartureAdd = () => {}
  }: { 
    open: boolean;
    routeId: string;
    onDepartureAdd?: (departure: PTDepartureExtended) => void;
  } = $props();

  let departureTime = $state('06:00');
  let vehicleId = $state('');
  let error = $state('');

  function resetForm() {
    departureTime = '06:00';
    vehicleId = '';
    error = '';
  }

  function validateForm(): boolean {
    if (!departureTime) {
      error = 'Departure time is required';
      return false;
    }

    if (!vehicleId.trim()) {
      error = 'Vehicle ID is required';
      return false;
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    const newDeparture: PTDepartureExtended = {
      id: generateId(),
      route_id: routeId,
      routeId: routeId,
      departure_time: departureTime,
      departureTime: departureTime,
      vehicle_id: vehicleId.trim(),
      raw_xml: ''
    };

    onDepartureAdd(newDeparture);
    resetForm();
    open = false;
  }

  function handleClose() {
    resetForm();
    open = false;
  }

  $effect(() => {
    if (open) {
      resetForm();
    }
  });
</script>

<Modal bind:open size="sm" title="Add Departure" class="bg-gray-900">
  <div class="space-y-4">
    <div>
      <Label for="departure-time" class="text-gray-300 mb-2">Departure Time</Label>
      <Input 
        id="departure-time"
        type="time" 
        bind:value={departureTime}
        class="bg-gray-700 text-white border-gray-600 focus:border-blue-500"
      />
    </div>

    <div>
      <Label for="vehicle-id" class="text-gray-300 mb-2">Vehicle ID</Label>
      <Input 
        id="vehicle-id"
        type="text" 
        bind:value={vehicleId}
        placeholder="e.g., bus_1, metro_1"
        class="bg-gray-700 text-white border-gray-600 focus:border-blue-500"
      />
    </div>

    {#if error}
      <Helper color="red" class="mt-2">{error}</Helper>
    {/if}
  </div>

  <div class="flex justify-end gap-2 mt-6">
    <Button color="alternative" onclick={handleClose}>Cancel</Button>
    <Button color="primary" onclick={handleCreate}>Create</Button>
  </div>
</Modal>