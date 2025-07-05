import React, { useEffect, useState } from 'react';
import { Layout, Button } from 'antd';
import { GetStartupConfig } from '@wailsjs/go/gui/App';
import './App.css';
import './index.css';
import WelcomeModal from './components/WelcomeModal';
import LoadingModal from './components/LoadingModal';
import MapView from './components/MapView';
import PersonDrawer from './components/PersonDrawer';
import PlanDrawer from './components/PlanDrawer';
import useUiStore from './store/uiStore';
import useProcessStore from './store/processStore';
import usePersonStore from './store/personStore';

const { Header: AntHeader, Content } = Layout;

function App() {
  const { showWelcomeModal, setShowWelcomeModal, setShowLoadingModal, viewMode } = useUiStore();
  const { selectedProcess, startProcessAndPollStatus, loadExistingProcess } = useProcessStore();
  const { selectedPerson, persons, loadPopulation } = usePersonStore();

  const [filePath, setFilePath] = useState('');
  const [editMode, setEditMode] = useState('');

  useEffect(() => {
    const fetchStartupConfig = async () => {
      try {
        const config = await GetStartupConfig();
        if (config) {
            setFilePath(config.filePath || '');
            setEditMode(config.editMode || '');
            if (config.filePath) {
                // Assuming we want to load initial data if a file path is present in config
                // This part is an assumption based on typical app flow.
            }
        }
      } catch (error) {
        console.error("Error fetching startup config:", error);
      }
    };
    fetchStartupConfig();
  }, []);
  
  // This effect will run once on mount to load initial population data if needed.
  // The logic inside loadPopulation in the store should be idempotent or handle this appropriately.
  useEffect(() => {
    if (selectedProcess?.table_name) {
        loadPopulation(selectedProcess.table_name);
    }
  }, [selectedProcess, loadPopulation]);


  const handleNewProcess = () => {
    startProcessAndPollStatus(filePath, editMode);
  };

  const handleLoadProcess = (process) => {
    loadExistingProcess(process);
    setEditMode(process.edit_mode || 'population');
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
      <Content style={{ overflow: 'hidden' }}>
        <div id="main-container">
          {editMode === 'population' && selectedProcess?.table_name && <PersonDrawer />}
          <main id="map-view-container">
            <MapView populationData={persons} />
          </main>
          {selectedPerson && <PlanDrawer />}
        </div>
        <WelcomeModal
          isVisible={showWelcomeModal}
          onNewProcess={handleNewProcess}
          onLoadProcess={handleLoadProcess}
          filePath={filePath}
          editMode={editMode}
          setFilePath={setFilePath}
        />
        <LoadingModal onCancel={handleCancelLoading} />
      </Content>
    </Layout>
  );
}

export default App;