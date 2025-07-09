<script lang="ts">
  import { Button, Modal } from "flowbite-svelte";
  import { ExclamationCircleOutline, CogOutline, DownloadOutline, CloseOutline } from "flowbite-svelte-icons";
  import { slide } from "svelte/transition";
  import { appState } from "../../store.svelte";
  import { ExitApplication } from "../../../wailsjs/go/gui/App";
  import SettingsModal from "../SettingsModal.svelte";
  
  let showExitConfirmation = $state(false);
  let showSettingsModal = $state(false);
  
  async function handleExitConfirmation() {
    try {
      await ExitApplication();
    } catch (error) {
      console.error('Failed to exit application:', error);
    }
  }
  
  function handleExport() {
    // TODO: Implement export functionality
    console.log('Export clicked');
  }
</script>

<div class="flex items-center justify-end h-full">
  {#if appState.display === 'EDITING'}
    <div class="flex space-x-3">
      <Button 
        color="primary"
        outline
        size="sm"
        onclick={handleExport}
      >
        <DownloadOutline class="w-4 h-4 me-2" />
        Export
      </Button>
      
      <Button 
        color="alternative"
        outline
        size="sm"
        onclick={() => showSettingsModal = true}
      >
        <CogOutline class="w-4 h-4 me-2" />
        Settings
      </Button>
      
      <Button 
        color="red"
        outline
        size="sm"
        onclick={() => showExitConfirmation = true}
      >
        <CloseOutline class="w-4 h-4 me-2" />
        Exit
      </Button>
    </div>
  {/if}
</div>

<!-- Exit confirmation modal -->
<Modal bind:open={showExitConfirmation} size="xs" autoclose transition={slide}>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">Are you sure you want to exit the application?</h3>
    <Button color="red" class="me-2" onclick={handleExitConfirmation}>Yes, Exit</Button>
    <Button color="alternative">Cancel</Button>
  </div>
</Modal>

<!-- Settings modal -->
<SettingsModal 
  open={showSettingsModal} 
  onClose={() => showSettingsModal = false} 
/>