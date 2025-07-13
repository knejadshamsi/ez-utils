import { changeTracker } from '../../../lib/changeTracker.svelte';
import type { Person } from './populationStore.svelte';
import type { SyncAction } from '../../../lib/changeTracker.svelte';
import { editingSession } from '../../../store.svelte';

function convertPersonToXML(person: Person): string {
  // Convert person data to XML format
  let xml = `<person id="${person.id}">`;
  
  // Add plans
  person.plans.forEach(plan => {
    xml += `<plan type="${plan.type}">`;
    
    plan.activities.forEach((activity, index) => {
      xml += `<activity type="${activity.type}" start_time="${activity.startTime}" end_time="${activity.endTime}" x="${activity.location[0]}" y="${activity.location[1]}" />`;
      
      // Add leg if not last activity
      if (index < plan.activities.length - 1 && plan.legs[index]) {
        const leg = plan.legs[index];
        xml += `<leg mode="${leg.mode}" duration="${leg.duration}" />`;
      }
    });
    
    xml += `</plan>`;
  });
  
  xml += `</person>`;
  return xml;
}

export function trackPersonChange(person: Person, action: 'create' | 'update' | 'delete') {
  if (!editingSession.tableName) return;
  
  // Remove any existing changes for this person
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'population' && change.elementType === 'person') {
        if ('personId' in change && change.personId === person.id) return false;
        if ('data' in change && change.data.id === person.id) return false;
      }
      return true;
    }
  );
  
  if (action === 'create') {
    const addAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'add',
      tableName: editingSession.tableName,
      data: {
        id: person.id,
        coords: '0,0', // Default coordinates, should be updated when placed on map
        rawXML: convertPersonToXML(person)
      }
    };
    changeTracker.pendingChanges.push(addAction);
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'update',
      tableName: editingSession.tableName,
      personId: person.id,
      planXML: convertPersonToXML(person)
    };
    changeTracker.pendingChanges.push(updateAction);
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'delete',
      tableName: editingSession.tableName,
      personId: person.id
    };
    changeTracker.pendingChanges.push(deleteAction);
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