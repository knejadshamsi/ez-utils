import React, { useState, useEffect } from 'react';
import { Layout, message, Button } from 'antd';
import { GetStartupConfig, ProcessPopulationFile, CheckProcessingStatus, GetPopulation } from '@wailsjs/go/gui/App';
import './App.css';
import WelcomeModal from './components/WelcomeModal';
import LoadingModal from './components/LoadingModal';
import MapView from './components/MapView';
import PersonDrawer from './components/PersonDrawer';
import PlanDrawer from './components/PlanDrawer';
import useAppStore from './store/appStore';

const { Header: AntHeader, Content } = Layout;

function App() {
  const { 
    showWelcomeModal, 
    setShowWelcomeModal, 
    showLoadingModal,
    setShowLoadingModal,
    setStartupConfig,
    setSelectedProcess,
    filePath,
    editMode,
    setEditMode,
    selectedProcess,
    currentTableName,
    startTelemetryPolling,
    stopTelemetryPolling,
    selectedPerson
  } = useAppStore();

  const [populationData, setPopulationData] = useState([]);

  useEffect(() => {
    const fetchStartupConfig = async () => {
      try {
        const config = await GetStartupConfig();
        setStartupConfig(config);
        
        // If a valid file path was provided at startup and no error, 
        // keep the welcome modal open to let user choose action
        // The modal will handle showing appropriate UI based on file path
      } catch (error) {
        console.error("Error fetching startup config:", error);
      }
    };
    fetchStartupConfig();
  }, [setStartupConfig]);

  const handleNewProcess = async () => {
    if (!filePath) {
      message.error('No file path specified');
      return;
    }

    setShowWelcomeModal(false);
    setShowLoadingModal(true, 'Starting new process...');

    try {
      const response = await ProcessPopulationFile(filePath);
      const processId = response.processId;
      
      // Start telemetry polling
      startTelemetryPolling(processId);
      
      // Poll for process completion
      const checkStatus = async () => {
        try {
          const status = await CheckProcessingStatus(processId);
          
          if (status === 'Completed') {
            stopTelemetryPolling();
            
            // Set the selected process with edit mode
            setSelectedProcess({
              id: processId,
              table_name: `population_data_${processId}`,
              status: 'Completed',
              edit_mode: editMode
            });
            
            // Load the data first
            try {
              await loadPopulationData(`population_data_${processId}`);
              // Close modal after data is loaded
              setShowLoadingModal(false);
              message.success('Process completed successfully!');
            } catch (error) {
              console.error('Error loading data after process completion:', error);
              setShowLoadingModal(false);
              message.error('Process completed but failed to load data');
            }
            return; // Stop polling
          } else if (status.includes('failed') || status.includes('error')) {
            stopTelemetryPolling();
            setShowLoadingModal(false);
            message.error(`Process failed: ${status}`);
            setShowWelcomeModal(true);
          } else {
            // Update loading message with status
            setShowLoadingModal(true, status);
            // Continue polling
            setTimeout(checkStatus, 1000);
          }
        } catch (error) {
          console.error('Error checking process status:', error);
          stopTelemetryPolling();
          setShowLoadingModal(false);
          message.error('Failed to check process status');
          setShowWelcomeModal(true);
        }
      };
      
      // Start polling after a short delay
      setTimeout(checkStatus, 1000);
      
    } catch (error) {
      console.error("Error starting new process:", error);
      stopTelemetryPolling();
      message.error('Failed to start new process');
      setShowLoadingModal(false);
      setShowWelcomeModal(true);
    }
  };

  const handleLoadProcess = async (process) => {
    setShowWelcomeModal(false);
    setShowLoadingModal(true, 'Loading process data...');
    setSelectedProcess(process);
    
    // Set edit mode from the process
    setEditMode(process.edit_mode || 'population');
    
    try {
      await loadPopulationData(process.table_name);
      setShowLoadingModal(false);
    } catch (error) {
      console.error("Error loading process:", error);
      message.error('Failed to load process data');
      setShowLoadingModal(false);
      setShowWelcomeModal(true);
    }
  };

  const loadPopulationData = async (tableName) => {
    try {
      const data = await GetPopulation(tableName);
      setPopulationData(data || []);
      // Set persons data in the store for map visualization
      useAppStore.getState().setPersons(data || []);
      message.success(`Loaded ${data.length} records`);
    } catch (error) {
      console.error("Error loading population data:", error);
      message.error('Failed to load population data');
      throw error;
    }
  };

  const handleCancelLoading = () => {
    setShowLoadingModal(false);
    setShowWelcomeModal(true);
  };

  const headerTitle = editMode ? `EZ-Utils - ${editMode.charAt(0).toUpperCase() + editMode.slice(1)} Editor` : 'EZ-Utils GUI';

  const handleExit = () => {
    window.location.reload();
  };

  return (
    <Layout style={{ height: '100%' }}>
      <AntHeader style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', backgroundColor: '#001529' }}>
        <div style={{ color: 'white', fontSize: '20px' }}>{headerTitle}</div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <Button onClick={() => console.log('Settings')}>Settings</Button>
          <Button onClick={handleExit} danger>Exit</Button>
        </div>
      </AntHeader>
      <Content style={{ position: 'relative', padding: 0, flexGrow: 1 }}>
        <MapView populationData={populationData} />
        {editMode === 'population' && selectedProcess?.table_name && <PersonDrawer />}
        {selectedPerson && <PlanDrawer />}
        <WelcomeModal
          isVisible={showWelcomeModal}
          onNewProcess={handleNewProcess}
          onLoadProcess={handleLoadProcess}
        />
        <LoadingModal onCancel={handleCancelLoading} />
      </Content>
    </Layout>
  );
}

export default App;