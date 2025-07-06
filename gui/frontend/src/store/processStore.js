import { create } from 'zustand';
import { message } from 'antd';
import { ProcessPopulationFile, CheckProcessingStatus, GetProcessTelemetry } from '@wailsjs/go/gui/App';
import useUiStore from './uiStore';
import usePersonStore from './personStore';

const useProcessStore = create((set, get) => ({
  // Process data
  selectedProcess: null,
  currentTableName: null,
  processStatus: 'Idle',

  // Telemetry data
  currentTelemetry: null,
  isPollingTelemetry: false,
  telemetryInterval: null,

  // Actions
  setSelectedProcess: (process) => set({
    selectedProcess: process,
    currentTableName: process ? `population_data_${process.id}` : null
  }),

  clearSelectedProcess: () => set({
    selectedProcess: null,
    currentTableName: null
  }),

  startProcessAndPollStatus: async (filePath, editMode) => {
    const { setShowLoadingModal, setShowWelcomeModal } = useUiStore.getState();
    const { loadPopulation } = usePersonStore.getState();
    const { startTelemetryPolling, stopTelemetryPolling, setSelectedProcess } = get();

    if (!filePath) {
      message.error('No file path specified');
      return;
    }

    setShowWelcomeModal(false);
    setShowLoadingModal(true, 'Starting new process...');

    try {
      const response = await ProcessPopulationFile(filePath);
      const processId = response.processId;

      startTelemetryPolling(processId);
      set({ processStatus: 'Processing' });

      const checkStatus = async () => {
        try {
          const status = await CheckProcessingStatus(processId);
          if (status === 'Completed') {
            stopTelemetryPolling();
            set({ processStatus: 'Completed' });
            
            const newProcess = {
              id: processId,
              table_name: `population_data_${processId}`,
              status: 'Completed',
              edit_mode: editMode
            };
            setSelectedProcess(newProcess);

            try {
              await loadPopulation(newProcess.table_name);
              setShowLoadingModal(false);
              message.success('Process completed successfully!');
            } catch (error) {
              setShowLoadingModal(false);
              message.error('Process completed but failed to load data');
            }
          } else if (status.includes('failed') || status.includes('error')) {
            stopTelemetryPolling();
            setShowLoadingModal(false);
            message.error(`Process failed: ${status}`);
            setShowWelcomeModal(true);
            set({ processStatus: 'Failed' });
          } else {
            setShowLoadingModal(true, status);
            setTimeout(checkStatus, 1000);
          }
        } catch (error) {
          stopTelemetryPolling();
          setShowLoadingModal(false);
          message.error('Failed to check process status');
          setShowWelcomeModal(true);
          set({ processStatus: 'Failed' });
        }
      };
      setTimeout(checkStatus, 1000);
    } catch (error) {
      stopTelemetryPolling();
      message.error('Failed to start new process');
      setShowLoadingModal(false);
      setShowWelcomeModal(true);
      set({ processStatus: 'Failed' });
    }
  },

  loadExistingProcess: async (process) => {
    const { setShowLoadingModal, setShowWelcomeModal } = useUiStore.getState();
    const { loadPopulation } = usePersonStore.getState();
    const { setSelectedProcess } = get();

    setShowWelcomeModal(false);
    setShowLoadingModal(true, 'Loading process data...');
    setSelectedProcess(process);
    
    try {
      await loadPopulation(process.table_name);
      setShowLoadingModal(false);
    } catch (error) {
      message.error('Failed to load process data');
      setShowLoadingModal(false);
      setShowWelcomeModal(true);
    }
  },

  // Telemetry actions
  startTelemetryPolling: (processId) => {
    const currentInterval = get().telemetryInterval;
    if (currentInterval) {
      clearInterval(currentInterval);
    }

    const interval = setInterval(async () => {
      try {
        const telemetry = await GetProcessTelemetry(processId);
        set({ currentTelemetry: telemetry });
      } catch (error) {
      }
    }, 1000);

    set({ isPollingTelemetry: true, telemetryInterval: interval });
  },

  stopTelemetryPolling: () => {
    const interval = get().telemetryInterval;
    if (interval) {
      clearInterval(interval);
    }
    set({
      isPollingTelemetry: false,
      telemetryInterval: null,
      currentTelemetry: null
    });
  },

  updateTelemetry: (telemetry) => set({ currentTelemetry: telemetry }),
}));

export default useProcessStore;