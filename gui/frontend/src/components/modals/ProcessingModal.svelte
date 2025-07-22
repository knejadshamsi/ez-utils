<script lang="ts">
  import { Modal, P, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline, ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { appState, commandArgs, editingSession } from "$lib/stores/app.svelte.ts";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { PTService } from "$lib/api/pt";
  import { 
    ProcessPopulationFile, 
    ProcessNetworkFile, 
    ProcessPTFile,
    GetProcessTelemetry,
    GetPTTelemetry,
    CheckProcessingStatus
  } from "@wailsjs/go/gui/App";

  // Type definitions for each workflow
  interface PopulationTelemetry {
    process_id: number;
    total_file_size: number;
    bytes_read: number;
    persons_extracted: number;
    error_count: number;
    last_updated: string;
  }

  interface NetworkTelemetry {
    process_id: number;
    total_file_size: number;
    bytes_read: number;
    nodes_read: number;
    links_read: number;
    error_count: number;
    last_updated: string;
  }

  interface PTTelemetry {
    process_id: number;
    total_file_size: number;
    bytes_read: number;
    stops_extracted: number;
    lines_extracted: number;
    routes_extracted: number;
    error_count: number;
    last_updated: string;
  }

  // Three states for each workflow
  let populationTelemetry = $state<PopulationTelemetry | null>(null);
  let networkTelemetry = $state<NetworkTelemetry | null>(null);
  let ptTelemetry = $state<PTTelemetry | null>(null);
  
  let pollingInterval = $state<ReturnType<typeof setInterval> | null>(null);
  let currentProcessId = $state<number | null>(null);
  let processingError = $state<string | null>(null);

  // Start processing based on file type
  async function startProcessing() {
    if (!commandArgs.filePath) return;
    
    processingError = null;

    try {
      let result;
      switch (commandArgs.fileEditMode) {
        case 'POPULATION':
          result = await ProcessPopulationFile(commandArgs.filePath);
          currentProcessId = result.process_id;
          break;
        case 'NETWORK':
          result = await ProcessNetworkFile(commandArgs.filePath);
          currentProcessId = result.process_id;
          break;
        case 'PT':
          result = await ProcessPTFile(commandArgs.filePath);
          currentProcessId = result.process_id;
          break;
        default:
          throw new Error(`Unknown file edit mode: ${commandArgs.fileEditMode}`);
      }
      
      if (!currentProcessId) {
        throw new Error('No process ID returned from backend');
      }
      
      // Start telemetry polling after getting process ID
      startTelemetryPolling();
      
    } catch (error) {
      console.error('Processing failed:', error);
      processingError = error instanceof Error ? error.message : 'Processing failed';
      appState.display = 'PROCESSING_FAILED';
      clearPolling();
    }
  }


  // Start telemetry polling for processing
  function startTelemetryPolling() {
    if (pollingInterval) clearInterval(pollingInterval);
    
    // Fetch telemetry immediately on start
    fetchTelemetryOnce();
    
    pollingInterval = setInterval(async () => {
      try {
        if (!currentProcessId) return;
        
        // Get telemetry data based on file type FIRST
        if (commandArgs.fileEditMode === 'PT') {
          const telemetry = await GetPTTelemetry(currentProcessId);
          ptTelemetry = telemetry;
        } else if (commandArgs.fileEditMode === 'POPULATION') {
          const telemetry = await GetProcessTelemetry(currentProcessId);
          populationTelemetry = telemetry;
        } else if (commandArgs.fileEditMode === 'NETWORK') {
          const telemetry = await GetProcessTelemetry(currentProcessId);
          networkTelemetry = telemetry;
        }
        
        // Then check processing status
        const status = await CheckProcessingStatus(currentProcessId);
        
        if (status === 'PROCESSING_SUCCESS') {
          appState.display = 'PROCESSING_SUCCESS';
          clearPolling();
          return;
        } else if (status === 'PROCESSING_FAILED') {
          appState.display = 'PROCESSING_FAILED';
          clearPolling();
          return;
        }
        
      } catch (error) {
        console.error('Failed to get telemetry:', error);
        // Don't stop polling on telemetry errors, just log them
      }
    }, 500); // Poll every 0.5 seconds for smoother updates
  }
  
  // Fetch telemetry once without status check
  async function fetchTelemetryOnce() {
    try {
      if (!currentProcessId) return;
      
      if (commandArgs.fileEditMode === 'PT') {
        const telemetry = await GetPTTelemetry(currentProcessId);
        ptTelemetry = telemetry;
      } else if (commandArgs.fileEditMode === 'POPULATION') {
        const telemetry = await GetProcessTelemetry(currentProcessId);
        populationTelemetry = telemetry;
      } else if (commandArgs.fileEditMode === 'NETWORK') {
        const telemetry = await GetProcessTelemetry(currentProcessId);
        networkTelemetry = telemetry;
      }
    } catch (error) {
      console.error('Failed to get initial telemetry:', error);
    }
  }

  // Clear polling interval
  function clearPolling() {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      pollingInterval = null;
    }
  }


  // Effect to start processing when modal opens
  $effect(() => {
    if (appState.display === 'PROCESSING') {
      startProcessing();
    }
  });

  // Cleanup on component unmount
  $effect(() => {
    return () => {
      clearPolling();
    };
  });
