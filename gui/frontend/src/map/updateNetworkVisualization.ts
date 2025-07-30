import { mapState, clearNetwork, createNetworkNodeIcon } from './mapState.svelte';
import { networkState } from '$lib/stores/network.svelte';
import { appState } from '$lib/stores/app.svelte';
import L from 'leaflet';
import type { NetworkNode as MapNetworkNode, NetworkLink as MapNetworkLink } from './types';
import { 
  handleAddNodeClick, 
  setupLinkMode, 
  setupMoveNodesMode,
  updateNodeStyle 
} from './networkEditor';
import { trackNodeChange, trackLinkChange } from '../lib/utils/networkChangeTracking';
import { nanoid } from 'nanoid';

function generateNodeId(): string {
  return `node_${nanoid(10)}`;
}

function generateLinkId(fromId: string, toId: string): string {
  return `link_${fromId}_${toId}_${nanoid(10)}`;
}

export function updateNetworkVisualization() {
  if (!mapState.map || !mapState.networkLayer) return;

  clearNetwork();

  networkState.nodes.forEach(node => {
    const mapNode: MapNetworkNode = {
      id: node.id,
      marker: null as any,
      position: L.latLng(node.lat, node.lng),
      attributes: {
        label: node.label || node.id,
        type: node.type || 'default',
        capacity: node.capacity
      }
    };

    const nodeIcon = createNetworkNodeIcon(
      node.id,
      mapState.network.defaultNodeColor
    );
    
    const marker = L.marker(mapNode.position, {
      draggable: false,
      icon: nodeIcon
    }).addTo(mapState.networkLayer);

    mapNode.marker = marker;
    mapState.network.nodes.push(mapNode);

    // Set up marker events
    setupNodeMarkerEvents(marker, mapNode);
  });

  // Add links from network store to map
  networkState.links.forEach(link => {
    const fromNode = mapState.network.nodes.find(n => n.id === link.from);
    const toNode = mapState.network.nodes.find(n => n.id === link.to);

    if (fromNode && toNode) {
      const polyline = L.polyline(
        [fromNode.position, toNode.position],
        {
          color: mapState.network.defaultLinkColor,
          weight: 3,
          opacity: 0.8,
          interactive: true,
          pane: 'networkPane'
        }
      ).addTo(mapState.networkLayer);

      const mapLink: MapNetworkLink = {
        id: link.id,
        polyline,
        fromNodeId: link.from,
        toNodeId: link.to,
        attributes: {
          label: link.label || link.id,
          lanes: link.lanes || 1,
          speed: link.speed || 50,
          capacity: link.capacity
        }
      };

      mapState.network.links.push(mapLink);

      // Set up link events
      setupLinkEvents(polyline, mapLink);
    }
  });

  // Update dragging state based on mode
  updateNetworkMode();
}

