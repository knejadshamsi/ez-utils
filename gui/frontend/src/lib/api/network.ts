import { GetNodesInBBox, GetLinksInBBox, UpdateNetworkNode, UpdateNetworkLink, DeleteNetworkNode, DeleteNetworkLink, CreateNetworkNode, CreateNetworkLink } from '@wailsjs/go/gui/App';
import type { gui } from '@wailsjs/go/models';
import { networkState, type NetworkNode, type NetworkLink } from '$lib/stores/network.svelte';
import { editingSession } from '$lib/stores/app.svelte.ts';

// Load network data for a given bounding box
export async function loadNetworkInBBox(bbox: gui.BoundingBox) {
  try {
    const processId = editingSession.processId;
    
    // Get nodes from backend
    const nodes = await GetNodesInBBox(processId, bbox);
    
    // Get node IDs for link query
    const nodeIds = nodes.map(n => n.id);
    
    // Get links connected to these nodes
    const links = await GetLinksInBBox(processId, nodeIds);
    
    // Update store with loaded data
    networkState.nodes = nodes.map(n => ({
      id: n.id,
      label: n.id,
      x: n.x,
      y: n.y
    }));
    
    networkState.links = links.map(l => ({
      id: l.id,
      label: l.id,
      from: l.from_node,
      to: l.to_node
    }));
    
  } catch (error) {
    console.error('Failed to load network data:', error);
    throw error;
  }
}

// Load all network data (for initial load)
export async function loadAllNetworkData() {
  // Create a large bounding box to get all data
  const bbox: gui.BoundingBox = {
    north: 90,
    south: -90,
    east: 180,
    west: -180
  };
  
  return loadNetworkInBBox(bbox);
}

// Update a node's properties
export async function updateNode(node: NetworkNode) {
  try {
    const processId = editingSession.processId;
    const rawXML = `<node id="${node.id}" x="${node.x}" y="${node.y}" />`;
    
    await UpdateNetworkNode(processId, node.id, node.x, node.y, rawXML);
    
    // Update local state
    const index = networkState.nodes.findIndex(n => n.id === node.id);
    if (index !== -1) {
      networkState.nodes[index] = node;
    }
  } catch (error) {
    console.error('Failed to update node:', error);
    throw error;
  }
}

// Update a link's properties
export async function updateLink(link: NetworkLink) {
  try {
    const processId = editingSession.processId;
    const rawXML = `<link id="${link.id}" from="${link.from}" to="${link.to}" />`;
    
    await UpdateNetworkLink(processId, link.id, rawXML);
    
    // Update local state
    const index = networkState.links.findIndex(l => l.id === link.id);
    if (index !== -1) {
      networkState.links[index] = link;
    }
  } catch (error) {
    console.error('Failed to update link:', error);
    throw error;
  }
}

// Delete a node (and its connected links)
export async function deleteNode(nodeId: string) {
  try {
    const processId = editingSession.processId;
    
    await DeleteNetworkNode(processId, nodeId);
    
    // Update local state
    networkState.nodes = networkState.nodes.filter(n => n.id !== nodeId);
    networkState.links = networkState.links.filter(l => l.from !== nodeId && l.to !== nodeId);
    
    // Clear selection if deleted node was selected
    if (networkState.selection.selectedNodeId === nodeId) {
      networkState.selection.selectedNodeId = null;
    }
  } catch (error) {
    console.error('Failed to delete node:', error);
    throw error;
  }
}

// Delete a link
export async function deleteLink(linkId: string) {
  try {
    const processId = editingSession.processId;
    
    await DeleteNetworkLink(processId, linkId);
    
    // Update local state
    networkState.links = networkState.links.filter(l => l.id !== linkId);
    
    // Clear selection if deleted link was selected
    if (networkState.selection.selectedLinkId === linkId) {
      networkState.selection.selectedLinkId = null;
    }
  } catch (error) {
    console.error('Failed to delete link:', error);
    throw error;
  }
}

// Create a new node
export async function createNode(x: number, y: number): Promise<string> {
  try {
    const processId = editingSession.processId;
    const nodeId = `node_${Date.now()}`;
    const rawXML = `<node id="${nodeId}" x="${x}" y="${y}" />`;
    
    await CreateNetworkNode(processId, nodeId, x, y, rawXML);
    
    // Add to local state
    const newNode: NetworkNode = {
      id: nodeId,
      label: nodeId,
      x,
      y
    };
    networkState.nodes = [...networkState.nodes, newNode];
    
    return nodeId;
  } catch (error) {
    console.error('Failed to create node:', error);
    throw error;
  }
}

// Create a new link
export async function createLink(fromNodeId: string, toNodeId: string): Promise<string> {
  try {
    const processId = editingSession.processId;
    const linkId = `link_${fromNodeId}_${toNodeId}_${Date.now()}`;
    const rawXML = `<link id="${linkId}" from="${fromNodeId}" to="${toNodeId}" />`;
    
    await CreateNetworkLink(processId, linkId, fromNodeId, toNodeId, rawXML);
    
    // Add to local state
    const newLink: NetworkLink = {
      id: linkId,
      label: linkId,
      from: fromNodeId,
      to: toNodeId
    };
    networkState.links = [...networkState.links, newLink];
    
    return linkId;
  } catch (error) {
    console.error('Failed to create link:', error);
    throw error;
  }
}