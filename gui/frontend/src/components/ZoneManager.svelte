<script lang="ts">
  import { Button, Modal, Input, Label } from 'flowbite-svelte';
  import { PlusOutline, TrashBinOutline, EditOutline } from 'flowbite-svelte-icons';
  import { GetAllZones, SaveZone, DeleteZone, CreateZoneFromGeoJSON } from '@wailsjs/go/gui/App';
  import { onMount } from 'svelte';
  
  let zones: any[] = [];
  let showModal = false;
  let editingZone: any = null;
  let newZoneName = '';
  let isDrawingZone = false;
  
  onMount(() => {
    loadZones();
  });
  
  async function loadZones() {
    try {
      zones = await GetAllZones();
    } catch (error) {
      console.error('Failed to load zones:', error);
    }
  }
  
  function startDrawingZone() {
    isDrawingZone = true;
    newZoneName = '';
    showModal = true;
    
    // Dispatch event to map to start drawing
    window.dispatchEvent(new CustomEvent('startDrawingZone'));
  }
  
  function cancelDrawing() {
    isDrawingZone = false;
    showModal = false;
    
    // Dispatch event to map to cancel drawing
    window.dispatchEvent(new CustomEvent('cancelDrawingZone'));
  }
  
  async function saveDrawnZone(polygon: any[]) {
    if (!newZoneName.trim()) {
      alert('Please enter a zone name');
      return;
    }
    
    try {
      const zone = {
        name: newZoneName,
        polygon: polygon
      };
      
      await SaveZone(zone);
      await loadZones();
      
      showModal = false;
      isDrawingZone = false;
    } catch (error) {
      console.error('Failed to save zone:', error);
      alert('Failed to save zone');
    }
  }
  
  async function deleteZone(zoneId: string) {
    if (!confirm('Are you sure you want to delete this zone?')) {
      return;
    }
    
    try {
      await DeleteZone(zoneId);
      await loadZones();
    } catch (error) {
      console.error('Failed to delete zone:', error);
      alert('Failed to delete zone');
    }
  }
  
  // Listen for zone drawn event from map
  const handleZoneDrawn = async (event: CustomEvent) => {
    if (event.detail && event.detail.polygon) {
      await saveDrawnZone(event.detail.polygon);
    }
  };
  
  window.addEventListener('zoneDrawn', handleZoneDrawn);
  
  // Clean up event listener on component destroy
  $effect(() => {
    return () => {
      window.removeEventListener('zoneDrawn', handleZoneDrawn);
    };
  });
</script>

<div class="p-4">
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-lg font-semibold">Zone Management</h2>
    <Button size="sm" color="primary" onclick={startDrawingZone}>
      <PlusOutline class="w-4 h-4 mr-2" />
      Draw New Zone
    </Button>
  </div>
  
  <div class="space-y-2">
    {#each zones as zone}
      <div class="flex items-center justify-between p-3 bg-gray-800 rounded">
        <div>
          <h3 class="font-medium">{zone.name}</h3>
          <p class="text-sm text-gray-400">ID: {zone.id}</p>
        </div>
        <div class="flex gap-2">
          <Button size="xs" color="alternative">
            <EditOutline class="w-3 h-3" />
          </Button>
          <Button size="xs" color="red" onclick={() => deleteZone(zone.id)}>
            <TrashBinOutline class="w-3 h-3" />
          </Button>
        </div>
      </div>
    {/each}
    
    {#if zones.length === 0}
      <p class="text-gray-400 text-center py-8">
        No zones defined. Click "Draw New Zone" to create one.
      </p>
    {/if}
  </div>
</div>

<Modal bind:open={showModal} size="sm" autoclose={false}>
  <h3 class="text-xl font-medium mb-4">New Zone</h3>
  
  <div class="mb-4">
    <Label for="zone-name">Zone Name</Label>
    <Input 
      id="zone-name"
      bind:value={newZoneName} 
      placeholder="Enter zone name..." 
    />
  </div>
  
  <p class="text-sm text-gray-400 mb-4">
    Draw the zone boundary on the map by clicking to add points. 
    Double-click to complete the polygon.
  </p>
  
  <div class="flex justify-end gap-2">
    <Button color="alternative" onclick={cancelDrawing}>Cancel</Button>
  </div>
</Modal>