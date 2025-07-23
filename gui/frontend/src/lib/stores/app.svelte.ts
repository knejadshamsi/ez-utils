// Define the type for display states - flat state machine for clear state management
export type DisplayState = 
  | 'WELCOME'
  | 'PROCESSING'
  | 'PROCESSING_SUCCESS'
  | 'PROCESSING_FAILED'
  | 'LOADING'
  | 'LOADING_SUCCESS'
  | 'LOADING_FAILED'
  | 'EDITING'
  | 'EXPORTING'
  | 'EXPORTING_FAILED';

// Define the type for sidebar states
export type SidebarState = 'HIDDEN' | 'COLLAPSED' | 'EXPANDED';

// Define the type for file edit modes
export type FileEditMode = 'POPULATION' | 'NETWORK' | 'PT';

// Define the type for network modes (View/Edit/Create)
export type NetworkMode = 'VIEW' | 'EDIT' | 'CREATE';

// Define the type for validation status
export type ValidationStatus = 'NOT' | 'VALIDATING' | 'FAIL' | 'VALIDATED_POPULATION' | 'VALIDATED_NETWORK' | 'VALIDATED_PT';

// Create the reactive state for UI using Svelte 5 runes
export const appState = $state<{ 
  display: DisplayState,
  primarySidebar: SidebarState,
  secondarySidebar: SidebarState,
  networkMode: NetworkMode,
  processId: number
}>({ 
  display: 'WELCOME',
  primarySidebar: 'EXPANDED',
  secondarySidebar: 'HIDDEN',
  networkMode: 'VIEW',
  processId: 0
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


// Map component reference for cross-component communication
export const mapComponent = $state<{
  startAddingStop?: (lineId: string, routeId: string) => void
}>({});