import L from 'leaflet';
import type { NetworkNode, NetworkLink } from './types';
import { mapState, getCursorForMode, setTestMessage, createNetworkNodeIcon } from './mapState.svelte';

function getNodeDisplayId(nodeId: string): string {
  const match = nodeId.match(/node-(\d+)/);
  if (match && match[1]) {
    return match[1].slice(-4);
  }
  return nodeId;
}

export function handleAddNodeClick(event: L.LeafletMouseEvent) {
  if (!mapState.map || !mapState.networkLayer) return;
  
  const latlng = event.latlng;
  const timestamp = Date.now();
  const nodeId = `node-${timestamp}`;
  const nodeNumber = mapState.network.nodes.length + 1;
  
  const node: NetworkNode = {
    id: nodeId,
    marker: null as any,
    position: latlng,
    attributes: {
      label: `Node ${nodeNumber}`,
      type: 'default'
    }
  };
  
  const nodeIcon = createNetworkNodeIcon(nodeId, mapState.network.defaultNodeColor);
  
  const marker = L.marker(latlng, {
    draggable: false,
    icon: nodeIcon
  }).addTo(mapState.networkLayer);
  
  node.marker = marker;
  
  type NodeState = 'idle' | 'hovered' | 'dragging' | 'selected';
  let nodeState: NodeState = 'idle';
  
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
    const isSelected = mapState.network.selectedNodeId === node.id;
    updateNodeStyle(node, isSelected);
  });
  
  marker.on('mouseover', () => {
    if (nodeState === 'idle' && mapState.network.selectedNodeId !== node.id) {
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
}

export function setupLinkMode() {
  mapState.network.nodes.forEach(node => {
    node.marker.on('click', (e: L.LeafletMouseEvent) => {
      if (mapState.mode !== 'adding-links') return;
      
      if (!mapState.network.selectedNodeId) {
        mapState.network.selectedNodeId = node.id;
        updateNodeStyle(node, true);
      } else if (mapState.network.selectedNodeId === node.id) {
        mapState.network.selectedNodeId = null;
        updateNodeStyle(node, false);
      } else {
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
        
        updateNodeStyle(fromNode, false);
        mapState.network.selectedNodeId = null;
      }
      
      e.originalEvent.stopPropagation();
    });
  });
}

export function updateNodeStyle(node: NetworkNode, selected: boolean) {
  const color = selected ? '#f59e0b' : mapState.network.defaultNodeColor;
  const nodeIcon = createNetworkNodeIcon(node.id, color);
  if (selected) {
    nodeIcon.options.className = 'network-node-icon selected';
  }
  node.marker.setIcon(nodeIcon);
}

export function setupMoveNodesMode() {
  mapState.network.nodes.forEach(node => {
    node.marker.dragging?.enable();
    
    node.marker.off('drag');
    
    node.marker.on('drag', () => {
      const newPosition = node.marker.getLatLng();
      node.position = newPosition;
      
      mapState.network.links.forEach(link => {
        if (link.fromNodeId === node.id) {
          const toNode = mapState.network.nodes.find(n => n.id === link.toNodeId);
          if (toNode) {
            link.polyline.setLatLngs([newPosition, toNode.position]);
          }
        } else if (link.toNodeId === node.id) {
          const fromNode = mapState.network.nodes.find(n => n.id === link.fromNodeId);
          if (fromNode) {
            link.polyline.setLatLngs([fromNode.position, newPosition]);
          }
        }
      });
    });
  });
}

export function handleNodeMoveStart(_event: L.LeafletMouseEvent) {
}