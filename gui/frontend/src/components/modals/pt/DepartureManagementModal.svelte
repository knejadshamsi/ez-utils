<script lang="ts">
  import { Modal, Table, TableBody, TableHead, TableHeadCell, TableBodyRow, TableBodyCell, Input, Button, Helper } from 'flowbite-svelte';
  import { TrashBinOutline, PlusOutline, ClockOutline } from 'flowbite-svelte-icons';
  import { ptState } from '$lib/stores/pt.svelte';
  import type { Departure } from '$lib/stores/pt.svelte';

  interface Props {
    routeId: string;
  }

  let { routeId }: Props = $props();

  const departures = $derived(() => {
    return ptState.currentRouteData.departures.sort((a, b) => a.departureTime.localeCompare(b.departureTime));
  });

  let editingId = $state<string | null>(null);
  let editValues = $state<{[key: string]: {time: string, vehicleId: string}}>({});
  
  let newDepartureTime = $state('');
  let newDepartureVehicleId = $state('');
  let newDepartureError = $state('');

  function handleClose() {
    ptState.editMode = 'NORMAL';
    editingId = null;
    editValues = {};
    newDepartureTime = '';
    newDepartureVehicleId = '';
    newDepartureError = '';
  }

  function startEdit(departure: Departure) {
    editingId = departure.id;
    editValues[departure.id] = {
      time: departure.departureTime,
      vehicleId: departure.vehicleRefId || ''
    };
  }

  function cancelEdit() {
    editingId = null;
  }

  function saveEdit(departureId: string) {
    const values = editValues[departureId];
    if (!values || !validateTime(values.time)) return;

    ptState.updateDeparture(departureId, {
      departureTime: values.time,
      vehicleRefId: values.vehicleId || undefined
    });
    
    editingId = null;
  }

  function handleDelete(departureId: string) {
    ptState.deleteDeparture(departureId);
  }

  function validateTime(time: string): boolean {
    const timeRegex = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/;
    return timeRegex.test(time);
  }

  function addNewDeparture() {
    newDepartureError = '';
    
    if (!newDepartureTime) {
      newDepartureError = 'Time is required';
      return;
    }

    if (!validateTime(newDepartureTime)) {
      newDepartureError = 'Invalid time format. Use HH:MM:SS';
      return;
    }

    ptState.createDeparture(routeId, newDepartureTime, newDepartureVehicleId || undefined);
    
    newDepartureTime = '';
    newDepartureVehicleId = '';
  }

  function duplicateDeparture(departure: Departure) {
    let [hours, minutes, seconds] = departure.departureTime.split(':').map(Number);
    minutes += 5;
    if (minutes >= 60) {
      hours += 1;
      minutes -= 60;
    }
    if (hours >= 24) hours = 0;
    
    const newTime = `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    ptState.createDeparture(routeId, newTime, departure.vehicleRefId);
  }

  function handleKeydown(e: KeyboardEvent, departureId: string) {
    if (e.key === 'Enter') {
      saveEdit(departureId);
    } else if (e.key === 'Escape') {
      cancelEdit();
    }
  }
</script>

<Modal open={ptState.editMode === 'EDITING_DEPARTURES'} onclose={handleClose} size="lg" title="Manage Departures">
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <ClockOutline class="w-5 h-5" />
        <span class="text-lg font-medium">Departure Schedule</span>
      </div>
      <span class="text-sm text-gray-500">{departures().length} departures</span>
    </div>

    <Table striped={true}>
      <TableHead>
        <TableHeadCell class="w-1/3">Time</TableHeadCell>
        <TableHeadCell class="w-1/3">Vehicle ID</TableHeadCell>
        <TableHeadCell class="w-1/3">Actions</TableHeadCell>
      </TableHead>
      <TableBody>
        {#each departures() as departure (departure.id)}
          <TableBodyRow>
            <TableBodyCell>
              {#if editingId === departure.id}
                <Input
                  type="text"
                  size="sm"
                  placeholder="HH:MM:SS"
                  bind:value={editValues[departure.id].time}
                  onkeydown={(e) => handleKeydown(e, departure.id)}
                  class="w-full"
                />
              {:else}
                <span class="font-mono">{departure.departureTime}</span>
              {/if}
            </TableBodyCell>
            <TableBodyCell>
              {#if editingId === departure.id}
                <Input
                  type="text"
                  size="sm"
                  placeholder="Optional"
                  bind:value={editValues[departure.id].vehicleId}
                  onkeydown={(e) => handleKeydown(e, departure.id)}
                  class="w-full"
                />
              {:else}
                <span class="text-gray-500">{departure.vehicleRefId || '-'}</span>
              {/if}
            </TableBodyCell>
            <TableBodyCell>
              <div class="flex gap-1">
                {#if editingId === departure.id}
                  <Button size="xs" color="green" onclick={() => saveEdit(departure.id)}>
                    Save
                  </Button>
                  <Button size="xs" color="alternative" onclick={cancelEdit}>
                    Cancel
                  </Button>
                {:else}
                  <Button size="xs" color="blue" onclick={() => startEdit(departure)}>
                    Edit
                  </Button>
                  <Button size="xs" color="alternative" onclick={() => duplicateDeparture(departure)}>
                    Duplicate
                  </Button>
                  <Button size="xs" color="red" onclick={() => handleDelete(departure.id)}>
                    <TrashBinOutline size="xs" />
                  </Button>
                {/if}
              </div>
            </TableBodyCell>
          </TableBodyRow>
        {/each}
        
        <TableBodyRow class="bg-gray-50 dark:bg-gray-700">
          <TableBodyCell>
            <Input
              type="text"
              size="sm"
              placeholder="HH:MM:SS"
              bind:value={newDepartureTime}
              onkeydown={(e) => e.key === 'Enter' && addNewDeparture()}
            />
          </TableBodyCell>
          <TableBodyCell>
            <Input
              type="text"
              size="sm"
              placeholder="Vehicle ID (optional)"
              bind:value={newDepartureVehicleId}
              onkeydown={(e) => e.key === 'Enter' && addNewDeparture()}
            />
          </TableBodyCell>
          <TableBodyCell>
            <Button size="xs" color="green" onclick={addNewDeparture}>
              <PlusOutline size="xs" class="mr-1" />
              Add
            </Button>
          </TableBodyCell>
        </TableBodyRow>
      </TableBody>
    </Table>

    {#if newDepartureError}
      <Helper color="red">{newDepartureError}</Helper>
    {/if}

    {#if departures().length === 0}
      <p class="text-center text-gray-500 py-8">No departures scheduled. Add one above.</p>
    {/if}
  </div>
</Modal>