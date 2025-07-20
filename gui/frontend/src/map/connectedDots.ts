import L from 'leaflet';
import type { ConnectedPoint, ConnectedLine } from './types';
import { mapState, getCursorForMode, setTestMessage, createNumberedIcon } from './mapState.svelte';

// Connected dots specific handlers
export function handleConnectedDrawingClick(event: L.LeafletMouseEvent) {
  if (!mapState.map || !mapState.connectedDotsLayer) return;
  
  const latlng = event.latlng;
  const pointNumber = mapState.connectedDots.points.length + 1;
  
  // Create point object first to get the color
  const point: ConnectedPoint = {
    id: `point-${Date.now()}`,
    marker: null as any, // Will be set below
    position: latlng,
    color: mapState.connectedDots.defaultPointColor
  };
  
  // Create marker with color
  const circleIcon = createNumberedIcon(pointNumber, point.color);
  
  const marker = L.marker(latlng, {
    draggable: false,
    icon: circleIcon
  }).addTo(mapState.connectedDotsLayer);
  
  point.marker = marker;
  
  // Add to state
  mapState.connectedDots.points.push(point);
  
  // Create line if we have previous points
  if (mapState.connectedDots.points.length > 1) {
    const prevPoint = mapState.connectedDots.points[mapState.connectedDots.points.length - 2];
    const line: ConnectedLine = {
      id: `line-${Date.now()}`,
      polyline: null as any, // Will be set below
      startPointId: prevPoint.id,
      endPointId: point.id,
      color: mapState.connectedDots.defaultLineColor
    };
    
    const polyline = L.polyline(
      [prevPoint.position, latlng],
      {
        color: line.color,
        weight: 4,
        smoothFactor: 0,
        noClip: true
      }
    ).addTo(mapState.connectedDotsLayer);
    
    line.polyline = polyline;
    
    // Set up hover handlers for the line with state tracking
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
    
    mapState.connectedDots.lines.push(line);
  }
  
  // Set up drag handler for when in edit mode
  type MarkerState = 'idle' | 'hovered' | 'dragging';
  let markerState: MarkerState = 'idle';
  
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
      mapState.map.getContainer().style.cursor = getCursorForMode(mapState.mode);
    }
    // Reset icon to normal color after drag
    const normalIcon = createNumberedIcon(pointNumber, point.color);
    marker.setIcon(normalIcon);
  });
  
  // Set up hover handlers with state tracking
  marker.on('mouseover', () => {
    if (markerState === 'idle') {
      markerState = 'hovered';
      const hoverIcon = createNumberedIcon(pointNumber, mapState.connectedDots.hoverPointColor);
      marker.setIcon(hoverIcon);
    }
  });
  
  marker.on('mouseout', () => {
    if (markerState === 'hovered') {
      markerState = 'idle';
      const normalIcon = createNumberedIcon(pointNumber, point.color);
      marker.setIcon(normalIcon);
    }
  });
  
  console.log(`Added connected point ${pointNumber} at [${latlng.lng}, ${latlng.lat}]`);
}

export function handleConnectedEditMouseDown(_event: L.LeafletMouseEvent) {
  // Dragging is handled by Leaflet's built-in dragging
  // This is here for any additional edit mode logic
}

export function setupConnectedInspection() {
  // Set up click handlers for inspection mode
  mapState.connectedDots.points.forEach((point, index) => {
    point.marker.on('click', () => {
      if (mapState.mode !== 'inspecting-connected' || !mapState.map) return;
      
      const latlng = point.position;
      
      // Create container for reactive content
      const container = document.createElement('div');
      container.innerHTML = `
        <div style="text-align: center; min-width: 200px;">
          <strong>Point ${mapState.connectedDots.showNumbers ? index + 1 : ''}</strong><br>
          <div style="margin: 10px 0; padding: 8px; background: #f3f4f6; border-radius: 4px;">
            <div style="font-size: 12px; color: #6b7280; margin-bottom: 5px;">Global State:</div>
            <div class="state-display" style="font-weight: bold; color: #0ea5e9;">${mapState.testMessage}</div>
          </div>
          <div style="margin: 10px 0;">
            <input type="text" 
                   class="state-input"
                   placeholder="Update state" 
                   style="width: 100%; padding: 4px; border: 1px solid #e5e7eb; border-radius: 4px; margin-bottom: 5px;">
            <button class="update-btn" 
                    style="width: 100%; padding: 4px 8px; background: #10b981; color: white; border: none; border-radius: 4px; cursor: pointer;">
              Update from Popup
            </button>
          </div>
          <button onclick="alert('Coordinates: [${latlng.lng.toFixed(6)}, ${latlng.lat.toFixed(6)}]')" 
                  style="margin-top: 8px; padding: 4px 8px; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Show Coordinates
          </button>
        </div>
      `;
      
      // Add event listeners
      const input = container.querySelector('.state-input') as HTMLInputElement;
      const updateBtn = container.querySelector('.update-btn') as HTMLButtonElement;
      const stateDisplay = container.querySelector('.state-display') as HTMLDivElement;
      
      updateBtn.onclick = () => {
        if (input.value) {
          setTestMessage(input.value);
          // Update the display in the popup immediately
          stateDisplay.textContent = input.value;
          input.value = '';
        }
      };
      
      L.popup()
        .setLatLng(latlng)
        .setContent(container)
        .openOn(mapState.map);
    });
  });
  
  mapState.connectedDots.lines.forEach((line) => {
    line.polyline.on('click', (e: L.LeafletMouseEvent) => {
      if (mapState.mode !== 'inspecting-connected' || !mapState.map) return;
      
      const startIndex = mapState.connectedDots.points.findIndex(p => p.id === line.startPointId);
      const endIndex = mapState.connectedDots.points.findIndex(p => p.id === line.endPointId);
      
      const popupContent = `
        <div style="text-align: center;">
          <strong>Line Segment</strong><br>
          <button onclick="alert('Connected points: ${startIndex + 1} → ${endIndex + 1}')" 
                  style="margin-top: 8px; padding: 4px 8px; background: #10b981; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Show Connected Points
          </button>
        </div>
      `;
      
      L.popup()
        .setLatLng(e.latlng)
        .setContent(popupContent)
        .openOn(mapState.map);
    });
  });
}

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

export function clearConnectedDots() {
  if (mapState.connectedDotsLayer) {
    mapState.connectedDotsLayer.clearLayers();
  }
  mapState.connectedDots.points = [];
  mapState.connectedDots.lines = [];
}