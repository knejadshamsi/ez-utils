import { GetNodesInBBox, GetLinksInBBox } from '@wailsjs/go/gui/App';
import type { gui } from '@wailsjs/go/models';
import { networkState, type NetworkNode, type NetworkLink } from '$lib/stores/network.svelte';
import { appState } from '$lib/stores/app.svelte';
import { trackNodeChange, trackLinkChange } from '$lib/utils/networkChangeTracking';
import { updateNetworkVisualization } from '../../map/updateNetworkVisualization';

export async function loadNetworkInBBox(bbox: gui.BoundingBox, appendData: boolean = false) {
  try {
    const processId = appState.processId;
    
        const nodes = await GetNodesInBBox(processId, bbox);
    
        const nodeIds = nodes.map(n => n.id);
    
        const links = await GetLinksInBBox(processId, nodeIds);
    
        const newNodes = nodes.map(n => ({
      id: n.id,
      label: n.id,
      lng: n.lng,
      lat: n.lat
    }));
    
    const newLinks = links.map(l => ({
      id: l.id,
      label: l.id,
      from: l.from_node,
      to: l.to_node
    }));
    
    if (appendData) {
            const existingNodeIds = new Set(networkState.nodes.map(n => n.id));
      const existingLinkIds = new Set(networkState.links.map(l => l.id));
      
      const uniqueNewNodes = newNodes.filter(n => !existingNodeIds.has(n.id));
      const uniqueNewLinks = newLinks.filter(l => !existingLinkIds.has(l.id));
      
      networkState.nodes = [...networkState.nodes, ...uniqueNewNodes];
      networkState.links = [...networkState.links, ...uniqueNewLinks];
    } else {
            networkState.nodes = newNodes;
      networkState.links = newLinks;
    }
    
        updateNetworkVisualization();
    
  } catch (error) {
    console.error('Failed to load network data:', error);
    throw error;
  }
}

export async function loadAllNetworkData() {
    const bbox: gui.BoundingBox = {
    north: 90,
    south: -90,
    east: 180,
    west: -180
  };
  
  return loadNetworkInBBox(bbox);
}

