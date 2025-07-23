import { appState } from '$lib/stores/app.svelte.ts';

export function getCurrentProcessId(): number {
  return appState.processId;
}