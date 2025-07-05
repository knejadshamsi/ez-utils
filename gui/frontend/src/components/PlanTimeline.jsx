import React, { useState, useMemo } from 'react';
import { Timeline, Empty, Button, message, Alert, Space, Input, InputNumber, Select, Popover, Row, Col } from 'antd';
import { HomeOutlined, ShoppingOutlined, CoffeeOutlined, CarOutlined, PlusOutlined, DeleteOutlined, WarningOutlined, EnvironmentOutlined, EditOutlined } from '@ant-design/icons';
import usePersonStore from '../store/personStore';

const PlanTimeline = ({ plan, onUpdate }) => {
  const { setMapClickHandler } = usePersonStore();
  const [selectedItemIndex, setSelectedItemIndex] = useState(null);
  const [addingActivity, setAddingActivity] = useState(false);
  const [addMode, setAddMode] = useState(null);

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
    const parts = timeStr.split(':');
    return parts.length === 3 ? timeStr : '00:00:00';
  };

  const checkTimeOverlap = (planItems) => {
    const activities = planItems.filter(item => item.type === 'activity');
    for (let i = 0; i < activities.length - 1; i++) {
      const current = activities[i];
      const next = activities[i + 1];
      if (current.end_time && next.start_time && current.end_time > next.start_time) {
        return { hasOverlap: true, message: `Time conflict detected for activity starting at ${next.start_time}.` };
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

  const adjustTime = (timeStr, hours) => {
    const [h, m, s] = timeStr.split(':').map(Number);
    let newHour = h + hours;
    if (newHour < 0) newHour = 0;
    if (newHour > 23) newHour = 23;
    return `${String(newHour).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  };

  const calculateDefaultTimes = (insertIndex) => {
    const prevActivity = plan.slice(0, insertIndex).reverse().find(item => item.type === 'activity');
    const nextActivity = plan.slice(insertIndex).find(item => item.type === 'activity');
    
    if (prevActivity && nextActivity) return { startTime: prevActivity.end_time || '08:00:00', endTime: nextActivity.start_time || '09:00:00' };
    if (prevActivity) {
        const lastEnd = prevActivity.end_time || '17:00:00';
        return { startTime: lastEnd, endTime: adjustTime(lastEnd, 1) };
    }
    if (nextActivity) {
        const firstStart = nextActivity.start_time || '08:00:00';
        return { startTime: adjustTime(firstStart, -1), endTime: firstStart };
    }
    return { startTime: '08:00:00', endTime: '09:00:00' };
  };

  const handleAddActivity = (mode) => {
    setAddingActivity(true);
    setAddMode(mode);
    message.info('Click on the map to add activity location');
    
    setMapClickHandler((coords) => {
      let insertIndex = plan.length;
      if ((mode === 'before' || mode === 'after') && selectedItemIndex !== null) {
        insertIndex = (mode === 'before') ? selectedItemIndex : selectedItemIndex + 1;
      }
      
      const defaultTimes = calculateDefaultTimes(insertIndex);
      const newActivity = {
        type: 'activity', activity_type: 'home',
        start_time: defaultTimes.startTime, end_time: defaultTimes.endTime,
        x: parseFloat(coords.lng.toFixed(6)), y: parseFloat(coords.lat.toFixed(6))
      };
      
      const newPlan = [...plan];
      newPlan.splice(insertIndex, 0, newActivity);

      if(insertIndex > 0 && newPlan[insertIndex-1]?.type === 'activity') {
        newPlan.splice(insertIndex, 0, {type: 'leg', mode: 'car', dep_time: newActivity.start_time, duration:'00:15:00'})
      }
      if(insertIndex < newPlan.length - 1 && newPlan[insertIndex+1]?.type ==='activity') {
        newPlan.splice(insertIndex+1,0, {type: 'leg', mode: 'car', dep_time: newActivity.end_time, duration:'00:15:00'})
      }

      onUpdate(newPlan);
      
      const overlapCheck = checkTimeOverlap(newPlan);
      if (overlapCheck.hasOverlap) message.warning(overlapCheck.message, 5);
      
      setAddingActivity(false);
      setAddMode(null);
      setMapClickHandler(null);
      message.success('Activity added');
      setSelectedItemIndex(null); 
    });
  };

  const renderActivityEditor = (item, index) => (
    <div className="editor-grid">
        <Space direction="vertical" style={{ gridColumn: 'span 2' }}>
            <label>Activity Type</label>
            <Select value={item.activity_type} onChange={(value) => updateItem(index, { activity_type: value })} style={{ width: '100%' }}>
              <Select.Option value="home">Home</Select.Option>
              <Select.Option value="work">Work</Select.Option>
              <Select.Option value="shop">Shop</Select.Option>
              <Select.Option value="leisure">Leisure</Select.Option>
            </Select>
        </Space>
        <Space direction="vertical">
            <label>Start Time</label>
            <Input value={item.start_time} onChange={(e) => updateItem(index, { start_time: e.target.value })}/>
        </Space>
        <Space direction="vertical">
            <label>End Time</label>
            <Input value={item.end_time} onChange={(e) => updateItem(index, { end_time: e.target.value })}/>
        </Space>
        <Space direction="vertical">
            <label>X-Coordinate</label>
            <InputNumber value={item.x} onChange={(value) => updateItem(index, { x: value })} step={0.000001} precision={6} style={{ width: '100%' }} />
        </Space>
        <Space direction="vertical">
            <label>Y-Coordinate</label>
            <InputNumber value={item.y} onChange={(value) => updateItem(index, { y: value })} step={0.000001} precision={6} style={{ width: '100%' }}/>
        </Space>
        <div style={{gridColumn: 'span 2'}}>
            <Button block icon={<PlusOutlined />} onClick={() => handleAddActivity('before')} size="small" style={{marginBottom: '0.5rem'}}>Add New Activity Before</Button>
            <Button block icon={<PlusOutlined />} onClick={() => handleAddActivity('after')} size="small">Add New Activity After</Button>
        </div>
        <Button icon={<DeleteOutlined />} danger onClick={() => deleteItem(index)} style={{gridColumn: 'span 2'}}>
            Delete Activity
        </Button>
    </div>
  );

  const renderLegEditor = (item, index) => (
    <div className="editor-grid">
      <Space direction="vertical" style={{ gridColumn: 'span 2' }}>
        <label>Mode</label>
        <Select value={item.mode} onChange={(value) => updateItem(index, { mode: value })} style={{ width: '100%' }}>
            <Select.Option value="car">Car</Select.Option>
            <Select.Option value="pt">PT</Select.Option>
            <Select.Option value="walk">Walk</Select.Option>
            <Select.Option value="bike">Bike</Select.Option>
        </Select>
      </Space>
      <Space direction="vertical">
        <label>Departure Time</label>
        <Input value={item.dep_time} onChange={(e) => updateItem(index, { dep_time: e.target.value })}/>
      </Space>
      <Space direction="vertical">
        <label>Duration</label>
        <Input value={item.duration} onChange={(e) => updateItem(index, { duration: e.target.value })}/>
      </Space>
      <Button icon={<DeleteOutlined />} danger onClick={() => deleteItem(index)} style={{gridColumn: 'span 2'}}>
        Delete Leg
      </Button>
    </div>
  );
  
  const overlapInfo = useMemo(() => checkTimeOverlap(plan || []), [plan]);

  if (!plan || activities.length === 0) {
    return (
      <div id="empty-plan-container">
        <Empty description="No activities in this plan" id="empty-plan">
          <Button type="primary" icon={<PlusOutlined />} onClick={() => handleAddActivity()} disabled={addingActivity} block>
            {addingActivity ? 'Click Map to Add' : 'Add First Activity'}
          </Button>
        </Empty>
      </div>
    );
  }
  
  return (
    <div id="plan-timeline">
      {overlapInfo.hasOverlap && (<Alert message={overlapInfo.message} type="warning" showIcon style={{ marginBottom: '1rem' }} closable />)}
      
      <div style={{ marginBottom: '1rem' }}>
        <Button block icon={<PlusOutlined />} onClick={() => handleAddActivity()} disabled={addingActivity}>Add Activity</Button>
      </div>
      
      <Timeline mode="left">
        {plan.map((item, index) => {
          if (item.type === 'activity') {
            return (
              <Timeline.Item key={index} dot={React.cloneElement(getActivityIcon(item.activity_type), { className: "timeline-icon" })}>
                  <div className="timeline-activity-item">
                    <div className="timeline-item-main-content">
                        <div className="timeline-item-body">
                           <div>{formatTime(item.start_time)} - {formatTime(item.end_time)}</div>
                           <div><span className="timeline-item-label">Type:</span> {item.activity_type || 'N/A'}</div>
                           {item.x && item.y && <div><span className="timeline-item-label">Location:</span> {item.x}, {item.y}</div>}
                        </div>
                    </div>
                    <Popover content={renderActivityEditor(item, index)} title="Edit Activity" trigger="click" open={selectedItemIndex === index} onOpenChange={(open) => setSelectedItemIndex(open ? index : null)} placement="right">
                        <Button type="text" icon={<EditOutlined/>} />
                    </Popover>
                  </div>
              </Timeline.Item>
            );
          } else if (item.type === 'leg') {
            return (
              <Timeline.Item key={index} dot={<div className="timeline-leg-dot" />}>
                <div className="timeline-leg-item">
                  <div className="timeline-item-main-content">
                      <div className="timeline-item-body">
                          <div><span className="timeline-item-label">Mode:</span> {item.mode}</div>
                          <div><span className="timeline-item-label">Departs:</span> {item.dep_time}</div>
                          <div><span className="timeline-item-label">Duration:</span> {item.duration}</div>
                      </div>
                  </div>
                  <Popover content={renderLegEditor(item, index)} title="Edit Leg" trigger="click" open={selectedItemIndex === index} onOpenChange={(open) => setSelectedItemIndex(open ? index : null)} placement="right">
                      <Button type="text" icon={<EditOutlined/>} />
                   </Popover>
                </div>
              </Timeline.Item>
            );
          }
          return null;
        })}
      </Timeline>
    </div>
  );
};

export default React.memo(PlanTimeline);