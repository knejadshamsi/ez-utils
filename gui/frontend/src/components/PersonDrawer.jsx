import React, { useState } from 'react';
import { Drawer, Button, Space, message, Pagination, Modal } from 'antd';
import { PlusOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import PersonList from './PersonList';
import BboxFilter from './BboxFilter';
import useAppStore from '../store/appStore';

const PersonDrawer = () => {
  const { viewMode, setViewMode, persons, changeCounter, syncToDatabase } = useAppStore();
  const [addingPerson, setAddingPerson] = useState(false);
  const [middle, setMiddle] = useState(20);
  const pageSize = 40;
  
  const handleAddPerson = async () => {
    setAddingPerson(true);
    
    try {
      const createdPerson = await useAppStore.getState().createPerson();
      message.success(`Person ${createdPerson.id} created`);
      
      // Auto-select the new person to open PlanDrawer
      useAppStore.getState().setSelectedPerson(createdPerson);
      
    } catch (error) {
      console.error('Failed to add person:', error);
      message.error(error.message || 'Failed to add person');
    } finally {
      setAddingPerson(false);
    }
  };
  
  // Calculate current page and total pages
  const currentPage = Math.floor(middle / pageSize);
  const totalPages = Math.ceil((persons?.length || 0) / pageSize);
  
  const handleNextPage = () => {
    if (currentPage < totalPages - 1) {
      setMiddle(prev => prev + pageSize);
    }
  };
  
  const handlePrevPage = () => {
    if (currentPage > 0) {
      setMiddle(prev => prev - pageSize);
    }
  };
  
  const handlePageChange = (page) => {
    // Check if there are unsaved changes
    if (changeCounter > 0) {
      Modal.confirm({
        title: 'Unsaved Changes',
        icon: <ExclamationCircleOutlined />,
        content: `You have ${changeCounter} unsaved changes. Do you want to sync before changing pages?`,
        okText: 'Sync & Continue',
        cancelText: 'Continue Without Sync',
        onOk: async () => {
          await syncToDatabase();
          // page is 1-based from Ant Design Pagination
          const newMiddle = (page * pageSize) - (pageSize / 2);
          setMiddle(newMiddle);
        },
        onCancel: () => {
          // page is 1-based from Ant Design Pagination
          const newMiddle = (page * pageSize) - (pageSize / 2);
          setMiddle(newMiddle);
        }
      });
    } else {
      // page is 1-based from Ant Design Pagination
      const newMiddle = (page * pageSize) - (pageSize / 2);
      setMiddle(newMiddle);
    }
  };
  
  return (
    <Drawer
      title={
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>Population</span>
            <Button 
              size="small"
              type="default"
              icon={<PlusOutlined />}
              onClick={handleAddPerson}
              loading={addingPerson}
            >
              New Person
            </Button>
          </div>
          <Button.Group size="small" style={{ width: '100%' }}>
            <Button 
              type={viewMode === 'list' ? 'primary' : 'default'}
              onClick={() => setViewMode('list')}
              style={{ width: '50%' }}
            >
              List View
            </Button>
            <Button 
              type={viewMode === 'bbox' ? 'primary' : 'default'}
              onClick={() => setViewMode('bbox')}
              style={{ width: '50%' }}
            >
              Bbox Filter
            </Button>
          </Button.Group>
          {totalPages > 1 && (
            <Pagination
              simple
              current={currentPage + 1}
              total={persons?.length || 0}
              pageSize={pageSize}
              onChange={handlePageChange}
              size="small"
              style={{ margin: '0 auto' }}
            />
          )}
        </div>
      }
      placement="left"
      open={true}
      mask={false}
      width={320}
      closable={false}
      styles={{ 
        body: { padding: 0, display: 'flex', flexDirection: 'column' },
        wrapper: { top: '72px', height: 'calc(100% - 72px)' },
        header: { borderBottom: 'none' }
      }}
    >
      {viewMode === 'bbox' && <BboxFilter />}
      <PersonList 
        middle={middle}
        pageSize={pageSize}
        onRequestPageChange={setMiddle}
      />
    </Drawer>
  );
};

export default PersonDrawer;