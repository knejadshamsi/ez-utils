import { GetNodesInBBox, GetLinksInBBox } from '@wailsjs/go/gui/App';
import type { gui } from '@wailsjs/go/models';
import { networkState, type NetworkNode, type NetworkLink } from '$lib/stores/network.svelte';
import { editingSession } from '$lib/stores/app.svelte';
import { trackNodeChange, trackLinkChange } from '$lib/utils/networkChangeTracking';
import { updateNetworkVisualization } from '../../map/updateNetworkVisualization';

// Load network data for a given bounding box
export async function loadNetworkInBBox(bbox: gui.BoundingBox, appendData: boolean = false) {
  try {
    const processId = editingSession.processId;
    
    // Get nodes from backend
    const nodes = await GetNodesInBBox(processId, bbox);
    
    // Get node IDs for link query
    const nodeIds = nodes.map(n => n.id);
    
    // Get links connected to these nodes
    const links = await GetLinksInBBox(processId, nodeIds);
    
    // Transform data
    const newNodes = nodes.map(n => ({
      id: n.id,
      label: n.id,
      x: n.x,
      y: n.y
    }));
    
    const newLinks = links.map(l => ({
      id: l.id,
      label: l.id,
      from: l.from_node,
      to: l.to_node
    }));
    
    if (appendData) {
      // Add new data without duplicates
      const existingNodeIds = new Set(networkState.nodes.map(n => n.id));
      const existingLinkIds = new Set(networkState.links.map(l => l.id));
      
      const uniqueNewNodes = newNodes.filter(n => !existingNodeIds.has(n.id));
      const uniqueNewLinks = newLinks.filter(l => !existingLinkIds.has(l.id));
      
      networkState.nodes = [...networkState.nodes, ...uniqueNewNodes];
      networkState.links = [...networkState.links, ...uniqueNewLinks];
    } else {
      // Replace all data
      networkState.nodes = newNodes;
      networkState.links = newLinks;
    }
    
    // Update visualization after loading
    updateNetworkVisualization();
    
  } catch (error) {
    console.error('Failed to load network data:', error);
    throw error;
  }
}

// Load all network data (for initial load) - DEPRECATED, use loadInitialViewportData
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

// Load network data for initial viewport (50% scaled down)
export async function loadInitialViewportData(mapInstance: L.Map) {
  const bounds = mapInstance.getBounds();
  const center = mapInstance.getCenter();
  
  // Calculate 50% scaled bounds centered on current view
  const latSpan = bounds.getNorth() - bounds.getSouth();
  const lngSpan = bounds.getEast() - bounds.getWest();
  
  const scaledLatSpan = latSpan * 0.5;
  const scaledLngSpan = lngSpan * 0.5;
  
  const bbox: gui.BoundingBox = {
    north: center.lat + (scaledLatSpan / 2),
    south: center.lat - (scaledLatSpan / 2),
    east: center.lng + (scaledLngSpan / 2),
    west: center.lng - (scaledLngSpan / 2)
  };
  
  return loadNetworkInBBox(bbox);
}