// Set up events for node markers
function setupNodeMarkerEvents(marker: L.Marker, node: MapNetworkNode) {
  type NodeState = 'idle' | 'hovered' | 'dragging' | 'selected';
  let nodeState: NodeState = 'idle';

  // Drag events
  marker.on('dragstart', () => {
    nodeState = 'dragging';
    mapState.isDragging = true;
    
    // Set initial position at drag start
    const startPosition = marker.getLatLng();
    const storeNodeIndex = networkState.nodes.findIndex(n => n.id === node.id);
    if (storeNodeIndex !== -1) {
      networkState.nodes[storeNodeIndex] = {
        ...networkState.nodes[storeNodeIndex],
        lng: startPosition.lng,
        lat: startPosition.lat
      };
    }
    
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grabbing';
    }
  });

  marker.on('drag', () => {
    const newPosition = marker.getLatLng();
    node.position = newPosition;
    
    // Update network store coordinates in real-time
    const storeNodeIndex = networkState.nodes.findIndex(n => n.id === node.id);
    if (storeNodeIndex !== -1) {
      networkState.nodes[storeNodeIndex] = {
        ...networkState.nodes[storeNodeIndex],
        lng: newPosition.lng,
        lat: newPosition.lat
      };
    }
    
    // Update all connected links using the marker positions directly
    mapState.network.links.forEach(link => {
      if (link.fromNodeId === node.id || link.toNodeId === node.id) {
        // Find the actual nodes and get their current marker positions
        const fromNode = mapState.network.nodes.find(n => n.id === link.fromNodeId);
        const toNode = mapState.network.nodes.find(n => n.id === link.toNodeId);
        
        if (fromNode && toNode) {
          // Use the marker's current position for accurate positioning
          const fromPos = fromNode.marker.getLatLng();
          const toPos = toNode.marker.getLatLng();
          link.polyline.setLatLngs([fromPos, toPos]);
        }
      }
    });
  });

  marker.on('dragend', () => {
    nodeState = networkState.selection.selectedNodeId === node.id ? 'selected' : 'idle';
    mapState.isDragging = false;
    
    // Track the final position change
    const storeNodeIndex = networkState.nodes.findIndex(n => n.id === node.id);
    if (storeNodeIndex !== -1) {
      const updatedNode = networkState.nodes[storeNodeIndex];
      
      // Track the change
      trackNodeChange(updatedNode, 'update');
    }
    
    // Maintain selection color after drag
    const isSelected = networkState.selection.selectedNodeId === node.id;
    updateNodeStyle(node, isSelected);
    
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grab';
    }
  });

  // Hover events
  marker.on('mouseover', () => {
    if (nodeState === 'idle' && networkState.selection.selectedNodeId !== node.id) {
      nodeState = 'hovered';
      const hoverIcon = createNetworkNodeIcon(node.id, mapState.network.hoverNodeColor);
      marker.setIcon(hoverIcon);
    }
  });

  marker.on('mouseout', () => {
    if (nodeState === 'hovered') {
      nodeState = 'idle';
      // Check if this node is selected before resetting color
      if (networkState.selection.selectedNodeId !== node.id) {
        const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
        marker.setIcon(normalIcon);
      }
    }
  });

  // Click event for selection and link creation
  marker.on('click', (e: L.LeafletMouseEvent) => {
    
    if (appState.networkMode === 'EDIT') {
      // Check if clicking on already selected node
      if (networkState.selection.selectedNodeId === node.id) {
        // Deselect
        networkState.selection.selectedNodeId = null;
        updateNodeStyle(node, false);
        appState.secondarySidebar = 'HIDDEN';
      } else {
        // Select node in network store
        networkState.selection.selectedNodeId = node.id;
        networkState.selection.selectedLinkId = null;
        
        // Reset all links to default style
        mapState.network.links.forEach(l => {
          l.polyline.setStyle({ 
            color: mapState.network.defaultLinkColor,
            weight: 3,
            opacity: 0.8
          });
        });
        
        // Update visual selection
        mapState.network.nodes.forEach(n => {
          updateNodeStyle(n, n.id === node.id);
        });
        
        // Open secondary sidebar
        appState.secondarySidebar = 'EXPANDED';
      }
    } else if (appState.networkMode === 'CREATE') {
      // Import NetworkContent's state to check current action
      const networkContent = document.querySelector('[data-network-action]');
      const currentAction = networkContent?.getAttribute('data-network-action');
      
      if (currentAction === 'link') {
        // Handle link creation
        if (!networkState.selection.selectedNodeId) {
          // First node selection
          networkState.selection.selectedNodeId = node.id;
          updateNodeStyle(node, true);
        } else if (networkState.selection.selectedNodeId === node.id) {
          // Clicking same node - deselect
          networkState.selection.selectedNodeId = null;
          updateNodeStyle(node, false);
        } else {
          // Second node selection - create link
          const fromId = networkState.selection.selectedNodeId;
          const toId = node.id;
          
          // Generate unique link ID
          const linkId = generateLinkId(fromId, toId);
          
          // Create link in network store
          const newLink = {
            id: linkId,
            label: `Link ${networkState.links.length + 1}`,
            from: fromId,
            to: toId
          };
          networkState.links = [...networkState.links, newLink];
          
          // Track the change
          trackLinkChange(newLink, 'add');
          
          // Reset selection
          networkState.selection.selectedNodeId = null;
          
          // Update visualization to show new link
          updateNetworkVisualization();
          
        }
      }
    }
    e.originalEvent.stopPropagation();
  });
}

