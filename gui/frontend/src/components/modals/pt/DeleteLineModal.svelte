<script lang="ts">
  import { Modal, Button, P } from 'flowbite-svelte';
  import { ptState, type LineWithRoutes } from '$lib/stores/pt.svelte';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { appState } from '$lib/stores/app.svelte';

  let { open = $bindable(), line }: { open: boolean; line: LineWithRoutes } = $props();

  function handleDelete() {
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'line',
      action: 'delete',
      processId: getCurrentProcessId(),
      lineId: line.id
    });

    ptState.lines.delete(line.id);
    ptState.setSelectedLine(null);
    appState.secondarySidebar = 'HIDDEN';
    
    open = false;
  }

  function handleCancel() {
    open = false;
  }

  const affectedStopsCount = $derived(() => {
    const uniqueStopIds = new Set<string>();
    for (const route of line.routes) {
      for (const stop of route.stopSequence) {
        uniqueStopIds.add(stop.stopId);
      }
    }
    return uniqueStopIds.size;
  });
</script>

<Modal bind:open title="Delete Line" size="sm">
  <div class="space-y-4">
    <P>
      Are you sure you want to delete the line <strong>{line.name}</strong>?
    </P>
    
    <div class="bg-red-50 dark:bg-red-900/20 p-3 rounded-lg">
      <P class="text-sm text-red-800 dark:text-red-200">
        This will permanently delete:
      </P>
      <ul class="list-disc list-inside text-sm text-red-700 dark:text-red-300 mt-2">
        <li>{line.routes.length} route{line.routes.length !== 1 ? 's' : ''}</li>
        <li>{affectedStopsCount()} stop reference{affectedStopsCount() !== 1 ? 's' : ''}</li>
      </ul>
    </div>
    
    <P class="text-sm text-gray-600 dark:text-gray-400">
      This action cannot be undone.
    </P>
  </div>

  <div class="flex justify-end gap-2 mt-6">
    <Button color="alternative" onclick={handleCancel}>Cancel</Button>
    <Button color="red" onclick={handleDelete}>Delete Line</Button>
  </div>
</Modal>