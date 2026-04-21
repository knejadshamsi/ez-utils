<script lang="ts">
  import L from 'leaflet';

  import { mapViewport, drawers } from '$lib/stores/ui.svelte';
  import { network, NETWORK_ZOOM_THRESHOLD } from '$lib/stores/network.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import {
    moveNode as moveNetworkNode,
    createLink as createNetworkLink,
  } from '$lib/stores/network-actions';
  import {
    renderNetwork,
    renderDragOverlay,
    applyFocusHighlight,
    highlightCreateLinkFromNode,
    type NetworkLayers,
  } from '$lib/components/network/network-map';

  interface Props {
    map: L.Map;
    layers: NetworkLayers;
  }

  let { map, layers }: Props = $props();

  let nodeMarkers = new Map();
  let linkPolylines = new Map();

  // Re-render network layers whenever data, visibility, or source color changes.
  $effect(() => {
    const result = renderNetwork(
      layers,
      network.mapNodes,
      network.mapLinks,
      network.activeSourceColor,
      { nodes: network.nodeOpacity, links: network.linkOpacity },
      (nodeId) => {
        if (network.mapAction === 'add_link_from' || network.mapAction === 'add_link_to') {
          const pair = network.advanceAddLink(nodeId);
          if (pair) {
            const id = network.generateLinkId();
            void createNetworkLink({ id, fromNode: pair.from, toNode: pair.to });
          }
          return;
        }
        void network.selectNode(nodeId);
      },
      (linkId) => void network.selectLink(linkId),
    );
    nodeMarkers = result.nodeMarkers;
    linkPolylines = result.linkPolylines;
  });

  // Apply focus highlighting whenever selection changes.
  $effect(() => {
    const selection = network.selection;
    const detail = network.selectedLinkDetail;
    const connectedNodes = new Set<string>();
    const connectedLinks = new Set<string>();
    let selectedLinkId: string | null = null;
    let selectedNodeId: string | null = null;

    if (selection?.type === 'link') {
      selectedLinkId = selection.id;
      if (detail) {
        connectedNodes.add(detail.fromNode);
        connectedNodes.add(detail.toNode);
      }
    } else if (selection?.type === 'node') {
      selectedNodeId = selection.id;
      connectedNodes.add(selection.id);
      for (const link of network.links) {
        if (link.fromNode === selection.id) {
          connectedNodes.add(link.toNode);
          connectedLinks.add(link.id);
        } else if (link.toNode === selection.id) {
          connectedNodes.add(link.fromNode);
          connectedLinks.add(link.id);
        }
      }
    }

    applyFocusHighlight(
      nodeMarkers,
      linkPolylines,
      selectedLinkId,
      selectedNodeId,
      connectedNodes,
      connectedLinks,
      { nodes: network.nodeOpacity, links: network.linkOpacity },
    );
  });

  // Fly to selection when it changes.
  let previousFlyKey: string | null = null;
  $effect(() => {
    const selection = network.selection;
    const detail = network.selectedLinkDetail;

    let flyKey: string | null = null;
    if (selection?.type === 'link' && detail) {
      flyKey = `link:${detail.id}`;
    } else if (selection?.type === 'node') {
      flyKey = `node:${selection.id}`;
    }

    if (flyKey === previousFlyKey) return;
    previousFlyKey = flyKey;

    if (selection?.type === 'link' && detail) {
      const bounds = L.latLngBounds(
        [detail.fromLat, detail.fromLng],
        [detail.toLat, detail.toLng],
      );
      map.flyToBounds(bounds, { padding: [50, 50], maxZoom: 17 });
    } else if (selection?.type === 'node') {
      const node = network.nodesById.get(selection.id);
      if (node) {
        map.flyTo([node.lat, node.lng], Math.max(map.getZoom(), 16));
      }
    }
  });

  // Visual ring on the "from" node while in two-click create-link flow.
  $effect(() => {
    highlightCreateLinkFromNode(nodeMarkers, network.createLinkFromId);
  });

  // Drag overlay is attached only when free-move is on.
  $effect(() => {
    if (network.nodeDragEnabled) {
      renderDragOverlay(
        layers,
        network.mapNodes,
        network.mapLinks,
        nodeMarkers,
        linkPolylines,
        (nodeId, lng, lat) => {
          void moveNetworkNode(nodeId, lng, lat);
        },
      );
    } else {
      layers.dragLayer.clearLayers();
    }
  });

  // Open primary drawer for network at valid zoom level.
  $effect(() => {
    if (sources.activeKind === 'network' && mapViewport.state.zoom >= NETWORK_ZOOM_THRESHOLD) {
      drawers.open('primary');
    }
  });
</script>
