<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { ptState, type RouteWithTiming } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';

  let { open = $bindable(), lineId }: { open: boolean; lineId: string } = $props();
  let direction = $state('');
  let firstDeparture = $state('06:00');
  let error = $state('');

  function resetForm() {
    direction = '';
    firstDeparture = '06:00';
    error = '';
  }

  function validateForm(): boolean {
    if (!direction.trim()) {
      error = 'Direction is required';
      return false;
    }

    if (!firstDeparture.match(/^\d{2}:\d{2}$/)) {
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

    const line = ptState.lines.get(lineId);
    if (!line) {
      error = 'Line not found';
      return;
    }

    const newRoute: RouteWithTiming = {
      id: generateId(),
      lineId: lineId,
      line_id: lineId,
      direction: direction.trim(),
      telemetryId: '',
      raw_xml: '',
      firstDeparture: firstDeparture,
      stopSequence: []
    };

    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'route',
      action: 'add',
      processId: getCurrentProcessId(),
      route: newRoute
    });

    line.routes.push(newRoute);
    
    open = false;
    resetForm();
  }

  function handleCancel() {
    open = false;
    resetForm();
  }
</script>

<Modal bind:open title="New Route" size="sm">
  <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }}>
    <div class="space-y-4">
      <div>
        <Label for="direction" class="mb-2">Direction *</Label>
        <Input
          id="direction"
          bind:value={direction}
          placeholder="e.g., Eastbound, Northbound, Inbound"
          required
        />
      </div>

      <div>
        <Label for="firstDeparture" class="mb-2">First Departure *</Label>
        <Input
          id="firstDeparture"
          bind:value={firstDeparture}
          placeholder="HH:MM"
          pattern="^\d{2}:\d{2}$"
          required
        />
        <Helper class="mt-1">24-hour format (e.g., 06:00, 14:30)</Helper>
      </div>

      {#if error}
        <Helper color="red">{error}</Helper>
      {/if}
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      <Button type="submit" color="primary">Create</Button>
    </div>
  </form>
</Modal>