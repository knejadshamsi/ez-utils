import { editingSession } from '../../../store.svelte';

export function getCurrentProcessId(): number {
  return editingSession.processId;
}