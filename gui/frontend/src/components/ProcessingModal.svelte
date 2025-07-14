<script lang="ts">
  import { Modal, P, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline } from "flowbite-svelte-icons";
  import { appState, commandArgs, type ProcessStatus } from "../store.svelte";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { PTService } from "../services/edit/pt/ptService";
  import { 
    ProcessPopulationFile, 
    ProcessNetworkFile, 
    ProcessPTFile,
    GetProcessTelemetry,
    GetPTTelemetry,
    CheckProcessingStatus
  } from "../../wailsjs/go/gui/App";

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
  
  let isCompleted = $state(false);
  let pollingInterval = $state<ReturnType<typeof setInterval> | null>(null);
  let currentProcessId = $state<number | null>(null);

  // Start processing based on file type
  async function startProcessing() {
    if (!commandArgs.filePath) return;
    
    isCompleted = false;

    try {
      let result;
      switch (commandArgs.fileEditMode) {
        case 'POPULATION':
          result = await ProcessPopulationFile(commandArgs.filePath);
          currentProcessId = result.processId || result.process_id;
          break;
        case 'NETWORK':
          result = await ProcessNetworkFile(commandArgs.filePath);
          currentProcessId = result.processId || result.process_id;
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
      clearPolling();
    }
  }

  // Start loading existing process
  async function startLoading() {
    if (!welcomeModalState.selectedProcessId) return;
    
    isCompleted = false;
    currentProcessId = welcomeModalState.selectedProcessId;

    try {
      // Check if processing is complete
      const status = await CheckProcessingStatus(currentProcessId);
      
      if (status === 'COMPLETED') {
        isCompleted = true;
      } else if (status === 'FAILED') {
        throw new Error('Process failed during processing');
      } else {
        throw new Error(`Process is still ${status}. Please wait for processing to complete.`);
      }
      
    } catch (error) {
      console.error('Loading failed:', error);
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
        
        if (status === 'COMPLETED') {
          isCompleted = true;
          clearPolling();
          return;
        } else if (status === 'FAILED') {
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

  // Handle begin editing button click
  async function handleBeginEditing() {
    clearPolling();
    
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

  // Effect to start operations when modal opens
  $effect(() => {
    if (appState.display === 'PROCESSING') {
      startProcessing();
    } else if (appState.display === 'LOADING') {
      startLoading();
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
  open={appState.display === 'LOADING' || appState.display === 'PROCESSING'} 
  permanent
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="md"
  placement="center"
>
  {#if appState.display === 'PROCESSING'}
    {#snippet header()}
      <h3 class="text-xl font-semibold text-gray-900 dark:text-white text-center flex items-center justify-center gap-3">
        {#if !isCompleted}
          <Spinner size="6" />
        {:else}
          <CheckCircleOutline class="w-6 h-6 text-green-500" />
        {/if}
        {isCompleted ? `${commandArgs.fileEditMode} File Processed` : `Processing ${commandArgs.fileEditMode} File`}
      </h3>
    {/snippet}
  {/if}
  
  <div class="space-y-6">
    {#if appState.display === 'LOADING'}
      <!-- Loading State -->
      <div class="text-center space-y-4">
        <div class="flex items-center justify-center gap-3">
          {#if !isCompleted}
            <Spinner size="8" />
            <P class="text-xl text-gray-600 dark:text-gray-400">
              Loading {commandArgs.fileEditMode} process data...
            </P>
          {:else}
            <CheckCircleOutline class="w-8 h-8 text-green-500" />
            <P class="text-xl text-gray-600 dark:text-gray-400">
              Process loaded successfully
            </P>
          {/if}
        </div>
      </div>
      
    {:else if appState.display === 'PROCESSING'}
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
    
    <!-- Begin Editing Button -->
    {#if isCompleted}
      <div class="text-center">
        <Button 
          color="primary" 
          size="lg"
          onclick={handleBeginEditing}
        >
          Begin Editing
        </Button>
      </div>
    {/if}
  </div>
</Modal>