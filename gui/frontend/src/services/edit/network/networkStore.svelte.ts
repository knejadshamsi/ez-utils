// Network-specific state management
export interface NetworkNode {
  id: string;
  label: string;
  x: number;
  y: number;
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

// Create reactive state for network operations
export const networkState = $state<{
  nodes: NetworkNode[];
  links: NetworkLink[];
  selection: NetworkSelection;
  pendingChanges: Map<string, any>;
  isDrawingPolygon: boolean;
  isAddingNode: boolean;
  isAddingLink: boolean;
}>({
  nodes: [],
  links: [],
  selection: {
    selectedNodeId: null,
    selectedLinkId: null,
    selectedPolygon: null
  },
  pendingChanges: new Map(),
  isDrawingPolygon: false,
  isAddingNode: false,
  isAddingLink: false
});

// Helper functions for network operations
export function selectNode(nodeId: string) {
  networkState.selection.selectedNodeId = nodeId;
  networkState.selection.selectedLinkId = null;
}

export function selectLink(linkId: string) {
  networkState.selection.selectedLinkId = linkId;
  networkState.selection.selectedNodeId = null;
}

export function clearSelection() {
  networkState.selection.selectedNodeId = null;
  networkState.selection.selectedLinkId = null;
  networkState.selection.selectedPolygon = null;
}

export function getNodeById(nodeId: string): NetworkNode | undefined {
  return networkState.nodes.find(n => n.id === nodeId);
}

export function getLinkById(linkId: string): NetworkLink | undefined {
  return networkState.links.find(l => l.id === linkId);
}

// Validation functions
export function getLinksConnectedToNode(nodeId: string): NetworkLink[] {
  return networkState.links.filter(link => link.from === nodeId || link.to === nodeId);
}

export function canDeleteNode(nodeId: string): { canDelete: boolean; affectedLinks: string[] } {
  const connectedLinks = getLinksConnectedToNode(nodeId);
  return {
    canDelete: true,
    affectedLinks: connectedLinks.map(l => l.id)
  };
}

export function validateLink(linkId: string): { isValid: boolean; errors: string[] } {
  const link = getLinkById(linkId);
  if (!link) return { isValid: false, errors: ['Link not found'] };
  
  const errors: string[] = [];
  
  if (link.from === link.to) {
    errors.push('Link cannot connect a node to itself');
  }
  
  if (!getNodeById(link.from)) {
    errors.push('From node does not exist');
  }
  
  if (!getNodeById(link.to)) {
    errors.push('To node does not exist');
  }
  
  return { isValid: errors.length === 0, errors };
}