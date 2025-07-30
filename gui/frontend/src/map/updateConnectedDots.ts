import { populationState, activityTypeConfig } from '$lib/stores/population.svelte';
import { clearConnectedDots, mapState, createNumberedIcon } from './mapState.svelte';
import type { ConnectedPoint, ConnectedLine } from './types';
import L from 'leaflet';

// Update connected dots based on current state
export function updateConnectedDots() {
  if (!mapState.map || !mapState.connectedDotsLayer) return;
  
  // Clear existing dots
  clearConnectedDots();
  
  // Check if we should show plans
  if (!populationState.visibility.plans) return;
  
  // Get selected person
  const selectedPersonId = populationState.selectedPersonId;
  if (!selectedPersonId) return;
  
  const person = populationState.persons.get(selectedPersonId);
  if (!person || !person.plans || person.plans.length === 0) return;
  
  // Get current plan
  const currentPlan = person.plans[populationState.currentPlanIndex] || person.plans[0];
  if (!currentPlan.activities || currentPlan.activities.length === 0) return;
  
  // Add activities as connected dots
  currentPlan.activities.forEach((activity, index) => {
    const config = activityTypeConfig[activity.type];
    const color = config?.color || '#6b7280';
    
    // Create point
    const point: ConnectedPoint = {
      id: `person_${person.id}_act_${activity.id}`,
      marker: null as any,
      position: L.latLng(activity.location[1], activity.location[0]),
      color: color
    };
    
    // Create marker
    const circleIcon = createNumberedIcon(index + 1, color);
    const marker = L.marker(point.position, {
      draggable: false, // Will be enabled based on mode
      icon: circleIcon,
      pane: 'connectedDotsPane'
    }).addTo(mapState.connectedDotsLayer);
    
    // Set up marker events
    setupMarkerEvents(marker, point, activity, index + 1);
    
    point.marker = marker;
    mapState.connectedDots.points.push(point);
    
    // Create line to previous activity
    if (index > 0) {
      const prevPoint = mapState.connectedDots.points[index - 1];
      const line: ConnectedLine = {
        id: `person_${person.id}_leg_${index}`,
        polyline: null as any,
        startPointId: prevPoint.id,
        endPointId: point.id,
        color: mapState.connectedDots.defaultLineColor
      };
      
      const polyline = L.polyline(
        [prevPoint.position, point.position],
        {
          color: line.color,
          weight: 4,
          smoothFactor: 0,
          noClip: true,
          pane: 'connectedDotsPane'
        }
      ).addTo(mapState.connectedDotsLayer);
      
      // Set up line hover events
      setupLineEvents(polyline, line);
      
      line.polyline = polyline;
      mapState.connectedDots.lines.push(line);
    }
  });
  
  // Enable/disable dragging based on current mode
  updateDraggingState();
}

// Set up marker events
function setupMarkerEvents(marker: L.Marker, point: ConnectedPoint, activity: any, number: number) {
  type MarkerState = 'idle' | 'hovered' | 'dragging';
  let markerState: MarkerState = 'idle';
  
  // Drag events
  marker.on('dragstart', () => {
    markerState = 'dragging';
    mapState.isDragging = true;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grabbing';
    }
  });
  
  marker.on('drag', () => {
    const currentPos = marker.getLatLng();
    updateConnectedLines(point.id, currentPos);
    
    // Update activity location in real-time during drag
    const newCoords: [number, number] = [currentPos.lng, currentPos.lat];
    
    if (populationState.selectedPersonId) {
      const person = populationState.persons.get(populationState.selectedPersonId);
      if (person) {
        // Create updated person with new activity location
        const updatedPerson = {
          ...person,
          plans: person.plans.map((plan, planIndex) => 
            planIndex === populationState.currentPlanIndex
              ? {
                  ...plan,
                  activities: plan.activities.map(a => 
                    a.id === activity.id 
                      ? { ...a, location: newCoords }
                      : a
                  )
                }
              : plan
          )
        };
        
        // Update the store to trigger real-time UI updates
        populationState.persons.set(person.id, updatedPerson);
      }
    }
  });
  
  marker.on('dragend', () => {
    markerState = 'idle';
    mapState.isDragging = false;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grab';
    }
    
    // Update activity location
    const newPos = marker.getLatLng();
    const newCoords: [number, number] = [newPos.lng, newPos.lat];
    
    // Update the person in the store to trigger reactivity
    if (populationState.selectedPersonId) {
      const person = populationState.persons.get(populationState.selectedPersonId);
      if (person) {
        // Create updated person with new activity location
        const updatedPerson = {
          ...person,
          plans: person.plans.map((plan, planIndex) => 
            planIndex === populationState.currentPlanIndex
              ? {
                  ...plan,
                  activities: plan.activities.map(a => 
                    a.id === activity.id 
                      ? { ...a, location: newCoords }
                      : a
                  )
                }
              : plan
          )
        };
        
        // Update the store to trigger reactivity
        populationState.persons.set(person.id, updatedPerson);
        
        // Track change for sync
        import('$lib/utils/populationChangeTracking').then(({ trackActivityChange }) => {
          trackActivityChange(person.id, updatedPerson);
        });
      }
    }
  });
  
  // Hover events
  marker.on('mouseover', () => {
    if (markerState === 'idle') {
      markerState = 'hovered';
      const hoverIcon = createNumberedIcon(number, mapState.connectedDots.hoverPointColor);
      marker.setIcon(hoverIcon);
    }
  });
  
  marker.on('mouseout', () => {
    if (markerState === 'hovered') {
      markerState = 'idle';
      const normalIcon = createNumberedIcon(number, point.color || mapState.connectedDots.defaultPointColor);
      marker.setIcon(normalIcon);
    }
  });
}

// Set up line events
function setupLineEvents(polyline: L.Polyline, line: ConnectedLine) {
  type LineState = 'idle' | 'hovered';
  let lineState: LineState = 'idle';
  
  polyline.on('mouseover', () => {
    if (lineState === 'idle') {
      lineState = 'hovered';
      polyline.setStyle({ color: mapState.connectedDots.hoverLineColor });
    }
  });
  
  polyline.on('mouseout', () => {
    lineState = 'idle';
    polyline.setStyle({ color: line.color });
  });
}

// Update connected lines when a point is dragged
function updateConnectedLines(pointId: string, newPosition: L.LatLng) {
  // Update the point's position
  const point = mapState.connectedDots.points.find(p => p.id === pointId);
  if (point) {
    point.position = newPosition;
  }
  
  // Update any connected lines
  mapState.connectedDots.lines.forEach(line => {
    if (line.startPointId === pointId || line.endPointId === pointId) {
      const startPoint = mapState.connectedDots.points.find(p => p.id === line.startPointId);
      const endPoint = mapState.connectedDots.points.find(p => p.id === line.endPointId);
      
      if (startPoint && endPoint) {
        line.polyline.setLatLngs([startPoint.position, endPoint.position]);
      }
    }
  });
}

// Update dragging state based on current mode
function updateDraggingState() {
  
  mapState.connectedDots.points.forEach(point => {
    if (populationState.sidebarInteraction === 'DRAGGING_ACTIVITIES') {
      point.marker.dragging?.enable();
    } else {
      point.marker.dragging?.disable();
    }
  });
}