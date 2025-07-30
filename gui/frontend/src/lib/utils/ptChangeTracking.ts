// Simplified change tracking for PT module
// Changes are tracked per route and saved automatically when switching routes

interface PTChanges {
  stops: {
    [stopId: string]: {
      action: 'save' | 'delete';
      data: any;
      isNew: boolean;
    }
  };
  departures: {
    [departureId: string]: {
      action: 'save' | 'delete';
      data: any;
      isNew: boolean;
    }
  };
  lines: {
    [lineId: string]: {
      action: 'save' | 'delete';
      data: any;
      isNew: boolean;
    }
  };
  routes: {
    [routeId: string]: {
      action: 'save' | 'delete';
      data: any;
      isNew: boolean;
    }
  };
}

// Initialize global change tracking
if (!(window as any).__ptChanges) {
  (window as any).__ptChanges = {
    stops: {},
    departures: {},
    lines: {},
    routes: {}
  } as PTChanges;
}

function getChanges(): PTChanges {
  return (window as any).__ptChanges;
}

export function trackStopChange(stop: any, action: 'save' | 'delete') {
  const changes = getChanges();
  
  if (action === 'delete') {
    // If it was new and not saved, just remove from tracking
    if (changes.stops[stop.stopId]?.isNew) {
      delete changes.stops[stop.stopId];
    } else {
      changes.stops[stop.stopId] = {
        action: 'delete',
        data: stop,
        isNew: false
      };
    }
  } else {
    // Check if this is a new stop (not in backend yet)
    const isNew = stop.stopId.startsWith('stop_') && !changes.stops[stop.stopId]?.isNew === false;
    
    changes.stops[stop.stopId] = {
      action: 'save',
      data: stop,
      isNew
    };
  }
}

export function trackLineChange(line: any, action: 'save' | 'delete') {
  const changes = getChanges();
  
  if (action === 'delete') {
    // If it was new and not saved, just remove from tracking
    if (changes.lines[line.id]?.isNew) {
      delete changes.lines[line.id];
    } else {
      changes.lines[line.id] = {
        action: 'delete',
        data: line,
        isNew: false
      };
    }
  } else {
    // Check if this is a new line
    const isNew = line.id.startsWith('line_') && !changes.lines[line.id]?.isNew === false;
    
    changes.lines[line.id] = {
      action: 'save',
      data: line,
      isNew
    };
  }
}

export function trackRouteChange(route: any, action: 'save' | 'delete') {
  const changes = getChanges();
  
  if (action === 'delete') {
    // If it was new and not saved, just remove from tracking
    if (changes.routes[route.id]?.isNew) {
      delete changes.routes[route.id];
    } else {
      changes.routes[route.id] = {
        action: 'delete',
        data: route,
        isNew: false
      };
    }
  } else {
    // Check if this is a new route
    const isNew = route.id.startsWith('route_') && !changes.routes[route.id]?.isNew === false;
    
    changes.routes[route.id] = {
      action: 'save',
      data: route,
      isNew
    };
  }
}

export function trackDepartureChange(departure: any, action: 'save' | 'delete') {
  const changes = getChanges();
  
  if (action === 'delete') {
    // If it was new and not saved, just remove from tracking
    if (changes.departures[departure.id]?.isNew) {
      delete changes.departures[departure.id];
    } else {
      changes.departures[departure.id] = {
        action: 'delete',
        data: departure,
        isNew: false
      };
    }
  } else {
    // Check if this is a new departure
    const isNew = departure.id.startsWith('dep_') && !changes.departures[departure.id]?.isNew === false;
    
    changes.departures[departure.id] = {
      action: 'save',
      data: departure,
      isNew
    };
  }
}

export function clearRouteChanges(routeId: string) {
  const changes = getChanges();
  
  // Clear stops for this route
  Object.keys(changes.stops).forEach(stopId => {
    if (changes.stops[stopId].data.routeId === routeId) {
      delete changes.stops[stopId];
    }
  });
  
  // Clear departures for this route
  Object.keys(changes.departures).forEach(depId => {
    if (changes.departures[depId].data.routeId === routeId) {
      delete changes.departures[depId];
    }
  });
}

export function clearAllChanges() {
  (window as any).__ptChanges = {
    stops: {},
    departures: {},
    lines: {},
    routes: {}
  };
}