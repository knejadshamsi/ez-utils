import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Button, Space, message, Alert } from 'antd';
import { EnvironmentOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';
import { GetPopulation } from '../../wailsjs/go/gui/App';
import { generateHighestPlusOneId, generateRandomId, detectIdPattern } from '../utils/idGenerator';

const AddPersonModal = ({ visible, onClose }) => {
  const { selectedProcess, setMapClickHandler, mapClickHandler } = useAppStore();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [locationMode, setLocationMode] = useState(false);
  const [idPattern, setIdPattern] = useState({ type: 'numeric', canUseHighestPlusOne: true });
  const [existingPersons, setExistingPersons] = useState([]);
  
  useEffect(() => {
    if (visible && selectedProcess) {
      loadExistingPersons();
    }
    
    return () => {
      // Clean up map click handler
      setMapClickHandler(null);
      setLocationMode(false);
    };
  }, [visible, selectedProcess]);
  
  const loadExistingPersons = async () => {
    try {
      const persons = await GetPopulation(selectedProcess.table_name);
      setExistingPersons(persons || []);
      const pattern = detectIdPattern(persons || []);
      setIdPattern(pattern);
    } catch (error) {
      console.error('Failed to load existing persons:', error);
    }
  };
  
  const handleMapClick = () => {
    setLocationMode(true);
    message.info('Click on the map to set the person location');
    
    const handler = (coords) => {
      form.setFieldsValue({
        x: coords.lng.toFixed(6),
        y: coords.lat.toFixed(6)
      });
      setLocationMode(false);
      setMapClickHandler(null);
      message.success('Location set successfully');
    };
    
    setMapClickHandler(handler);
  };
  
  const generateId = (method) => {
    if (method === 'highest') {
      const newId = generateHighestPlusOneId(existingPersons);
      if (newId) {
        form.setFieldsValue({ id: newId });
      } else {
        message.error('Cannot use Highest+1 method: non-numeric IDs found');
      }
    } else {
      const newId = generateRandomId(existingPersons);
      form.setFieldsValue({ id: newId });
    }
  };
  
  const handleSubmit = async (values) => {
    setLoading(true);
    try {
      await useAppStore.getState().createPerson({
        id: values.id,
        x: values.x,
        y: values.y,
        withHomeActivity: true,
        existingPersons
      });
      
      message.success('Person added successfully');
      
      // Reset form and close modal
      form.resetFields();
      onClose();
      
    } catch (error) {
      console.error('Failed to add person:', error);
      message.error(error.message || 'Failed to add person');
    } finally {
      setLoading(false);
    }
  };
  
  const handleCancel = () => {
    if (locationMode) {
      setLocationMode(false);
      setMapClickHandler(null);
    }
    form.resetFields();
    onClose();
  };
  
  return (
    <Modal
      title="Add New Person"
      open={visible}
      onCancel={handleCancel}
      footer={null}
      width={500}
      maskClosable={!locationMode}
      closable={!locationMode}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{
          id: '',
          x: '',
          y: ''
        }}
      >
        <Form.Item
          name="id"
          label="Person ID"
          rules={[{ required: true, message: 'Person ID is required' }]}
        >
          <Input placeholder="Enter person ID" />
        </Form.Item>
        
        <Space style={{ marginBottom: 16 }}>
          <Button 
            onClick={() => generateId('highest')}
            disabled={!idPattern.canUseHighestPlusOne}
          >
            Highest + 1
          </Button>
          <Button onClick={() => generateId('random')}>
            Random
          </Button>
        </Space>
        
        {!idPattern.canUseHighestPlusOne && (
          <Alert
            message="Non-numeric IDs detected"
            description="The 'Highest + 1' method is disabled because non-numeric person IDs were found."
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}
        
        <Form.Item label="Initial Location">
          <Space>
            <Form.Item
              name="x"
              rules={[{ required: true, message: 'X coordinate is required' }]}
              style={{ marginBottom: 0 }}
            >
              <Input placeholder="X coordinate" style={{ width: 150 }} />
            </Form.Item>
            
            <Form.Item
              name="y"
              rules={[{ required: true, message: 'Y coordinate is required' }]}
              style={{ marginBottom: 0 }}
            >
              <Input placeholder="Y coordinate" style={{ width: 150 }} />
            </Form.Item>
            
            <Button 
              icon={<EnvironmentOutlined />}
              onClick={handleMapClick}
              type={locationMode ? 'primary' : 'default'}
            >
              Set on Map
            </Button>
          </Space>
        </Form.Item>
        
        <Form.Item style={{ marginBottom: 0, marginTop: 24 }}>
          <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
            <Button onClick={handleCancel}>
              Cancel
            </Button>
            <Button type="primary" htmlType="submit" loading={loading}>
              Add Person
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default AddPersonModal;