</script>

<Modal 
  open={appState.display === 'PROCESSING' || appState.display === 'PROCESSING_SUCCESS' || appState.display === 'PROCESSING_FAILED'} 
  permanent
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="md"
  placement="center"
>
  {#snippet header()}
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white text-center flex items-center justify-center gap-3">
      {#if appState.display === 'PROCESSING'}
        <Spinner size="6" />
        Processing {commandArgs.fileEditMode} File
      {:else if appState.display === 'PROCESSING_SUCCESS'}
        <CheckCircleOutline class="w-6 h-6 text-green-500" />
        {commandArgs.fileEditMode} File Processed
      {:else if appState.display === 'PROCESSING_FAILED'}
        <ExclamationCircleOutline class="w-6 h-6 text-red-500" />
        Processing Failed
      {/if}
    </h3>
  {/snippet}
  
  <div class="space-y-6">
    {#if appState.display === 'PROCESSING'}
      <!-- Processing State -->
      <div class="space-y-4">
        
        {#if commandArgs.fileEditMode === 'POPULATION'}
          {#if populationTelemetry}
            <div class="space-y-3">
              <P class="text-gray-600 dark:text-gray-400">
                Processing Population Data
              </P>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Persons Extracted: {populationTelemetry.persons_extracted.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Bytes Read: {populationTelemetry.bytes_read.toLocaleString()} / {populationTelemetry.total_file_size.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Errors: {populationTelemetry.error_count}
              </div>
            </div>
          {/if}
          
        {:else if commandArgs.fileEditMode === 'NETWORK'}
          {#if networkTelemetry}
            <div class="space-y-3">
              <P class="text-gray-600 dark:text-gray-400">
                Processing Network Data
              </P>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Nodes Read: {networkTelemetry.nodes_read.toLocaleString()}
              </div>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Links Read: {networkTelemetry.links_read.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Bytes Read: {networkTelemetry.bytes_read.toLocaleString()} / {networkTelemetry.total_file_size.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Errors: {networkTelemetry.error_count}
              </div>
            </div>
          {/if}
          
        {:else if commandArgs.fileEditMode === 'PT'}
          {#if ptTelemetry}
            <div class="space-y-3">
              <P class="text-gray-600 dark:text-gray-400">
                Processing PT Data
              </P>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Stops Extracted: {ptTelemetry.stops_extracted.toLocaleString()}
              </div>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Lines Extracted: {ptTelemetry.lines_extracted.toLocaleString()}
              </div>
              <div class="text-sm text-gray-600 dark:text-gray-400">
                Routes Extracted: {ptTelemetry.routes_extracted.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Bytes Read: {ptTelemetry.bytes_read.toLocaleString()} / {ptTelemetry.total_file_size.toLocaleString()}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                Errors: {ptTelemetry.error_count}
              </div>
            </div>
          {/if}
        {/if}
      </div>
    {/if}
    
    {#if appState.display === 'PROCESSING_SUCCESS'}
      <!-- Success State -->
      <div class="text-center space-y-4">
        <P class="text-gray-600 dark:text-gray-400">
          File processing completed successfully.
        </P>
        <Button 
          color="primary" 
          size="lg"
          onclick={() => {
            // Set the process ID for loading
            if (currentProcessId) {
              welcomeModalState.selectedProcessId = currentProcessId;
            }
            appState.display = 'LOADING';
          }}
        >
          Load Data
        </Button>
      </div>
    {:else if appState.display === 'PROCESSING_FAILED'}
      <!-- Failed State -->
      <div class="text-center space-y-4">
        <P color="red" class="text-red-600 dark:text-red-400">
          {processingError || 'Processing failed. Please check the file and try again.'}
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