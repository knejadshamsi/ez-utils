import React from 'react';
import { Layout, Space } from 'antd';
import SyncButton from './SyncButton';
import useAppStore from '../store/appStore';

const { Header: AntHeader } = Layout;

const Header = () => {
  const { selectedProcess } = useAppStore();
  
  return (
    <AntHeader style={{ 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'space-between',
      backgroundColor: '#001529',
      paddingLeft: '24px',
      paddingRight: '24px'
    }}>
      <div style={{ color: 'white', fontSize: '20px' }}>EZ-Utils Population Editor</div>
      {selectedProcess && (
        <Space>
          <SyncButton />
        </Space>
      )}
    </AntHeader>
  );
};

export default Header;