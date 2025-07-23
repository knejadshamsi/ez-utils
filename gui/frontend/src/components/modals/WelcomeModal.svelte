<script lang="ts">
  import { Modal, P } from "flowbite-svelte";
  import { appState, commandArgs, type FileEditMode } from "$lib/stores/app.svelte";
  import { onMount } from "svelte";
  import { GetStartupConfig, GetProcessesByFile, ValidatePopulationXML, ValidateNetworkXML, ValidatePTXML, StartProcessing } from "@wailsjs/go/gui/App";
  import SelectEditMode from "./welcome-modal/SelectEditMode.svelte";
  import FileSelection from "./welcome-modal/FileSelection.svelte";
  import ProcessTable from "./welcome-modal/ProcessTable.svelte";
  import WelcomeModalFooter from "./welcome-modal/WelcomeModalFooter.svelte";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";

  
  async function createNewProcess() {
    try {
      appState.display = 'PROCESSING';
      const result = await StartProcessing(commandArgs.filePath, commandArgs.fileEditMode);
      appState.processId = result.process_id;
    } catch (error) {
      appState.display = 'PROCESSING_FAILED';
    }
  }

  async function showProcessSelectionTable() {
    commandArgs.isFileEditModeProvided = true;
    welcomeModalState.isProcessFetching = true;
    
    try {
      const processes = await GetProcessesByFile(commandArgs.filePath);
      
      if (processes?.length > 0) {
        welcomeModalState.processes = processes;
        appState.processId = processes[0].process_id;
        welcomeModalState.showTable = true;
      } else {
        welcomeModalState.showTable = false;
      }
    } catch (error) {
      welcomeModalState.showTable = false;
    } finally {
      welcomeModalState.isProcessFetching = false;
    }
  }

  function returnToFileSelection() {
    const state = welcomeModalState;
    state.showTable = false;
    state.currentPage = 1;
    state.processes = [];
    appState.processId = 0;
  }

  async function validateFile() {
    if (!commandArgs.filePath) return;
    
    const validators = {
      POPULATION: ValidatePopulationXML,
      NETWORK: ValidateNetworkXML,
      PT: ValidatePTXML
    };
    
    const validator = validators[commandArgs.fileEditMode];
    
    commandArgs.validationStatus = 'VALIDATING';
    try {
      const result = await validator(commandArgs.filePath);
      commandArgs.validationStatus = result.isValid 
        ? `VALIDATED_${commandArgs.fileEditMode}` 
        : 'FAIL';
    } catch (error) {
      commandArgs.validationStatus = 'FAIL';
    }
  }

  onMount(async () => {
    try {
      const config = await GetStartupConfig();
      
      commandArgs.isFileEditModeProvided = !!config.editMode;
      if (config.editMode) {
        commandArgs.fileEditMode = config.editMode.toUpperCase() as FileEditMode;
      }
      
      commandArgs.isFilePathProvided = !!config.filePath;
      if (config.filePath) {
        commandArgs.filePath = config.filePath;
        await validateFile();
      }
    } catch (error) {
      commandArgs.isFileEditModeProvided = false;
      commandArgs.isFilePathProvided = false;
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
      onNext={showProcessSelectionTable}
      onGoBack={returnToFileSelection}
      onLoadExisting={() => appState.display = 'LOADING'}
      onNewProcess={createNewProcess}
    />
  {/snippet}
</Modal>