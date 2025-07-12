// Define the type for display states
export type DisplayState = 'WELCOME' | 'LOADING' | 'PROCESSING' | 'EDITING' | 'EXPORTING';

// Define the type for sidebar states
export type SidebarState = 'HIDDEN' | 'COLLAPSED' | 'EXPANDED';

// Define the type for file edit modes
export type FileEditMode = 'POPULATION' | 'NETWORK' | 'PT';

// Define the type for validation status
export type ValidationStatus = 'NOT' | 'VALIDATING' | 'FAIL' | 'VALIDATED_POPULATION' | 'VALIDATED_NETWORK' | 'VALIDATED_PT';

// Define the type for process status
export type ProcessStatus = 'PENDING' | 'INITIALIZING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';

// Create the reactive state for UI using Svelte 5 runes
export const appState = $state<{ 
  display: DisplayState,
  primarySidebar: SidebarState,
  secondarySidebar: SidebarState
}>({ 
  display: 'EDITING', // TEMPORARY: Changed from 'WELCOME' for development - CHANGE BACK TO 'WELCOME' BEFORE COMMIT
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
  fileEditMode: 'POPULATION', // TEMPORARY: Default to POPULATION for development
  filePath: '/dummy/population.xml', // TEMPORARY: Dummy path for development
  isFileEditModeProvided: true, // TEMPORARY: Set to true for development
  isFilePathProvided: true, // TEMPORARY: Set to true for development
  validationStatus: 'VALIDATED_POPULATION' // TEMPORARY: Skip validation for development
});

// TEMPORARY: Add state for current editing session - for development
export const editingSession = $state<{
  processId: number,
  tableName: string
}>({
  processId: 999, // TEMPORARY: Dummy process ID for development
  tableName: 'population_data_999' // TEMPORARY: Dummy table name for development
});