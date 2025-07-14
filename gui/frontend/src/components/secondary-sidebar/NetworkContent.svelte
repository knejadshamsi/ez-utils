<script lang="ts">
  import { Button, Label, Input, Select, Hr, Heading, P, Helper } from 'flowbite-svelte';
  import { CloseOutline, TrashBinOutline, ExclamationCircleOutline } from 'flowbite-svelte-icons';
  import { appState } from '$lib/stores/app.svelte.ts';
  import { networkState, getNodeById, getLinkById, canDeleteNode, clearSelection } from '$lib/stores/network.svelte';
  import { updateNode, updateLink, deleteNode as deleteNetworkNode, deleteLink as deleteNetworkLink } from '$lib/api/network';
  
  $: selectedNode = networkState.selection.selectedNodeId ? getNodeById(networkState.selection.selectedNodeId) : null;
  $: selectedLink = networkState.selection.selectedLinkId ? getLinkById(networkState.selection.selectedLinkId) : null;
  $: deletionCheck = networkState.selection.selectedNodeId ? canDeleteNode(networkState.selection.selectedNodeId) : null;
  
  function closePanel() {
    clearSelection();
    appState.secondarySidebar = 'HIDDEN';
  }
  
  async function deleteEntity() {
    try {
      if (networkState.selection.selectedNodeId) {
        await deleteNetworkNode(networkState.selection.selectedNodeId);
        appState.secondarySidebar = 'HIDDEN';
      } else if (networkState.selection.selectedLinkId) {
        await deleteNetworkLink(networkState.selection.selectedLinkId);
        appState.secondarySidebar = 'HIDDEN';
      }
    } catch (error) {
      console.error('Failed to delete entity:', error);
    }
  }
  
  async function updateNodeProperty(property: keyof typeof selectedNode, value: any) {
    if (selectedNode) {
      const nodeIndex = networkState.nodes.findIndex(n => n.id === selectedNode.id);
      if (nodeIndex !== -1) {
        const updatedNode = { ...networkState.nodes[nodeIndex], [property]: value };
        networkState.nodes[nodeIndex] = updatedNode;
        
        // Debounce API calls
        clearTimeout(updateNodeProperty.timeout);
        updateNodeProperty.timeout = setTimeout(async () => {
          try {
            await updateNode(updatedNode);
          } catch (error) {
            console.error('Failed to update node:', error);
          }
        }, 500);
      }
    }
  }
  updateNodeProperty.timeout = null;
  
  async function updateLinkProperty(property: keyof typeof selectedLink, value: any) {
    if (selectedLink) {
      const linkIndex = networkState.links.findIndex(l => l.id === selectedLink.id);
      if (linkIndex !== -1) {
        const updatedLink = { ...networkState.links[linkIndex], [property]: value };
        networkState.links[linkIndex] = updatedLink;
        
        // Debounce API calls
        clearTimeout(updateLinkProperty.timeout);
        updateLinkProperty.timeout = setTimeout(async () => {
          try {
            await updateLink(updatedLink);
          } catch (error) {
            console.error('Failed to update link:', error);
          }
        }, 500);
      }
    }
  }
  updateLinkProperty.timeout = null;
</script>

