import { clearConnectedDots, mapState, createNumberedIcon } from './mapState.svelte';
import { lines, routes, stops, departures, sequence, selected } from '@workflow/pt/state.svelte';
import { updateStop } from '@workflow/pt/crud.svelte';
import type { ConnectedPoint, ConnectedLine } from './types';
import L from 'leaflet';

// Transport mode colors
const modeColors = {
  BUS: '#ff7800',    // Orange
  METRO: '#004CFF',  // Blue
  TRAM: '#00ff00'    // Green
};

export function updatePTVisualization() {
  if (!mapState.map || !mapState.connectedDotsLayer) return;
  
  // Clear existing dots
  clearConnectedDots();
  
  // Get selected route data
  if (!selected.routeId) return;
  
  const line = selected.lineId ? lines[selected.lineId] : null;
  if (!line) return;
  
  const route = routes[selected.routeId];
  if (!route) return;
  
  const color = modeColors[selected.mode] || '#666666';
  
  // Get stops for this route using sequence
  const routeStops = sequence
    .map(stopId => stops[stopId])
    .filter(stop => stop !== undefined);
  
  if (routeStops.length > 0) {
    // Create dots for each stop in the route
    routeStops.forEach((stop, index) => {
      if (!stop || (stop.lng === 0 && stop.lat === 0)) return;
      
      // Create point
      const point: ConnectedPoint = {
        id: `pt_${selected.lineId}_stop_${stop.stopId}`,
        marker: null as any,
        position: L.latLng(stop.lat || 0, stop.lng || 0),
        color: color
      };
      
      // Create marker with mode-specific shape
      const circleIcon = createNumberedIcon(index + 1, color, selected.mode);
      const marker = L.marker(point.position, {
        draggable: selected.editMode === 'DRAGGING_STOP',
        icon: circleIcon,
        pane: 'connectedDotsPane'
      }).addTo(mapState.connectedDotsLayer!);
      
      // Store stop data for drag events
      (marker as any).stopData = {
        stopId: stop.stopId,
        stopName: stop.stopName,
        lineId: selected.lineId,
        routeId: selected.routeId
      };
      
      // Set up marker events
      setupMarkerEvents(marker, point, stop, index + 1);
      
      point.marker = marker;
      
      // Always add point to array (needed for route lines)
      mapState.connectedDots.points.push(point);
      
      // Create line to previous stop
      if (index > 0) {
        const prevPoint = mapState.connectedDots.points[index - 1];
        const connectedLine: ConnectedLine = {
          id: `pt_${selected.lineId}_route_${selected.routeId}_segment_${index}`,
          polyline: null as any,
          startPointId: prevPoint.id,
          endPointId: point.id,
          color: color
        };
        
        const polyline = L.polyline(
          [prevPoint.position, point.position],
          {
            color: connectedLine.color,
            weight: 4,
            smoothFactor: 0,
            noClip: true,
            pane: 'connectedDotsPane'
          }
        ).addTo(mapState.connectedDotsLayer!);
        
        // Set up line hover events
        setupLineEvents(polyline, connectedLine);
        
        connectedLine.polyline = polyline;
        mapState.connectedDots.lines.push(connectedLine);
      }
    });
  }
  
  // Enable/disable dragging based on current mode
  updateDraggingState();
}

// Set up marker events
function setupMarkerEvents(marker: L.Marker, point: ConnectedPoint, stop: any, number: number) {
  type MarkerState = 'idle' | 'hovered' | 'dragging';
  let markerState: MarkerState = 'idle';
  
  // Get mode from stopData
  const stopData = (marker as any).stopData;
  const mode = selected.mode;
  
  // Drag events
  marker.on('dragstart', () => {
    markerState = 'dragging';
    mapState.isDragging = true;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grabbing';
    }
  });
  
  marker.on('drag', () => {
    updateConnectedLines(point.id, marker.getLatLng());
  });
  
  marker.on('dragend', () => {
    markerState = 'idle';
    mapState.isDragging = false;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grab';
    }
    
    // Update stop location
    const newPos = marker.getLatLng();
    
    if (stop?.stopId) {
      // Update the stop using the new store's update method
      updateStop(stop.stopId, {
        lat: newPos.lat,
        lng: newPos.lng
      });
    }
  });
  
  // Hover events
  marker.on('mouseover', () => {
    if (markerState === 'idle') {
      markerState = 'hovered';
      const hoverIcon = createNumberedIcon(number, mapState.connectedDots.hoverPointColor, mode);
      marker.setIcon(hoverIcon);
    }
  });
  
  marker.on('mouseout', () => {
    if (markerState === 'hovered') {
      markerState = 'idle';
      const normalIcon = createNumberedIcon(number, point.color || mapState.connectedDots.defaultPointColor, mode);
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
    if (point.marker) {
      if (selected.editMode === 'DRAGGING_STOP') {
        point.marker.dragging?.enable();
      } else {
        point.marker.dragging?.disable();
      }
    }
  });
}

// Export type-safe state interface
export type PTMode = 'IDLE' | 'DRAGGING_STOP' | 'ADDING_STOP' | 'SELECTING_LOCATION';

export function getPTMode(): PTMode {
  if (selected.editMode === 'DRAGGING_STOP') return 'DRAGGING_STOP';
  if (selected.editMode === 'ADDING_STOP') return 'ADDING_STOP';
  if (selected.editMode === 'SELECTING_STOP_LOCATION') return 'SELECTING_LOCATION';
  return 'IDLE';
}