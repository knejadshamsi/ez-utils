import L from 'leaflet';
import type { MapState, ToolMode, CursorStyle, MapEventType } from './types';
import {
  handleConnectedDrawingClick,
  handleConnectedEditMouseDown,
  setupConnectedInspection,
  clearConnectedDots as clearConnectedDotsImpl
} from './connectedDots';
import {
  handleAddNodeClick,
  setupLinkMode,
  setupAttributeMode,
  setupMoveNodesMode,
  handleNodeMoveStart,
  updateNodeStyle
} from './networkEditor';
import {
  clearAllZones as clearPolygonsImpl
} from './polygonDrawing';

// Create the global map state using Svelte 5 runes
export const mapState = $state<MapState>({
  mode: 'idle',
  cursor: 'default',
  map: null,
  
  connectedDotsLayer: null,
  networkLayer: null,
  polygonLayer: null,
  
  connectedDots: {
    points: [],
    lines: [],
    showNumbers: true,
    defaultPointColor: '#2563eb',  // Blue
    defaultLineColor: '#3b82f6',    // Lighter blue
    hoverPointColor: '#1d4ed8',     // Darker blue
    hoverLineColor: '#2563eb'       // Darker blue
  },
  
  network: {
    nodes: [],
    links: [],
    selectedNodeId: null,
    defaultNodeColor: '#ef4444',    // Red
    defaultLinkColor: '#ef4444',    // Red
    hoverNodeColor: '#dc2626',      // Darker red
    hoverLinkColor: '#dc2626'       // Darker red
  },
  
  polygon: {
    currentVertices: [],
    drawnPolygons: [],
    isDrawing: false,
    defaultFillColor: '#8b5cf6',    // Purple
    defaultStrokeColor: '#7c3aed',  // Darker purple
    hoverFillColor: '#a78bfa',      // Lighter purple
    hoverStrokeColor: '#6d28d9'     // Even darker purple
  },
  
  isDragging: false,
  testMessage: 'Hello from global state!'
});

// Helper to get cursor style for a given mode
export function getCursorForMode(mode: ToolMode): CursorStyle {
  switch (mode) {
    case 'drawing-connected':
    case 'adding-nodes':
    case 'drawing-polygon':
      return 'crosshair';
    case 'editing-connected':
    case 'moving-nodes':
    case 'editing-polygon':
      return 'grab';
    case 'inspecting-connected':
    case 'adding-links':
    case 'editing-attributes':
    case 'inspecting-polygon':
      return 'pointer';
    default:
      return 'default';
  }
}

// Mode management
export function setMode(mode: ToolMode) {
  const previousMode = mapState.mode;
  console.log(`Changing mode from ${previousMode} to ${mode}`);
  
  // Clean up previous mode
  cleanupMode(previousMode);
  
  // Set new mode
  mapState.mode = mode;
  mapState.cursor = getCursorForMode(mode);
  
  // Update map cursor
  if (mapState.map) {
    mapState.map.getContainer().style.cursor = mapState.cursor;
  }
  
  // Initialize new mode
  initializeMode(mode);
}

function cleanupMode(mode: ToolMode) {
  if (!mapState.map) return;
  
  // Close any open popups
  mapState.map.closePopup();
  
  // Mode-specific cleanup
  switch (mode) {
    case 'editing-connected':
      // Disable dragging on all markers
      mapState.connectedDots.points.forEach(point => {
        point.marker.dragging?.disable();
      });
      break;
      
    case 'inspecting-connected':
      // Remove click handlers from connected elements
      mapState.connectedDots.points.forEach(point => {
        point.marker.off('click');
      });
      mapState.connectedDots.lines.forEach(line => {
        line.polyline.off('click');
      });
      break;
      
    case 'adding-links':
      // Clear selection
      mapState.network.selectedNodeId = null;
      // Remove highlight from all nodes
      mapState.network.nodes.forEach(node => {
        updateNodeStyle(node, false);
      });
      // Remove click handlers
      mapState.network.nodes.forEach(node => {
        node.marker.off('click');
      });
      break;
      
    case 'editing-attributes':
      // Remove click handlers from network elements
      mapState.network.nodes.forEach(node => {
        node.marker.off('click');
      });
      mapState.network.links.forEach(link => {
        link.polyline.off('click');
      });
      break;
      
    case 'moving-nodes':
      // Disable dragging on all nodes
      mapState.network.nodes.forEach(node => {
        node.marker.dragging?.disable();
      });
      break;
      
  }
}

function initializeMode(mode: ToolMode) {
  switch (mode) {
    case 'drawing-connected':
      // Clear existing items when starting new drawing
      clearConnectedDots();
      break;
      
    case 'editing-connected':
      // Enable dragging on all markers
      mapState.connectedDots.points.forEach(point => {
        point.marker.dragging?.enable();
      });
      break;
      
    case 'inspecting-connected':
      setupConnectedInspection();
      break;
      
    case 'adding-nodes':
      console.log('Add nodes mode activated');
      break;
      
    case 'adding-links':
      setupLinkMode();
      break;
      
    case 'editing-attributes':
      setupAttributeMode();
      break;
      
    case 'moving-nodes':
      setupMoveNodesMode();
      break;
      
    case 'drawing-polygon':
      mapState.polygon.isDrawing = true;
      console.log('Polygon drawing mode activated');
      break;
      
  }
}

// Connected dots management
export function clearConnectedDots() {
  clearConnectedDotsImpl();
}

