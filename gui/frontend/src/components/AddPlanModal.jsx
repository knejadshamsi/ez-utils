import React, { useState } from 'react';
import { Modal, Button, Space, message, Empty, Input } from 'antd';
import { PlusOutlined, EnvironmentOutlined, CodeOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';
import { XMLBuilder, XMLParser } from 'fast-xml-parser';

const AddPlanModal = ({ visible, onClose }) => {
  const { setMapClickHandler } = useAppStore();
  const [activities, setActivities] = useState([]);
  const [addingActivity, setAddingActivity] = useState(false);
  const [xmlMode, setXmlMode] = useState(false);
  const [xmlContent, setXmlContent] = useState('');
  
  const handleAddActivity = () => {
    setAddingActivity(true);
    message.info('Click on the map to add an activity location');
    
    setMapClickHandler((coords) => {
      const newActivity = {
        '@_type': 'home',
        '@_start_time': '08:00:00',
        '@_x': coords.lng.toFixed(6),
        '@_y': coords.lat.toFixed(6)
      };
      
      setActivities([...activities, newActivity]);
      setAddingActivity(false);
      setMapClickHandler(null);
      message.success('Activity added');
    });
  };
  
  const handleRemoveActivity = (index) => {
    const newActivities = activities.filter((_, i) => i !== index);
    setActivities(newActivities);
  };
  
  const toggleXmlMode = () => {
    if (xmlMode) {
      // Parse XML back to activities
      try {
        const parser = new XMLParser({
          ignoreAttributes: false,
          attributeNamePrefix: '@_',
          parseAttributeValue: true
        });
        
        const parsed = parser.parse(xmlContent);
        if (parsed.plan && parsed.plan.activity) {
          const parsedActivities = Array.isArray(parsed.plan.activity) 
            ? parsed.plan.activity 
            : [parsed.plan.activity];
          setActivities(parsedActivities);
          setXmlMode(false);
        }
      } catch (error) {
        message.error('Invalid XML format');
      }
    } else {
      // Convert activities to XML
      const builder = new XMLBuilder({
        ignoreAttributes: false,
        attributeNamePrefix: '@_',
        format: true,
        indentBy: '  '
      });
      
      const planObj = {
        plan: {
          '@_selected': 'no',
          activity: activities
        }
      };
      
      const xml = builder.build(planObj);
      setXmlContent(xml);
      setXmlMode(true);
    }
  };
  
  const handleSave = () => {
    if (activities.length === 0 && !xmlMode) {
      message.error('Please add at least one activity');
      return;
    }
    
    let finalActivities = activities;
    
    if (xmlMode) {
      // Parse XML if in XML mode
      try {
        const parser = new XMLParser({
          ignoreAttributes: false,
          attributeNamePrefix: '@_',
          parseAttributeValue: true
        });
        
        const parsed = parser.parse(xmlContent);
        if (parsed.plan && parsed.plan.activity) {
          finalActivities = Array.isArray(parsed.plan.activity) 
            ? parsed.plan.activity 
            : [parsed.plan.activity];
        }
      } catch (error) {
        message.error('Invalid XML format');
        return;
      }
    }
    
    const newPlan = {
      '@_selected': 'no',
      activity: finalActivities
    };
    
    // TODO: Implement plan saving logic
    message.success('Plan created successfully');
    handleCancel();
  };
  
  const handleCancel = () => {
    if (addingActivity) {
      setAddingActivity(false);
      setMapClickHandler(null);
    }
    setActivities([]);
    setXmlMode(false);
    setXmlContent('');
    onClose();
  };
  
  return (
    <Modal
      title="Add New Plan"
      open={visible}
      onCancel={handleCancel}
      width={600}
      footer={[
        <Button key="cancel" onClick={handleCancel}>
          Cancel
        </Button>,
        <Button 
          key="save" 
          type="primary" 
          onClick={handleSave}
          disabled={addingActivity}
        >
          Save Plan
        </Button>
      ]}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space>
          <Button 
            icon={<PlusOutlined />}
            onClick={handleAddActivity}
            disabled={xmlMode || addingActivity}
            type={addingActivity ? 'primary' : 'default'}
          >
            {addingActivity ? 'Click Map to Add' : 'Add Activity'}
          </Button>
          <Button 
            icon={<CodeOutlined />}
            onClick={toggleXmlMode}
          >
            {xmlMode ? 'Visual Mode' : 'XML Mode'}
          </Button>
        </Space>
        
        {xmlMode ? (
          <Input.TextArea
            value={xmlContent}
            onChange={(e) => setXmlContent(e.target.value)}
            rows={15}
            style={{ fontFamily: 'monospace', fontSize: '12px' }}
          />
        ) : activities.length === 0 ? (
          <Empty 
            description="No activities yet"
            style={{ padding: '40px 0' }}
          />
        ) : (
          <div style={{ maxHeight: '400px', overflow: 'auto' }}>
            {activities.map((activity, index) => (
              <div 
                key={index} 
                style={{ 
                  padding: '12px', 
                  border: '1px solid #f0f0f0', 
                  borderRadius: '4px',
                  marginBottom: '8px'
                }}
              >
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <div>
                    <div>
                      <strong>Activity {index + 1}</strong>
                    </div>
                    <div style={{ fontSize: '12px', color: '#666' }}>
                      Type: {activity['@_type']}, Time: {activity['@_start_time']}
                    </div>
                    <div style={{ fontSize: '12px', color: '#999' }}>
                      Location: {activity['@_x']}, {activity['@_y']}
                    </div>
                  </div>
                  <Button 
                    size="small" 
                    danger
                    onClick={() => handleRemoveActivity(index)}
                  >
                    Remove
                  </Button>
                </Space>
              </div>
            ))}
          </div>
        )}
      </Space>
    </Modal>
  );
};

export default AddPlanModal;