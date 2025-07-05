import React, { useState, useEffect, useMemo } from 'react';
import { Timeline, Empty, Button, message, Alert, Space, Input, InputNumber, Select } from 'antd';
import { HomeOutlined, ShoppingOutlined, CoffeeOutlined, CarOutlined, PlusOutlined, DeleteOutlined, WarningOutlined, EnvironmentOutlined, ClockCircleOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';

const PlanTimeline = ({ plan, onUpdate }) => {
  const { setMapClickHandler } = useAppStore();
  const [selectedItemIndex, setSelectedItemIndex] = useState(null);
  const [addingActivity, setAddingActivity] = useState(false);
  const [addMode, setAddMode] = useState(null); // 'before', 'after', or null
  
  // Extract activities from plan array
  const activities = plan ? plan.filter(item => item.type === 'activity') : [];
  
  const getActivityIcon = (type) => {
    const iconMap = {
      'home': <HomeOutlined />,
      'work': <CoffeeOutlined />,
      'shop': <ShoppingOutlined />,
      'leisure': <CarOutlined />
    };
    return iconMap[type] || <CarOutlined />;
  };
  
  const formatTime = (timeStr) => {
    if (!timeStr) return '00:00:00';
    // Ensure time is in HH:MM:SS format
    const parts = timeStr.split(':');
    if (parts.length === 3) {
      return timeStr;
    }
    return '00:00:00';
  };
  
  const checkTimeOverlap = (planItems) => {
    const activities = planItems.filter(item => item.type === 'activity');
    for (let i = 0; i < activities.length - 1; i++) {
      const current = activities[i];
      const next = activities[i + 1];
      
      if (current.end_time && next.start_time) {
        if (current.end_time > next.start_time) {
          return {
            hasOverlap: true,
            index1: i,
            index2: i + 1,
            message: `Activity at ${current.start_time} ends at ${current.end_time}, but next activity starts at ${next.start_time}. Agent cannot be at different places at the same time!`
          };
        }
      }
    }
    return { hasOverlap: false };
  };
  
  const updateItem = (index, updates) => {
    const newPlan = [...plan];
    newPlan[index] = { ...newPlan[index], ...updates };
    onUpdate(newPlan);
  };
  
  const deleteItem = (index) => {
    const newPlan = plan.filter((_, i) => i !== index);
    onUpdate(newPlan);
    message.success('Item deleted');
  };
  
  // Helper function to add/subtract hours from time string
  const adjustTime = (timeStr, hours) => {
    const [h, m, s] = timeStr.split(':').map(Number);
    let newHour = h + hours;
    
    // Handle day boundaries
    if (newHour < 0) newHour = 0;
    if (newHour > 23) newHour = 23;
    
    return `${String(newHour).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  };
  
  const calculateDefaultTimes = (insertIndex) => {
    const prevActivity = plan.slice(0, insertIndex).reverse().find(item => item.type === 'activity');
    const nextActivity = plan.slice(insertIndex).find(item => item.type === 'activity');
    
    if (prevActivity && nextActivity) {
      // Insert between two activities
      return {
        startTime: prevActivity.end_time || prevActivity.start_time || '08:00:00',
        endTime: nextActivity.start_time || '09:00:00'
      };
    } else if (prevActivity) {
      // Insert after last activity
      const lastEnd = prevActivity.end_time || prevActivity.start_time || '17:00:00';
      return {
        startTime: lastEnd,
        endTime: adjustTime(lastEnd, 1)
      };
    } else if (nextActivity) {
      // Insert before first activity
      const firstStart = nextActivity.start_time || '08:00:00';
      return {
        startTime: adjustTime(firstStart, -1),
        endTime: firstStart
      };
    } else {
      // No activities
      return {
        startTime: '08:00:00',
        endTime: '09:00:00'
      };
    }
  };
  
  const handleAddActivity = (mode = null) => {
    setAddingActivity(true);
    setAddMode(mode);
    message.info('Click on the map to add activity location');
    
    setMapClickHandler((coords) => {
      let insertIndex = plan.length;
      
      if (mode === 'before' && selectedItemIndex !== null) {
        insertIndex = selectedItemIndex;
      } else if (mode === 'after' && selectedItemIndex !== null) {
        insertIndex = selectedItemIndex + 1;
      }
      
      const defaultTimes = calculateDefaultTimes(insertIndex);
      
      const newActivity = {
        type: 'activity',
        activity_type: 'home',
        start_time: defaultTimes.startTime,
        end_time: defaultTimes.endTime,
        x: parseFloat(coords.lng.toFixed(6)),
        y: parseFloat(coords.lat.toFixed(6))
      };
      
      // If inserting between activities, add a leg
      const newPlan = [...plan];
      if (insertIndex > 0 && insertIndex < plan.length) {
        // Check if there's already a leg between activities
        const prevIsActivity = insertIndex > 0 && plan[insertIndex - 1].type === 'activity';
        const nextIsActivity = insertIndex < plan.length && plan[insertIndex].type === 'activity';
        
        if (prevIsActivity && nextIsActivity) {
          // Need to add a leg before the new activity
          newPlan.splice(insertIndex, 0, {
            type: 'leg',
            mode: 'car',
            dep_time: defaultTimes.startTime,
            duration: '00:30:00'
          }, newActivity);
        } else {
          newPlan.splice(insertIndex, 0, newActivity);
        }
      } else {
        // Add leg if not at the beginning
        if (insertIndex > 0) {
          newPlan.push({
            type: 'leg',
            mode: 'car',
            dep_time: defaultTimes.startTime,
            duration: '00:30:00'
          });
        }
        newPlan.push(newActivity);
      }
      
      onUpdate(newPlan);
      
      // Check for overlaps
      const overlapCheck = checkTimeOverlap(newPlan);
      if (overlapCheck.hasOverlap) {
        message.warning(overlapCheck.message, 5);
      }
      
      setAddingActivity(false);
      setAddMode(null);
      setMapClickHandler(null);
      message.success('Activity added');
    });
  };
  
  const renderActivityEditor = (item, index) => {
    return (
      <div style={{ 
        background: '#f5f5f5', 
        padding: '12px', 
        borderRadius: '4px',
        marginTop: '8px'
      }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            <Select
              value={item.activity_type}
              onChange={(value) => updateItem(index, { activity_type: value })}
              style={{ width: 120 }}
            >
              <Select.Option value="home">Home</Select.Option>
              <Select.Option value="work">Work</Select.Option>
              <Select.Option value="shop">Shop</Select.Option>
              <Select.Option value="leisure">Leisure</Select.Option>
            </Select>
            <Button
              icon={<DeleteOutlined />}
              size="small"
              danger
              onClick={() => deleteItem(index)}
            >
              Delete
            </Button>
          </Space>
          
          <Space>
            <ClockCircleOutlined />
            <Input
              value={item.start_time}
              onChange={(e) => updateItem(index, { start_time: e.target.value })}
              placeholder="Start time"
              style={{ width: 100 }}
            />
            <span>to</span>
            <Input
              value={item.end_time}
              onChange={(e) => updateItem(index, { end_time: e.target.value })}
              placeholder="End time"
              style={{ width: 100 }}
            />
          </Space>
          
          <Space>
            <EnvironmentOutlined />
            <InputNumber
              value={item.x}
              onChange={(value) => updateItem(index, { x: value })}
              placeholder="X"
              style={{ width: 120 }}
              step={0.000001}
              precision={6}
            />
            <InputNumber
              value={item.y}
              onChange={(value) => updateItem(index, { y: value })}
              placeholder="Y"
              style={{ width: 120 }}
              step={0.000001}
              precision={6}
            />
          </Space>
        </Space>
      </div>
    );
  };
  
  const renderLegEditor = (item, index) => {
    return (
      <div style={{ 
        background: '#f0f0f0', 
        padding: '8px', 
        borderRadius: '4px',
        marginTop: '8px',
        marginLeft: '20px'
      }}>
        <Space>
          <Select
            value={item.mode}
            onChange={(value) => updateItem(index, { mode: value })}
            style={{ width: 100 }}
          >
            <Select.Option value="car">Car</Select.Option>
            <Select.Option value="pt">PT</Select.Option>
            <Select.Option value="walk">Walk</Select.Option>
            <Select.Option value="bike">Bike</Select.Option>
          </Select>
          <Input
            value={item.dep_time}
            onChange={(e) => updateItem(index, { dep_time: e.target.value })}
            placeholder="Dep. time"
            style={{ width: 100 }}
          />
          <Input
            value={item.duration}
            onChange={(e) => updateItem(index, { duration: e.target.value })}
            placeholder="Duration"
            style={{ width: 100 }}
          />
          <Button
            icon={<DeleteOutlined />}
            size="small"
            danger
            onClick={() => deleteItem(index)}
          />
        </Space>
      </div>
    );
  };
  
  // Check for overlaps whenever plan changes
  const overlapInfo = useMemo(() => {
    return checkTimeOverlap(plan || []);
  }, [plan]);
  
  if (!plan || activities.length === 0) {
    return (
      <div style={{ padding: '16px', height: '100%' }}>
        <Empty
          description="No activities in this plan"
          style={{ padding: '40px' }}
        >
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => handleAddActivity()}
            disabled={addingActivity}
            block
          >
            {addingActivity ? 'Click Map to Add' : 'Add First Activity'}
          </Button>
        </Empty>
      </div>
    );
  }
  
  let activityCounter = -1;
  
  return (
    <div style={{ padding: '16px', height: '100%' }}>
      {overlapInfo.hasOverlap && (
        <Alert
          message="Time Conflict Detected"
          description={overlapInfo.message}
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          style={{ marginBottom: '16px' }}
          closable
        />
      )}
      
      <div style={{ marginBottom: '16px' }}>
        <Space wrap>
          <Button 
            icon={<PlusOutlined />}
            onClick={() => handleAddActivity('before')}
            disabled={addingActivity}
            type={addingActivity && addMode === 'before' ? 'primary' : 'default'}
            size="small"
          >
            {addingActivity && addMode === 'before' ? 'Click Map' : 'Add Before'}
          </Button>
          <Button 
            icon={<PlusOutlined />}
            onClick={() => handleAddActivity('after')}
            disabled={addingActivity}
            type={addingActivity && addMode === 'after' ? 'primary' : 'default'}
            size="small"
          >
            {addingActivity && addMode === 'after' ? 'Click Map' : 'Add After'}
          </Button>
          <Button 
            icon={<PlusOutlined />}
            onClick={() => handleAddActivity()}
            disabled={addingActivity}
            size="small"
          >
            Add at End
          </Button>
        </Space>
      </div>
      
      <Timeline mode="left">
        {plan.map((item, index) => {
          if (item.type === 'activity') {
            activityCounter++;
            const isOverlapping = overlapInfo.hasOverlap && 
              (activityCounter === overlapInfo.index1 || activityCounter === overlapInfo.index2);
            
            return (
              <Timeline.Item
                key={index}
                dot={isOverlapping ? <WarningOutlined style={{ color: '#ff4d4f' }} /> : getActivityIcon(item.activity_type)}
                color={isOverlapping ? 'red' : selectedItemIndex === index ? 'blue' : 'gray'}
                style={{ cursor: 'pointer' }}
                onClick={() => setSelectedItemIndex(selectedItemIndex === index ? null : index)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ marginBottom: '8px' }}>
                      <strong>{formatTime(item.start_time)}</strong>
                      {item.end_time && ` - ${formatTime(item.end_time)}`}
                    </div>
                    <div style={{ color: '#666' }}>
                      Type: {item.activity_type || 'unknown'}
                    </div>
                    {item.x && item.y && (
                      <div style={{ color: '#999', fontSize: '12px' }}>
                        Location: {item.x}, {item.y}
                      </div>
                    )}
                  </div>
                </div>
                
                {selectedItemIndex === index && renderActivityEditor(item, index)}
              </Timeline.Item>
            );
          } else if (item.type === 'leg') {
            return (
              <div key={index} style={{ marginLeft: '50px', marginBottom: '16px' }}>
                <div 
                  style={{ 
                    color: '#888', 
                    fontSize: '12px',
                    cursor: 'pointer' 
                  }}
                  onClick={() => setSelectedItemIndex(selectedItemIndex === index ? null : index)}
                >
                  → {item.mode} | {item.dep_time} | {item.duration}
                </div>
                {selectedItemIndex === index && renderLegEditor(item, index)}
              </div>
            );
          }
          return null;
        })}
      </Timeline>
    </div>
  );
};

export default PlanTimeline;