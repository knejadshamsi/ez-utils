import React, { useState, useEffect } from 'react';
import { Drawer, Button, Spin, Empty, Tabs } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import PlanTimeline from './PlanTimeline';
import useAppStore from '../store/appStore';

const PlanDrawer = () => {
  const { selectedPerson, parsePersonPlans, updatePersonLocally } = useAppStore();
  const [loading, setLoading] = useState(false);
  const [activeKey, setActiveKey] = useState('0');
  
  const plans = selectedPerson?.plans || [];
  
  useEffect(() => {
    if (selectedPerson && selectedPerson.plans === null) {
      // Parse plans on demand when drawer opens
      setLoading(true);
      setTimeout(() => {
        parsePersonPlans(selectedPerson.id);
        setActiveKey('0');
        setLoading(false);
      }, 100);
    }
  }, [selectedPerson, parsePersonPlans]);
  
  // Update map route when plan is selected
  useEffect(() => {
    const planIndex = parseInt(activeKey);
    if (plans && plans.length > 0 && planIndex >= 0 && planIndex < plans.length) {
      const selectedPlan = plans[planIndex];
      const activities = selectedPlan ? selectedPlan.filter(item => item.type === 'activity') : [];
      
      // Create route path from activities
      if (activities.length > 1) {
        const path = activities
          .filter(a => a.x && a.y)
          .map(a => [a.x, a.y]);
        
        if (path.length > 1) {
          useAppStore.getState().setSelectedPlanRoute({ path });
        }
      }
    }
    
    return () => {
      useAppStore.getState().setSelectedPlanRoute(null);
    };
  }, [plans, activeKey]);
  
  const updatePersonPlans = (personId, newPlans) => {
    // Extract coordinates from first activity of first plan
    let newCoords = selectedPerson?.coords || "0,0";
    if (newPlans.length > 0 && newPlans[0].length > 0) {
      const firstActivity = newPlans[0].find(item => 
        item.type === 'activity' && item.x && item.y
      );
      if (firstActivity) {
        newCoords = `${firstActivity.x},${firstActivity.y}`;
      }
    }
    
    // Update person with new plans and coordinates
    updatePersonLocally(personId, {
      plans: newPlans,
      coords: newCoords
    });
  };
  
  const handlePlanUpdate = (updatedPlanArray) => {
    if (!selectedPerson) return;
    const planIndex = parseInt(activeKey);
    const newPlans = [...plans];
    newPlans[planIndex] = updatedPlanArray;
    updatePersonPlans(selectedPerson.id, newPlans);
  };
  
  const handleAddPlan = () => {
    if (!selectedPerson) return;
    const newPlan = []; // Empty array for new plan
    const newPlans = [...plans, newPlan];
    updatePersonPlans(selectedPerson.id, newPlans);
    
    // Switch to the new plan tab
    setActiveKey(String(newPlans.length - 1));
  };
  
  const handleRemovePlan = (targetKey) => {
    if (!selectedPerson) return;
    const planIndex = parseInt(targetKey);
    const newPlans = plans.filter((_, index) => index !== planIndex);
    updatePersonPlans(selectedPerson.id, newPlans);
    
    // Adjust active tab if needed
    if (newPlans.length > 0) {
      if (planIndex < parseInt(activeKey)) {
        setActiveKey(String(parseInt(activeKey) - 1));
      } else if (planIndex === parseInt(activeKey)) {
        setActiveKey('0');
      }
    }
  };
  
  if (!selectedPerson) {
    return null;
  }
  
  return (
    <Drawer
      title={`Plans - Person ${selectedPerson.id}`}
      placement="right"
      open={!!selectedPerson}
      onClose={() => useAppStore.getState().setSelectedPerson(null)}
      width={400}
      mask={false}
      styles={{ 
        body: { padding: 0 },
        wrapper: { top: '72px', height: 'calc(100% - 72px)' },
        header: { borderBottom: 'none' }
      }}
    >
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '40px' }}>
          <Spin size="large" />
        </div>
      ) : !plans || plans.length === 0 ? (
        <Empty
          description="No plans found"
          style={{ padding: '40px' }}
        >
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleAddPlan}
          >
            Add First Plan
          </Button>
        </Empty>
      ) : (
        <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', fontSize: '12px', color: '#888' }}>
            Auto-saves locally. Use sync button to save to database.
          </div>
          
          <Tabs
            activeKey={activeKey}
            onChange={setActiveKey}
            type="editable-card"
            onEdit={(targetKey, action) => {
              if (action === 'add') {
                handleAddPlan();
              } else if (action === 'remove') {
                handleRemovePlan(targetKey);
              }
            }}
            style={{ 
              height: '100%',
              display: 'flex',
              flexDirection: 'column'
            }}
            tabBarStyle={{ 
              marginBottom: 0,
              flexShrink: 0
            }}
            items={plans.map((plan, index) => ({
              key: String(index),
              label: `Plan ${index + 1}`,
              children: (
                <div style={{ height: '100%', overflow: 'auto' }}>
                  <PlanTimeline
                    plan={plan}
                    onUpdate={handlePlanUpdate}
                  />
                </div>
              )
            }))}
          />
        </div>
      )}
    </Drawer>
  );
};

export default PlanDrawer;