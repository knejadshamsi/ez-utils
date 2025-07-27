<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { ptState } from '$lib/stores/pt.svelte';

  let { 
    open = $bindable(), 
    routeId
  }: { 
    open: boolean;
    routeId: string;
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

    if (!departureTime.match(/^\d{2}:\d{2}$/)) {
      error = 'Invalid time format. Use HH:MM';
      return false;
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    ptState.createDeparture(routeId, departureTime, vehicleId.trim() || undefined);
    
    open = false;
    resetForm();
  }

  function handleCancel() {
    open = false;
    resetForm();
  }

  $effect(() => {
    if (open) {
      resetForm();
    }
  });
</script>

<Modal bind:open size="sm" title="Add Departure">
  <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }}>
    <div class="space-y-4">
      <div>
        <Label for="departure-time" class="mb-2">Departure Time *</Label>
        <Input 
          id="departure-time"
          type="time" 
          bind:value={departureTime}
          required
        />
      </div>

      <div>
        <Label for="vehicle-id" class="mb-2">Vehicle ID (Optional)</Label>
        <Input 
          id="vehicle-id"
          type="text" 
          bind:value={vehicleId}
          placeholder="e.g., bus_1, metro_1"
        />
      </div>

      {#if error}
        <Helper color="red">{error}</Helper>
      {/if}
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      <Button type="submit" color="primary">Add Departure</Button>
    </div>
  </form>
</Modal>