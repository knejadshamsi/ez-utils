import { changeTracker } from '$lib/changeTracker.svelte';
import type { SyncAction } from '$lib/changeTracker.svelte';
import { getCurrentProcessId } from '$lib/utils/processId';
import type { Line, Stop, Departure } from '$lib/stores/pt.svelte';

export function trackStopChange(stop: Stop, action: 'add' | 'update' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  // Remove any existing changes for this stop
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'stop') {
        if ('stopId' in change && change.stopId === stop.stopId) return false;
        if ('stop' in change && change.stop.stopId === stop.stopId) return false;
      }
      return true;
    }
  );
  
  if (action === 'add') {
    const addAction: SyncAction = {
      type: 'pt',
      elementType: 'stop',
      action: 'add',
      processId,
      stop
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'pt',
      elementType: 'stop',
      action: 'update',
      processId,
      stopId: stop.stopId,
      update: stop
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'stop',
      action: 'delete',
      processId,
      stopId: stop.stopId
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
  }
}

export function trackLineChange(line: Line, action: 'add' | 'update' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
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
  
  if (action === 'add') {
    const addAction: SyncAction = {
      type: 'pt',
      elementType: 'line',
      action: 'add',
      processId,
      line
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'pt',
      elementType: 'line',
      action: 'update',
      processId,
      lineId: line.id,
      update: line
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'line',
      action: 'delete',
      processId,
      lineId: line.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
  }
}

// Route is now embedded in Line, so we need lineId too
export function trackRouteChange(route: {id: string, name: string, lineId: string}, action: 'add' | 'update' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
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
  
  if (action === 'add') {
    const addAction: SyncAction = {
      type: 'pt',
      elementType: 'route',
      action: 'add',
      processId,
      route
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'pt',
      elementType: 'route',
      action: 'update',
      processId,
      routeId: route.id,
      update: route
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'route',
      action: 'delete',
      processId,
      routeId: route.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
  }
}

// RouteStop doesn't exist anymore - just use Stop
export function trackRouteStopChange(routeStop: Stop, action: 'add' | 'update' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
  // For route stops, we need to filter by both routeId and stopOrder
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === 'routeStop') {
        if ('routeStop' in change && change.routeStop.routeId === routeStop.routeId && 
            change.routeStop.sequence === routeStop.sequence) return false;
        if ('routeId' in change && 'stopOrder' in change && 
            change.routeId === routeStop.routeId && change.stopOrder === routeStop.sequence) return false;
      }
      return true;
    }
  );
  
  if (action === 'add') {
    const addAction: SyncAction = {
      type: 'pt',
      elementType: 'routeStop',
      action: 'add',
      processId,
      routeStop
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'pt',
      elementType: 'routeStop',
      action: 'update',
      processId,
      routeId: routeStop.routeId,
      stopOrder: routeStop.sequence,
      update: routeStop
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'routeStop',
      action: 'delete',
      processId,
      routeId: routeStop.routeId,
      stopOrder: routeStop.sequence
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
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
  
  const batchAction: SyncAction = {
    type: 'pt',
    elementType: 'stop',
    action: 'batchUpdate',
    processId,
    updates
  };
  
  changeTracker.pendingChanges = [...changeTracker.pendingChanges, batchAction];
}

export function removePTChanges(elementType: 'stop' | 'line' | 'route' | 'routeStop', id: string, routeId?: string) {
  changeTracker.pendingChanges = changeTracker.pendingChanges.filter(
    (change) => {
      if (change.type === 'pt' && change.elementType === elementType) {
        switch (elementType) {
          case 'stop':
            if ('stopId' in change && change.stopId === id) return false;
            if ('stop' in change && change.stop.id === id) return false;
            break;
          case 'line':
            if ('lineId' in change && change.lineId === id) return false;
            if ('line' in change && change.line.id === id) return false;
            break;
          case 'route':
            if ('routeId' in change && change.routeId === id) return false;
            if ('route' in change && change.route.id === id) return false;
            break;
          case 'routeStop':
            if ('routeId' in change && 'stopOrder' in change && 
                change.routeId === routeId && change.stopOrder.toString() === id) return false;
            if ('routeStop' in change && change.routeStop.routeId === routeId && 
                change.routeStop.sequence.toString() === id) return false;
            break;
        }
      }
      return true;
    }
  );
}

export function trackDepartureChange(departure: Departure, action: 'add' | 'update' | 'delete') {
  const processId = getCurrentProcessId();
  if (!processId) return;
  
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
  
  if (action === 'add') {
    const addAction: SyncAction = {
      type: 'pt',
      elementType: 'departure',
      action: 'add',
      processId,
      departure
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, addAction];
  } else if (action === 'update') {
    const updateAction: SyncAction = {
      type: 'pt',
      elementType: 'departure',
      action: 'update',
      processId,
      departureId: departure.id,
      update: departure
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, updateAction];
  } else if (action === 'delete') {
    const deleteAction: SyncAction = {
      type: 'pt',
      elementType: 'departure',
      action: 'delete',
      processId,
      departureId: departure.id
    };
    changeTracker.pendingChanges = [...changeTracker.pendingChanges, deleteAction];
  }
}