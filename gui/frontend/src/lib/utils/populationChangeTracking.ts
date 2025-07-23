import { changeTracker } from '$lib/changeTracker.svelte';
import type { Person, PersonAttribute, Plan, Activity } from '$lib/stores/population.svelte';
import type { SyncAction } from '$lib/changeTracker.svelte';
import { appState } from '$lib/stores/app.svelte.ts';

function convertPersonToXML(person: Person): string {
  // Convert person data to XML format - compact for storage
  let xml = `<person id="${person.id}">`;
  
  // Add attributes if any exist and are included
  const includedAttributes = person.attributes.filter(attr => attr.included);
  if (includedAttributes.length > 0) {
    xml += `<attributes>`;
    includedAttributes.forEach(attr => {
      xml += `<attribute name="${attr.name}" class="${attr.type}">${attr.value}</attribute>`;
    });
    xml += `</attributes>`;
  }
  
  // Add plans
  person.plans.forEach(plan => {
    xml += `<plan selected="yes">`;
    
    plan.activities.forEach((activity, index) => {
      // Build activity attributes
      let activityAttrs = `type="${activity.type}"`;
      activityAttrs += ` x="${activity.location[0]}" y="${activity.location[1]}"`;
      
      // Add time attributes based on what's provided
      if (activity.startTime) {
        activityAttrs += ` start_time="${activity.startTime}"`;
      }
      if (activity.endTime) {
        activityAttrs += ` end_time="${activity.endTime}"`;
      }
      // TODO: Support duration attribute when UI allows it
      
      xml += `<activity ${activityAttrs}/>`;
      
      // Add leg between activities (not after last activity)
      if (index < plan.activities.length - 1) {
        const leg = plan.legs[index];
        // Only add leg if it has a mode AND it's not "person's choice"
        if (leg && leg.mode && leg.mode !== "person's choice") {
          xml += `<leg mode="${leg.mode}"`;
          if (leg.duration) {
            xml += ` duration="${leg.duration}"`;
          }
          xml += `/>`; 
        }
        // If no leg or "person's choice", don't add any leg element
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
  if (!appState.processId) return;
  
  const tableName = `population_data_${appState.processId}`;
  
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
      tableName: tableName,
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
      tableName: tableName,
      personId: person.id,
      planXML: convertPersonToXML(person)
    };
    console.log('[trackPersonChange] Creating update action:', updateAction);
    console.log('[trackPersonChange] Current pending changes before:', changeTracker.pendingChanges.length);
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
    console.log('[trackPersonChange] Current pending changes after:', changeTracker.pendingChanges.length);
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'population',
      elementType: 'person',
      action: 'delete',
      tableName: tableName,
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