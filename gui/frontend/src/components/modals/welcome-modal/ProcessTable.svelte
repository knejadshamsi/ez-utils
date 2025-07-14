<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Radio, Button, P, Modal } from "flowbite-svelte";
  import { ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { welcomeModalState, getPaginatedProcesses, getTotalPages } from "./welcome.svelte";
  import { GetProcessesByFile, SyncChanges } from "@wailsjs/go/gui/App";
  import { commandArgs } from "$lib/stores/app.svelte.ts";
  import { changeTracker } from "../../../lib/changeTracker.svelte";
  
  let error = $state<string | null>(null);
  let showDeleteConfirmation = $state(false);
  let processToDelete = $state<number | null>(null);

  // Show delete confirmation modal
  function showDeleteModal(processId: number) {
    processToDelete = processId;
    showDeleteConfirmation = true;
  }
  
  // Handle delete confirmation
  async function confirmDelete() {
    if (!processToDelete) return;
    
    showDeleteConfirmation = false;
    const processId = processToDelete;
    processToDelete = null;
    
    try {
      console.log('Deleting process:', processId);
      
      // Queue the delete action
      changeTracker.pendingChanges.push({
        type: 'process',
        elementType: 'process',
        action: 'delete',
        processId: processId
      });
      
      console.log('Syncing changes:', changeTracker.pendingChanges);
      
      // Sync immediately for delete operations
      await SyncChanges(changeTracker.pendingChanges);
      changeTracker.pendingChanges = [];
      
      console.log('Delete successful, refreshing process list');
      
      // Refresh the process list
      const updatedProcesses = await GetProcessesByFile(commandArgs.filePath);
      console.log('Updated processes:', updatedProcesses);
      welcomeModalState.processes = updatedProcesses || [];
      
      // If we deleted the selected process, clear selection
      if (welcomeModalState.selectedProcessId === processId) {
        welcomeModalState.selectedProcessId = welcomeModalState.processes.length > 0 ? welcomeModalState.processes[0].process_id : null;
      }
      
      
      // Reset to first page if current page is now empty
      if (welcomeModalState.currentPage > getTotalPages()) {
        welcomeModalState.currentPage = Math.max(1, getTotalPages());
      }
    } catch (err) {
      console.error('Delete failed:', err);
      error = err instanceof Error ? err.message : 'Failed to delete process';
    }
  }
  
  // Cancel delete
  function cancelDelete() {
    showDeleteConfirmation = false;
    processToDelete = null;
  }
</script>

{#if welcomeModalState.processes.length === 0}
  <!-- No records found message -->
  <div class="text-center py-8">
    <P class="text-gray-600 dark:text-gray-400 mb-4">
      No previous processes found for this file.
    </P>
    <P class="text-sm text-gray-500 dark:text-gray-500">
      Click "New Process" to start processing this file.
    </P>
  </div>
{:else}
  <!-- Process table -->
  <div>
    <P class="text-gray-600 dark:text-gray-400 mb-4">
      Select a previous process to continue editing, or create a new process.
    </P>
    
    <Table striped class="mt-4">
      <TableHead>
        <TableHeadCell class="w-10"></TableHeadCell>
        <TableHeadCell>Process ID</TableHeadCell>
        <TableHeadCell>Status</TableHeadCell>
        <TableHeadCell>Timestamp</TableHeadCell>
        <TableHeadCell class="w-20">Actions</TableHeadCell>
      </TableHead>
      <TableBody>
        {#each getPaginatedProcesses() as process}
          <TableBodyRow>
            <TableBodyCell class="w-10">
              <Radio 
                name="processSelection"
                value={process.process_id}
                checked={welcomeModalState.selectedProcessId === process.process_id}
                onchange={() => welcomeModalState.selectedProcessId = process.process_id}
              />
            </TableBodyCell>
            <TableBodyCell>{process.process_id}</TableBodyCell>
            <TableBodyCell>
              <span class={process.status === 'COMPLETED' ? 'text-green-600 font-medium' : process.status === 'FAILED' ? 'text-red-600 font-medium' : 'text-yellow-600'}>
                {process.status}
              </span>
            </TableBodyCell>
            <TableBodyCell>{new Date(process.timestamp).toLocaleString()}</TableBodyCell>
            <TableBodyCell>
              <Button 
                size="xs" 
                color="red" 
                outline
                onclick={() => showDeleteModal(process.process_id)}
              >
                Delete
              </Button>
            </TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    </Table>
    
    {#if getTotalPages() > 1}
      <div class="mt-4 flex justify-center">
        <Button
          size="sm"
          color="alternative"
          disabled={welcomeModalState.currentPage === 1}
          onclick={() => welcomeModalState.currentPage--}
        >
          Previous
        </Button>
        <span class="mx-4 text-sm text-gray-600 dark:text-gray-400">
          Page {welcomeModalState.currentPage} of {getTotalPages()}
        </span>
        <Button
          size="sm"
          color="alternative"
          disabled={welcomeModalState.currentPage === getTotalPages()}
          onclick={() => welcomeModalState.currentPage++}
        >
          Next
        </Button>
      </div>
    {/if}
    
    {#if error}
      <P class="mt-4 text-red-600 dark:text-red-400">{error}</P>
    {/if}
  </div>
{/if}

<!-- Delete Confirmation Modal -->
<Modal bind:open={showDeleteConfirmation} size="xs" autoclose={false}>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
      Are you sure you want to delete process #{processToDelete}?
    </h3>
    <P class="mb-5 text-sm text-gray-500 dark:text-gray-400">
      This action cannot be undone. All data associated with this process will be permanently deleted.
    </P>
    <div class="flex justify-center gap-4">
      <Button color="red" onclick={confirmDelete}>
        Yes, Delete Process
      </Button>
      <Button color="alternative" onclick={cancelDelete}>
        Cancel
      </Button>
    </div>
  </div>
</Modal>