import React, { useState, useEffect } from 'react';
import { Drawer, Button, Spin, Empty, Tabs } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import PlanTimeline from './PlanTimeline';
import { calculateRouteFromPlan } from '../utils/mapUtils';
import usePersonStore from '../store/personStore';

const PlanDrawer = () => {
  const { selectedPerson, parsePersonPlans, updatePersonLocally, setSelectedPerson, setSelectedPlanRoute } = usePersonStore();
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
  
  const planIndex = parseInt(activeKey, 10);
  const selectedPlan = (plans && plans.length > 0 && planIndex >= 0 && planIndex < plans.length)
    ? plans[planIndex]
    : null;

  // Update map route when plan is selected
  useEffect(() => {
    const route = calculateRouteFromPlan(selectedPlan);
    setSelectedPlanRoute(route);

    return () => {
      setSelectedPlanRoute(null);
    };
  }, [selectedPlan, setSelectedPlanRoute]);
  
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
      title={<span className="drawer-title-text">{`Plans - Person ${selectedPerson.id}`}</span>}
      placement="right"
      open={!!selectedPerson}
      onClose={() => setSelectedPerson(null)}
      width={450}
      mask={false}
      className="plan-drawer"
    >
      {loading ? (
        <div className="spinner-container">
          <Spin size="large" />
        </div>
      ) : !plans || plans.length === 0 ? (
        <div className="empty-plans-container">
          <Empty description="No plans found for this person.">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAddPlan}
            >
              Add First Plan
            </Button>
          </Empty>
        </div>
      ) : (
        <div className="plan-tabs-container">
          <Tabs
            activeKey={activeKey}
            onChange={setActiveKey}
            type="editable-card"
            onEdit={(targetKey, action) => {
              if (action === 'add') handleAddPlan();
              else if (action === 'remove') handleRemovePlan(targetKey);
            }}
            className="plan-tabs"
            items={plans.map((plan, index) => ({
              key: String(index),
              label: `Plan ${index + 1}`,
              children: (
                <div className="plan-timeline-container">
                  <PlanTimeline plan={plan} onUpdate={handlePlanUpdate} />
                </div>
              ),
            }))}
          />
        </div>
      )}
    </Drawer>
  );
};

export default PlanDrawer;