import { changeTracker } from '../../../lib/changeTracker.svelte';
import type { Person } from './populationStore.svelte';
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
    (change: any) => !(change.actionType?.includes('population.person') && 
                      (change.personId === person.id || change.data?.id === person.id))
  );
  
  if (action === 'create') {
    changeTracker.pendingChanges.push({
      actionType: 'population.person.add',
      tableName: editingSession.tableName,
      data: {
        id: person.id,
        coords: '0,0', // Default coordinates, should be updated when placed on map
        rawXML: convertPersonToXML(person)
      }
    } as any);
  } else if (action === 'update') {
    changeTracker.pendingChanges.push({
      actionType: 'population.person.update',
      tableName: editingSession.tableName,
      personId: person.id,
      planXML: convertPersonToXML(person)
    } as any);
  } else if (action === 'delete') {
    changeTracker.pendingChanges.push({
      actionType: 'population.person.delete',
      tableName: editingSession.tableName,
      personId: person.id
    } as any);
  }
}

export function trackActivityChange(personId: string, person: Person) {
  trackPersonChange(person, 'update');
}

export function removePersonChanges(personId: string) {
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change: any) => !(change.actionType?.includes('population.person') && 
                      (change.personId === personId || change.data?.id === personId))
  );
}