// Set up events for links
function setupLinkEvents(polyline: L.Polyline, link: MapNetworkLink) {
  type LinkState = 'idle' | 'hovered';
  let linkState: LinkState = 'idle';

  polyline.on('mouseover', () => {
    if (linkState === 'idle') {
      linkState = 'hovered';
      polyline.setStyle({ color: mapState.network.hoverLinkColor, weight: 4 });
    }
  });

  polyline.on('mouseout', () => {
    linkState = 'idle';
    polyline.setStyle({ color: mapState.network.defaultLinkColor, weight: 3 });
  });

  // Click event for selection
  polyline.on('click', (e: L.LeafletMouseEvent) => {
    if (appState.networkMode === 'EDIT') {
      // Check if clicking on already selected link
      if (networkState.selection.selectedLinkId === link.id) {
        // Deselect
        networkState.selection.selectedLinkId = null;
        link.polyline.setStyle({ 
          color: mapState.network.defaultLinkColor,
          weight: 3,
          opacity: 0.8
        });
        appState.secondarySidebar = 'HIDDEN';
      } else {
        // Select link in network store
        networkState.selection.selectedLinkId = link.id;
        networkState.selection.selectedNodeId = null;
        
        // Reset all nodes to default style
        mapState.network.nodes.forEach(n => {
          updateNodeStyle(n, false);
        });
        
        // Update visual selection for all links
        mapState.network.links.forEach(l => {
          if (l.id === link.id) {
            l.polyline.setStyle({ 
              color: '#f59e0b', // Orange for selected
              weight: 5,
              opacity: 1
            });
          } else {
            l.polyline.setStyle({ 
              color: mapState.network.defaultLinkColor,
              weight: 3,
              opacity: 0.8
            });
          }
        });
        
        // Open secondary sidebar
        appState.secondarySidebar = 'EXPANDED';
      }
    }
    e.originalEvent.stopPropagation();
  });
}

// Update network visualization based on current mode
function updateNetworkMode() {
  // In EDIT mode, allow dragging
  // In CREATE mode, disable dragging
  // In VIEW mode, disable dragging
  const isDraggingEnabled = appState.networkMode === 'EDIT';
  
  mapState.network.nodes.forEach(node => {
    if (isDraggingEnabled) {
      node.marker.dragging?.enable();
    } else {
      node.marker.dragging?.disable();
    }
  });
}

// Handle map click for adding nodes
export function handleNetworkMapClick(e: L.LeafletMouseEvent) {
  if (appState.networkMode === 'CREATE') {
    const networkContent = document.querySelector('[data-network-action]');
    const currentAction = networkContent?.getAttribute('data-network-action');
    
    if (currentAction === 'node') {
      const nodeId = generateNodeId();
      const latlng = e.latlng;
      
      // Add to network store
      const newNode = {
        id: nodeId,
        label: `Node ${networkState.nodes.length + 1}`,
        lng: latlng.lng,
        lat: latlng.lat
      };
      networkState.nodes = [...networkState.nodes, newNode];
      
      // Track the change
      trackNodeChange(newNode, 'add');
      
      // Update visualization
      updateNetworkVisualization();
    }
  }
}

// Export utility functions
export { generateNodeId, generateLinkId, updateNetworkMode };