export async function loadInitialViewportData(mapInstance: L.Map) {
  const bounds = mapInstance.getBounds();
  const center = mapInstance.getCenter();
  
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

export async function loadNetworkInPolygon(polygonCoords: number[][]) {
  try {
        if (!Array.isArray(polygonCoords)) {
      throw new Error('Invalid polygon: coordinates must be an array');
    }
    if (polygonCoords.length === 0) {
      throw new Error('Invalid polygon: empty coordinates array');
    }
    if (polygonCoords.length < 3) {
      throw new Error('Invalid polygon: minimum 3 points required');
    }
    
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
    
        const { changeTracker } = await import('$lib/changeTracker.svelte');
    
    if (changeTracker.pendingChanges.length > 0) {
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
    
        const lats = polygonCoords.map(coord => coord[0]);
    const lngs = polygonCoords.map(coord => coord[1]);
    
    const bbox: gui.BoundingBox = {
      north: Math.max(...lats),
      south: Math.min(...lats),
      east: Math.max(...lngs),
      west: Math.min(...lngs)
    };
    
        if (bbox.north === bbox.south && bbox.east === bbox.west) {
      throw new Error('Invalid polygon: all points are the same location');
    }
    
    return loadNetworkInBBox(bbox, true);   } catch (error) {
    console.error('Failed to load network data for polygon:', error);
    throw error;
  }
}

export function updateNode(node: NetworkNode) {
  try {
        if (!node || !node.id) {
      throw new Error('Invalid node: missing id');
    }
    if (typeof node.lng !== 'number' || typeof node.lat !== 'number' || !isFinite(node.lng) || !isFinite(node.lat)) {
      throw new Error('Invalid node: coordinates must be finite numbers');
    }
    if (!appState.processId) {
      throw new Error('No active process for node updates');
    }
    
        const index = networkState.nodes.findIndex(n => n.id === node.id);
    if (index === -1) {
      throw new Error(`Node ${node.id} not found`);
    }
    
    networkState.nodes[index] = node;
    
        trackNodeChange(node, 'update');
    
        updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to update node locally:', error);
    throw error;
  }
}

export function updateLink(link: NetworkLink) {
  try {
        if (!link || !link.id) {
      throw new Error('Invalid link: missing id');
    }
    if (!link.from || !link.to) {
      throw new Error('Invalid link: missing from/to nodes');
    }
    if (link.from === link.to) {
      throw new Error('Invalid link: cannot connect node to itself');
    }
    if (!appState.processId) {
      throw new Error('No active process for link updates');
    }
    
        const fromExists = networkState.nodes.some(n => n.id === link.from);
    const toExists = networkState.nodes.some(n => n.id === link.to);
    if (!fromExists || !toExists) {
      throw new Error(`Invalid link: referenced nodes ${link.from} or ${link.to} do not exist`);
    }
    
        const index = networkState.links.findIndex(l => l.id === link.id);
    if (index === -1) {
      throw new Error(`Link ${link.id} not found`);
    }
    
    networkState.links[index] = link;
    
        trackLinkChange(link, 'update');
    
        updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to update link locally:', error);
    throw error;
  }
}

export function deleteNode(nodeId: string) {
  try {
        const nodeToDelete = networkState.nodes.find(n => n.id === nodeId);
    if (!nodeToDelete) {
      throw new Error(`Node ${nodeId} not found`);
    }
    
        networkState.nodes = networkState.nodes.filter(n => n.id !== nodeId);
    networkState.links = networkState.links.filter(l => l.from !== nodeId && l.to !== nodeId);
    
        if (networkState.selection.selectedNodeId === nodeId) {
      networkState.selection.selectedNodeId = null;
    }
    
        trackNodeChange(nodeToDelete, 'delete');
    
        updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to delete node locally:', error);
    throw error;
  }
}

export function deleteLink(linkId: string) {
  try {
        const linkToDelete = networkState.links.find(l => l.id === linkId);
    if (!linkToDelete) {
      throw new Error(`Link ${linkId} not found`);
    }
    
        networkState.links = networkState.links.filter(l => l.id !== linkId);
    
        if (networkState.selection.selectedLinkId === linkId) {
      networkState.selection.selectedLinkId = null;
    }
    
        trackLinkChange(linkToDelete, 'delete');
    
        updateNetworkVisualization();
  } catch (error) {
    console.error('Failed to delete link locally:', error);
    throw error;
  }
}

export function createNode(lng: number, lat: number): string {
  try {
        if (typeof lng !== 'number' || typeof lat !== 'number' || !isFinite(lng) || !isFinite(lat)) {
      throw new Error('Invalid coordinates: must be finite numbers');
    }
    if (!appState.processId) {
      throw new Error('No active process for node creation');
    }
    
        const nodeId = `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
        if (networkState.nodes.some(n => n.id === nodeId)) {
      throw new Error('Node ID collision detected, please retry');
    }
    
        const newNode: NetworkNode = {
      id: nodeId,
      label: nodeId,
      lng,
      lat
    };
    networkState.nodes = [...networkState.nodes, newNode];
    
        trackNodeChange(newNode, 'add');
    
        updateNetworkVisualization();
    
    return nodeId;
  } catch (error) {
    console.error('Failed to create node locally:', error);
    throw error;
  }
}

export function createLink(fromNodeId: string, toNodeId: string): string {
  try {
        if (!fromNodeId || !toNodeId) {
      throw new Error('Invalid input: missing node IDs');
    }
    if (fromNodeId === toNodeId) {
      throw new Error('Invalid link: cannot connect node to itself');
    }
    if (!appState.processId) {
      throw new Error('No active process for link creation');
    }
    
        const fromNode = networkState.nodes.find(n => n.id === fromNodeId);
    const toNode = networkState.nodes.find(n => n.id === toNodeId);
    
    if (!fromNode || !toNode) {
      throw new Error(`Cannot create link: nodes ${fromNodeId} or ${toNodeId} not found`);
    }
    
        const duplicateLink = networkState.links.find(l => 
      (l.from === fromNodeId && l.to === toNodeId) || 
      (l.from === toNodeId && l.to === fromNodeId)
    );
    if (duplicateLink) {
      throw new Error(`Link already exists between nodes ${fromNodeId} and ${toNodeId}`);
    }
    
        const linkId = `link_${fromNodeId}_${toNodeId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
        const newLink: NetworkLink = {
      id: linkId,
      label: linkId,
      from: fromNodeId,
      to: toNodeId
    };
    networkState.links = [...networkState.links, newLink];
    
        trackLinkChange(newLink, 'add');
    
        updateNetworkVisualization();
    
    return linkId;
  } catch (error) {
    console.error('Failed to create link locally:', error);
    throw error;
  }
}