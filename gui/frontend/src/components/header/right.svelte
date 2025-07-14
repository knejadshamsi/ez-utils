<script lang="ts">
  import { Button, Modal, Badge } from "flowbite-svelte";
  import { ExclamationCircleOutline, CogOutline, DownloadOutline, CloseOutline, FloppyDiskOutline } from "flowbite-svelte-icons";
  import { slide } from "svelte/transition";
  import { appState, commandArgs, editingSession } from "$lib/stores/app.svelte.ts";
  import { ExitApplication, SyncChanges, SaveFile, ExportPopulationFile, ExportNetworkFile, ExportPTFile, GetExportInfo } from "@wailsjs/go/gui/App";
  import { changeTracker } from "../../lib/changeTracker.svelte";
  import { showSuccess, showError, showInfo, showWarning } from "../../lib/toast.svelte";
  import SettingsModal from "../modals/SettingsModal.svelte";
  
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
  
  async function handleExport() {
    try {
      // Get export info based on current edit mode
      const exportInfo = await GetExportInfo(commandArgs.fileEditMode);
      
      if (!exportInfo.canExport) {
        showWarning('Export is not available for this file type');
        return;
      }
      
      // Check if we have pending changes
      if (changeTracker.pendingChanges.length > 0) {
        const shouldSave = confirm('You have unsaved changes. Would you like to save them before exporting?');
        if (shouldSave) {
          await handleSync();
        }
      }
      
      // Open save file dialog
      const defaultFileName = exportInfo.defaultFileName || 'export';
      const fileType = exportInfo.exportTypes?.[0] || 'xml';
      const outputPath = await SaveFile(defaultFileName, fileType);
      
      if (!outputPath) {
        // User cancelled
        return;
      }
      
      // Export based on file type
      showInfo('Exporting data...');
      
      switch (commandArgs.fileEditMode) {
        case 'POPULATION':
          if (!editingSession.tableName) {
            showError('No population table available for export');
            return;
          }
          await ExportPopulationFile(editingSession.tableName, outputPath);
          break;
          
        case 'NETWORK':
          if (!editingSession.processId) {
            showError('No network process available for export');
            return;
          }
          // Export entire network (empty bounding box array means all data)
          await ExportNetworkFile(editingSession.processId, outputPath, []);
          break;
          
        case 'PT':
          if (!editingSession.processId) {
            showError('No PT process available for export');
            return;
          }
          await ExportPTFile(editingSession.processId, outputPath);
          break;
          
        default:
          showError('Unknown file type for export');
          return;
      }
      
      showSuccess(`Data exported successfully to ${outputPath.split('/').pop() || outputPath}`);
    } catch (error) {
      console.error('Export failed:', error);
      showError(`Export failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  async function handleSync() {
    if (changeTracker.pendingChanges.length === 0) {
      showInfo('No changes to sync');
      return;
    }
    
    changeTracker.isSyncing = true;
    const changeCount = changeTracker.pendingChanges.length;
    
    try {
      console.log('=== SYNC DEBUG INFO ===');
      console.log('Total pending changes:', changeCount);
      console.log('Pending changes structure:', JSON.stringify(changeTracker.pendingChanges, null, 2));
      console.log('Individual changes:');
      changeTracker.pendingChanges.forEach((change, index) => {
        console.log(`Change ${index}:`, change);
      });
      console.log('Calling SyncChanges with:', changeTracker.pendingChanges);
      
      const result = await SyncChanges(changeTracker.pendingChanges);
      console.log('SyncChanges result:', result);
      console.log('Successfully synced', changeCount, 'changes');
      
      // Clear pending changes after successful sync
      changeTracker.pendingChanges = [];
      
      // Show success notification
      showSuccess(`Successfully saved ${changeCount} change${changeCount > 1 ? 's' : ''}`);
    } catch (error) {
      console.error('Sync failed:', error);
      // Check if error has details about the specific failure
      if (error instanceof Error) {
        console.error('Error details:', error.message);
        if (error.stack) {
          console.error('Stack trace:', error.stack);
        }
      }
      // Show error notification with specific error message
      showError(`Failed to save changes: ${error instanceof Error ? error.message : 'Unknown error'}`);
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