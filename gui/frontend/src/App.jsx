import React, { useState, useEffect } from 'react';
import { Layout, message } from 'antd';
import { GetStartupConfig, ProcessPopulationFile, CheckProcessingStatus, GetPopulation } from '@wailsjs/go/gui/App';
import './App.css';
import WelcomeModal from './components/WelcomeModal';
import LoadingModal from './components/LoadingModal';
import MapView from './components/MapView';
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
    selectedProcess,
    currentTableName,
    startTelemetryPolling,
    stopTelemetryPolling
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
            setShowLoadingModal(false);
            message.success('Process completed successfully!');
            
            // Set the selected process
            setSelectedProcess({
              id: processId,
              table_name: `population_data_${processId}`,
              status: 'Completed'
            });
            
            // Load the data
            await loadPopulationData(`population_data_${processId}`);
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

  const headerTitle = editMode || 'EZ-Utils GUI';

  return (
    <Layout style={{ height: '100%' }}>
      <AntHeader style={{ display: 'flex', alignItems: 'center', backgroundColor: '#001529' }}>
        <div style={{ color: 'white', fontSize: '20px' }}>{headerTitle}</div>
      </AntHeader>
      <Content style={{ position: 'relative', padding: 0, flexGrow: 1 }}>
        <MapView populationData={populationData} />
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