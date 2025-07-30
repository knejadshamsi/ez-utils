<script lang="ts">
  import { Button, Label, ButtonGroup, Hr, Heading, P } from 'flowbite-svelte';
  import { PlusOutline } from 'flowbite-svelte-icons';
  import { appState } from '$lib/stores/app.svelte';
  import { networkState, selectNode as selectNetworkNode, selectLink as selectNetworkLink } from '$lib/stores/network.svelte';
  import { mapState, createNetworkNodeIcon } from '../../map/mapState.svelte';
  import { startPolygonDrawing, stopPolygonDrawing, clearAllZones } from '../../map/polygonDrawing';
  import { loadNetworkInBBox } from '$lib/api/network';
  import { updateNetworkMode } from '../../map/updateNetworkVisualization';
  import type { gui } from '@wailsjs/go/models';
  
  let drawButtonText = $state('Draw Polygon');
  
  // Make node dragging reactive to mode changes
  $effect(() => {
    // This will run whenever appState.networkMode changes
    if (mapState.map && mapState.networkLayer) {
      updateNetworkMode();
    }
  });
  
  // Auto-start polygon drawing in VIEW mode
  $effect(() => {
    if (appState.networkMode === 'VIEW' && mapState.map && mapState.mode !== 'drawing-polygon') {
      handleDrawPolygon();
    }
  });
  
  function handleNodeSelect(nodeId: string) {
    // In CREATE mode, switch to EDIT and show secondary sidebar
    if (appState.networkMode === 'CREATE') {
      appState.networkMode = 'EDIT';
      selectNetworkNode(nodeId);
      appState.secondarySidebar = 'EXPANDED';
      
      // Update visual selection
      mapState.network.nodes.forEach(node => {
        if (node.id === nodeId) {
          const selectedIcon = createNetworkNodeIcon(node.id, '#f59e0b');
          node.marker.setIcon(selectedIcon);
        } else {
          const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
          node.marker.setIcon(normalIcon);
        }
      });
      
      // Scroll to selected node in list
      scrollToNode(nodeId);
      return;
    }
    
    // Check if clicking on already selected node
    if (networkState.selection.selectedNodeId === nodeId) {
      // Deselect
      networkState.selection.selectedNodeId = null;
      appState.secondarySidebar = 'HIDDEN';
      
      // Reset node color
      const node = mapState.network.nodes.find(n => n.id === nodeId);
      if (node) {
        const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
        node.marker.setIcon(normalIcon);
      }
    } else {
      // Select new node
      selectNetworkNode(nodeId);
      appState.secondarySidebar = 'EXPANDED';
      
      // Reset all links to default color first
      mapState.network.links.forEach(link => {
        link.polyline.setStyle({ 
          color: mapState.network.defaultLinkColor,
          weight: 3,
          opacity: 0.8
        });
      });
      
      // Update map visualization to show selection
      mapState.network.nodes.forEach(node => {
        if (node.id === nodeId) {
          // Selected node - orange color
          const selectedIcon = createNetworkNodeIcon(node.id, '#f59e0b');
          node.marker.setIcon(selectedIcon);
        } else {
          // Unselected nodes - default color
          const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
          node.marker.setIcon(normalIcon);
        }
      });
      
      // Scroll to selected node in list
      scrollToNode(nodeId);
    }
  }
  
  function handleLinkSelect(linkId: string) {
    // In CREATE mode, switch to EDIT and show secondary sidebar
    if (appState.networkMode === 'CREATE') {
      appState.networkMode = 'EDIT';
      selectNetworkLink(linkId);
      appState.secondarySidebar = 'EXPANDED';
      
      // Update visual selection
      mapState.network.links.forEach(link => {
        if (link.id === linkId) {
          link.polyline.setStyle({ 
            color: '#f59e0b',
            weight: 5,
            opacity: 1
          });
        } else {
          link.polyline.setStyle({ 
            color: mapState.network.defaultLinkColor,
            weight: 3,
            opacity: 0.8
          });
        }
      });
      
      // Scroll to selected link in list
      scrollToLink(linkId);
      return;
    }
    
    // Check if clicking on already selected link
    if (networkState.selection.selectedLinkId === linkId) {
      // Deselect
      networkState.selection.selectedLinkId = null;
      appState.secondarySidebar = 'HIDDEN';
      
      // Reset link color
      const link = mapState.network.links.find(l => l.id === linkId);
      if (link) {
        link.polyline.setStyle({ 
          color: mapState.network.defaultLinkColor,
          weight: 3,
          opacity: 0.8
        });
      }
    } else {
      // Select new link
      selectNetworkLink(linkId);
      appState.secondarySidebar = 'EXPANDED';
      
      // Reset all nodes to default color first
      mapState.network.nodes.forEach(node => {
        const normalIcon = createNetworkNodeIcon(node.id, mapState.network.defaultNodeColor);
        node.marker.setIcon(normalIcon);
      });
      
      // Update link visual selection
      mapState.network.links.forEach(link => {
        if (link.id === linkId) {
          link.polyline.setStyle({ 
            color: '#f59e0b', // Orange for selected
            weight: 5,
            opacity: 1
          });
        } else {
          link.polyline.setStyle({ 
            color: mapState.network.defaultLinkColor,
            weight: 3,
            opacity: 0.8
          });
        }
      });
      
      // Scroll to selected link in list
      scrollToLink(linkId);
    }
  }
  
  async function handleDrawPolygon() {
    if (!mapState.map) return;
    
    if (mapState.mode === 'drawing-polygon') {
      // Cancel drawing
      stopPolygonDrawing();
      drawButtonText = networkState.selection.selectedPolygon ? 'Redraw Polygon' : 'Draw Polygon';
      return;
    }
    
    // Clear existing data and zones
    networkState.nodes = [];
    networkState.links = [];
    clearAllZones();
    
    // Start drawing
    drawButtonText = 'Drawing...';
    
    startPolygonDrawing(mapState.map, async (zoneId, coordinates) => {
      // Calculate bounding box
      const lats = coordinates.map(c => c[1]);
      const lngs = coordinates.map(c => c[0]);
      
      const bbox: gui.BoundingBox = {
        north: Math.max(...lats),
        south: Math.min(...lats),
        east: Math.max(...lngs),
        west: Math.min(...lngs)
      };
      
      // Store polygon
      networkState.selection.selectedPolygon = coordinates;
      
      // Load network data
      try {
        await loadNetworkInBBox(bbox);
        // Switch to EDIT mode after successful load
        if (appState.networkMode === 'VIEW') {
          appState.networkMode = 'EDIT';
        }
      } catch (error) {
      }
      
      // Update UI
      drawButtonText = 'Redraw Polygon';
    });
  }
  
  let currentCreateAction = $state<'node' | 'link' | null>(null);
  let nodesListElement: HTMLDivElement;
  let linksListElement: HTMLDivElement;
  
  // Auto-scroll when selection changes from map
  $effect(() => {
    if (networkState.selection.selectedNodeId && nodesListElement) {
      scrollToNode(networkState.selection.selectedNodeId);
    }
  });
  
  $effect(() => {
    if (networkState.selection.selectedLinkId && linksListElement) {
      scrollToLink(networkState.selection.selectedLinkId);
    }
  });
  
  function scrollToNode(nodeId: string) {
    if (!nodesListElement) return;
    const nodeButton = nodesListElement.querySelector(`[data-node-id="${nodeId}"]`);
    if (nodeButton) {
      nodeButton.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }
  
  function scrollToLink(linkId: string) {
    if (!linksListElement) return;
    const linkButton = linksListElement.querySelector(`[data-link-id="${linkId}"]`);
    if (linkButton) {
      linkButton.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }
  
  function startAddingNode() {
    currentCreateAction = currentCreateAction === 'node' ? null : 'node';
  }
  
  function startAddingLink() {
    currentCreateAction = currentCreateAction === 'link' ? null : 'link';
    if (currentCreateAction === 'link') {
      networkState.selection.selectedNodeId = null;
    }
  }
</script>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-800 p-4" data-network-action={currentCreateAction}>
  <!-- Part 1: Mode Toggle -->
  <div class="flex-shrink-0">
    <Label class="mb-2 text-center block">Network Mode</Label>
    <div class="flex justify-center">
      <ButtonGroup>
      <Button
        size="sm"
        color={appState.networkMode === 'VIEW' ? 'primary' : 'alternative'}
        onclick={() => appState.networkMode = 'VIEW'}
      >
        View
      </Button>
      <Button
        size="sm"
        color={appState.networkMode === 'EDIT' ? 'primary' : 'alternative'}
        onclick={() => appState.networkMode = 'EDIT'}
      >
        Edit
      </Button>
      <Button
        size="sm"
        color={appState.networkMode === 'CREATE' ? 'primary' : 'alternative'}
        onclick={() => appState.networkMode = 'CREATE'}
      >
        Create
      </Button>
    </ButtonGroup>
    </div>
  </div>
  
  <!-- Part 2: Mode-specific Secondary Menu -->
  <div class="flex-shrink-0 mt-4 mb-4">
    {#if appState.networkMode === 'VIEW'}
      <Button 
        class="w-full" 
        color={mapState.mode === 'drawing-polygon' ? 'yellow' : 'primary'}
        onclick={handleDrawPolygon}
      >
        {drawButtonText}
      </Button>
    {:else if appState.networkMode === 'EDIT'}
      <!-- Nothing for edit mode -->
      <P class="text-sm text-gray-600 dark:text-gray-400 text-center">
        Click nodes or links to edit
      </P>
    {:else if appState.networkMode === 'CREATE'}
      <div class="space-y-2">
        <Button 
          class="w-full" 
          color={currentCreateAction === 'node' ? 'yellow' : 'primary'}
          onclick={startAddingNode}
        >
          <PlusOutline class="w-4 h-4 mr-2" />
          {currentCreateAction === 'node' ? 'Stop Adding Nodes' : 'Add Nodes'}
        </Button>
        <Button 
          class="w-full" 
          color={currentCreateAction === 'link' ? 'yellow' : 'alternative'}
          onclick={startAddingLink}
        >
          <PlusOutline class="w-4 h-4 mr-2" />
          {currentCreateAction === 'link' ? 'Stop Adding Links' : 'Add Links'}
        </Button>
      </div>
    {/if}
  </div>
  
  <!-- Part 3: Lists (equal split for nodes and links) -->
  <div class="flex-1 flex flex-col min-h-0">
    <!-- Nodes Section - 50% -->
    <div class="flex-1 flex flex-col min-h-0 mb-4">
      <Label class="mb-2 flex-shrink-0">Nodes ({networkState.nodes.length})</Label>
      <div class="flex-1 overflow-y-auto border rounded p-2 dark:border-gray-700" bind:this={nodesListElement}>
        {#if networkState.nodes.length === 0}
          <P class="text-sm text-gray-500 dark:text-gray-400 text-center">No nodes</P>
        {:else}
          <div class="space-y-1">
            {#each networkState.nodes as node}
              <button
                class="w-full text-left px-3 py-2 text-sm rounded hover:bg-gray-100 dark:hover:bg-gray-700 {networkState.selection.selectedNodeId === node.id ? 'bg-blue-100 dark:bg-blue-900' : ''}"
                onclick={() => handleNodeSelect(node.id)}
                data-node-id={node.id}
              >
                <div class="flex items-center gap-2">
                  <span class="w-3 h-3 bg-red-500"></span>
                  <span class="truncate text-white">{node.label || node.id}</span>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    
    <!-- Links Section - 50% -->
    <div class="flex-1 flex flex-col min-h-0">
      <Label class="mb-2 flex-shrink-0">Links ({networkState.links.length})</Label>
      <div class="flex-1 overflow-y-auto border rounded p-2 dark:border-gray-700" bind:this={linksListElement}>
        {#if networkState.links.length === 0}
          <P class="text-sm text-gray-500 dark:text-gray-400 text-center">No links</P>
        {:else}
          <div class="space-y-1">
            {#each networkState.links as link}
              <button
                class="w-full text-left px-3 py-2 text-sm rounded hover:bg-gray-100 dark:hover:bg-gray-700 {networkState.selection.selectedLinkId === link.id ? 'bg-blue-100 dark:bg-blue-900' : ''}"
                onclick={() => handleLinkSelect(link.id)}
                data-link-id={link.id}
              >
                <div class="flex items-center gap-2">
                  <span class="w-8 h-0.5 bg-gray-500"></span>
                  <span class="truncate text-xs text-white">{link.label || link.id}</span>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>