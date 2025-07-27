<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { TrashBinOutline } from 'flowbite-svelte-icons';
  import { ptState } from '$lib/stores/pt.svelte';

  let { open = $bindable(), line, onclose }: { 
    open: boolean, 
    line: any,
    onclose: () => void 
  } = $props();
  
  let name = $state('');
  let error = $state('');

  $effect(() => {
    if (line) {
      name = line.name || '';
      error = '';
    }
  });

  function validateForm(): boolean {
    if (!name.trim()) {
      error = 'Line name is required';
      return false;
    }

    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    if (ptData) {
      const existingLine = ptData.lines.find(
        l => l.id !== line.id && l.name.toLowerCase() === name.trim().toLowerCase()
      );
      
      if (existingLine) {
        error = 'A line with this name already exists';
        return false;
      }
    }

    return true;
  }

  function handleSave() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    ptState.updateLine(line.id, { name: name.trim() });
    
    open = false;
    onclose();
  }

  function handleDelete() {
    if (!confirm(`Are you sure you want to delete line "${line.name}"? This will also delete all its routes, stops, and departures.`)) {
      return;
    }

    ptState.deleteLine(line.id);
    
    open = false;
    onclose();
  }

  function handleCancel() {
    open = false;
    onclose();
  }

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