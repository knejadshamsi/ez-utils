<script lang="ts">
  import { Modal, Button, Label, Input, Helper, P } from 'flowbite-svelte';
  import { generateId } from '$lib/utils/generateId';
  import type { PTDepartureExtended } from '$lib/utils/ptModelMapping';
  import AutoComplete from 'simple-svelte-autocomplete';
  import { SelectFile } from '@wailsjs/go/gui/App';

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
  let tvFilePath = $state('');
  let vehicleIds = $state<string[]>([]);
  let selectedVehicle = $state<string>('');

  async function handleSelectTVFile() {
    try {
      const filePath = await SelectFile();
      if (filePath && filePath.trim() !== '') {
        tvFilePath = filePath;
        await parseTVFile(filePath);
      }
    } catch (error) {
      vehicleIds = [];
    }
  }

  async function parseTVFile(filePath: string) {
    try {
      const response = await fetch(`file://${filePath}`);
      const text = await response.text();
      const parser = new DOMParser();
      const xmlDoc = parser.parseFromString(text, 'text/xml');
      
      const parseError = xmlDoc.querySelector('parsererror');
      if (parseError) {
        throw new Error('Invalid XML format');
      }
      
      const departures = xmlDoc.querySelectorAll('departure[vehicleRefId]');
      const vehicleIdSet = new Set<string>();
      
      departures.forEach(departure => {
        const vehicleRefId = departure.getAttribute('vehicleRefId');
        if (vehicleRefId) {
          vehicleIdSet.add(vehicleRefId);
        }
      });
      
      vehicleIds = Array.from(vehicleIdSet).sort();
      selectedVehicle = '';
    } catch (error) {
      vehicleIds = [];
    }
  }

  function handleClearTVFile() {
    tvFilePath = '';
    vehicleIds = [];
    selectedVehicle = '';
  }

  function resetForm() {
    departureTime = '06:00';
    vehicleId = '';
    selectedVehicle = '';
    error = '';
    tvFilePath = '';
    vehicleIds = [];
  }

  function validateForm(): boolean {
    if (!departureTime) {
      error = 'Departure time is required';
      return false;
    }

    const finalVehicleId = selectedVehicle || vehicleId.trim();
    if (!finalVehicleId) {
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

    const finalVehicleId = selectedVehicle || vehicleId.trim();

    const newDeparture: PTDepartureExtended = {
      id: generateId(),
      route_id: routeId,
      routeId: routeId,
      departure_time: departureTime,
      departureTime: departureTime,
      vehicle_id: finalVehicleId,
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

<Modal bind:open size="md" title="Add Departure" class="bg-gray-900">
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
      <Label class="text-gray-300 mb-2">Transit Vehicles File (Optional)</Label>
      <P class="mb-2 text-sm text-gray-400">
        Select a TV XML file to load vehicle IDs for autocomplete search.
      </P>
      
      {#if tvFilePath}
        <!-- Display selected file path -->
        <div class="rounded-lg border border-gray-600 bg-gray-800 p-3 mb-3">
          <P class="text-sm text-gray-300 break-all">
            {tvFilePath}
          </P>
          <Button 
            size="sm" 
            color="alternative" 
            class="mt-2"
            onclick={handleClearTVFile}
          >
            Clear File
          </Button>
        </div>
      {:else}
        <!-- Show file selection button -->
        <div class="flex justify-center mb-3">
          <Button 
            color="alternative" 
            outline
            onclick={handleSelectTVFile}
            class="w-48"
          >
            Select TV XML File
          </Button>
        </div>
      {/if}
    </div>

    <div>
      <Label class="text-gray-300 mb-2">Vehicle ID</Label>
      
      {#if vehicleIds.length > 0}
        <div class="mb-2">
          <AutoComplete
            items={vehicleIds}
            bind:selectedItem={selectedVehicle}
            placeholder="Search vehicle IDs..."
            maxItemsToShowInList={10}
            minCharactersToSearch={0}
            class="bg-gray-700 text-white border-gray-600 focus:border-blue-500 rounded-lg"
          />
          <Helper class="text-gray-400 text-sm mt-1">
            {vehicleIds.length} vehicle(s) loaded from TV file
          </Helper>
        </div>
      {/if}
      
      <Input 
        type="text" 
        bind:value={vehicleId}
        placeholder="Or enter vehicle ID manually (e.g., bus_1, metro_1)"
        class="bg-gray-700 text-white border-gray-600 focus:border-blue-500"
      />
      <Helper class="text-gray-400 text-sm mt-1">
        You can search from loaded vehicles above or enter manually
      </Helper>
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