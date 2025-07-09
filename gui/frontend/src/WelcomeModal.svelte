<script lang="ts">
  import { Modal, P } from "flowbite-svelte";
  import { appState, commandArgs, type FileEditMode } from "./store.svelte";
  import { onMount } from "svelte";
  import { GetStartupConfig, GetProcessesByFile } from "../wailsjs/go/gui/App";
  import SelectEditMode from "./components/welcome-modal/SelectEditMode.svelte";
  import FileSelection from "./components/welcome-modal/FileSelection.svelte";
  import ProcessTable from "./components/welcome-modal/ProcessTable.svelte";
  import WelcomeModalFooter from "./components/welcome-modal/WelcomeModalFooter.svelte";
  import { welcomeModalState } from "./components/welcome-modal/welcome.svelte";

  // Handle form submission
  function handleStart() {
    appState.display = 'EDITING';
  }
  
  // Handle creating new process
  function handleNewProcess() {
    appState.display = 'PROCESSING';
  }

  // Handle next button click
  async function handleNext() {
    console.log('handleNext called', { filePath: commandArgs.filePath, isFilePathProvided: commandArgs.isFilePathProvided });
    // Mark both as provided since user is proceeding with their choices
    commandArgs.isFileEditModeProvided = true;
    welcomeModalState.isProcessFetching = true;
    
    // Fetch processes for the selected file
    console.log('Calling GetProcessesByFile with:', commandArgs.filePath);
    const processes = await GetProcessesByFile(commandArgs.filePath);
    console.log('GetProcessesByFile returned:', processes);
    
    if (processes === null) {
      console.log('Error case: processes is null');
      // Handle error - don't show table
      welcomeModalState.showTable = false;
    } else {
      console.log('Success case: processes array length:', processes.length);
      // Set processes (empty array or filled array)
      welcomeModalState.processes = processes;
      
      if (welcomeModalState.processes.length > 0) {
        // Auto-select the first process
        welcomeModalState.selectedProcessId = (welcomeModalState.processes[0] as any).process_id;
      }
      
      welcomeModalState.showTable = true;
    }
    
    welcomeModalState.isProcessFetching = false;
  }

  // Handle go back button click
  function handleGoBack() {
    welcomeModalState.showTable = false;
    welcomeModalState.selectedProcessId = null;
    welcomeModalState.currentPage = 1;
    welcomeModalState.processes = [];
  }

  // Receive command arguments from Wails/Go backend
  onMount(async () => {
    try {
      const config = await GetStartupConfig();
      
      // Handle edit mode
      if (config.editMode && config.editMode !== '') {
        commandArgs.fileEditMode = config.editMode.toUpperCase() as FileEditMode;
        commandArgs.isFileEditModeProvided = true;
      } else {
        commandArgs.isFileEditModeProvided = false;
      }
      
      // Handle file path
      if (config.filePath && config.filePath !== '') {
        commandArgs.filePath = config.filePath;
        commandArgs.isFilePathProvided = true;
      } else {
        commandArgs.isFilePathProvided = false;
      }
    } catch (error) {
      console.error('Failed to get startup config from backend:', error);
    }
  });
</script>

<Modal 
  open={appState.display === 'WELCOME'} 
  permanent
  size="lg"
  placement="center"
  class="!w-full max-w-2xl"
  title="Welcome to Ez-utils GUI"
>
  <div class="space-y-6">
    {#if !welcomeModalState.showTable}
      <P class="text-gray-600 dark:text-gray-400">
        Ez-utils GUI is a visual editor for transportation simulation files.
        Configure your editing session below to get started.
      </P>

      <SelectEditMode />
      <FileSelection />
    {:else}
      <ProcessTable />
    {/if}
  </div>

  {#snippet footer()}
    <WelcomeModalFooter 
      onNext={handleNext}
      onGoBack={handleGoBack}
      onStart={handleStart}
      onNewProcess={handleNewProcess}
    />
  {/snippet}
</Modal>