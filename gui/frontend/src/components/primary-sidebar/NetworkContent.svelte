<script lang="ts">
  import { Button, Label, Checkbox, Hr, Heading, P } from 'flowbite-svelte';
  import { PlusOutline, TrashBinOutline } from 'flowbite-svelte-icons';
  import { appState } from '$lib/stores/app.svelte.ts';
  import { networkState, selectNode as selectNetworkNode, selectLink as selectNetworkLink } from '$lib/stores/network.svelte';
  
  let showNodesOnMap = true;
  let showLinksOnMap = true;
  
  function handleNodeSelect(nodeId: string) {
    selectNetworkNode(nodeId);
    appState.secondarySidebar = 'EXPANDED';
  }
  
  function handleLinkSelect(linkId: string) {
    selectNetworkLink(linkId);
    appState.secondarySidebar = 'EXPANDED';
  }
  
  function startDrawingPolygon() {
    console.log('Draw Selection Polygon clicked');
    networkState.isDrawingPolygon = true;
    console.log('networkState.isDrawingPolygon set to:', networkState.isDrawingPolygon);
  }
  
  function startAddingNode() {
    networkState.isAddingNode = true;
    networkState.isAddingLink = false;
  }
  
  function startAddingLink() {
    networkState.isAddingLink = true;
    networkState.isAddingNode = false;
  }
</script>

