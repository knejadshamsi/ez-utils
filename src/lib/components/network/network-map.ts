import L from 'leaflet';

import type { NetworkLinkPayload, NetworkNodePayload } from '$lib/components/network/types';

export interface NetworkLayers {
  nodeLayer: L.LayerGroup;
  linkLayer: L.LayerGroup;
  dragLayer: L.LayerGroup;
}

export interface NetworkLayerOpacity {
  /** Range [0, 1]. 0 means skip rendering entirely. */
  nodes: number;
  links: number;
}

export function ensureNetworkLayers(map: L.Map): NetworkLayers {
  return {
    linkLayer: L.layerGroup().addTo(map),
    nodeLayer: L.layerGroup().addTo(map),
    dragLayer: L.layerGroup().addTo(map),
  };
}

export function clearNetworkLayers(layers: NetworkLayers) {
  layers.linkLayer.clearLayers();
  layers.nodeLayer.clearLayers();
  layers.dragLayer.clearLayers();
}

/**
 * Renders nodes as circleMarkers and links as polylines on their respective
 * layers. Returns a map of link id -> polyline so callers (e.g. the drag
 * overlay) can mutate endpoints live without a full re-render.
 */
export function renderNetwork(
  layers: NetworkLayers,
  nodes: NetworkNodePayload[],
  links: NetworkLinkPayload[],
  sourceColor: string,
  opacity: NetworkLayerOpacity,
  onNodeClick: (nodeId: string) => void,
  onLinkClick: (linkId: string) => void,
): {
  nodeMarkers: Map<string, L.CircleMarker>;
  linkPolylines: Map<string, L.Polyline>;
} {
  layers.nodeLayer.clearLayers();
  layers.linkLayer.clearLayers();

  const linkPolylines = new Map<string, L.Polyline>();
  if (opacity.links > 0) {
    for (const link of links) {
      const polyline = L.polyline(
        [[link.fromLat, link.fromLng], [link.toLat, link.toLng]],
        { color: sourceColor, weight: 2, opacity: 0.8 * opacity.links },
      );
      polyline.on('click', () => onLinkClick(link.id));
      polyline.addTo(layers.linkLayer);
      linkPolylines.set(link.id, polyline);
    }
  }

  const nodeMarkers = new Map<string, L.CircleMarker>();
  if (opacity.nodes > 0) {
    for (const node of nodes) {
      const base = node.ghost ? 0.3 : 0.85;
      const marker = L.circleMarker([node.lat, node.lng], {
        radius: node.ghost ? 2 : 4,
        color: sourceColor,
        weight: 1,
        fillColor: sourceColor,
        opacity: base * opacity.nodes,
        fillOpacity: base * opacity.nodes,
        interactive: !node.ghost,
      });
      if (!node.ghost) {
        marker.on('click', () => onNodeClick(node.id));
      }
      marker.addTo(layers.nodeLayer);
      nodeMarkers.set(node.id, marker);
    }
  }

  return { nodeMarkers, linkPolylines };
}

/**
 * Dims every node and link that is not the selected link or one of its
 * endpoints. When `selectedLinkId` is null, clears any previous dimming.
 */
export function applyFocusHighlight(
  nodeMarkers: Map<string, L.CircleMarker>,
  linkPolylines: Map<string, L.Polyline>,
  selectedLinkId: string | null,
  selectedNodeId: string | null,
  connectedNodeIds: Set<string>,
  connectedLinkIds: Set<string>,
  opacity: NetworkLayerOpacity,
) {
  const dimFactor = 0.18;
  const fullNode = 0.85 * opacity.nodes;
  const fullLink = 0.8 * opacity.links;
  const dimNode = fullNode * dimFactor;
  const dimLink = fullLink * dimFactor;

  const hasSelection = selectedLinkId || selectedNodeId;

  for (const [id, polyline] of linkPolylines) {
    let on = true;
    if (selectedLinkId) {
      on = id === selectedLinkId;
    } else if (selectedNodeId) {
      on = connectedLinkIds.has(id);
    }
    polyline.setStyle({ opacity: hasSelection ? (on ? fullLink : dimLink) : fullLink });
  }

  for (const [id, marker] of nodeMarkers) {
    let on = true;
    if (selectedLinkId) {
      on = connectedNodeIds.has(id);
    } else if (selectedNodeId) {
      on = connectedNodeIds.has(id);
    }
    const v = hasSelection ? (on ? fullNode : dimNode) : fullNode;
    marker.setStyle({ opacity: v, fillOpacity: v });
  }
}

/**
 * Adds a visible ring around the given node to mark it as the current
 * "from" node during the two-click create-link flow. Pass null to clear.
 */
export function highlightCreateLinkFromNode(
  nodeMarkers: Map<string, L.CircleMarker>,
  fromNodeId: string | null,
) {
  for (const [id, marker] of nodeMarkers) {
    if (id === fromNodeId) {
      marker.setStyle({ weight: 3, radius: 7 });
    } else {
      marker.setStyle({ weight: 1, radius: 4 });
    }
  }
}

/**
 * Attaches invisible draggable markers per non-ghost node. While dragging,
 * updates the node's circleMarker and every connected link's polyline endpoint
 * in real time, mirroring the activity-drag pattern from the population map.
 */
export function renderDragOverlay(
  layers: NetworkLayers,
  nodes: NetworkNodePayload[],
  links: NetworkLinkPayload[],
  nodeMarkers: Map<string, L.CircleMarker>,
  linkPolylines: Map<string, L.Polyline>,
  onDragEnd: (nodeId: string, lng: number, lat: number) => void,
) {
  layers.dragLayer.clearLayers();

  // Index links by endpoint node so drag can find affected polylines quickly.
  const linksByNode = new Map<string, { link: NetworkLinkPayload; endpoint: 'from' | 'to' }[]>();
  for (const link of links) {
    if (!linksByNode.has(link.fromNode)) linksByNode.set(link.fromNode, []);
    if (!linksByNode.has(link.toNode)) linksByNode.set(link.toNode, []);
    linksByNode.get(link.fromNode)!.push({ link, endpoint: 'from' });
    linksByNode.get(link.toNode)!.push({ link, endpoint: 'to' });
  }

  for (const node of nodes) {
    if (node.ghost) continue;
    const handle = L.marker([node.lat, node.lng], { draggable: true, opacity: 0 });
    handle.on('drag', (event: L.LeafletEvent) => {
      const pos = (event.target as L.Marker).getLatLng();
      nodeMarkers.get(node.id)?.setLatLng(pos);
      const affected = linksByNode.get(node.id);
      if (!affected) return;
      for (const { link, endpoint } of affected) {
        const polyline = linkPolylines.get(link.id);
        if (!polyline) continue;
        const coords = polyline.getLatLngs() as L.LatLng[];
        if (endpoint === 'from') coords[0] = pos;
        else coords[1] = pos;
        polyline.setLatLngs(coords);
      }
    });
    handle.on('dragend', (event: L.LeafletEvent) => {
      const pos = (event.target as L.Marker).getLatLng();
      onDragEnd(node.id, pos.lng, pos.lat);
    });
    handle.addTo(layers.dragLayer);
  }
}
