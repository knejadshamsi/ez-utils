import React, { useState, useEffect, useRef } from 'react';
import { List, Empty, Spin, Badge, Button, Space, message } from 'antd';
import { UserOutlined, DeleteOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons';
import useUiStore from '../store/uiStore';
import useProcessStore from '../store/processStore';
import usePersonStore from '../store/personStore';
import { DeletePerson } from '../../wailsjs/go/gui/App';

const PersonList = ({ middle, pageSize, onRequestPageChange }) => {
  const { viewMode } = useUiStore();
  const { selectedProcess } = useProcessStore();
  const {
    persons,
    loadPopulation,
    visiblePersons,
    togglePersonVisibility,
    selectedPerson,
    setSelectedPerson,
    bboxFilter,
    setPersons,
  } = usePersonStore();
  
  const [loading, setLoading] = useState(false);
  const [deletingPersonId, setDeletingPersonId] = useState(null);
  const listRef = useRef(null);

  useEffect(() => {
    if (selectedProcess?.table_name) {
      loadPersons();
    }
  }, [selectedProcess, viewMode, bboxFilter]);

  // Watch for newly selected person and scroll to it
  useEffect(() => {
    if (!selectedPerson) return;
  
    const personIndex = persons.findIndex(p => p.id === selectedPerson.id);
    if (personIndex === -1) return;
  
    const targetPage = Math.floor(personIndex / pageSize);
    const newMiddle = (targetPage * pageSize) + (pageSize / 2);
  
    if (newMiddle !== middle) {
      // Defer scrolling until after the page has changed
      if (onRequestPageChange) {
        onRequestPageChange(newMiddle);
      }
    } else {
      // Scroll immediately if already on the correct page
      const element = document.getElementById(`person-${selectedPerson.id}`);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }
  }, [selectedPerson, persons, pageSize, middle, onRequestPageChange]);
  
  // Don't auto-select all persons - let user manually toggle visibility
  
  const loadPersons = async () => {
    setLoading(true);
    try {
      await loadPopulation(selectedProcess.table_name, viewMode === 'bbox' ? bboxFilter : null);
    } catch (error) {
      console.error('Failed to load persons:', error);
      message.error('Failed to load persons');
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
      <div className="spinner-container">
        <Spin size="large" />
      </div>
    );
  }
  
  if (persons.length === 0) {
    return (
      <Empty
        description={viewMode === 'bbox' ? "No persons in bbox" : "No persons found"}
        className="empty-person-list"
      />
    );
  }
  
  return (
    <div className="person-list-container">
      <List
        dataSource={paginatedPersons}
        className="person-list"
        renderItem={(person) => {
          const isVisible = visiblePersons.has(person.id);
          const isSelected = selectedPerson?.id === person.id;
          const isDeleting = deletingPersonId === person.id;
          
          return (
            <List.Item
              id={`person-${person.id}`}
              key={person.id}
              className={isSelected ? 'person-list-item selected' : 'person-list-item'}
              onClick={() => handlePersonClick(person)}
            >
            <div className="person-list-item-content">
              <div
                  onClick={(e) => {
                      e.stopPropagation();
                      handleVisibilityToggle(person);
                  }}
                  className="person-list-item-avatar"
              >
                  <Badge dot={isVisible} color="green" offset={[-3, 28]}>
                      <UserOutlined className={`user-icon ${isVisible ? 'visible' : ''}`} />
                  </Badge>
              </div>

              <List.Item.Meta
                className="person-list-item-meta"
                title={`Person ${person.id}`}
                description={`Location: ${person.coords}`}
              />
              {isSelected && (
                <Space size="small" align="end">
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

export default React.memo(PersonList);