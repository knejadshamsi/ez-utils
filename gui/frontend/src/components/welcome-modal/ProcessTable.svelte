<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Radio, Button, P } from "flowbite-svelte";
  import { welcomeModalState, getPaginatedProcesses, getTotalPages } from "./welcome.svelte";
  import { DeleteProcess, GetProcessesByFile } from "../../../wailsjs/go/gui/App";
  import { commandArgs } from "../../store.svelte";
  
  let error = $state<string | null>(null);

  // Handle delete process
  async function handleDeleteProcess(processId: number) {
    if (!confirm('Are you sure you want to delete this process?')) {
      return;
    }
    
    try {
      await DeleteProcess(processId);
      // Refresh the process list
      welcomeModalState.processes = await GetProcessesByFile(commandArgs.filePath);
      
      // If we deleted the selected process, clear selection
      if (welcomeModalState.selectedProcessId === processId) {
        welcomeModalState.selectedProcessId = welcomeModalState.processes.length > 0 ? welcomeModalState.processes[0].process_id : null;
      }
      
      
      // Reset to first page if current page is now empty
      if (welcomeModalState.currentPage > getTotalPages()) {
        welcomeModalState.currentPage = Math.max(1, getTotalPages());
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete process';
    }
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
                on:change={() => welcomeModalState.selectedProcessId = process.process_id}
              />
            </TableBodyCell>
            <TableBodyCell>{process.process_id}</TableBodyCell>
            <TableBodyCell>
              <span class={process.status === 'Completed' ? 'text-green-600' : process.status === 'failed' ? 'text-red-600' : 'text-yellow-600'}>
                {process.status}
              </span>
            </TableBodyCell>
            <TableBodyCell>{new Date(process.timestamp).toLocaleString()}</TableBodyCell>
            <TableBodyCell>
              <Button 
                size="xs" 
                color="red" 
                outline
                on:click={() => handleDeleteProcess(process.process_id)}
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
          on:click={() => welcomeModalState.currentPage--}
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
          on:click={() => welcomeModalState.currentPage++}
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