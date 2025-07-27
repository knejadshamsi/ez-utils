<script lang="ts">
  import { Modal, Button, Label, Input, Helper } from 'flowbite-svelte';
  import { ptState } from '$lib/stores/pt.svelte';

  let { open = $bindable(), lineId }: { open: boolean; lineId: string } = $props();
  let name = $state('');
  let error = $state('');

  function resetForm() {
    name = '';
    error = '';
  }

  function validateForm(): boolean {
    if (!name.trim()) {
      error = 'Route name is required';
      return false;
    }

    return true;
  }

  function handleCreate() {
    error = '';
    
    if (!validateForm()) {
      return;
    }

    ptState.createRoute(lineId, name.trim());
    
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

<Modal bind:open title="New Route" size="sm">
  <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }}>
    <div class="space-y-4">
      <div>
        <Label for="name" class="mb-2">Route Name *</Label>
        <Input
          id="name"
          bind:value={name}
          placeholder="e.g., Downtown, Airport Express"
          required
        />
      </div>

      {#if error}
        <Helper color="red">{error}</Helper>
      {/if}
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      <Button type="submit" color="primary">Create Route</Button>
    </div>
  </form>
</Modal>