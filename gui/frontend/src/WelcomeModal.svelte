<script lang="ts">
  import { Modal, P } from "flowbite-svelte";
  import { appState, commandArgs, type FileEditMode } from "./store.svelte";
  import { onMount } from "svelte";
  import { GetStartupConfig, GetProcessesByFile, ValidatePopulationXML, ValidateNetworkXML, ValidatePTXML } from "../wailsjs/go/gui/App";
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

  // Get the correct validator function based on file edit mode
  function getValidatorFunction() {
    switch (commandArgs.fileEditMode) {
      case 'POPULATION':
        return ValidatePopulationXML;
      case 'NETWORK':
        return ValidateNetworkXML;
      case 'PT':
        return ValidatePTXML;
      default:
        throw new Error(`Unknown file edit mode: ${commandArgs.fileEditMode}`);
    }
  }

  // Validate file (used for startup files)
  async function validateFile() {
    if (!commandArgs.filePath) return;
    
    commandArgs.validationStatus = 'VALIDATING';
    
    try {
      const validator = getValidatorFunction();
      const result = await validator(commandArgs.filePath);
      commandArgs.validationStatus = result.isValid ? `VALIDATED_${commandArgs.fileEditMode}` : 'FAIL';
      
      if (!result.isValid) {
        console.error('Validation failed:', result.error);
      }
    } catch (error) {
      console.error('Validation error:', error);
      commandArgs.validationStatus = 'FAIL';
    }
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
        
        // Auto-validate the startup file
        await validateFile();
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
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="lg"
  placement="center"
  class="!w-full max-w-2xl"
>
  {#snippet header()}
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
      Welcome to Ez-utils GUI
    </h3>
  {/snippet}
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