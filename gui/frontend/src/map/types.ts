import type L from 'leaflet';

// Tool modes - strictly typed
export type ToolMode = 
  | 'idle' 
  | 'drawing-connected' 
  | 'editing-connected' 
  | 'inspecting-connected'
  | 'adding-nodes' 
  | 'adding-links' 
  | 'editing-attributes'
  | 'moving-nodes'
  | 'drawing-polygon'
  | 'editing-polygon'
  | 'inspecting-polygon';

// Cursor styles that map to CSS cursor values
export type CursorStyle = 'default' | 'crosshair' | 'grab' | 'grabbing' | 'pointer' | 'not-allowed';

// Connected dots specific types
export interface ConnectedPoint {
  id: string;
  marker: L.Marker;
  position: L.LatLng;
  color?: string;
}

export interface ConnectedLine {
  id: string;
  polyline: L.Polyline;
  startPointId: string;
  endPointId: string;
  color?: string;
}

// Polygon specific types
export interface PolygonVertex {
  id: string;
  marker: L.Marker;
  position: L.LatLng;
  index: number;
}

export interface DrawnPolygon {
  id: string;
  polygon: L.Polygon;
  vertices: PolygonVertex[];
  fillColor: string;
  strokeColor: string;
}

// Network editor specific types
export interface NetworkNode {
  id: string;
  marker: L.Marker;
  position: L.LatLng;
  attributes: {
    label?: string;
    type?: string;
    capacity?: number;
    [key: string]: any;
  };
}

export interface NetworkLink {
  id: string;
  polyline: L.Polyline;
  fromNodeId: string;
  toNodeId: string;
  attributes: {
    label?: string;
    capacity?: number;
    lanes?: number;
    speed?: number;
    [key: string]: any;
  };
}

// Events
export type MapEventType = 'click' | 'mousedown' | 'mousemove' | 'mouseup' | 'mouseover' | 'mouseout';

export interface MapEventHandler {
  type: MapEventType;
  handler: (e: L.LeafletEvent) => void;
}

// Main map state interface
export interface MapState {
  mode: ToolMode;
  cursor: CursorStyle;
  map: L.Map | null;
  
  // Layer groups for different tools
  connectedDotsLayer: L.FeatureGroup | null;
  networkLayer: L.FeatureGroup | null;
  polygonLayer: L.FeatureGroup | null;
  
  // Tool-specific state
  connectedDots: {
    points: ConnectedPoint[];
    lines: ConnectedLine[];
    showNumbers: boolean;
    defaultPointColor: string;
    defaultLineColor: string;
    hoverPointColor: string;
    hoverLineColor: string;
  };
  
  network: {
    nodes: NetworkNode[];
    links: NetworkLink[];
    selectedNodeId: string | null;
    defaultNodeColor: string;
    defaultLinkColor: string;
    hoverNodeColor: string;
    hoverLinkColor: string;
  };
  
  polygon: {
    currentVertices: PolygonVertex[];
    drawnPolygons: DrawnPolygon[];
    isDrawing: boolean;
    defaultFillColor: string;
    defaultStrokeColor: string;
    hoverFillColor: string;
    hoverStrokeColor: string;
  };
  
  // Shared state
  isDragging: boolean;
  
  // Test global state
  testMessage: string;
}

// Configuration options
export interface MapConfig {
  center: { lng: number; lat: number };
  zoom: number;
  style: string;
}