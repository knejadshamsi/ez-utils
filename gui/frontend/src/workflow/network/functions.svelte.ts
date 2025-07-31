// Network functions
import type { NetworkNode, NetworkLink } from './types';
import { networkState } from './state.svelte';

// Selection functions
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

// Data access functions
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

export function canDeleteNode(nodeId: string): { canDelete: boolean; affectedLinks: NetworkLink[] } {
  const connectedLinks = getLinksConnectedToNode(nodeId);
  return {
    canDelete: true,
    affectedLinks: connectedLinks
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

// Find reverse link (B→A for A→B)
export function findReverseLink(linkId: string): NetworkLink | null {
  const link = getLinkById(linkId);
  if (!link) return null;
  
  return networkState.links.find(l => 
    l.from === link.to && l.to === link.from
  ) || null;
}