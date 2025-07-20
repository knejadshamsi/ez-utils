import { changeTracker } from '$lib/changeTracker.svelte';
import { editingSession } from '$lib/stores/app.svelte.ts';
import type { NetworkNode, NetworkLink } from '$lib/stores/network.svelte';

// Track node changes
export function trackNodeChange(node: NetworkNode, operation: 'add' | 'update' | 'delete') {
  if (!editingSession.processId) return;
  
  if (operation === 'update') {
    const change = {
      type: 'network' as const,
      elementType: 'node' as const,
      action: 'update' as const,
      processId: editingSession.processId,
      nodeId: node.id,
      x: node.x,
      y: node.y
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'delete') {
    const change = {
      type: 'network' as const,
      elementType: 'node' as const,
      action: 'delete' as const,
      processId: editingSession.processId,
      nodeId: node.id,
      deleteConnectedLinks: true
    };
    changeTracker.pendingChanges.push(change);
  }
  // Note: 'add' operations for nodes are handled differently in the network API
  
  console.log(`[NetworkChangeTracking] Tracked ${operation} for node ${node.id}`);
}

// Track link changes
export function trackLinkChange(link: NetworkLink, operation: 'add' | 'update' | 'delete') {
  if (!editingSession.processId) return;
  
  if (operation === 'add') {
    const change = {
      type: 'network' as const,
      elementType: 'link' as const,
      action: 'create' as const,
      processId: editingSession.processId,
      linkId: link.id,
      fromNode: link.from,
      toNode: link.to,
      rawXML: `<link id="${link.id}" from="${link.from}" to="${link.to}" />`
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'update') {
    const change = {
      type: 'network' as const,
      elementType: 'link' as const,
      action: 'update' as const,
      processId: editingSession.processId,
      linkId: link.id,
      rawXML: `<link id="${link.id}" from="${link.from}" to="${link.to}" />`
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'delete') {
    const change = {
      type: 'network' as const,
      elementType: 'link' as const,
      action: 'delete' as const,
      processId: editingSession.processId,
      linkId: link.id
    };
    changeTracker.pendingChanges.push(change);
  }
  
  console.log(`[NetworkChangeTracking] Tracked ${operation} for link ${link.id}`);
}