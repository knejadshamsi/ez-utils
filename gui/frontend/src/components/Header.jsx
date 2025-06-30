import React from 'react';
import { Layout } from 'antd';

const { Header: AntHeader } = Layout;

const Header = () => {
  return (
    <AntHeader style={{ display: 'flex', alignItems: 'center', backgroundColor: '#001529' }}>
      <div style={{ color: 'white', fontSize: '20px' }}>EZ-Utils Population Editor hahah</div>
    </AntHeader>
  );
};

export default Header;