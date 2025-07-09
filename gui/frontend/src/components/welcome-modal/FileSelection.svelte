<script lang="ts">
  import { Fileupload, Label, Helper, P, Button } from "flowbite-svelte";
  import { commandArgs } from "../../store.svelte";
  
  // Local state for file selection
  let selectedFile = $state<File | null>(null);

  // Handle file selection from file input
  function handleFileChange(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      selectedFile = target.files[0];
      commandArgs.filePath = target.files[0].name;
      commandArgs.isFilePathProvided = true;
    }
  }

  // Handle change file button click
  function handleChangeFile() {
    selectedFile = null;
    commandArgs.filePath = '';
    commandArgs.isFilePathProvided = false;
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
      <P class="text-sm text-gray-700 dark:text-gray-300">
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
    <!-- Show file upload if no file provided -->
    <Helper color="yellow" class="mb-3">
      Please provide a file to continue
    </Helper>
    <Fileupload 
      id="file-upload" 
      onchange={handleFileChange}
      accept=".xml,.json,.csv"
    />
    {#if selectedFile}
      <Helper color="green" class="mt-2">
        Selected: {selectedFile.name}
      </Helper>
    {/if}
  {/if}
</div>