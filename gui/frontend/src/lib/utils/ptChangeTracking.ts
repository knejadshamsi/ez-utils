import { changeTracker } from '$lib/changeTracker.svelte';
import type { SyncAction } from '$lib/changeTracker.svelte';
import { getCurrentProcessId } from '$lib/utils/processId';
import type { Line, Stop, Departure } from '$lib/stores/pt.svelte';

export function trackStopChange(stop: Stop, action: 'save' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  if (action === 'delete') {
    // For delete, we only need to remove from route_stops (which cascades)
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'routeStop',
      action: 'delete',
      processId,
      stopId: stop.stopId,
      routeId: stop.routeId
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
    return;
  }
  
  if (action === 'save') {
    // Save involves two operations:
    // 1. Save the stop entity
    const stopEntity = {
      stopId: stop.stopId,
      stopName: stop.stopName,
      lat: stop.lat,
      lng: stop.lng,
      arrivalOffset: stop.arrivalOffset,
      departureOffset: stop.departureOffset,
      stopType: stop.stopType,
      wheelchairAccessible: stop.wheelchairAccessible,
      timingPoint: stop.timingPoint
    };
    
    const saveStopAction: SyncAction = {
      type: 'pt',
      elementType: 'stop',
      action: 'save',
      processId,
      stop: stopEntity
    };
    
    // 2. Save the route-stop junction
    const routeStop = {
      linkId: `rs_${stop.routeId}_${stop.stopId}`,
      routeId: stop.routeId,
      stopId: stop.stopId,
      sequence: stop.sequence
    };
    
    const saveRouteStopAction: SyncAction = {
      type: 'pt',
      elementType: 'routeStop',
      action: 'save',
      processId,
      routeStop: routeStop
    };
    
    // Remove any existing changes for this stop/route-stop
    changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
      (change) => {
        if (change.type === 'pt') {
          if (change.elementType === 'stop' && 'stop' in change && change.stop.stopId === stop.stopId) return false;
          if (change.elementType === 'routeStop' && 'routeStop' in change && change.routeStop.stopId === stop.stopId && change.routeStop.routeId === stop.routeId) return false;
        }
        return true;
      }
    );
    
    // Add both actions
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, saveStopAction, saveRouteStopAction];
  }
}

export function trackLineChange(line: Line, action: 'save' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  if (action === 'delete') {
    // Check if there's an existing 'add' action for this line
    const hasAddAction = changeTracker.pendingChanges.some(
      (change) => change.type === 'pt' && 
                  change.elementType === 'line' && 
                  change.action === 'add' &&
                  ('line' in change && change.line.id === line.id)
    );
    
    // Remove any existing changes for this line
    changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
      (change) => {
        if (change.type === 'pt' && change.elementType === 'line') {
          if ('lineId' in change && change.lineId === line.id) return false;
          if ('line' in change && change.line.id === line.id) return false;
        }
        return true;
      }
    );
    
    // If it was newly added (not yet in database), just remove it from pending changes
    if (hasAddAction) {
      return; // Don't add a delete action
    }
    
    // Otherwise, add a delete action
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'line',
      action: 'delete',
      processId,
      lineId: line.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
    return;
  }
  
  // Remove any existing changes for this line
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'line') {
        if ('lineId' in change && change.lineId === line.id) return false;
        if ('line' in change && change.line.id === line.id) return false;
      }
      return true;
    }
  );
  
  if (action === 'save') {
    // Strip sourceState before sending to backend
    const { sourceState, ...lineWithoutState } = line;
    const saveAction: SyncAction = {
      type: 'pt',
      elementType: 'line',
      action: 'save',
      processId,
      line: lineWithoutState
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, saveAction];
  }
}

