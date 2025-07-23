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
  setupMoveNodesMode,
  handleNodeMoveStart,
  updateNodeStyle
} from './networkEditor';
import {
  clearAllZones as clearPolygonsImpl
} from './polygonDrawing';

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
    defaultPointColor: '#2563eb',
    defaultLineColor: '#3b82f6',
    hoverPointColor: '#1d4ed8',
    hoverLineColor: '#2563eb'
  },
  
  network: {
    nodes: [],
    links: [],
    selectedNodeId: null,
    defaultNodeColor: '#ef4444',
    defaultLinkColor: '#ef4444',
    hoverNodeColor: '#dc2626',
    hoverLinkColor: '#dc2626'
  },
  
  polygon: {
    currentVertices: [],
    drawnPolygons: [],
    isDrawing: false,
    defaultFillColor: '#8b5cf6',
    defaultStrokeColor: '#7c3aed',
    hoverFillColor: '#a78bfa',
    hoverStrokeColor: '#6d28d9'
  },
  
  isDragging: false,
  testMessage: 'Hello from global state!'
});

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

export function setMode(mode: ToolMode) {
  const previousMode = mapState.mode;
  
  cleanupMode(previousMode);
  
  mapState.mode = mode;
  mapState.cursor = getCursorForMode(mode);
  
  if (mapState.map) {
    mapState.map.getContainer().style.cursor = mapState.cursor;
  }
  
  initializeMode(mode);
}

function cleanupMode(mode: ToolMode) {
  if (!mapState.map) return;
  
  mapState.map.closePopup();
  
  switch (mode) {
    case 'editing-connected':
      mapState.connectedDots.points.forEach(point => {
        point.marker.dragging?.disable();
      });
      break;
      
    case 'inspecting-connected':
      mapState.connectedDots.points.forEach(point => {
        point.marker.off('click');
      });
      mapState.connectedDots.lines.forEach(line => {
        line.polyline.off('click');
      });
      break;
      
    case 'adding-links':
      mapState.network.selectedNodeId = null;
      mapState.network.nodes.forEach(node => {
        updateNodeStyle(node, false);
      });
      mapState.network.nodes.forEach(node => {
        node.marker.off('click');
      });
      break;
      
    case 'editing-attributes':
      mapState.network.nodes.forEach(node => {
        node.marker.off('click');
      });
      mapState.network.links.forEach(link => {
        link.polyline.off('click');
      });
      break;
      
    case 'moving-nodes':
      mapState.network.nodes.forEach(node => {
        node.marker.dragging?.disable();
      });
      break;
      
  }
}

function initializeMode(mode: ToolMode) {
  switch (mode) {
    case 'drawing-connected':
      clearConnectedDots();
      break;
      
    case 'editing-connected':
      mapState.connectedDots.points.forEach(point => {
        point.marker.dragging?.enable();
      });
      break;
      
    case 'inspecting-connected':
      setupConnectedInspection();
      break;
      
    case 'adding-nodes':
      break;
      
    case 'adding-links':
      setupLinkMode();
      break;
      
    case 'editing-attributes':
      break;
      
    case 'moving-nodes':
      setupMoveNodesMode();
      break;
      
    case 'drawing-polygon':
      mapState.polygon.isDrawing = true;
      break;
      
  }
}

export function clearConnectedDots() {
  clearConnectedDotsImpl();
}

export function setShowNumbers(show: boolean) {
  mapState.connectedDots.showNumbers = show;
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
  
  mapState.network.nodes.forEach(node => {
    const icon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
    node.marker.setIcon(icon);
  });
  
  mapState.network.links.forEach(link => {
    link.polyline.setStyle({ color: mapState.network.defaultLinkColor });
  });
}

export function createNumberedIcon(number: number, color: string, mode?: string): L.DivIcon {
  const showNumber = mapState.connectedDots.showNumbers;
  
  let shapeClass = 'circle-marker';
  let iconSize: [number, number] = [30, 30];
  let iconAnchor: [number, number] = [15, 15];
  
  if (mode === 'bus') {
    shapeClass = 'circle-marker';
  } else if (mode === 'metro') {
    shapeClass = 'square-marker';
  } else if (mode === 'tram') {
    shapeClass = 'rectangle-marker';
    iconSize = [40, 30];
    iconAnchor = [20, 15];
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

export function createNetworkNodeIcon(id: string, color: string): L.DivIcon {
  const minSize = 40;
  const charWidth = 7;
  const padding = 16;
  const width = Math.max(minSize, id.length * charWidth + padding);
  
  return L.divIcon({
    className: 'network-node-icon',
    html: `<div class="network-node-square" style="background: ${color}; width: ${width}px;">
      <span>${id}</span>
    </div>`,
    iconSize: [width, minSize],
    iconAnchor: [width / 2, minSize / 2]
  });
}

export function handleMapEvent(event: L.LeafletEvent, type: MapEventType) {
  if (!mapState.map) return;
  
  
  if (type === 'click') {
    import('./updateNetworkVisualization').then(module => {
      module.handleNetworkMapClick(event as L.LeafletMouseEvent);
    });
  }
  
  switch (mapState.mode) {
    case 'drawing-connected':
      if (type === 'click') handleConnectedDrawingClick(event as L.LeafletMouseEvent);
      break;
      
    case 'editing-connected':
      if (type === 'mousedown') handleConnectedEditMouseDown(event as L.LeafletMouseEvent);
      break;
      
    case 'inspecting-connected':
      break;
      
    case 'adding-nodes':
      if (type === 'click') handleAddNodeClick(event as L.LeafletMouseEvent);
      break;
      
    case 'adding-links':
      break;
      
    case 'editing-attributes':
      break;
      
    case 'moving-nodes':
      if (type === 'mousedown') handleNodeMoveStart(event as L.LeafletMouseEvent);
      break;
      
  }
}

export function setupCentralEventHandlers() {
  if (!mapState.map) return;
  
  const eventTypes: MapEventType[] = ['click', 'mousedown', 'mousemove', 'mouseup'];
  
  eventTypes.forEach(type => {
    mapState.map!.on(type, (e) => handleMapEvent(e, type));
  });
  
}

export function clearPolygons() {
  clearPolygonsImpl();
}

export function setTestMessage(message: string) {
  mapState.testMessage = message;
}

export { updateNodeStyle };