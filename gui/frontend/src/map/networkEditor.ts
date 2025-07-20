import L from 'leaflet';
import type { NetworkNode, NetworkLink } from './types';
import { mapState, getCursorForMode, setTestMessage, createNetworkNodeIcon } from './mapState.svelte';

// Helper to extract display ID from node ID
function getNodeDisplayId(nodeId: string): string {
  // Extract just the timestamp numbers from "node-1234567890"
  const match = nodeId.match(/node-(\d+)/);
  if (match && match[1]) {
    return match[1].slice(-4); // Last 4 digits
  }
  return nodeId; // Fallback to full ID if pattern doesn't match
}

// Network-specific handlers
export function handleAddNodeClick(event: L.LeafletMouseEvent) {
  if (!mapState.map || !mapState.networkLayer) return;
  
  const latlng = event.latlng;
  const timestamp = Date.now();
  const nodeId = `node-${timestamp}`;
  const nodeNumber = mapState.network.nodes.length + 1;
  
  // Create node object with default attributes
  const node: NetworkNode = {
    id: nodeId,
    marker: null as any, // Will be set below
    position: latlng,
    attributes: {
      label: `Node ${nodeNumber}`,
      type: 'default'
    }
  };
  
  // Create node marker with custom icon and color
  const nodeIcon = createNetworkNodeIcon(nodeId, mapState.network.defaultNodeColor);
  
  const marker = L.marker(latlng, {
    draggable: false,  // Will be enabled in move mode
    icon: nodeIcon
  }).addTo(mapState.networkLayer);
  
  node.marker = marker;
  
  // Set up hover handlers for the node with state tracking
  type NodeState = 'idle' | 'hovered' | 'dragging' | 'selected';
  let nodeState: NodeState = 'idle';
  
  // Track dragging state
  marker.on('dragstart', () => {
    nodeState = 'dragging';
    mapState.isDragging = true;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = 'grabbing';
    }
  });
  
  marker.on('dragend', () => {
    nodeState = mapState.network.selectedNodeId === node.id ? 'selected' : 'idle';
    mapState.isDragging = false;
    if (mapState.map) {
      mapState.map.getContainer().style.cursor = getCursorForMode(mapState.mode);
    }
    // Reset icon to appropriate color after drag
    const isSelected = mapState.network.selectedNodeId === node.id;
    updateNodeStyle(node, isSelected);
  });
  
  marker.on('mouseover', () => {
    if (nodeState === 'idle' && mapState.network.selectedNodeId !== node.id) { // Don't change color if selected
      nodeState = 'hovered';
      const hoverIcon = createNetworkNodeIcon(node.id, mapState.network.hoverNodeColor);
      marker.setIcon(hoverIcon);
    }
  });
  
  marker.on('mouseout', () => {
    if (nodeState === 'hovered') {
      nodeState = 'idle';
      const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
      marker.setIcon(normalIcon);
    }
  });
  
  mapState.network.nodes.push(node);
  console.log(`Added network node ${nodeNumber} at [${latlng.lng}, ${latlng.lat}]`);
}

export function setupLinkMode() {
  // Set up click handlers for nodes to create links
  mapState.network.nodes.forEach(node => {
    node.marker.on('click', (e: L.LeafletMouseEvent) => {
      if (mapState.mode !== 'adding-links') return;
      
      if (!mapState.network.selectedNodeId) {
        // First node selection
        mapState.network.selectedNodeId = node.id;
        updateNodeStyle(node, true);
        console.log(`Selected first node: ${node.attributes.label}`);
      } else if (mapState.network.selectedNodeId === node.id) {
        // Clicking the same node - deselect it
        mapState.network.selectedNodeId = null;
        updateNodeStyle(node, false);
        console.log(`Deselected node: ${node.attributes.label}`);
      } else {
        // Second node selection - create link
        const fromNode = mapState.network.nodes.find(n => n.id === mapState.network.selectedNodeId);
        if (!fromNode || !mapState.networkLayer) return;
        
        const linkId = `link-${Date.now()}`;
        const polyline = L.polyline(
          [fromNode.position, node.position],
          {
            color: mapState.network.defaultLinkColor,
            weight: 3,
            opacity: 0.8
          }
        ).addTo(mapState.networkLayer);
        
        const link: NetworkLink = {
          id: linkId,
          polyline,
          fromNodeId: fromNode.id,
          toNodeId: node.id,
          attributes: {
            label: `${fromNode.attributes.label} → ${node.attributes.label}`,
            lanes: 1,
            speed: 50
          }
        };
        
        // Set up hover handlers for the link with state tracking
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
        
        mapState.network.links.push(link);
        console.log(`Created link from ${fromNode.attributes.label} to ${node.attributes.label}`);
        
        // Reset selection
        updateNodeStyle(fromNode, false);
        mapState.network.selectedNodeId = null;
      }
      
      e.originalEvent.stopPropagation();
    });
  });
}

// Attribute editing is now handled via sidebar, not popups

// Popover functions removed - using sidebar for editing

export function updateNodeStyle(node: NetworkNode, selected: boolean) {
  const color = selected ? '#f59e0b' : mapState.network.defaultNodeColor;
  const nodeIcon = createNetworkNodeIcon(node.id, color);
  if (selected) {
    nodeIcon.options.className = 'network-node-icon selected';
  }
  node.marker.setIcon(nodeIcon);
}

export function setupMoveNodesMode() {
  // Enable dragging on all network nodes
  mapState.network.nodes.forEach(node => {
    node.marker.dragging?.enable();
    
    // Remove any existing drag handlers first to prevent duplicates
    node.marker.off('drag');
    
    // Update position and connected links when dragging
    node.marker.on('drag', () => {
      const newPosition = node.marker.getLatLng();
      node.position = newPosition;
      
      // Update all connected links
      mapState.network.links.forEach(link => {
        if (link.fromNodeId === node.id) {
          // This node is the 'from' node of the link
          const toNode = mapState.network.nodes.find(n => n.id === link.toNodeId);
          if (toNode) {
            link.polyline.setLatLngs([newPosition, toNode.position]);
          }
        } else if (link.toNodeId === node.id) {
          // This node is the 'to' node of the link
          const fromNode = mapState.network.nodes.find(n => n.id === link.fromNodeId);
          if (fromNode) {
            link.polyline.setLatLngs([fromNode.position, newPosition]);
          }
        }
      });
    });
  });
  
  console.log('Move nodes mode activated');
}

export function handleNodeMoveStart(_event: L.LeafletMouseEvent) {
  // This is handled by Leaflet's built-in dragging
  // We can add additional logic here if needed
}