<div class="p-4 h-full overflow-y-auto bg-white dark:bg-gray-800">
  <div class="flex items-center justify-between mb-4">
    <Heading tag="h6">
      {#if selectedNode}
        Node Properties
      {:else if selectedLink}
        Link Properties
      {:else}
        No Selection
      {/if}
    </Heading>
    <Button size="xs" color="alternative" on:click={closePanel}>
      <CloseOutline class="w-4 h-4" />
    </Button>
  </div>
  
  {#if selectedNode}
    <!-- Node Properties -->
    <div class="space-y-4">
      <div>
        <Label for="node-id" class="mb-2">ID</Label>
        <Input id="node-id" value={selectedNode.id} disabled />
      </div>
      
      <div>
        <Label for="node-label" class="mb-2">Label</Label>
        <Input 
          id="node-label" 
          value={selectedNode.label} 
          on:input={(e) => updateNodeProperty('label', e.target.value)}
        />
      </div>
      
      <div class="grid grid-cols-2 gap-4">
        <div>
          <Label for="node-x" class="mb-2">X Coordinate</Label>
          <Input 
            id="node-x" 
            type="number" 
            value={selectedNode.x} 
            on:input={(e) => updateNodeProperty('x', parseFloat(e.target.value))}
          />
        </div>
        <div>
          <Label for="node-y" class="mb-2">Y Coordinate</Label>
          <Input 
            id="node-y" 
            type="number" 
            value={selectedNode.y} 
            on:input={(e) => updateNodeProperty('y', parseFloat(e.target.value))}
          />
        </div>
      </div>
      
      <div>
        <Label for="node-type" class="mb-2">Type</Label>
        <Select 
          id="node-type" 
          value={selectedNode.type || 'intersection'}
          on:change={(e) => updateNodeProperty('type', e.target.value)}
        >
          <option value="intersection">Intersection</option>
          <option value="terminal">Terminal</option>
          <option value="waypoint">Waypoint</option>
        </Select>
      </div>
      
      <div>
        <Label for="node-capacity" class="mb-2">Capacity</Label>
        <Input 
          id="node-capacity" 
          type="number" 
          value={selectedNode.capacity || 0} 
          on:input={(e) => updateNodeProperty('capacity', parseInt(e.target.value))}
        />
      </div>
      
      <Hr />
      
      <!-- Validation Warning -->
      {#if deletionCheck && deletionCheck.affectedLinks.length > 0}
        <Helper color="yellow">
          <ExclamationCircleOutline class="w-4 h-4 inline mr-1" />
          Deleting this node will also delete {deletionCheck.affectedLinks.length} connected link(s)
        </Helper>
      {/if}
      
    </div>
    
  {:else if selectedLink}
    <!-- Link Properties -->
    <div class="space-y-4">
      <div>
        <Label for="link-id" class="mb-2">ID</Label>
        <Input id="link-id" value={selectedLink.id} disabled />
      </div>
      
      <div>
        <Label for="link-label" class="mb-2">Label</Label>
        <Input 
          id="link-label" 
          value={selectedLink.label} 
          on:input={(e) => updateLinkProperty('label', e.target.value)}
        />
      </div>
      
      <div>
        <Label for="link-from" class="mb-2">From Node</Label>
        <Select 
          id="link-from" 
          value={selectedLink.from}
          on:change={(e) => updateLinkProperty('from', e.target.value)}
        >
          {#each networkState.nodes as node}
            <option value={node.id}>{node.label}</option>
          {/each}
        </Select>
      </div>
      
      <div>
        <Label for="link-to" class="mb-2">To Node</Label>
        <Select 
          id="link-to" 
          value={selectedLink.to}
          on:change={(e) => updateLinkProperty('to', e.target.value)}
        >
          {#each networkState.nodes as node}
            <option value={node.id}>{node.label}</option>
          {/each}
        </Select>
      </div>
      
      <div>
        <Label for="link-length" class="mb-2">Length (m)</Label>
        <Input 
          id="link-length" 
          type="number" 
          value={selectedLink.length || 0} 
          on:input={(e) => updateLinkProperty('length', parseFloat(e.target.value))}
        />
      </div>
      
      <div>
        <Label for="link-capacity" class="mb-2">Capacity (veh/hr)</Label>
        <Input 
          id="link-capacity" 
          type="number" 
          value={selectedLink.capacity || 0} 
          on:input={(e) => updateLinkProperty('capacity', parseInt(e.target.value))}
        />
      </div>
      
      <div class="grid grid-cols-2 gap-4">
        <div>
          <Label for="link-lanes" class="mb-2">Lanes</Label>
          <Input 
            id="link-lanes" 
            type="number" 
            value={selectedLink.lanes || 1} 
            min="1" 
            on:input={(e) => updateLinkProperty('lanes', parseInt(e.target.value))}
          />
        </div>
        <div>
          <Label for="link-speed" class="mb-2">Speed (km/h)</Label>
          <Input 
            id="link-speed" 
            type="number" 
            value={selectedLink.speed || 0} 
            on:input={(e) => updateLinkProperty('speed', parseFloat(e.target.value))}
          />
        </div>
      </div>
      
      <Hr />
      
      <!-- Validation Info -->
      <Helper color="blue">
        Links must connect exactly two different nodes
      </Helper>
    </div>
    
  {:else}
    <P class="text-gray-500 dark:text-gray-400">
      Select a node or link from the map or sidebar to view and edit its properties.
    </P>
  {/if}
  
  {#if selectedNode || selectedLink}
    <div class="mt-6 space-y-2">
      <Button class="w-full" color="red" on:click={deleteEntity}>
        <TrashBinOutline class="w-4 h-4 mr-2" />
        Delete {selectedNode ? 'Node' : 'Link'}
      </Button>
    </div>
  {/if}
</div>