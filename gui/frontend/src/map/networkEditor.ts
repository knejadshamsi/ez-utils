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
  const nodeIcon = createNetworkNodeIcon(timestamp.toString().slice(-4), mapState.network.defaultNodeColor);
  
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
      const hoverIcon = createNetworkNodeIcon(getNodeDisplayId(node.id), mapState.network.hoverNodeColor);
      marker.setIcon(hoverIcon);
    }
  });
  
  marker.on('mouseout', () => {
    if (nodeState === 'hovered') {
      nodeState = 'idle';
      const normalIcon = createNetworkNodeIcon(getNodeDisplayId(node.id), mapState.network.defaultNodeColor);
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

export function setupAttributeMode() {
  // Set up click handlers for nodes
  mapState.network.nodes.forEach(node => {
    node.marker.on('click', () => {
      if (mapState.mode !== 'editing-attributes' || !mapState.map) return;
      
      const popupContent = createNodeAttributePopup(node);
      
      L.popup()
        .setLatLng(node.position)
        .setContent(popupContent)
        .openOn(mapState.map);
    });
  });
  
  // Set up click handlers for links
  mapState.network.links.forEach(link => {
    link.polyline.on('click', (e: L.LeafletMouseEvent) => {
      if (mapState.mode !== 'editing-attributes' || !mapState.map) return;
      
      const popupContent = createLinkAttributePopup(link);
      
      L.popup()
        .setLatLng(e.latlng)
        .setContent(popupContent)
        .openOn(mapState.map);
    });
  });
}

function createNodeAttributePopup(node: NetworkNode): HTMLElement {
  const attrs = node.attributes;
  const container = document.createElement('div');
  container.innerHTML = `
    <div style="min-width: 200px;">
      <h4 style="margin: 0 0 10px 0;">Node Attributes</h4>
      <div style="margin: 10px 0; padding: 8px; background: #fef3c7; border-radius: 4px;">
        <div style="font-size: 12px; color: #92400e; margin-bottom: 5px;">Global State:</div>
        <div class="state-display" style="font-weight: bold; color: #d97706;">${mapState.testMessage}</div>
      </div>
      <table style="width: 100%; font-size: 12px;">
        <tr>
          <td><strong>ID:</strong></td>
          <td>${node.id}</td>
        </tr>
        <tr>
          <td><strong>Label:</strong></td>
          <td>${attrs.label || 'N/A'}</td>
        </tr>
        <tr>
          <td><strong>Type:</strong></td>
          <td>${attrs.type || 'default'}</td>
        </tr>
        <tr>
          <td><strong>Position:</strong></td>
          <td>[${node.position.lng.toFixed(4)}, ${node.position.lat.toFixed(4)}]</td>
        </tr>
      </table>
      <div style="margin: 10px 0;">
        <input type="text" 
               class="state-input"
               placeholder="Update state from node" 
               style="width: 100%; padding: 4px; border: 1px solid #e5e7eb; border-radius: 4px; margin-bottom: 5px;">
        <button class="update-btn" 
                style="width: 100%; padding: 4px 8px; background: #ef4444; color: white; border: none; border-radius: 4px; cursor: pointer;">
          Update from Node Popup
        </button>
      </div>
      <button onclick="alert('Edit functionality coming soon!')" 
              style="margin-top: 10px; width: 100%; padding: 5px; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer;">
        Edit Attributes
      </button>
    </div>
  `;
  
  // Add event listeners
  const input = container.querySelector('.state-input') as HTMLInputElement;
  const updateBtn = container.querySelector('.update-btn') as HTMLButtonElement;
  const stateDisplay = container.querySelector('.state-display') as HTMLDivElement;
  
  updateBtn.onclick = () => {
    if (input.value) {
      setTestMessage(input.value);
      // Update the display in the popup immediately
      stateDisplay.textContent = input.value;
      input.value = '';
    }
  };
  
  return container;
}

function createLinkAttributePopup(link: NetworkLink): string {
  const attrs = link.attributes;
  const fromNode = mapState.network.nodes.find(n => n.id === link.fromNodeId);
  const toNode = mapState.network.nodes.find(n => n.id === link.toNodeId);
  
  return `
    <div style="min-width: 200px;">
      <h4 style="margin: 0 0 10px 0;">Link Attributes</h4>
      <table style="width: 100%; font-size: 12px;">
        <tr>
          <td><strong>ID:</strong></td>
          <td>${link.id}</td>
        </tr>
        <tr>
          <td><strong>From:</strong></td>
          <td>${fromNode?.attributes.label || 'Unknown'}</td>
        </tr>
        <tr>
          <td><strong>To:</strong></td>
          <td>${toNode?.attributes.label || 'Unknown'}</td>
        </tr>
        <tr>
          <td><strong>Lanes:</strong></td>
          <td>${attrs.lanes || 1}</td>
        </tr>
        <tr>
          <td><strong>Speed:</strong></td>
          <td>${attrs.speed || 50} km/h</td>
        </tr>
      </table>
      <button onclick="alert('Edit functionality coming soon!')" 
              style="margin-top: 10px; width: 100%; padding: 5px; background: #10b981; color: white; border: none; border-radius: 4px; cursor: pointer;">
        Edit Attributes
      </button>
    </div>
  `;
}

export function updateNodeStyle(node: NetworkNode, selected: boolean) {
  const color = selected ? '#f59e0b' : mapState.network.defaultNodeColor;
  const nodeIcon = createNetworkNodeIcon(getNodeDisplayId(node.id), color);
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
    
    // Update connected links when dragging
    node.marker.on('drag', () => {
      const newPosition = node.marker.getLatLng();
      node.position = newPosition;
      
      // Update all connected links
      mapState.network.links.forEach(link => {
        if (link.fromNodeId === node.id || link.toNodeId === node.id) {
          const fromNode = mapState.network.nodes.find(n => n.id === link.fromNodeId);
          const toNode = mapState.network.nodes.find(n => n.id === link.toNodeId);
          
          if (fromNode && toNode) {
            link.polyline.setLatLngs([fromNode.position, toNode.position]);
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