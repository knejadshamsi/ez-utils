<script lang="ts">
  import { Button, Modal } from "flowbite-svelte";
  import { ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { slide } from "svelte/transition";
  import { welcomeModalState } from "./welcome.svelte";
  import { commandArgs, appState } from "$lib/stores/app.svelte.ts";
  import { ExitApplication } from "@wailsjs/go/gui/App";
  
  interface Props {
    onNext: () => Promise<void>;
    onGoBack: () => void;
    onStart: () => void;
    onNewProcess: () => void;
  }
  
  let { onNext, onGoBack, onStart, onNewProcess }: Props = $props();
  
  let showExitConfirmation = $state(false);
  
  async function handleExitConfirmation() {
    try {
      await ExitApplication();
    } catch (error) {
      console.error('Failed to exit application:', error);
    }
  }
</script>

<div class="w-full">
  
  <div class="flex w-full justify-between items-center">
    <Button 
      color="red" 
      outline
      disabled={welcomeModalState.isProcessFetching}
      onclick={() => showExitConfirmation = true}
    >
      Exit
    </Button>
    
    <div class="flex space-x-3">
    {#if !welcomeModalState.showTable}
      <Button 
        color="primary"
        disabled={!commandArgs.isFilePathProvided || commandArgs.validationStatus !== `VALIDATED_${commandArgs.fileEditMode}` || welcomeModalState.isProcessFetching}
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
          onclick={() => {
            appState.display = 'PROCESSING';
            onNewProcess();
          }}
        >
          New Process
        </Button>
      {#if welcomeModalState.processes.length > 0}
        <Button 
          color="primary"
          disabled={!welcomeModalState.selectedProcessId}
          onclick={() => {
            appState.display = 'LOADING';
            onStart();
          }}
        >
          Load
        </Button>
      {/if}
    {/if}
    </div>
  </div>
</div>

<Modal bind:open={showExitConfirmation} size="xs" autoclose transition={slide}>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">Are you sure you want to exit the application?</h3>
    <Button color="red" class="me-2" onclick={handleExitConfirmation}>Yes, Exit</Button>
    <Button color="alternative">Cancel</Button>
  </div>
</Modal>