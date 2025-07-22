<script lang="ts">
  import { Modal, P, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline, ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { appState, commandArgs, editingSession } from "$lib/stores/app.svelte.ts";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { PTService } from "$lib/api/pt";
  import { CheckProcessingStatus } from "@wailsjs/go/gui/App";

  let currentProcessId = $state<number | null>(null);
  let loadingError = $state<string | null>(null);

  // Start loading existing process
  async function startLoading() {
    if (!welcomeModalState.selectedProcessId) return;
    
    currentProcessId = welcomeModalState.selectedProcessId;
    loadingError = null;

    try {
      // Check if processing is complete
      const status = await CheckProcessingStatus(currentProcessId);
      
      if (status === 'COMPLETED') {
        appState.display = 'LOADING_SUCCESS';
      } else if (status === 'FAILED') {
        loadingError = 'Process failed during processing';
        appState.display = 'LOADING_FAILED';
      } else {
        loadingError = `Process is still ${status}. Please wait for processing to complete.`;
        appState.display = 'LOADING_FAILED';
      }
      
    } catch (error) {
      console.error('Loading failed:', error);
      loadingError = error instanceof Error ? error.message : 'Loading failed';
      appState.display = 'LOADING_FAILED';
    }
  }

  // Handle begin editing button click
  async function handleBeginEditing() {
    // Set the editing session values
    if (currentProcessId) {
      editingSession.processId = currentProcessId;
      
      // Set table name based on file edit mode
      switch (commandArgs.fileEditMode) {
        case 'POPULATION':
          editingSession.tableName = `population_data_${currentProcessId}`;
          break;
        case 'NETWORK':
          editingSession.tableName = `network_${currentProcessId}`;
          break;
        case 'PT':
          editingSession.tableName = `pt_${currentProcessId}`;
          break;
      }
      
      console.log('Editing session initialized:', {
        processId: editingSession.processId,
        tableName: editingSession.tableName,
        fileEditMode: commandArgs.fileEditMode
      });
    }
    
    // Load PT data if in PT editing mode
    if (commandArgs.fileEditMode === 'PT') {
      try {
        await PTService.loadPTData();
      } catch (error) {
        console.error('Failed to load PT data:', error);
      }
    }
    
    appState.display = 'EDITING';
  }

  // Effect to start loading when modal opens
  $effect(() => {
    if (appState.display === 'LOADING') {
      startLoading();
    }
  });
</script>

<Modal 
  open={appState.display === 'LOADING' || appState.display === 'LOADING_SUCCESS' || appState.display === 'LOADING_FAILED'} 
  permanent
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="md"
  placement="center"
>
  {#snippet header()}
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white text-center flex items-center justify-center gap-3">
      {#if appState.display === 'LOADING'}
        <Spinner size="6" />
        Loading Process Data
      {:else if appState.display === 'LOADING_SUCCESS'}
        <CheckCircleOutline class="w-6 h-6 text-green-500" />
        Process Loaded Successfully
      {:else if appState.display === 'LOADING_FAILED'}
        <ExclamationCircleOutline class="w-6 h-6 text-red-500" />
        Loading Failed
      {/if}
    </h3>
  {/snippet}
  
  <div class="space-y-6">
    {#if appState.display === 'LOADING'}
      <div class="text-center">
        <P class="text-gray-600 dark:text-gray-400">
          Loading process data from database...
        </P>
      </div>
      
    {:else if appState.display === 'LOADING_SUCCESS'}
      <div class="text-center space-y-4">
        <P class="text-gray-600 dark:text-gray-400">
          Process data loaded successfully. Ready to begin editing.
        </P>
        <Button 
          color="primary" 
          size="lg"
          onclick={handleBeginEditing}
        >
          Begin Editing
        </Button>
      </div>
      
    {:else if appState.display === 'LOADING_FAILED'}
      <div class="text-center space-y-4">
        <P color="red" class="text-red-600 dark:text-red-400">
          {loadingError || 'Failed to load process data'}
        </P>
        <Button 
          color="alternative" 
          onclick={() => appState.display = 'WELCOME'}
        >
          Back to Welcome
        </Button>
      </div>
    {/if}
  </div>
</Modal>