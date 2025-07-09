<script lang="ts">
  import { Button } from "flowbite-svelte";
  import { welcomeModalState } from "./welcome.svelte";
  import { commandArgs } from "../../store.svelte";
  
  export let onNext: () => Promise<void>;
  export let onGoBack: () => void;
  export let onStart: () => void;
  export let onNewProcess: () => void;
</script>

<div class="w-full">
  
  <div class="flex w-full justify-end space-x-3">
    {#if !welcomeModalState.showTable}
      <Button 
        color="primary"
        disabled={!commandArgs.isFilePathProvided || welcomeModalState.isProcessFetching}
        onclick={onNext}
      >
        {welcomeModalState.isProcessFetching ? 'Loading...' : 'Next'}
      </Button>
    {:else}
      <Button 
        color="alternative"
        onclick={onGoBack}
      >
        Go Back
      </Button>
        <Button 
          color="primary"
          onclick={onNewProcess}
        >
          New Process
        </Button>
      {#if welcomeModalState.processes.length > 0}
        <Button 
          color="primary"
          disabled={!welcomeModalState.selectedProcessId}
          onclick={onStart}
        >
          Start Editing
        </Button>
      {/if}
    {/if}
  </div>
</div>