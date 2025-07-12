<script lang="ts">
  import { Button, Modal, Badge } from "flowbite-svelte";
  import { ExclamationCircleOutline, CogOutline, DownloadOutline, CloseOutline, FloppyDiskOutline } from "flowbite-svelte-icons";
  import { slide } from "svelte/transition";
  import { appState } from "../../store.svelte";
  import { ExitApplication, SyncChanges } from "../../../wailsjs/go/gui/App";
  import { changeTracker } from "../../lib/changeTracker.svelte";
  import SettingsModal from "../SettingsModal.svelte";
  
  let showExitConfirmation = $state(false);
  let showSettingsModal = $state(false);
  
  async function handleExitApplication() {
    try {
      await ExitApplication();
    } catch (error) {
      console.error('Failed to exit application:', error);
    }
  }
  
  function handleGoBackToWelcome() {
    showExitConfirmation = false;
    // Reset app state to show welcome screen
    appState.display = 'WELCOME';
    // Clear any pending changes
    changeTracker.pendingChanges = [];
  }
  
  function handleExport() {
    // TODO: Implement export functionality
    console.log('Export clicked');
  }
  
  async function handleSync() {
    if (changeTracker.pendingChanges.length === 0) {
      console.log('No changes to sync');
      return;
    }
    
    changeTracker.isSyncing = true;
    
    try {
      const result = await SyncChanges(changeTracker.pendingChanges);
      console.log('Successfully synced', changeTracker.pendingChanges.length, 'changes');
      
      // Clear pending changes after successful sync
      changeTracker.pendingChanges = [];
    } catch (error) {
      console.error('Sync failed:', error);
      // TODO: Show error toast or modal
    } finally {
      changeTracker.isSyncing = false;
    }
  }
</script>

<div class="flex items-center justify-end h-full">
  {#if appState.display === 'EDITING'}
    <div class="flex space-x-3">
      <Button 
        color={changeTracker.pendingChanges.length === 0 ? "alternative" : "blue"}
        outline
        size="sm"
        class="py-1 min-w-[140px]"
        onclick={handleSync}
        disabled={changeTracker.isSyncing || changeTracker.pendingChanges.length === 0}
      >
        <FloppyDiskOutline class="w-4 h-4 me-2" />
        {#if changeTracker.pendingChanges.length === 0}
          No Changes
        {:else}
          Save ({changeTracker.pendingChanges.length})
        {/if}
      </Button>
      
      <Button 
        color="green"
        outline
        size="sm"
        class="py-1"
        onclick={handleExport}
      >
        <DownloadOutline class="w-4 h-4 me-2" />
        Export
      </Button>
      
      <Button 
        color="dark"
        outline
        size="sm"
        class="py-1"
        onclick={() => showSettingsModal = true}
      >
        <CogOutline class="w-4 h-4 me-2" />
        Settings
      </Button>
      
      <Button 
        color="red"
        outline
        size="sm"
        class="py-1"
        onclick={() => showExitConfirmation = true}
      >
        <CloseOutline class="w-4 h-4 me-2" />
        Exit
      </Button>
    </div>
  {/if}
</div>

<!-- Exit confirmation modal -->
<Modal bind:open={showExitConfirmation} size="xs" autoclose={false} transition={slide}>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
    <h3 class="mb-2 text-lg font-semibold text-gray-700 dark:text-gray-300">Exit Options</h3>
    <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">Choose how you want to exit:</p>
    
    <div class="flex flex-col gap-3">
      <Button 
        color="alternative" 
        class="w-full"
        onclick={handleGoBackToWelcome}
      >
        Go Back to Welcome Screen
      </Button>
      
      <Button 
        color="red" 
        class="w-full"
        onclick={handleExitApplication}
      >
        Exit Application
      </Button>
      
      <Button 
        color="light" 
        class="w-full"
        onclick={() => showExitConfirmation = false}
      >
        Cancel
      </Button>
    </div>
  </div>
</Modal>

<!-- Settings modal -->
<SettingsModal 
  open={showSettingsModal} 
  onClose={() => showSettingsModal = false} 
/>