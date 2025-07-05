import React, { useState, useEffect } from 'react';
import { Modal, Button, Spin, message, Table, Popconfirm, Typography, Alert, Upload, Drawer } from 'antd';
import { DeleteOutlined, UploadOutlined, BugOutlined } from '@ant-design/icons';
import { GetProcessesByFile, DeleteProcess, ExitApplication, SelectFile, GetDebugLogs } from '@wailsjs/go/gui/App';

const { Text } = Typography;

const WelcomeModal = ({ isVisible, onNewProcess, onLoadProcess, filePath, editMode, setFilePath }) => {
  const [loading, setLoading] = useState(true);
  const [processes, setProcesses] = useState([]);
  const [selectedProcessId, setSelectedProcessId] = useState(null);
  const [dbError, setDbError] = useState(null);
  const [retrying, setRetrying] = useState(false);
  const [showDebugDrawer, setShowDebugDrawer] = useState(false);
  const [debugLogs, setDebugLogs] = useState({ cliLog: '', guiLog: '' });
  
  const fetchProcesses = async () => {
    if (!filePath) return;
    setLoading(true);
    setDbError(null);
    try {
      // For now, we get all processes but filter by edit mode on frontend
      // TODO: Update backend to accept editMode parameter
      const data = await GetProcessesByFile(filePath);
      // Filter processes by current edit mode
      // For backward compatibility, treat null/undefined edit_mode as 'population'
      const filteredData = editMode ? 
        (data || []).filter(p => (p.edit_mode || 'population') === editMode) : 
        (data || []);
      setProcesses(filteredData);
    } catch (error) {
      console.error('Error fetching processes:', error);
      setDbError('Cannot establish connection with database');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isVisible && filePath) {
      fetchProcesses();
    } else if (isVisible && !filePath) {
      setLoading(false);
    }
  }, [isVisible, filePath]);

  const handleDeleteProcess = async (processId) => {
    try {
      await DeleteProcess(processId);
      message.success('Process deleted successfully.');
      fetchProcesses();
      if (selectedProcessId === processId) {
        setSelectedProcessId(null);
      }
    } catch (error) {
      console.error('Error deleting process:', error);
      message.error(`Failed to delete process: ${error}`);
    }
  };

  const handleClose = async () => {
    try {
      await ExitApplication();
    } catch (error) {
      console.error('Error exiting application:', error);
      window.close();
    }
  };

  const handleLoadProcess = () => {
    const selectedProcess = processes.find(p => p.id === selectedProcessId);
    if (selectedProcess) {
      onLoadProcess(selectedProcess);
    }
  };

  const handleFileSelect = async () => {
    try {
      const selectedPath = await SelectFile();
      if (selectedPath) {
        setFilePath(selectedPath);
      }
    } catch (error) {
      console.error('Error selecting file:', error);
      message.error('Failed to select file');
    }
  };

  const handleRetry = async () => {
    setRetrying(true);
    await fetchProcesses();
    setRetrying(false);
  };

  const fetchDebugLogs = async () => {
    try {
      const logs = await GetDebugLogs();
      setDebugLogs(logs);
      setShowDebugDrawer(true);
    } catch (error) {
      message.error('Failed to fetch debug logs');
    }
  };

  const columns = [
    {
      title: '',
      key: 'checkbox',
      width: 50,
      render: () => null,
    },
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: 'Action',
      key: 'action',
      width: 80,
      align: 'center',
      render: (_, record) => (
        <Popconfirm
          title="Delete this process?"
          description="This will permanently delete the associated data."
          onConfirm={() => handleDeleteProcess(record.id)}
          okText="Yes, Delete"
          cancelText="No"
        >
          <Button danger icon={<DeleteOutlined />} size="small" />
        </Popconfirm>
      ),
    },
  ];

  const rowSelection = {
    type: 'radio',
    selectedRowKeys: selectedProcessId ? [selectedProcessId] : [],
    onChange: (selectedRowKeys) => {
      setSelectedProcessId(selectedRowKeys[0] || null);
    },
  };

  const renderContent = () => {
    // Scenario 4: Database connection error
    if (dbError) {
      return (
        <div style={{ textAlign: 'center' }}>
          <Alert
            message={dbError}
            type="error"
            showIcon
            style={{ marginBottom: 0 }}
          />
        </div>
      );
    }

    // Scenario 1: No file specified
    if (!filePath) {
      return (
        <div style={{ textAlign: 'center', padding: '1.5rem 0' }}>
          <Text style={{ display: 'block', marginBottom: '1.5rem', fontSize: '1rem' }}>
            You haven't specified a file yet. To get started please:
          </Text>
          <Button
            icon={<UploadOutlined />}
            size="large"
            onClick={handleFileSelect}
            style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}
          >
            Upload a file
          </Button>
        </div>
      );
    }

    // Scenario 2: File path known - First time process
    if (!loading && processes.length === 0) {
      return (
        <>
          <div style={{ marginBottom: '1rem' }}>
            <Text strong>File Path: </Text>
            <Text code>{filePath}</Text>
          </div>
          <div style={{ textAlign: 'center', padding: '1.5rem 0' }}>
            <Text style={{ fontSize: '1rem' }}>
              We were not able to find any previous process of this file. You need to{' '}
              <span style={{ backgroundColor: '#f3f4f6', color: '#3b82f6', padding: '0.125rem 0.5rem', borderRadius: '0.375rem', fontWeight: 500 }}>
                start New process
              </span>
            </Text>
          </div>
        </>
      );
    }

    // Scenario 3: File path known - Previous process exists
    if (!loading && processes.length > 0) {
      return (
        <>
          <div style={{ marginBottom: '1rem' }}>
            <Text strong>File Path: </Text>
            <Text code>{filePath}</Text>
          </div>
          <div style={{ marginBottom: '1rem' }}>
            <Text style={{ fontSize: '1rem' }}>
              We detected previous process of this file, either select to continue editing or start a new process
            </Text>
          </div>
          <Table
            columns={columns}
            dataSource={processes}
            rowKey="id"
            rowSelection={rowSelection}
            pagination={{
              pageSize: 10,
              showSizeChanger: false,
              showTotal: (total) => `Total ${total} processes`,
            }}
            size="small"
            onRow={(record) => ({
              onClick: () => setSelectedProcessId(record.id),
            })}
          />
        </>
      );
    }

    // Loading state
    return (
      <div style={{ textAlign: 'center', padding: '2rem' }}>
        <Spin tip="Loading processes..." />
      </div>
    );
  };

  const renderFooter = () => {
    // Scenario 4: Database error - only Cancel and Retry
    if (dbError) {
      return [
        <Button key="cancel" onClick={handleClose}>
          Cancel
        </Button>,
        <Button key="retry" type="primary" onClick={handleRetry} loading={retrying}>
          Retry
        </Button>,
      ];
    }

    // Scenario 1: No file - no footer buttons
    if (!filePath) {
      return null;
    }

    // Scenario 2: First time - Cancel and Start New Process
    if (!loading && processes.length === 0) {
      return [
        <Button key="cancel" onClick={handleClose}>
          Cancel
        </Button>,
        <Button key="new" type="primary" onClick={onNewProcess}>
          Start New Process
        </Button>,
      ];
    }

    // Scenario 3: Has processes - all three buttons
    if (!loading && processes.length > 0) {
      return [
        <Button key="cancel" onClick={handleClose}>
          Cancel
        </Button>,
        <Button key="new" onClick={onNewProcess}>
          Start New Process
        </Button>,
        <Button
          key="continue"
          type="primary"
          disabled={!selectedProcessId}
          onClick={handleLoadProcess}
        >
          Continue Last
        </Button>,
      ];
    }

    // Loading state - no buttons
    return null;
  };

  return (
    <>
      <Modal
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>Welcome To Ez-utils {editMode ? editMode.charAt(0).toUpperCase() + editMode.slice(1) : 'Population'} Edit v3</span>
            <Button
              icon={<BugOutlined />}
              size="small"
              onClick={fetchDebugLogs}
              style={{ marginRight: '2.5rem' }}
            >
              Debug Logs
            </Button>
          </div>
        }
        open={isVisible}
        onCancel={handleClose}
        width={600}
        footer={renderFooter()}
        maskClosable={false}
        closable={true}
        centered
      >
        <div style={{ padding: '1.5rem' }}>{renderContent()}</div>
      </Modal>
      
      <Drawer
        title="Debug Logs"
        placement="right"
        width={800}
        onClose={() => setShowDebugDrawer(false)}
        open={showDebugDrawer}
      >
        <div style={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
          <h3>CLI Log (ez-utils.log):</h3>
          <pre style={{ backgroundColor: '#f3f4f6', padding: '0.625rem', borderRadius: '0.375rem', maxHeight: '18rem', overflow: 'auto' }}>
            {debugLogs.cliLog || 'No CLI logs available'}
          </pre>
          
          <h3 style={{ marginTop: '1.25rem' }}>GUI Log (ez-utils-gui.log):</h3>
          <pre style={{ backgroundColor: '#f3f4f6', padding: '0.625rem', borderRadius: '0.375rem', maxHeight: '18rem', overflow: 'auto' }}>
            {debugLogs.guiLog || 'No GUI logs available'}
          </pre>
          
          <h3 style={{ marginTop: '1.25rem' }}>Current Configuration:</h3>
          <pre style={{ backgroundColor: '#f3f4f6', padding: '0.625rem', borderRadius: '0.375rem' }}>
            {JSON.stringify({ filePath, editMode }, null, 2)}
          </pre>
        </div>
      </Drawer>
    </>
  );
};

export default WelcomeModal;