import React, { useState, useEffect, useRef } from 'react';
import { List, Empty, Spin, Badge, Button, Space, message } from 'antd';
import { UserOutlined, DeleteOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';
import { GetPopulation, GetPopulationByBbox, DeletePerson } from '../../wailsjs/go/gui/App';

const PersonList = ({ middle, pageSize, onRequestPageChange }) => {
  const { 
    selectedProcess, 
    viewMode, 
    visiblePersons, 
    togglePersonVisibility,
    selectedPerson,
    setSelectedPerson,
    bboxFilter,
    persons: globalPersons
  } = useAppStore();
  
  const [persons, setPersons] = useState([]);
  const [loading, setLoading] = useState(false);
  const [deletingPersonId, setDeletingPersonId] = useState(null);
  const listRef = useRef(null);
  
  useEffect(() => {
    if (selectedProcess?.table_name) {
      loadPersons();
    }
  }, [selectedProcess, viewMode, bboxFilter]);
  
  // Watch for updates from global store (e.g., when new person is added or updated)
  useEffect(() => {
    if (globalPersons && globalPersons.length > 0) {
      setPersons(globalPersons);
    }
  }, [globalPersons]);
  
  // Watch for newly selected person and scroll to it
  useEffect(() => {
    if (selectedPerson && persons.length > 0) {
      const personIndex = persons.findIndex(p => p.id === selectedPerson.id);
      if (personIndex !== -1) {
        // Calculate which middle value would show this person
        const targetPage = Math.floor(personIndex / pageSize);
        const targetMiddle = (targetPage * pageSize) + (pageSize / 2);
        
        if (targetMiddle !== middle) {
          // Request parent to change to the right page
          if (onRequestPageChange) {
            onRequestPageChange(targetMiddle);
          }
        }
        
        // Scroll to the person item after page change
        setTimeout(() => {
          const element = document.getElementById(`person-${selectedPerson.id}`);
          if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 100);
      }
    }
  }, [selectedPerson, persons, pageSize, middle, onRequestPageChange]);
  
  // Don't auto-select all persons - let user manually toggle visibility
  
  const loadPersons = async () => {
    setLoading(true);
    try {
      // Check for local modifications that should be preserved
      const currentPersons = useAppStore.getState().persons;
      const localModifiedPersons = currentPersons.filter(p => p.plans !== null);
      const localModifiedMap = new Map(localModifiedPersons.map(p => [p.id, p]));
      
      let data;
      if (viewMode === 'bbox' && bboxFilter) {
        data = await GetPopulationByBbox(
          selectedProcess.table_name,
          bboxFilter.minLat,
          bboxFilter.minLng,
          bboxFilter.maxLat,
          bboxFilter.maxLng
        );
      } else {
        data = await GetPopulation(selectedProcess.table_name);
      }
      
      // Merge with local modifications, ensuring plans field exists
      const mergedData = (data || []).map(person => {
        const localVersion = localModifiedMap.get(person.id);
        if (localVersion) {
          // Preserve local modifications
          return localVersion;
        }
        // Ensure new persons have plans field set to null
        return { ...person, plans: null };
      });
      
      setPersons(mergedData);
      useAppStore.getState().setPersons(mergedData);
    } catch (error) {
      console.error('Failed to load persons:', error);
      setPersons([]);
    } finally {
      setLoading(false);
    }
  };
  
  const handlePersonClick = (person) => {
    setSelectedPerson(person);
    // Automatically make the person visible when selected
    if (!visiblePersons.has(person.id)) {
      togglePersonVisibility(person.id);
    }
  };
  
  const handleVisibilityToggle = (person) => {
    togglePersonVisibility(person.id);
  };
  
  const handleDeletePerson = async (person) => {
    try {
      await DeletePerson(selectedProcess.table_name, person.id);
      message.success(`Person ${person.id} deleted successfully`);
      
      // If this was the selected person, clear selection
      if (selectedPerson?.id === person.id) {
        setSelectedPerson(null);
      }
      
      // Update local state immediately
      const updatedPersons = persons.filter(p => p.id !== person.id);
      setPersons(updatedPersons);
      useAppStore.getState().setPersons(updatedPersons);
      
      // Adjust middle if we deleted the last person on current view
      const minIndex = middle - (pageSize / 2);
      const maxIndex = middle + (pageSize / 2);
      if (updatedPersons.length > 0 && minIndex >= updatedPersons.length) {
        // Move to previous page
        const newMiddle = Math.max(pageSize / 2, middle - pageSize);
        if (onRequestPageChange) {
          onRequestPageChange(newMiddle);
        }
      }
    } catch (error) {
      console.error('Failed to delete person:', error);
      message.error('Failed to delete person');
      // Reload on error to ensure consistency
      loadPersons();
    } finally {
      setDeletingPersonId(null);
    }
  };
  
  // Calculate the slice range using middle
  const minIndex = Math.max(0, middle - (pageSize / 2));
  const maxIndex = middle + (pageSize / 2);
  const paginatedPersons = persons.slice(minIndex, maxIndex);
  
  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '40px' }}>
        <Spin size="large" />
      </div>
    );
  }
  
  if (persons.length === 0) {
    return (
      <Empty 
        description={viewMode === 'bbox' ? "No persons in bbox" : "No persons found"}
        style={{ padding: '40px' }}
      />
    );
  }
  
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <List
        dataSource={paginatedPersons}
        style={{ flex: 1, overflow: 'auto', padding: '0 16px' }}
        renderItem={(person) => {
          const isVisible = visiblePersons.has(person.id);
          const isSelected = selectedPerson?.id === person.id;
          const isDeleting = deletingPersonId === person.id;
          
          return (
            <List.Item
              id={`person-${person.id}`}
              key={person.id}
              style={{ 
                cursor: 'pointer',
                backgroundColor: isSelected ? '#e6f4ff' : 'white',
                padding: '12px 16px',
                margin: '4px 0',
                borderRadius: '4px',
                border: isSelected ? '1px solid #1890ff' : '1px solid transparent'
              }}
              onClick={() => handlePersonClick(person)}
            >
            <div style={{ display: 'flex', alignItems: 'center', width: '100%' }}>
              <List.Item.Meta
                style={{ flex: 1 }}
              avatar={
                <div 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleVisibilityToggle(person);
                  }}
                  style={{ cursor: 'pointer' }}
                >
                  <Badge 
                    dot={isVisible} 
                    color="#52c41a"
                    offset={[-5, 5]}
                  >
                    <UserOutlined style={{ 
                      fontSize: '20px', 
                      color: isVisible ? '#1890ff' : '#bfbfbf' 
                    }} />
                  </Badge>
                </div>
              }
              title={`Person ${person.id}`}
              description={`Location: ${person.coords}`}
              />
              {isSelected && (
                <Space size="small">
                  {!isDeleting ? (
                    <Button
                      danger
                      size="small"
                      icon={<DeleteOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        setDeletingPersonId(person.id);
                      }}
                    />
                  ) : (
                    <>
                      <Button
                        type="primary"
                        size="small"
                        icon={<CheckOutlined />}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeletePerson(person);
                        }}
                      />
                      <Button
                        size="small"
                        icon={<CloseOutlined />}
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeletingPersonId(null);
                        }}
                      />
                    </>
                  )}
                </Space>
              )}
            </div>
            </List.Item>
          );
        }}
      />
    </div>
  );
};

export default PersonList;