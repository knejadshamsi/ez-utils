<script lang="ts">
  import { Label, Helper, P, Button, Spinner } from "flowbite-svelte";
  import { InfoCircleOutline, ExclamationCircleOutline, CheckCircleOutline } from "flowbite-svelte-icons";
  import { commandArgs } from "$lib/stores/app.svelte.ts";
  import { ValidatePopulationXML, ValidateNetworkXML, ValidatePTXML, SelectFile } from "@wailsjs/go/gui/App";
  
  // Handle file selection using Wails native file dialog
  async function handleSelectFile() {
    try {
      const filePath = await SelectFile();
      if (filePath && filePath.trim() !== '') {
        commandArgs.filePath = filePath;
        commandArgs.isFilePathProvided = true;
        
        // Auto-validate the selected file
        await validateFile();
      }
    } catch (error) {
      console.error('Error selecting file:', error);
      commandArgs.validationStatus = 'FAIL';
    }
  }

  // Handle change file button click
  function handleChangeFile() {
    commandArgs.filePath = '';
    commandArgs.isFilePathProvided = false;
    commandArgs.validationStatus = 'NOT';
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

  // Validate the selected file
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
</script>

<div>
  <Label class="mb-3 block text-sm font-medium text-gray-900 dark:text-white">
    File Path
  </Label>
  <P class="mb-2 text-sm text-gray-500 dark:text-gray-400">
    Select the simulation file you want to edit.
  </P>
  
  {#if commandArgs.isFilePathProvided}
    <!-- Display existing file path -->
    <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-800">
      <P class="text-sm text-gray-700 dark:text-gray-300 break-all">
        {commandArgs.filePath}
      </P>
      <Button 
        size="sm" 
        color="alternative" 
        class="mt-3"
        onclick={handleChangeFile}
      >
        Change File
      </Button>
    </div>
  {:else}
    <!-- Show file selection button -->
    <div class="flex justify-center">
      <Button 
        color="alternative" 
        outline
        onclick={handleSelectFile}
        class="w-48"
      >
        Select XML File
      </Button>
    </div>
  {/if}

  <!-- Validation Status Display -->
  <div class="mt-4">
    {#if commandArgs.validationStatus === 'NOT'}
      <Helper color="gray" class="flex items-center gap-2">
        <InfoCircleOutline class="w-4 h-4" />
        Please select a file. Note that the file will be validated for proper XML structure.
      </Helper>
    {:else if commandArgs.validationStatus === 'VALIDATING'}
      <Helper color="blue" class="flex items-center gap-2">
        <Spinner size="4" />
        Validating XML structure. This might take a few moments.
      </Helper>
    {:else if commandArgs.validationStatus === 'FAIL'}
      <Helper color="red" class="flex items-center gap-2">
        <ExclamationCircleOutline class="w-4 h-4" />
        Validation failed. Please provide another file with proper XML structure.
      </Helper>
    {:else if commandArgs.validationStatus === `VALIDATED_${commandArgs.fileEditMode}`}
      <Helper color="green" class="flex items-center gap-2">
        <CheckCircleOutline class="w-4 h-4" />
        File successfully validated for {commandArgs.fileEditMode}.
      </Helper>
    {:else if commandArgs.validationStatus.startsWith('VALIDATED_')}
      <Helper color="yellow" class="flex items-center gap-2">
        <ExclamationCircleOutline class="w-4 h-4" />
        File is validated for {commandArgs.validationStatus.replace('VALIDATED_', '')}. Please reselect the file or select a different file.
      </Helper>
    {/if}
  </div>
</div>