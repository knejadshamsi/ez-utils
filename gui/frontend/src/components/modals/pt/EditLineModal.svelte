<script lang="ts">
  import { Modal, Button, Label, Input, Select, Helper } from 'flowbite-svelte';
  import { TrashBinOutline } from 'flowbite-svelte-icons';
  import { ptState, TransportMode } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';

  let { open = $bindable(), line, onclose }: { 
    open: boolean, 
    line: any,
    onclose: () => void 
  } = $props();
  
  let name = $state('');
  let number = $state('');
  let mode = $state<TransportMode>(TransportMode.Bus);
  let color = $state('#FF0000');
  let error = $state('');

  $effect(() => {
    if (line) {
      name = line.name || '';
      number = line.number || '';
      mode = line.mode || TransportMode.Bus;
      color = line.color || '#FF0000';
      error = '';
    }
  });

  function validateForm(): boolean {
    if (!name.trim()) {
      error = 'Line name is required';
      return false;
    }

    const existingLine = Array.from(ptState.lines.values()).find(
      l => l.id !== line.id && l.name.toLowerCase() === name.trim().toLowerCase() && l.mode === mode
    );
    
    if (existingLine) {
      error = 'A line with this name already exists for this mode';
      return false;
    }

    return true;
  }

  function handleSave() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    // Update line properties
    line.name = name.trim();
    line.number = number.trim();
    line.mode = mode;
    line.color = color;

    // Add to change tracker
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'line',
      action: 'update',
      processId: getCurrentProcessId(),
      lineId: line.id,
      update: {
        name: line.name,
        number: line.number,
        mode: line.mode,
        color: line.color
      }
    });

    // Update state to trigger reactivity
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, { ...line });
    ptState.lines = newLines;
    
    open = false;
    onclose();
  }

  function handleDelete() {
    if (!confirm(`Are you sure you want to delete line "${line.name}"? This will also delete all its routes.`)) {
      return;
    }

    // Add to change tracker
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'line',
      action: 'delete',
      processId: getCurrentProcessId(),
      lineId: line.id
    });

    // Remove from state
    const newLines = new Map(ptState.lines);
    newLines.delete(line.id);
    ptState.lines = newLines;

    // Clear selection if this line was selected
    if (ptState.selectedLineId === line.id) {
      ptState.setSelectedLine(null);
    }
    
    open = false;
    onclose();
  }

  function handleCancel() {
    open = false;
    onclose();
  }

  const modeOptions = [
    { value: TransportMode.Bus, name: '🚌 Bus' },
    { value: TransportMode.Rail, name: '🚇 Rail' },
    { value: TransportMode.Tram, name: '🚊 Tram' }
  ];
</script>

<Modal bind:open title="Edit Line" size="sm">
  <form onsubmit={(e) => { e.preventDefault(); handleSave(); }}>
    <div class="space-y-4">
      <div>
        <Label for="name" class="mb-2">Name *</Label>
        <Input
          id="name"
          bind:value={name}
          placeholder="Enter line name"
          required
        />
      </div>

      <div>
        <Label for="number" class="mb-2">Number</Label>
        <Input
          id="number"
          bind:value={number}
          placeholder="Optional line number"
        />
      </div>

      <div>
        <Label for="mode" class="mb-2">Mode *</Label>
        <Select
          id="mode"
          bind:value={mode}
          items={modeOptions}
        />
      </div>

      <div>
        <Label for="color" class="mb-2">Color</Label>
        <div class="flex gap-2">
          <Input
            id="color"
            type="color"
            bind:value={color}
            class="w-20 h-10"
          />
          <Input
            bind:value={color}
            placeholder="#FF0000"
            class="flex-1"
          />
        </div>
      </div>

      {#if error}
        <Helper color="red">{error}</Helper>
      {/if}
    </div>

    <div class="flex justify-between items-center mt-6">
      <Button 
        color="red" 
        onclick={handleDelete}
        class="flex items-center gap-2"
      >
        <TrashBinOutline size="sm" />
        Delete Line
      </Button>
      
      <div class="flex gap-2">
        <Button color="alternative" onclick={handleCancel}>Cancel</Button>
        <Button type="submit" color="primary">Save Changes</Button>
      </div>
    </div>
  </form>
</Modal>