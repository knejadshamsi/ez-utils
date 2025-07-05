import React, { useState } from 'react';
import { Input, Button, Space, Form } from 'antd';
import { SearchOutlined, ClearOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';

const BboxFilter = () => {
  const { setBboxFilter } = useAppStore();
  const [form] = Form.useForm();
  
  const handleSearch = (values) => {
    const { minLat, minLng, maxLat, maxLng } = values;
    
    // Convert string values to numbers
    const bbox = {
      minLat: parseFloat(minLat),
      minLng: parseFloat(minLng),
      maxLat: parseFloat(maxLat),
      maxLng: parseFloat(maxLng)
    };
    
    // Validate all values are numbers
    if (Object.values(bbox).some(isNaN)) {
      return;
    }
    
    setBboxFilter(bbox);
  };
  
  const handleClear = () => {
    form.resetFields();
    setBboxFilter(null);
  };
  
  return (
    <div className="p-4 border-b">
      <Form
        form={form}
        onFinish={handleSearch}
        layout="vertical"
        initialValues={{
          minLat: '',
          minLng: '',
          maxLat: '',
          maxLng: ''
        }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
            <Form.Item
              name="minLat"
              label="Min Lat"
              rules={[
                { required: true, message: 'Required' },
                { pattern: /^-?\d*\.?\d+$/, message: 'Invalid number' }
              ]}
              style={{ marginBottom: 8 }}
            >
              <Input placeholder="-90.0" />
            </Form.Item>
            
            <Form.Item
              name="minLng"
              label="Min Lng"
              rules={[
                { required: true, message: 'Required' },
                { pattern: /^-?\d*\.?\d+$/, message: 'Invalid number' }
              ]}
              style={{ marginBottom: 8 }}
            >
              <Input placeholder="-180.0" />
            </Form.Item>
            
            <Form.Item
              name="maxLat"
              label="Max Lat"
              rules={[
                { required: true, message: 'Required' },
                { pattern: /^-?\d*\.?\d+$/, message: 'Invalid number' }
              ]}
              style={{ marginBottom: 8 }}
            >
              <Input placeholder="90.0" />
            </Form.Item>
            
            <Form.Item
              name="maxLng"
              label="Max Lng"
              rules={[
                { required: true, message: 'Required' },
                { pattern: /^-?\d*\.?\d+$/, message: 'Invalid number' }
              ]}
              style={{ marginBottom: 8 }}
            >
              <Input placeholder="180.0" />
            </Form.Item>
          </div>
          
          <Space style={{ width: '100%' }}>
            <Button 
              type="primary" 
              icon={<SearchOutlined />}
              htmlType="submit"
              block
              style={{ flex: 1 }}
            >
              Search
            </Button>
            <Button 
              icon={<ClearOutlined />}
              onClick={handleClear}
              block
              style={{ flex: 1 }}
            >
              Clear
            </Button>
          </Space>
        </Space>
      </Form>
    </div>
  );
};

export default BboxFilter;