import React from 'react';
import { Layout, Space } from 'antd';
import SyncButton from './SyncButton';
import useProcessStore from '../store/processStore';

const { Header: AntHeader } = Layout;

const Header = () => {
  const { selectedProcess } = useProcessStore();
  
  return (
    <AntHeader id="app-header">
      <div id="app-title">EZ-Utils Population Editor</div>
      {selectedProcess && (
        <Space>
          <SyncButton />
        </Space>
      )}
    </AntHeader>
  );
};

export default Header;