import React from 'react';
import { Layout, Button } from 'antd';
import SyncButton from './SyncButton';
import useProcessStore from '../store/processStore';

const { Header: AntHeader } = Layout;

const Header = ({ title, onSettings, onExit }) => {
  const { selectedProcess } = useProcessStore();
  
  return (
    <AntHeader style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', backgroundColor: '#001529' }}>
      <div style={{ color: 'white', fontSize: '20px' }}>{title}</div>
      <div style={{ display: 'flex', gap: '8px' }}>
        {selectedProcess && <SyncButton />}
        <Button onClick={onSettings}>Settings</Button>
        <Button onClick={onExit} danger>Exit</Button>
      </div>
    </AntHeader>
  );
};

export default Header;