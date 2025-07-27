<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { ptState } from '$lib/stores/pt.svelte';

  let { open = $bindable() }: { open: boolean } = $props();
  let name = $state('');
  let error = $state('');

  function resetForm() {
    name = '';
    error = '';
  }

  function validateForm(): boolean {
    if (!name.trim()) {
      error = 'Line name is required';
      return false;
    }

    const ptData = ptState.ptData.find(pd => pd.mode === ptState.selected.mode);
    if (ptData) {
      const existingLine = ptData.lines.find(
        line => line.name.toLowerCase() === name.trim().toLowerCase()
      );
      
      if (existingLine) {
        error = 'A line with this name already exists for this mode';
        return false;
      }
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    ptState.createLine(name.trim());
    
    open = false;
    resetForm();
  }

  function handleCancel() {
    open = false;
    resetForm();
  }

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