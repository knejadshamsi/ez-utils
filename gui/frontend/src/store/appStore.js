import { create } from 'zustand';

const useAppStore = create((set) => ({
  // Startup configuration
  filePath: '',
  editMode: '',
  startupError: null,
  
  // Modal states
  showWelcomeModal: true,
  showLoadingModal: false,
  loadingMessage: 'Loading...',
  
  // Process data
  selectedProcess: null,
  currentTableName: null,
  
  // Telemetry data
  currentTelemetry: null,
  isPollingTelemetry: false,
  telemetryInterval: null,
  
  // Actions
  setStartupConfig: (config) => set({
    filePath: config.filePath || '',
    editMode: config.editMode || '',
    startupError: config.error || null
  }),
  
  setShowWelcomeModal: (show) => set({ showWelcomeModal: show }),
  
  setShowLoadingModal: (show, message = 'Loading...') => set({ 
    showLoadingModal: show, 
    loadingMessage: message 
  }),
  
  setSelectedProcess: (process) => set({ 
    selectedProcess: process,
    currentTableName: process ? `population_data_${process.id}` : null
  }),
  
  clearSelectedProcess: () => set({ 
    selectedProcess: null,
    currentTableName: null
  }),
  
  // Telemetry actions
  startTelemetryPolling: (processId) => {
    const { GetProcessTelemetry } = window.go.gui.App;
    
    // Clear any existing interval
    const currentInterval = useAppStore.getState().telemetryInterval;
    if (currentInterval) {
      clearInterval(currentInterval);
    }
    
    // Start polling
    const interval = setInterval(async () => {
      try {
        const telemetry = await GetProcessTelemetry(processId);
        set({ currentTelemetry: telemetry });
      } catch (error) {
        console.error('Failed to fetch telemetry:', error);
      }
    }, 1000);
    
    set({ isPollingTelemetry: true, telemetryInterval: interval });
  },
  
  stopTelemetryPolling: () => {
    const interval = useAppStore.getState().telemetryInterval;
    if (interval) {
      clearInterval(interval);
    }
    set({ 
      isPollingTelemetry: false, 
      telemetryInterval: null,
      currentTelemetry: null
    });
  },
  
  updateTelemetry: (telemetry) => set({ currentTelemetry: telemetry })
}));

export default useAppStore;