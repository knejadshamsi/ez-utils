import { editingSession } from '$lib/stores/app.svelte.ts';

export function getCurrentProcessId(): number {
  return editingSession.processId;
}