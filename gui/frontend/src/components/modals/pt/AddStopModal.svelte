<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { ptState } from '$lib/stores/pt.svelte';
  import { nanoid } from 'nanoid';

  let { open = $bindable(), routeId }: { open: boolean; routeId: string } = $props();
  let stopName = $state('');
  let stopId = $state('');
  let arrivalOffset = $state('00:00');
  let departureOffset = $state('00:02');
  let lat = $state(0);
  let lng = $state(0);
  let error = $state('');

  function resetForm() {
    stopName = '';
    stopId = `stop_${nanoid(10)}`;
    arrivalOffset = '00:00';
    departureOffset = '00:02';
    lat = 0;
    lng = 0;
    error = '';
  }

  $effect(() => {
    if (open) {
      resetForm();
    }
  });

  function validateForm(): boolean {
    if (!stopName.trim()) {
      error = 'Stop name is required';
      return false;
    }

    if (!arrivalOffset.match(/^\d{2}:\d{2}$/)) {
      error = 'Invalid arrival time format. Use HH:MM';
      return false;
    }

    if (!departureOffset.match(/^\d{2}:\d{2}$/)) {
      error = 'Invalid departure time format. Use HH:MM';
      return false;
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    ptState.createStop({
      routeId: routeId,
      stopId: stopId,
      arrivalOffset: arrivalOffset,
      departureOffset: departureOffset,
      stopName: stopName.trim(),
      lat: lat,
      lng: lng
    });
    
    open = false;
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
        <Label for="stopName" class="mb-2">Stop Name *</Label>
        <Input
          id="stopName"
          bind:value={stopName}
          placeholder="Enter stop name"
          required
        />
      </div>

      <div>
        <Label for="arrival" class="mb-2">Arrival Offset (HH:MM) *</Label>
        <Input
          id="arrival"
          bind:value={arrivalOffset}
          placeholder="00:00"
          pattern="^\d{2}:\d{2}$"
          required
        />
      </div>

      <div>
        <Label for="departure" class="mb-2">Departure Offset (HH:MM) *</Label>
        <Input
          id="departure"
          bind:value={departureOffset}
          placeholder="00:02"
          pattern="^\d{2}:\d{2}$"
          required
        />
      </div>

      <div class="grid grid-cols-2 gap-2">
        <div>
          <Label for="lat" class="mb-2">Latitude</Label>
          <Input
            id="lat"
            type="number"
            bind:value={lat}
            step="0.000001"
            placeholder="0.0"
          />
        </div>
        <div>
          <Label for="lng" class="mb-2">Longitude</Label>
          <Input
            id="lng"
            type="number"
            bind:value={lng}
            step="0.000001"
            placeholder="0.0"
          />
        </div>
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