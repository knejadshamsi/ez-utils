import { SyncChanges } from '../../wailsjs/go/gui/App';
import { changeTracker } from './changeTracker.svelte';
import { showSuccess, showError } from './toast.svelte';

/**
 * Sync all pending changes to the backend
 * @returns Promise<boolean> - true if sync was successful, false otherwise
 */
export async function syncChanges(): Promise<boolean> {
  if (changeTracker.pendingChanges.length === 0) {
    return true; // Nothing to sync
  }
  
  changeTracker.isSyncing = true;
  const changeCount = changeTracker.pendingChanges.length;
  
  try {
    console.log('Syncing changes:', changeTracker.pendingChanges);
    await SyncChanges(changeTracker.pendingChanges);
    console.log('Successfully synced', changeCount, 'changes');
    
    // Clear pending changes after successful sync
    changeTracker.pendingChanges = [];
    
    // Show success notification
    showSuccess(`Successfully saved ${changeCount} change${changeCount > 1 ? 's' : ''}`);
    return true;
  } catch (error) {
    console.error('Sync failed:', error);
    // Show error notification
    showError(`Failed to save changes: ${error instanceof Error ? error.message : 'Unknown error'}`);
    return false;
  } finally {
    changeTracker.isSyncing = false;
  }
}