// Load network data for a drawn polygon
export async function loadNetworkInPolygon(polygonCoords: number[][]) {
  try {
    // Validate polygon input
    if (!Array.isArray(polygonCoords)) {
      throw new Error('Invalid polygon: coordinates must be an array');
    }
    if (polygonCoords.length === 0) {
      throw new Error('Invalid polygon: empty coordinates array');
    }
    if (polygonCoords.length < 3) {
      throw new Error('Invalid polygon: minimum 3 points required');
    }
    
    // Validate coordinate format and values
    for (let i = 0; i < polygonCoords.length; i++) {
      const coord = polygonCoords[i];
      if (!Array.isArray(coord) || coord.length !== 2) {
        throw new Error(`Invalid polygon: coordinate ${i} must be [lat, lng] array`);
      }
      const [lat, lng] = coord;
      if (typeof lat !== 'number' || typeof lng !== 'number' || !isFinite(lat) || !isFinite(lng)) {
        throw new Error(`Invalid polygon: coordinate ${i} must contain finite numbers`);
      }
      if (lat < -90 || lat > 90) {
        throw new Error(`Invalid polygon: latitude ${lat} at coordinate ${i} out of valid range [-90, 90]`);
      }
      if (lng < -180 || lng > 180) {
        throw new Error(`Invalid polygon: longitude ${lng} at coordinate ${i} out of valid range [-180, 180]`);
      }
    }
    
    // Check for unsaved changes before loading new data
    const { changeTracker } = await import('$lib/changeTracker.svelte');
    
    if (changeTracker.pendingChanges.length > 0) {
      // Filter for network changes only
      const networkChanges = changeTracker.pendingChanges.filter(change => change.type === 'network');
      
      if (networkChanges.length > 0) {
        const { showConfirmation } = await import('$lib/stores/confirmationModal.svelte');
        
        const shouldContinue = await new Promise<boolean>((resolve) => {
          showConfirmation({
            title: "Unsaved Changes Detected",
            message: `You have ${networkChanges.length} unsaved network change(s). Loading new data will keep your changes.`,
            confirmText: "Continue",
            cancelText: "Cancel",
            onConfirm: () => resolve(true),
            onCancel: () => resolve(false)
          });
        });
        
        if (!shouldContinue) {
          throw new Error('Data loading cancelled by user');
        }
      }
    }
    
    // Convert polygon to bounding box
    const lats = polygonCoords.map(coord => coord[0]);
    const lngs = polygonCoords.map(coord => coord[1]);
    
    const bbox: gui.BoundingBox = {
      north: Math.max(...lats),
      south: Math.min(...lats),
      east: Math.max(...lngs),
      west: Math.min(...lngs)
    };
    
    // Validate bounding box
    if (bbox.north === bbox.south && bbox.east === bbox.west) {
      throw new Error('Invalid polygon: all points are the same location');
    }
    
    return loadNetworkInBBox(bbox, true); // Append data, don't replace
  } catch (error) {
    console.error('Failed to load network data for polygon:', error);
    throw error;
  }
}

// Update a node's properties (local only)
export function updateNode(node: NetworkNode) {
  try {
    // Validate input
    if (!node || !node.id) {
      throw new Error('Invalid node: missing id');
    }
    if (typeof node.x !== 'number' || typeof node.y !== 'number' || !isFinite(node.x) || !isFinite(node.y)) {
      throw new Error('Invalid node: coordinates must be finite numbers');
    }
    if (!editingSession.processId) {
      throw new Error('No active process for node updates');
    }
    
    // Find and update local state
    const index = networkState.nodes.findIndex(n => n.id === node.id);
    if (index === -1) {
      throw new Error(`Node ${node.id} not found`);
    }
    
    networkState.nodes[index] = node;
    
    // Track the change for sync
    trackNodeChange(node, 'update');
    
    // Update visualization
    updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to update node locally:', error);
    throw error;
  }
}

// Update a link's properties (local only)
export function updateLink(link: NetworkLink) {
  try {
    // Validate input
    if (!link || !link.id) {
      throw new Error('Invalid link: missing id');
    }
    if (!link.from || !link.to) {
      throw new Error('Invalid link: missing from/to nodes');
    }
    if (link.from === link.to) {
      throw new Error('Invalid link: cannot connect node to itself');
    }
    if (!editingSession.processId) {
      throw new Error('No active process for link updates');
    }
    
    // Validate that referenced nodes exist
    const fromExists = networkState.nodes.some(n => n.id === link.from);
    const toExists = networkState.nodes.some(n => n.id === link.to);
    if (!fromExists || !toExists) {
      throw new Error(`Invalid link: referenced nodes ${link.from} or ${link.to} do not exist`);
    }
    
    // Find and update local state
    const index = networkState.links.findIndex(l => l.id === link.id);
    if (index === -1) {
      throw new Error(`Link ${link.id} not found`);
    }
    
    networkState.links[index] = link;
    
    // Track the change for sync
    trackLinkChange(link, 'update');
    
    // Update visualization
    updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to update link locally:', error);
    throw error;
  }
}

