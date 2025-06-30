import React, { useState, useEffect } from 'react';
import { Modal, Button, Spin, message, Table, Popconfirm, Typography } from 'antd';
import { DeleteOutlined, CheckCircleOutlined, SyncOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { GetProcessesByFile, DeleteProcess } from '@wailsjs/go/gui/App';

const { Text } = Typography;

const WelcomeModal = ({ isVisible, onClose, onNewProcess, onLoadProcess, filePath }) => {
  const [loading, setLoading] = useState(true);
  const [processes, setProcesses] = useState([]);
  const [selectedProcess, setSelectedProcess] = useState(null);

  const fetchProcesses = async () => {
    if (!filePath) return;
    setLoading(true);
    try {
      const data = await GetProcessesByFile(filePath);
      setProcesses(data || []);
    } catch (error) {
      console.error('Error fetching processes:', error);
      message.error('Failed to load process list.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isVisible && filePath) {
      fetchProcesses();
    }
  }, [isVisible, filePath]);

  const handleDeleteProcess = async (processId) => {
    try {
      await DeleteProcess(processId);
      message.success('Process deleted successfully.');
      fetchProcesses(); // Refresh the list
      setSelectedProcess(null); // Deselect after deletion
    } catch (error) {
      console.error('Error deleting process:', error);
      message.error(`Failed to delete process: ${error}`);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'Completed':
        return <CheckCircleOutlined style={{ color: 'green' }} />;
      case 'started':
      case 'processing':
        return <SyncOutlined spin style={{ color: 'blue' }} />;
      default:
        return <ExclamationCircleOutlined style={{ color: 'red' }} />;
    }
  };


  const columns = [
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => (
        <div style={{ textAlign: 'center' }}>
          {getStatusIcon(status)} {status}
        </div>
      )
    },
    { title: 'Table Name', dataIndex: 'table_name', key: 'table_name' },
    { title: 'Record Count', dataIndex: 'record_count', key: 'record_count', align: 'center' },
    {
      title: 'Created At',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: 'Action',
      key: 'action',
      align: 'center',
      render: (_, record) => (
        <Popconfirm
          title="Delete this process?"
          description="This will permanently delete the associated data."
          onConfirm={() => handleDeleteProcess(record.id)}
          okText="Yes, Delete"
          cancelText="No"
        >
          <Button danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const rowSelection = {
    type: 'radio',
    selectedRowKeys: selectedProcess ? [selectedProcess.id] : [],
    onChange: (_, selectedRows) => setSelectedProcess(selectedRows[0]),
  };

  const handleLoadProcess = () => {
    if (selectedProcess) {
      onLoadProcess(selectedProcess);
      onClose();
    }
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div style={{ textAlign: 'center', padding: '32px' }}>
          <Spin tip="Loading..." />
        </div>
      );
    }
    if (processes.length === 0) {
      return (
        <Text>
          No previously processed records found for this file. Click "Start New Process" to begin.
        </Text>
      );
    }
    return (
      <Table
        columns={columns}
        dataSource={processes}
        rowKey="id"
        rowSelection={rowSelection}
        pagination={false}
        onRow={(record) => ({
          onClick: () => setSelectedProcess(record),
        })}
      />
    );
  };

  return (
    <Modal
      title={`Select a Process (${Math.random().toString(36).substring(7)})`}
      open={isVisible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="exit" onClick={() => window.close()}>
          Exit
        </Button>,
        <Button key="new" onClick={onNewProcess}>
          Start New Process
        </Button>,
        <Button
          key="load"
          type="primary"
          disabled={!selectedProcess}
          onClick={handleLoadProcess}
        >
          Load Selected Process
        </Button>,
      ]}
    >
      {renderContent()}
    </Modal>
  );
};

export default WelcomeModal;