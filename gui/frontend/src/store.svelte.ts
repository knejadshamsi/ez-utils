// Define the type for display states
export type DisplayState = 'WELCOME' | 'PROCESSING' | 'EDITING' | 'EXPORTING';

// Define the type for sidebar states
export type SidebarState = 'HIDDEN' | 'COLLAPSED' | 'EXPANDED';

// Define the type for file edit modes
export type FileEditMode = 'POPULATION' | 'NETWORK' | 'PT';

// Define the type for validation status
export type ValidationStatus = 'NOT' | 'VALIDATING' | 'FAIL' | 'VALIDATED_POPULATION' | 'VALIDATED_NETWORK' | 'VALIDATED_PT';

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
  filePath: string,
  isFileEditModeProvided: boolean,
  isFilePathProvided: boolean,
  validationStatus: ValidationStatus
}>({
  fileEditMode: 'POPULATION',
  filePath: '',
  isFileEditModeProvided: false,
  isFilePathProvided: false,
  validationStatus: 'NOT'
});