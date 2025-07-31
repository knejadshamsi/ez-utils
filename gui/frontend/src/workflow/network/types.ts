// Network workflow type definitions

export interface NetworkNode {
  id: string;
  label: string;
  lng: number;
  lat: number;
  type?: string;
  capacity?: number;
}

export interface NetworkLink {
  id: string;
  label: string;
  from: string;
  to: string;
  length?: number;
  capacity?: number;
  lanes?: number;
  speed?: number;
}

export interface NetworkSelection {
  selectedNodeId: string | null;
  selectedLinkId: string | null;
  selectedPolygon: number[][] | null;
}

export interface NodeData {
  id: string;
  lng: number;
  lat: number;
  rawXML: string;
}

export interface LinkData {
  id: string;
  fromNode: string;
  toNode: string;
  rawXML: string;
}

export interface BoundingBox {
  north: number;
  south: number;
  east: number;
  west: number;
}

export interface NodeResult {
  id: string;
  lng: number;
  lat: number;
  rawXML: string;
}

export interface LinkResult {
  id: string;
  fromNode: string;
  toNode: string;
  rawXML: string;
}

export interface NetworkEditState {
  selectedNodes: Set<string>;
  selectedLinks: Set<string>;
  isDrawing: boolean;
  drawingMode: 'node' | 'link' | null;
}

export interface NetworkFilters {
  showNodes: boolean;
  showLinks: boolean;
  nodeTypes: string[];
  linkTypes: string[];
}

export interface NetworkStats {
  nodeCount: number;
  linkCount: number;
  connectedComponents: number;
}