// Delete a node (and its connected links) - local only
export function deleteNode(nodeId: string) {
  try {
    // Find node before deletion for change tracking
    const nodeToDelete = networkState.nodes.find(n => n.id === nodeId);
    if (!nodeToDelete) {
      throw new Error(`Node ${nodeId} not found`);
    }
    
    // Update local state - remove node and connected links
    networkState.nodes = networkState.nodes.filter(n => n.id !== nodeId);
    networkState.links = networkState.links.filter(l => l.from !== nodeId && l.to !== nodeId);
    
    // Clear selection if deleted node was selected
    if (networkState.selection.selectedNodeId === nodeId) {
      networkState.selection.selectedNodeId = null;
    }
    
    // Track the deletion for sync
    trackNodeChange(nodeToDelete, 'delete');
    
    // Update visualization
    updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to delete node locally:', error);
    throw error;
  }
}

// Delete a link - local only
export function deleteLink(linkId: string) {
  try {
    // Find link before deletion for change tracking
    const linkToDelete = networkState.links.find(l => l.id === linkId);
    if (!linkToDelete) {
      throw new Error(`Link ${linkId} not found`);
    }
    
    // Update local state
    networkState.links = networkState.links.filter(l => l.id !== linkId);
    
    // Clear selection if deleted link was selected
    if (networkState.selection.selectedLinkId === linkId) {
      networkState.selection.selectedLinkId = null;
    }
    
    // Track the deletion for sync
    trackLinkChange(linkToDelete, 'delete');
    
    // Update visualization
    updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to delete link locally:', error);
    throw error;
  }
}

// Create a new node - local only
export function createNode(x: number, y: number): string {
  try {
    // Validate input
    if (typeof x !== 'number' || typeof y !== 'number' || !isFinite(x) || !isFinite(y)) {
      throw new Error('Invalid coordinates: must be finite numbers');
    }
    if (!editingSession.processId) {
      throw new Error('No active process for node creation');
    }
    
    // Generate unique ID with better collision avoidance
    const nodeId = `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // Check for ID collision (should be extremely rare)
    if (networkState.nodes.some(n => n.id === nodeId)) {
      throw new Error('Node ID collision detected, please retry');
    }
    
    // Add to local state
    const newNode: NetworkNode = {
      id: nodeId,
      label: nodeId,
      x,
      y
    };
    networkState.nodes = [...networkState.nodes, newNode];
    
    // Track the creation for sync
    trackNodeChange(newNode, 'add');
    
    // Update visualization
    updateNetworkVisualization();
    
    return nodeId;
  } catch (error) {
    console.error('Failed to create node locally:', error);
    throw error;
  }
}

// Create a new link - local only
export function createLink(fromNodeId: string, toNodeId: string): string {
  try {
    // Validate input
    if (!fromNodeId || !toNodeId) {
      throw new Error('Invalid input: missing node IDs');
    }
    if (fromNodeId === toNodeId) {
      throw new Error('Invalid link: cannot connect node to itself');
    }
    if (!editingSession.processId) {
      throw new Error('No active process for link creation');
    }
    
    // Validate that both nodes exist
    const fromNode = networkState.nodes.find(n => n.id === fromNodeId);
    const toNode = networkState.nodes.find(n => n.id === toNodeId);
    
    if (!fromNode || !toNode) {
      throw new Error(`Cannot create link: nodes ${fromNodeId} or ${toNodeId} not found`);
    }
    
    // Check for duplicate links
    const duplicateLink = networkState.links.find(l => 
      (l.from === fromNodeId && l.to === toNodeId) || 
      (l.from === toNodeId && l.to === fromNodeId)
    );
    if (duplicateLink) {
      throw new Error(`Link already exists between nodes ${fromNodeId} and ${toNodeId}`);
    }
    
    // Generate unique ID with better collision avoidance
    const linkId = `link_${fromNodeId}_${toNodeId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // Add to local state
    const newLink: NetworkLink = {
      id: linkId,
      label: linkId,
      from: fromNodeId,
      to: toNodeId
    };
    networkState.links = [...networkState.links, newLink];
    
    // Track the creation for sync
    trackLinkChange(newLink, 'add');
    
    // Update visualization
    updateNetworkVisualization();
    
    return linkId;
  } catch (error) {
    console.error('Failed to create link locally:', error);
    throw error;
  }
}