export function setShowNumbers(show: boolean) {
  mapState.connectedDots.showNumbers = show;
  // Update existing markers
  mapState.connectedDots.points.forEach((point, index) => {
    const icon = createNumberedIcon(index + 1, point.color || mapState.connectedDots.defaultPointColor);
    point.marker.setIcon(icon);
  });
}

export function setConnectedDotsColors(
  pointColor?: string, 
  lineColor?: string, 
  hoverPointColor?: string, 
  hoverLineColor?: string
) {
  if (pointColor) mapState.connectedDots.defaultPointColor = pointColor;
  if (lineColor) mapState.connectedDots.defaultLineColor = lineColor;
  if (hoverPointColor) mapState.connectedDots.hoverPointColor = hoverPointColor;
  if (hoverLineColor) mapState.connectedDots.hoverLineColor = hoverLineColor;
  
  // Update existing elements
  mapState.connectedDots.points.forEach((point, index) => {
    if (!point.color) {
      const icon = createNumberedIcon(index + 1, mapState.connectedDots.defaultPointColor);
      point.marker.setIcon(icon);
    }
  });
  
  mapState.connectedDots.lines.forEach(line => {
    if (!line.color) {
      line.polyline.setStyle({ color: mapState.connectedDots.defaultLineColor });
    }
  });
}

// Network management
export function clearNetwork() {
  if (mapState.networkLayer) {
    mapState.networkLayer.clearLayers();
  }
  mapState.network.nodes = [];
  mapState.network.links = [];
  mapState.network.selectedNodeId = null;
}

export function setNetworkColors(
  nodeColor?: string,
  linkColor?: string,
  hoverNodeColor?: string,
  hoverLinkColor?: string
) {
  if (nodeColor) mapState.network.defaultNodeColor = nodeColor;
  if (linkColor) mapState.network.defaultLinkColor = linkColor;
  if (hoverNodeColor) mapState.network.hoverNodeColor = hoverNodeColor;
  if (hoverLinkColor) mapState.network.hoverLinkColor = hoverLinkColor;
  
  // Update existing elements
  mapState.network.nodes.forEach(node => {
    const icon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
    node.marker.setIcon(icon);
  });
  
  mapState.network.links.forEach(link => {
    link.polyline.setStyle({ color: mapState.network.defaultLinkColor });
  });
}

// Helper function to create numbered icon
export function createNumberedIcon(number: number, color: string, mode?: string): L.DivIcon {
  const showNumber = mapState.connectedDots.showNumbers;
  
  // Determine shape class based on mode
  let shapeClass = 'circle-marker'; // default
  let iconSize: [number, number] = [30, 30];
  let iconAnchor: [number, number] = [15, 15];
  
  if (mode === 'bus') {
    shapeClass = 'circle-marker'; // stays circle
  } else if (mode === 'metro') {
    shapeClass = 'square-marker';
  } else if (mode === 'tram') {
    shapeClass = 'rectangle-marker';
    iconSize = [40, 30]; // wider for rectangle
    iconAnchor = [20, 15]; // center it
  }
  
  return L.divIcon({
    className: 'custom-circle-icon',
    html: `<div class="${shapeClass}" style="background: ${color}">
      ${showNumber ? `<span>${number}</span>` : ''}
    </div>`,
    iconSize: iconSize,
    iconAnchor: iconAnchor
  });
}

// Helper function to create network node icon
export function createNetworkNodeIcon(id: string, color: string): L.DivIcon {
  return L.divIcon({
    className: 'network-node-icon',
    html: `<div class="network-node" style="background: ${color}">
      <span>${id}</span>
    </div>`,
    iconSize: [30, 30],
    iconAnchor: [15, 15]
  });
}

// Central event dispatcher
export function handleMapEvent(event: L.LeafletEvent, type: MapEventType) {
  if (!mapState.map) return;
  
  console.log(`Map event: ${type} in mode: ${mapState.mode}`);
  
  switch (mapState.mode) {
    case 'drawing-connected':
      if (type === 'click') handleConnectedDrawingClick(event as L.LeafletMouseEvent);
      break;
      
    case 'editing-connected':
      if (type === 'mousedown') handleConnectedEditMouseDown(event as L.LeafletMouseEvent);
      break;
      
    case 'inspecting-connected':
      // Handled by individual element click handlers
      break;
      
    case 'adding-nodes':
      if (type === 'click') handleAddNodeClick(event as L.LeafletMouseEvent);
      break;
      
    case 'adding-links':
      // Handled by node click handlers
      break;
      
    case 'editing-attributes':
      // Handled by element click handlers
      break;
      
    case 'moving-nodes':
      if (type === 'mousedown') handleNodeMoveStart(event as L.LeafletMouseEvent);
      break;
      
  }
}

// Export function to set up central event handlers
export function setupCentralEventHandlers() {
  if (!mapState.map) return;
  
  const eventTypes: MapEventType[] = ['click', 'mousedown', 'mousemove', 'mouseup'];
  
  eventTypes.forEach(type => {
    mapState.map!.on(type, (e) => handleMapEvent(e, type));
  });
  
  console.log('Central event handlers set up');
}

// Polygon management
export function clearPolygons() {
  clearPolygonsImpl();
}

// Update test message
export function setTestMessage(message: string) {
  mapState.testMessage = message;
}

// Re-export from other modules
export { updateNodeStyle };