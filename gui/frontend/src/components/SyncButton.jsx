import React from 'react';
import { Button, Tooltip, Badge } from 'antd';
import { SyncOutlined, LoadingOutlined } from '@ant-design/icons';
import usePersonStore from '../store/personStore';
import useProcessStore from '../store/processStore';

const SyncButton = () => {
  const { changeCounter, syncThreshold, isSyncing, syncToDatabase, lastSyncTime } = usePersonStore();
  const { selectedProcess } = useProcessStore();
  
  const syncProgress = (changeCounter / syncThreshold) * 100;
  const needsSync = changeCounter > 0;
  
  // Format last sync time
  const formatLastSync = () => {
    if (!lastSyncTime) return 'Never synced';
    const now = Date.now();
    const diff = now - lastSyncTime;
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return 'Just now';
    if (minutes === 1) return '1 minute ago';
    if (minutes < 60) return `${minutes} minutes ago`;
    const hours = Math.floor(minutes / 60);
    if (hours === 1) return '1 hour ago';
    if (hours < 24) return `${hours} hours ago`;
    return 'More than 24 hours ago';
  };
  
  const tooltipContent = (
    <div>
      <div>Changes: {changeCounter}/{syncThreshold}</div>
      <div>Last sync: {formatLastSync()}</div>
      {changeCounter >= syncThreshold && <div style={{ color: '#f59e0b' }}>Auto-sync threshold reached!</div>}
    </div>
  );

  const buttonStyle = needsSync && !isSyncing ? {
    background: `linear-gradient(to right, #1890ff ${syncProgress}%, transparent ${syncProgress}%)`,
    borderColor: '#1890ff'
  } : {};

  return (
    <Tooltip title={tooltipContent}>
      <Badge count={needsSync ? changeCounter : 0} offset={[-5, 5]}>
        <Button
          type={needsSync ? "primary" : "default"}
          icon={isSyncing ? <LoadingOutlined spin /> : <SyncOutlined />}
          onClick={() => syncToDatabase(selectedProcess.table_name)}
          disabled={isSyncing || changeCounter === 0 || !selectedProcess}
          style={buttonStyle}
        >
          {isSyncing ? "Syncing..." : "Sync"}
        </Button>
      </Badge>
    </Tooltip>
  );
};

export default SyncButton;