// Define the type for display states
export type DisplayState = 'WELCOME' | 'PROCESSING' | 'EDITING' | 'EXPORTING';

// Define the type for sidebar states
export type SidebarState = 'HIDDEN' | 'COLLAPSED' | 'EXPANDED';

// Define the type for file edit modes
export type FileEditMode = 'POPULATION' | 'NETWORK' | 'PT';

// Create the reactive state for UI using Svelte 5 runes
export const appState = $state<{ 
  display: DisplayState,
  primarySidebar: SidebarState,
  secondarySidebar: SidebarState
}>({ 
  display: 'WELCOME',
  primarySidebar: 'EXPANDED',
  secondarySidebar: 'HIDDEN'
});

// Create separate state for command arguments
export const commandArgs = $state<{
  fileEditMode: FileEditMode,
  filePath: string | null
}>({
  fileEditMode: 'POPULATION',
  filePath: null
});