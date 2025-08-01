<script lang="ts">
  import { Button, Label, Input, Select, Hr, Heading, P, Helper, Toggle, Accordion, AccordionItem } from 'flowbite-svelte';
  import { CloseOutline, TrashBinOutline, ExclamationCircleOutline, ChevronDownOutline, ChevronRightOutline, PlusOutline } from 'flowbite-svelte-icons';
  import { appState } from '$lib/stores/app.svelte';
  import { networkState } from '@workflow/network/state.svelte';
  import { getNodeById, getLinkById, canDeleteNode, clearSelection, findReverseLink } from '@workflow/network/functions.svelte';
  import { updateNode, updateLink, deleteNode, deleteLink, createLink } from '@workflow/network/network';
  import { mapState } from '../../map/mapState.svelte';
  
  let nodeDraggingEnabled = $state(true);
  let showDeletionDetails = $state(false);
  
  const selectedNode = $derived(
    networkState.selection.selectedNodeId 
      ? getNodeById(networkState.selection.selectedNodeId) 
      : null
  );
  
  // Reset deletion details when selection changes
  $effect(() => {
    if (networkState.selection.selectedNodeId || networkState.selection.selectedLinkId) {
      showDeletionDetails = false;
    }
  });
  
  const selectedLink = $derived(
    networkState.selection.selectedLinkId 
      ? getLinkById(networkState.selection.selectedLinkId) 
      : null
  );
  
  const reverseLink = $derived(
    networkState.selection.selectedLinkId 
      ? findReverseLink(networkState.selection.selectedLinkId)
      : null
  );
  
  const deletionCheck = $derived(
    networkState.selection.selectedNodeId 
      ? canDeleteNode(networkState.selection.selectedNodeId) 
      : null
  );
  
  function closePanel() {
    clearSelection();
    appState.secondarySidebar = 'HIDDEN';
  }
  
  function deleteEntity() {
    try {
      if (networkState.selection.selectedNodeId) {
        deleteNode(networkState.selection.selectedNodeId);
        appState.secondarySidebar = 'HIDDEN';
      } else if (networkState.selection.selectedLinkId) {
        deleteLink(networkState.selection.selectedLinkId);
        appState.secondarySidebar = 'HIDDEN';
      }
    } catch (error) {
      console.error('Failed to delete entity:', error);
    }
  }
  
  function updateNodeProperty(property: keyof typeof selectedNode, value: any) {
    if (selectedNode) {
      const updatedNode = { ...selectedNode, [property]: value };
      
      try {
        // updateNode handles both state update and change tracking
        updateNode(updatedNode);
      } catch (error) {
        console.error('Failed to update node:', error);
      }
    }
  }
  
  function updateLinkProperty(property: keyof typeof selectedLink, value: any) {
    if (selectedLink) {
      const updatedLink = { ...selectedLink, [property]: value };
      
      try {
        // updateLink handles both state update and change tracking
        updateLink(updatedLink);
      } catch (error) {
        console.error('Failed to update link:', error);
      }
    }
  }
  
  function updateReverseLinkProperty(property: keyof typeof reverseLink, value: any) {
    if (reverseLink) {
      const updatedLink = { ...reverseLink, [property]: value };
      
      try {
        // updateLink handles both state update and change tracking
        updateLink(updatedLink);
      } catch (error) {
        console.error('Failed to update reverse link:', error);
      }
    }
  }
  
  function createReverseLink() {
    if (selectedLink) {
      try {
        createLink(selectedLink.to, selectedLink.from);
      } catch (error) {
        console.error('Failed to create reverse link:', error);
      }
    }
  }
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="p-4 pb-0">
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
      <Button size="xs" color="alternative" onclick={closePanel}>
        <CloseOutline class="w-4 h-4" />
      </Button>
    </div>
  </div>
  
  <div class="flex-1 overflow-y-auto p-4 pt-0">
  
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
          oninput={(e) => updateNodeProperty('label', e.target.value)}
          disabled={appState.networkMode === 'VIEW'}
        />
      </div>
      
      <div class="grid grid-cols-2 gap-4">
        <div>
          <Label for="node-lng" class="mb-2">Longitude</Label>
          <Input 
            id="node-lng" 
            type="number" 
            value={selectedNode.lng} 
            oninput={(e) => updateNodeProperty('lng', parseFloat(e.target.value))}
            disabled={appState.networkMode === 'VIEW'}
          />
        </div>
        <div>
          <Label for="node-lat" class="mb-2">Latitude</Label>
          <Input 
            id="node-lat" 
            type="number" 
            value={selectedNode.lat} 
            oninput={(e) => updateNodeProperty('lat', parseFloat(e.target.value))}
            disabled={appState.networkMode === 'VIEW'}
          />
        </div>
      </div>
      
      <div class="flex items-center justify-between">
        <Label for="node-dragging" class="mb-0">Enable Node Dragging</Label>
        <Toggle
          id="node-dragging"
          checked={nodeDraggingEnabled}
          onchange={() => {
            nodeDraggingEnabled = !nodeDraggingEnabled;
            // Toggle dragging for this specific node
            const node = mapState.network.nodes.find(n => n.id === selectedNode.id);
            if (node && node.marker) {
              if (nodeDraggingEnabled) {
                node.marker.dragging.enable();
              } else {
                node.marker.dragging.disable();
              }
            }
          }}
          disabled={appState.networkMode === 'VIEW'}
        />
      </div>
      
      <div>
        <Label for="node-type" class="mb-2">Type</Label>
        <Select 
          id="node-type" 
          value={selectedNode.type || 'intersection'}
          onchange={(e) => updateNodeProperty('type', e.target.value)}
          disabled={appState.networkMode === 'VIEW'}
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
          oninput={(e) => updateNodeProperty('capacity', parseInt(e.target.value))}
          disabled={appState.networkMode === 'VIEW'}
        />
      </div>
      
      <Hr />
      
      <!-- Validation Warning -->
      {#if deletionCheck && deletionCheck.affectedLinks.length > 0}
        <div class="space-y-2">
          <Helper color="yellow">
            <ExclamationCircleOutline class="w-4 h-4 inline mr-1" />
            Deleting this node will also delete {deletionCheck.affectedLinks.length} connected link{deletionCheck.affectedLinks.length > 1 ? 's' : ''}
            <button 
              class="ml-2 text-yellow-600 hover:text-yellow-700 dark:text-yellow-400 dark:hover:text-yellow-300"
              onclick={() => showDeletionDetails = !showDeletionDetails}
            >
              {#if showDeletionDetails}
                <ChevronDownOutline class="w-3 h-3 inline" />
              {:else}
                <ChevronRightOutline class="w-3 h-3 inline" />
              {/if}
              Details
            </button>
          </Helper>
          
          {#if showDeletionDetails}
            <div class="ml-6 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-md">
              <p class="text-sm font-medium text-yellow-800 dark:text-yellow-300 mb-2">Links to be deleted:</p>
              <ul class="space-y-1">
                {#each deletionCheck.affectedLinks as link}
                  <li class="text-sm text-yellow-700 dark:text-yellow-400">
                    • {link.label || link.id}
                    <span class="text-xs text-yellow-600 dark:text-yellow-500">
                      ({getNodeById(link.from)?.label || link.from} → {getNodeById(link.to)?.label || link.to})
                    </span>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </div>
      {/if}
      
    </div>
    
  {:else if selectedLink}
    <!-- Link Properties with Bidirectional Support -->
    <div class="space-y-4">
      <Accordion>
        <!-- Forward Direction -->
        <AccordionItem open>
          {#snippet header()}
            <span class="font-medium" title="{getNodeById(selectedLink.from)?.label || selectedLink.from} → {getNodeById(selectedLink.to)?.label || selectedLink.to}">
              {(getNodeById(selectedLink.from)?.label || selectedLink.from).substring(0, 15)}{(getNodeById(selectedLink.from)?.label || selectedLink.from).length > 15 ? '...' : ''} → {(getNodeById(selectedLink.to)?.label || selectedLink.to).substring(0, 15)}{(getNodeById(selectedLink.to)?.label || selectedLink.to).length > 15 ? '...' : ''}
            </span>
          {/snippet}
          
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
                oninput={(e) => updateLinkProperty('label', e.target.value)}
                disabled={appState.networkMode === 'VIEW'}
              />
            </div>
            
            <div>
              <Label for="link-from" class="mb-2">From Node</Label>
              <Select 
                id="link-from" 
                value={selectedLink.from}
                onchange={(e) => updateLinkProperty('from', e.target.value)}
                disabled={appState.networkMode === 'VIEW'}
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
                onchange={(e) => updateLinkProperty('to', e.target.value)}
                disabled={appState.networkMode === 'VIEW'}
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
                oninput={(e) => updateLinkProperty('length', parseFloat(e.target.value))}
                disabled={appState.networkMode === 'VIEW'}
              />
            </div>
            
            <div>
              <Label for="link-capacity" class="mb-2">Capacity (veh/hr)</Label>
              <Input 
                id="link-capacity" 
                type="number" 
                value={selectedLink.capacity || 0} 
                oninput={(e) => updateLinkProperty('capacity', parseInt(e.target.value))}
                disabled={appState.networkMode === 'VIEW'}
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
                  oninput={(e) => updateLinkProperty('lanes', parseInt(e.target.value))}
                  disabled={appState.networkMode === 'VIEW'}
                />
              </div>
              <div>
                <Label for="link-speed" class="mb-2">Speed (km/h)</Label>
                <Input 
                  id="link-speed" 
                  type="number" 
                  value={selectedLink.speed || 0} 
                  oninput={(e) => updateLinkProperty('speed', parseFloat(e.target.value))}
                  disabled={appState.networkMode === 'VIEW'}
                />
              </div>
            </div>
          </div>
        </AccordionItem>
        
        <!-- Reverse Direction -->
        <AccordionItem>
          {#snippet header()}
            <span class="font-medium" title="{getNodeById(selectedLink.to)?.label || selectedLink.to} → {getNodeById(selectedLink.from)?.label || selectedLink.from}">
              {(getNodeById(selectedLink.to)?.label || selectedLink.to).substring(0, 15)}{(getNodeById(selectedLink.to)?.label || selectedLink.to).length > 15 ? '...' : ''} → {(getNodeById(selectedLink.from)?.label || selectedLink.from).substring(0, 15)}{(getNodeById(selectedLink.from)?.label || selectedLink.from).length > 15 ? '...' : ''}
            </span>
            {#if !reverseLink}
              <span class="text-sm text-gray-500">(Create)</span>
            {/if}
          {/snippet}
          
          {#if reverseLink}
            <div class="space-y-4">
              <div>
                <Label for="reverse-link-id" class="mb-2">ID</Label>
                <Input id="reverse-link-id" value={reverseLink.id} disabled />
              </div>
              
              <div>
                <Label for="reverse-link-label" class="mb-2">Label</Label>
                <Input 
                  id="reverse-link-label" 
                  value={reverseLink.label} 
                  oninput={(e) => updateReverseLinkProperty('label', e.target.value)}
                  disabled={appState.networkMode === 'VIEW'}
                />
              </div>
              
              <div>
                <Label for="reverse-link-length" class="mb-2">Length (m)</Label>
                <Input 
                  id="reverse-link-length" 
                  type="number" 
                  value={reverseLink.length || 0} 
                  oninput={(e) => updateReverseLinkProperty('length', parseFloat(e.target.value))}
                  disabled={appState.networkMode === 'VIEW'}
                />
              </div>
              
              <div>
                <Label for="reverse-link-capacity" class="mb-2">Capacity (veh/hr)</Label>
                <Input 
                  id="reverse-link-capacity" 
                  type="number" 
                  value={reverseLink.capacity || 0} 
                  oninput={(e) => updateReverseLinkProperty('capacity', parseInt(e.target.value))}
                  disabled={appState.networkMode === 'VIEW'}
                />
              </div>
              
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <Label for="reverse-link-lanes" class="mb-2">Lanes</Label>
                  <Input 
                    id="reverse-link-lanes" 
                    type="number" 
                    value={reverseLink.lanes || 1} 
                    min="1" 
                    oninput={(e) => updateReverseLinkProperty('lanes', parseInt(e.target.value))}
                    disabled={appState.networkMode === 'VIEW'}
                  />
                </div>
                <div>
                  <Label for="reverse-link-speed" class="mb-2">Speed (km/h)</Label>
                  <Input 
                    id="reverse-link-speed" 
                    type="number" 
                    value={reverseLink.speed || 0} 
                    oninput={(e) => updateReverseLinkProperty('speed', parseFloat(e.target.value))}
                    disabled={appState.networkMode === 'VIEW'}
                  />
                </div>
              </div>
            </div>
          {:else}
            <div class="text-center py-4">
              <P class="text-gray-500 mb-4">No reverse link exists</P>
              <Button 
                color="primary" 
                size="sm"
                onclick={createReverseLink}
                disabled={appState.networkMode === 'VIEW'}
              >
                <PlusOutline class="w-4 h-4 mr-2" />
                Create Reverse Link
              </Button>
            </div>
          {/if}
        </AccordionItem>
      </Accordion>
    </div>
    
  {:else}
    <P class="text-gray-500 dark:text-gray-400">
      Select a node or link from the map or sidebar to view and edit its properties.
    </P>
  {/if}
  </div>
  
  {#if selectedNode || selectedLink}
    <div class="p-4 pt-2 border-t border-gray-200 dark:border-gray-700">
      <Button class="w-full" color="red" onclick={deleteEntity}>
        <TrashBinOutline class="w-4 h-4 mr-2" />
        Delete {selectedNode ? 'Node' : 'Link'}
      </Button>
    </div>
  {/if}
</div>