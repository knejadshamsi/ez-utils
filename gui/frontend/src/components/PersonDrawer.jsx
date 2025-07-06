import React, { useState, useCallback } from 'react';
import { Drawer, Button, Space, message, Pagination, Modal } from 'antd';
import { PlusOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import PersonList from './PersonList';
import BboxFilter from './BboxFilter';
import useUiStore from '../store/uiStore';
import usePersonStore from '../store/personStore';
import useProcessStore from '../store/processStore';

const PersonDrawer = () => {
  const { viewMode, setViewMode } = useUiStore();
  const { persons, changeCounter, syncToDatabase, createPerson, setSelectedPerson } = usePersonStore();
  const { selectedProcess } = useProcessStore();
  const [addingPerson, setAddingPerson] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 40;
  
  const handleAddPerson = useCallback(async () => {
    setAddingPerson(true);
    
    try {
      const createdPerson = await createPerson(selectedProcess.table_name);
      message.success(`Person ${createdPerson.id} created`);
      
      // Auto-select the new person to open PlanDrawer
      setSelectedPerson(createdPerson);
      
    } catch (error) {
      console.error('Failed to add person:', error);
      message.error(error.message || 'Failed to add person');
    } finally {
      setAddingPerson(false);
    }
  }, [createPerson, selectedProcess, setSelectedPerson]);
  
  const totalPages = Math.ceil((persons?.length || 0) / pageSize);

  const handlePageChange = (newPage) => {
    const doChange = () => setPage(newPage);

    if (changeCounter > 0) {
      Modal.confirm({
        title: 'Unsaved Changes',
        icon: <ExclamationCircleOutlined />,
        content: `You have ${changeCounter} unsaved changes. Do you want to sync before changing pages?`,
        okText: 'Sync & Continue',
        cancelText: 'Continue Without Sync',
        onOk: async () => {
          await syncToDatabase(selectedProcess.table_name);
          doChange();
        },
        onCancel: doChange
      });
    } else {
      doChange();
    }
  };

  const onRequestPageChange = (newMiddle) => {
    const newPage = Math.floor(newMiddle / pageSize) + 1;
    setPage(newPage);
  };
  
  return (
    <Drawer
      title={
        <div className="person-drawer-title">
          <div className="drawer-title-row">
            <span className="drawer-title-text">Population</span>
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
            <div className="pagination-container">
              <Pagination
                simple
                current={page}
                total={persons?.length || 0}
                pageSize={pageSize}
                onChange={handlePageChange}
                size="small"
              />
            </div>
          )}
        </div>
      }
      placement="left"
      open={true}
      mask={false}
      width={360}
      closable={false}
      className="person-drawer"
    >
      <div className="drawer-content">
        {viewMode === 'bbox' ? (
          <div className="bbox-filter-container">
            <BboxFilter />
          </div>
        ) : null}
        <PersonList
          middle={(page * pageSize) - (pageSize / 2)}
          pageSize={pageSize}
          onRequestPageChange={onRequestPageChange}
        />
      </div>
    </Drawer>
  );
};

export default PersonDrawer;