import React, { useState, useEffect } from 'react';
import { Layout } from 'antd';
import { GetStartupFile, ProcessFile } from '@wailsjs/go/gui/App';
import './App.css';
import WelcomeModal from './components/WelcomeModal';
import MapView from './components/MapView';
const { Header: AntHeader, Content } = Layout;

function App() {
  const [showWelcomeModal, setShowWelcomeModal] = useState(true);
  const [filePath, setFilePath] = useState('');

  useEffect(() => {
    // This just ensures we have the file path for the WelcomeModal to use
    const fetchStartupFile = async () => {
      try {
        const file = await GetStartupFile();
        setFilePath(file);
      } catch (error) {
        console.error("Error fetching startup file:", error);
      }
    };
    fetchStartupFile();
  }, []);

  // These functions are placeholders to allow the WelcomeModal to operate without error.
  // They can be wired up to new functionality later.
  const handleNewProcess = async () => {
    console.log("Requesting new process for:", filePath);
    // This would trigger the backend processing and likely close the modal.
    try {
      await ProcessFile(filePath);
      // For now, we'll just close the modal as a demonstration.
      setShowWelcomeModal(false);
    } catch(error) {
        console.error("Error starting new process: ", error)
    }
  };

  const handleLoadProcess = (process) => {
    console.log("Loading existing process:", process);
    // This would load the selected data onto the map.
    setShowWelcomeModal(false);
  };

  const handleCloseWelcomeModal = () => {
    // Closing the modal without selecting a process can simply close the app.
    window.close();
  };

  return (
    <Layout style={{ height: '100%' }}>
      <AntHeader style={{ display: 'flex', alignItems: 'center', backgroundColor: '#001529' }}>
        <div style={{ color: 'white', fontSize: '20px' }}>EZ-Utils Population Editor</div>
      </AntHeader>
      <Content style={{ position: 'relative', padding: 0, flexGrow: 1 }}>
        <MapView />
        <WelcomeModal
          isVisible={showWelcomeModal}
          onClose={handleCloseWelcomeModal}
          onNewProcess={handleNewProcess}
          onLoadProcess={handleLoadProcess}
          filePath={filePath}
        />
      </Content>
    </Layout>
  );
}

export default App;