// Route is now a separate entity with lineId
export function trackRouteChange(route: {id: string, name: string, lineId: string, sourceState?: string}, action: 'save' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  if (action === 'delete') {
    // Check if there's an existing 'add' action for this route
    const hasAddAction = changeTracker.pendingChanges.some(
      (change) => change.type === 'pt' && 
                  change.elementType === 'route' && 
                  change.action === 'add' &&
                  ('route' in change && change.route.id === route.id)
    );
    
    // Remove any existing changes for this route
    changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
      (change) => {
        if (change.type === 'pt' && change.elementType === 'route') {
          if ('routeId' in change && change.routeId === route.id) return false;
          if ('route' in change && change.route.id === route.id) return false;
        }
        return true;
      }
    );
    
    // If it was newly added (not yet in database), just remove it from pending changes
    if (hasAddAction) {
      return; // Don't add a delete action
    }
    
    // Otherwise, add a delete action
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'route',
      action: 'delete',
      processId,
      routeId: route.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
    return;
  }
  
  // Remove any existing changes for this route
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'route') {
        if ('routeId' in change && change.routeId === route.id) return false;
        if ('route' in change && change.route.id === route.id) return false;
      }
      return true;
    }
  );
  
  if (action === 'save') {
    // Strip sourceState before sending to backend
    const { sourceState, ...routeWithoutState } = route;
    const saveAction: SyncAction = {
      type: 'pt',
      elementType: 'route',
      action: 'save',
      processId,
      route: routeWithoutState
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, saveAction];
  }
}


export function trackBatchStopUpdates(updates: Record<string, Stop>) {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  // Remove any existing individual stop changes for the stops being batch updated
  const stopIds = Object.keys(updates);
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'stop') {
        if ('stopId' in change && stopIds.includes(change.stopId)) return false;
        if ('stop' in change && stopIds.includes(change.stop.stopId)) return false;
      }
      return true;
    }
  );
  
  // Strip sourceState from all stops in the batch
  const updatesWithoutState: Record<string, any> = {};
  for (const [id, stop] of Object.entries(updates)) {
    const { sourceState, ...stopWithoutState } = stop;
    updatesWithoutState[id] = stopWithoutState;
  }
  
  const batchAction: SyncAction = {
    type: 'pt',
    elementType: 'stop',
    action: 'batchUpdate',
    processId,
    updates: updatesWithoutState
  };
  
  changeTracker.pendingChanges = [...changeTracker.pendingChanges, batchAction];
}

export function removePTChanges(elementType: 'stop' | 'line' | 'route', id: string) {
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === elementType) {
        switch (elementType) {
          case 'stop':
            if ('stopId' in change && change.stopId === id) return false;
            if ('stop' in change && change.stop.stopId === id) return false;
            break;
          case 'line':
            if ('lineId' in change && change.lineId === id) return false;
            if ('line' in change && change.line.id === id) return false;
            break;
          case 'route':
            if ('routeId' in change && change.routeId === id) return false;
            if ('route' in change && change.route.id === id) return false;
            break;
        }
      }
      return true;
    }
  );
}

export function trackDepartureChange(departure: Departure, action: 'save' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  if (action === 'delete') {
    // Check if there's an existing 'add' action for this departure
    const hasAddAction = changeTracker.pendingChanges.some(
      (change) => change.type === 'pt' && 
                  change.elementType === 'departure' && 
                  change.action === 'add' &&
                  ('departure' in change && change.departure.id === departure.id)
    );
    
    // Remove any existing changes for this departure
    changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
      (change) => {
        if (change.type === 'pt' && change.elementType === 'departure') {
          if ('departureId' in change && change.departureId === departure.id) return false;
          if ('departure' in change && change.departure.id === departure.id) return false;
        }
        return true;
      }
    );
    
    // If it was newly added (not yet in database), just remove it from pending changes
    if (hasAddAction) {
      return; // Don't add a delete action
    }
    
    // Otherwise, add a delete action
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'departure',
      action: 'delete',
      processId,
      departureId: departure.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
    return;
  }
  
  // Remove any existing changes for this departure
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'departure') {
        if ('departureId' in change && change.departureId === departure.id) return false;
        if ('departure' in change && change.departure.id === departure.id) return false;
      }
      return true;
    }
  );
  
  if (action === 'save') {
    // Strip sourceState before sending to backend
    const { sourceState, ...departureWithoutState } = departure;
    const saveAction: SyncAction = {
      type: 'pt',
      elementType: 'departure',
      action: 'save',
      processId,
      departure: departureWithoutState
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, saveAction];
  }
}