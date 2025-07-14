import { changeTracker } from '$lib/changeTracker.svelte';
import type { Person } from '$lib/stores/population.svelte';
import type { SyncAction } from '$lib/changeTracker.svelte';
import { editingSession } from '$lib/stores/app.svelte.ts';

function convertPersonToXML(person: Person): string {
  // Convert person data to XML format
  let xml = `<person id="${person.id}">`;
  
  // Add plans
  person.plans.forEach(plan => {
    xml += `<plan type="${plan.type}">`;
    
    plan.activities.forEach((activity, index) => {
      xml += `<activity type="${activity.type}" start_time="${activity.startTime}" end_time="${activity.endTime}" x="${activity.location[0]}" y="${activity.location[1]}" />`;
      
      // Add leg if not last activity and not "person's choice"
      if (index < plan.activities.length - 1 && plan.legs[index]) {
        const leg = plan.legs[index];
        if (leg.mode !== "person's choice") {
          xml += `<leg mode="${leg.mode}" duration="${leg.duration}" />`;
        }
      }
    });
    
    xml += `</plan>`;
  });
  
  xml += `</person>`;
  return xml;
}

function extractCoordinatesFromPerson(person: Person): string {
  // Extract coordinates from the first activity of the first plan
  const firstActivity = person.plans[0]?.activities[0];
  if (firstActivity && firstActivity.location && (firstActivity.location[0] !== 0 || firstActivity.location[1] !== 0)) {
    return `${firstActivity.location[0]},${firstActivity.location[1]}`;
  }
  return '0,0';
}

export function trackPersonChange(person: Person, action: 'create' | 'update' | 'delete') {
  if (!editingSession.tableName) return;
  
  // Check if there's already a pending "add" action for this person
  const existingAddAction = changeTracker.pendingChanges.find(
    (change) => {
      if (change.type === 'population' && change.elementType === 'person' && change.action === 'add') {
        return 'data' in change && change.data.id === person.id;
      }
      return false;
    }
  );
  
  // If we have a pending "add" action and this is an "update", 
  // just update the "add" action instead of creating a separate "update"
  if (existingAddAction && action === 'update') {
    // Update the existing add action with new data
    const addActionIndex = changeTracker.pendingChanges.indexOf(existingAddAction);
    if (addActionIndex !== -1) {
      const updatedAddAction: SyncAction = {
        ...existingAddAction,
        data: {
          id: person.id,
          coords: extractCoordinatesFromPerson(person),
          rawXML: convertPersonToXML(person)
        }
      };
      changeTracker.pendingChanges[addActionIndex] = updatedAddAction;
      return;
    }
  }
  
  // Remove any existing changes for this person (only if not preserving add action)
  if (!(existingAddAction && action === 'update')) {
    changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
      (change) => {
        if (change.type === 'population' && change.elementType === 'person') {
          if ('personId' in change && change.personId === person.id) return false;
          if ('data' in change && change.data.id === person.id) return false;
        }
        return true;
      }
    );
  }
  
  if (action === 'create') {
    const addAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'add',
      tableName: editingSession.tableName,
      data: {
        id: person.id,
        coords: extractCoordinatesFromPerson(person),
        rawXML: convertPersonToXML(person)
      }
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'update',
      tableName: editingSession.tableName,
      personId: person.id,
      planXML: convertPersonToXML(person)
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'delete',
      tableName: editingSession.tableName,
      personId: person.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
  }
}

export function trackActivityChange(personId: string, person: Person) {
  trackPersonChange(person, 'update');
}

export function removePersonChanges(personId: string) {
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'population' && change.elementType === 'person') {
        if ('personId' in change && change.personId === personId) return false;
        if ('data' in change && change.data.id === personId) return false;
      }
      return true;
    }
  );
}