<div class="p-4 h-full flex flex-col bg-gray-50 dark:bg-gray-800">
  {#if appState.networkMode === 'VIEW'}
    <!-- View Mode -->
    <Heading tag="h6" class="mb-4 flex-shrink-0">View Mode</Heading>
    
    <div class="flex-1 flex flex-col min-h-0">
      <!-- Fixed top section -->
      <div class="flex-shrink-0 space-y-4">
        <Button 
          class="w-full" 
          color={networkState.isDrawingPolygon ? "red" : "primary"} 
          onclick={startDrawingPolygon}
          disabled={networkState.isDrawingPolygon}
        >
          {#if networkState.isDrawingPolygon}
            Drawing... Click points to create polygon
          {:else}
            <PlusOutline class="w-4 h-4 mr-2" />
            Draw Selection Polygon
          {/if}
        </Button>
        
        <P class="text-sm text-gray-600 dark:text-gray-400">
          Draw a polygon on the map to select nodes and links for viewing.
        </P>
        
        <Hr />
      </div>
      
      <!-- Flexible selection section -->
      <div class="flex-1 flex flex-col min-h-0 mt-4">
        <!-- Selection Info Section -->
        <div class="flex-shrink-0">
          <Label>Selection Info</Label>
          {#if networkState.selection.selectedPolygon}
            <P class="text-sm text-gray-600 dark:text-gray-400 mb-4">
              Area selected with {networkState.selection.selectedPolygon.length} vertices
            </P>
          {:else}
            <P class="text-sm text-gray-600 dark:text-gray-400">
              No area selected
            </P>
          {/if}
        </div>
        
        {#if networkState.selection.selectedPolygon}
          <!-- Nodes and Links Container - fills remaining space -->
          <div class="flex-1 flex flex-col space-y-4 min-h-0">
            <!-- Selected Nodes - 50% of remaining space -->
            <div class="flex-1 flex flex-col min-h-0">
              <Label class="mb-2 flex-shrink-0">Selected Nodes ({networkState.nodes.length})</Label>
              <div class="flex-1 overflow-y-auto border rounded p-2 dark:border-gray-700">
                {#each networkState.nodes as node}
                  <div class="flex items-center py-1">
                    <Checkbox id="node-{node.id}" checked disabled />
                    <Label for="node-{node.id}" class="ml-2 text-sm">{node.label}</Label>
                  </div>
                {/each}
                {#if networkState.nodes.length === 0}
                  <P class="text-sm text-gray-500 dark:text-gray-400">No nodes in selection</P>
                {/if}
              </div>
            </div>
            
            <!-- Selected Links - 50% of remaining space -->
            <div class="flex-1 flex flex-col min-h-0">
              <Label class="mb-2 flex-shrink-0">Selected Links ({networkState.links.length})</Label>
              <div class="flex-1 overflow-y-auto border rounded p-2 dark:border-gray-700">
                {#each networkState.links as link}
                  <div class="flex items-center py-1">
                    <Checkbox id="link-{link.id}" checked disabled />
                    <Label for="link-{link.id}" class="ml-2 text-sm">{link.label}</Label>
                  </div>
                {/each}
                {#if networkState.links.length === 0}
                  <P class="text-sm text-gray-500 dark:text-gray-400">No links in selection</P>
                {/if}
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
    
  {:else if appState.networkMode === 'EDIT'}
    <!-- Edit Mode -->
    <Heading tag="h6" class="mb-4">Edit Mode</Heading>
    
    <div class="space-y-4">
      <!-- Nodes Section -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <Label>Nodes ({networkState.nodes.length})</Label>
          <Checkbox bind:checked={showNodesOnMap}>Show on map</Checkbox>
        </div>
        
        <div class="space-y-1 max-h-48 overflow-y-auto">
          {#each networkState.nodes as node}
            <button
              class="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 {networkState.selection.selectedNodeId === node.id ? 'bg-blue-100 dark:bg-blue-900' : ''}"
              onclick={() => handleNodeSelect(node.id)}
            >
              <div class="flex items-center gap-2">
                <span class="w-4 h-4 bg-blue-500 rounded-full"></span>
                <span>{node.label}</span>
              </div>
            </button>
          {/each}
          {#if networkState.nodes.length === 0}
            <P class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
              No nodes loaded
            </P>
          {/if}
        </div>
      </div>
      
      <Hr />
      
      <!-- Links Section -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <Label>Links ({networkState.links.length})</Label>
          <Checkbox bind:checked={showLinksOnMap}>Show on map</Checkbox>
        </div>
        
        <div class="space-y-1 max-h-48 overflow-y-auto">
          {#each networkState.links as link}
            <button
              class="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 {networkState.selection.selectedLinkId === link.id ? 'bg-blue-100 dark:bg-blue-900' : ''}"
              onclick={() => handleLinkSelect(link.id)}
            >
              <div class="flex items-center gap-2">
                <span class="w-4 h-1 bg-gray-500 rounded-full"></span>
                <span>{link.label}</span>
              </div>
            </button>
          {/each}
          {#if networkState.links.length === 0}
            <P class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
              No links loaded
            </P>
          {/if}
        </div>
      </div>
      
      <Hr />
      
      <P class="text-sm text-gray-600 dark:text-gray-400">
        Click on any node or link to edit its properties.
      </P>
    </div>
    
  {:else if appState.networkMode === 'CREATE'}
    <!-- Create Mode -->
    <Heading tag="h6" class="mb-4">Create Mode</Heading>
    
    <div class="space-y-4">
      <Button class="w-full" color="primary" onclick={startAddingNode}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Node
      </Button>
      
      <Button class="w-full" color="alternative" onclick={startAddingLink}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Link
      </Button>
      
      <Hr />
      
      <div class="space-y-2">
        <Label>Instructions</Label>
        <P class="text-sm text-gray-600 dark:text-gray-400">
          • Click "Add Node" then click on the map to place a new node
        </P>
        <P class="text-sm text-gray-600 dark:text-gray-400">
          • Click "Add Link" then click two nodes to connect them
        </P>
      </div>
      
      <Hr />
      
      <div class="space-y-2">
        <Label>Validation Rules</Label>
        <P class="text-sm text-gray-500 dark:text-gray-400">
          1. All nodes must have at least one link
        </P>
        <P class="text-sm text-gray-500 dark:text-gray-400">
          2. All links must connect exactly two nodes
        </P>
      </div>
    </div>
  {/if}
</div>