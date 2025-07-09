<script lang="ts">
  import { Label, P, Helper } from "flowbite-svelte";
  import { commandArgs, type FileEditMode } from "../../store.svelte";
  import EzRadioGroup from "../EzRadioGroup.svelte";
  import EzRadioOption from "../EzRadioOption.svelte";

  // Handle file edit mode selection
  function handleEditModeSelect(value: string) {
    commandArgs.fileEditMode = value as FileEditMode;
    // User has made a selection, so mark it as provided
    commandArgs.isFileEditModeProvided = true;
  }
</script>

<div>
  <Label class="mb-3 block text-sm font-medium text-gray-900 dark:text-white">
    Select Edit Mode
  </Label>
  <P class="mb-2 text-sm text-gray-500 dark:text-gray-400">
    Choose the type of data you want to edit: population, network, or public transit.
  </P>
  
  {#if !commandArgs.isFileEditModeProvided}
    <Helper color="yellow" class="mb-3">
      No edit mode was provided. Using POPULATION as default. Change if needed.
    </Helper>
  {/if}
  
  <div class="flex justify-center">
    <EzRadioGroup 
      name="fileEditMode"
      bind:value={commandArgs.fileEditMode}
      onchange={handleEditModeSelect}
    >
      <EzRadioOption 
        name="fileEditMode" 
        value="POPULATION" 
        checked={commandArgs.fileEditMode === 'POPULATION'}
      >
        Population
      </EzRadioOption>
      <EzRadioOption 
        name="fileEditMode" 
        value="NETWORK" 
        checked={commandArgs.fileEditMode === 'NETWORK'}
      >
        Network
      </EzRadioOption>
      <EzRadioOption 
        name="fileEditMode" 
        value="PT" 
        checked={commandArgs.fileEditMode === 'PT'}
      >
        Public Transit
      </EzRadioOption>
    </EzRadioGroup>
  </div>
</div>