import { appState } from '$lib/stores/app.svelte';

export function getCurrentProcessId(): number {
  return appState.processId;
}