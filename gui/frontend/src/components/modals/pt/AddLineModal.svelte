<script lang="ts">
  import { Modal, Button, Label, Input, Select, Helper } from 'flowbite-svelte';
  import { ptState, TransportMode, type LineWithRoutes } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';
  import { trackLineChange } from '$lib/utils/ptChangeTracking';

  let { open = $bindable() }: { open: boolean } = $props();
  let name = $state('');
  let number = $state('');
  let mode = $state<TransportMode>(TransportMode.Bus);
  let color = $state('#FF0000');
  let error = $state('');

  function resetForm() {
    name = '';
    number = '';
    mode = TransportMode.Bus;
    color = '#FF0000';
    error = '';
  }

  function validateForm(): boolean {
    if (!name.trim()) {
      error = 'Line name is required';
      return false;
    }

    const existingLine = Array.from(ptState.lines.values()).find(
      line => line.name.toLowerCase() === name.trim().toLowerCase() && line.mode === mode
    );
    
    if (existingLine) {
      error = 'A line with this name already exists for this mode';
      return false;
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    const newLine: LineWithRoutes = {
      id: generateId(),
      name: name.trim(),
      number: number.trim(),
      mode: mode,
      color: color,
      agencyId: '',
      telemetryId: '',
      raw_xml: '',
      routes: []
    };

    // Create new Map to trigger Svelte 5 reactivity
    const newLines = new Map(ptState.lines);
    newLines.set(newLine.id, newLine);
    ptState.lines = newLines;
    
    // Track the new line
    trackLineChange(newLine, 'add');
    ptState.setSelectedLine(newLine.id);
    
    open = false;
    resetForm();
  }

  function handleCancel() {
    open = false;
    resetForm();
  }

  const modeOptions = [
    { value: TransportMode.Bus, name: '🚌 Bus' },
    { value: TransportMode.Metro, name: '🚇 Metro' },
    { value: TransportMode.Tram, name: '🚊 Tram' }
  ];
</script>

<Modal bind:open title="New Line" size="sm">
  <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }}>
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

    <div class="flex justify-end gap-2 mt-6">
      <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      <Button type="submit" color="primary">Create</Button>
    </div>
  </form>
</Modal>