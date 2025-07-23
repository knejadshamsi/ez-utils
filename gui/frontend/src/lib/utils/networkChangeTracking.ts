import { changeTracker } from '$lib/changeTracker.svelte';
import { appState } from '$lib/stores/app.svelte.ts';
import type { NetworkNode, NetworkLink } from '$lib/stores/network.svelte';

// Track node changes
export function trackNodeChange(node: NetworkNode, operation: 'add' | 'update' | 'delete') {
  // Validate prerequisites
  if (!appState.processId) {
    console.warn('[NetworkChangeTracking] No active process ID - change not tracked');
    return;
  }
  
  // Validate node input
  if (!node || !node.id) {
    console.error('[NetworkChangeTracking] Invalid node: missing id');
    return;
  }
  
  if (operation === 'add') {
    const change = {
      type: 'network' as const,
      elementType: 'node' as const,
      action: 'create' as const,
      processId: appState.processId,
      nodeId: node.id,
      lng: node.lng,
      lat: node.lat,
      rawXML: `<node id="${node.id}" x="${node.lng}" y="${node.lat}" />`
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'update') {
    const change = {
      type: 'network' as const,
      elementType: 'node' as const,
      action: 'update' as const,
      processId: appState.processId,
      nodeId: node.id,
      lng: node.lng,
      lat: node.lat,
      rawXML: `<node id="${node.id}" x="${node.lng}" y="${node.lat}" />`
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'delete') {
    const change = {
      type: 'network' as const,
      elementType: 'node' as const,
      action: 'delete' as const,
      processId: appState.processId,
      nodeId: node.id,
      deleteConnectedLinks: true
    };
    changeTracker.pendingChanges.push(change);
  }
  
  console.log(`[NetworkChangeTracking] Tracked ${operation} for node ${node.id}`);
}

// Track link changes
export function trackLinkChange(link: NetworkLink, operation: 'add' | 'update' | 'delete') {
  // Validate prerequisites
  if (!appState.processId) {
    console.warn('[NetworkChangeTracking] No active process ID - change not tracked');
    return;
  }
  
  // Validate link input
  if (!link || !link.id) {
    console.error('[NetworkChangeTracking] Invalid link: missing id');
    return;
  }
  if (!link.from || !link.to) {
    console.error('[NetworkChangeTracking] Invalid link: missing from/to nodes');
    return;
  }
  
  if (operation === 'add') {
    const change = {
      type: 'network' as const,
      elementType: 'link' as const,
      action: 'create' as const,
      processId: appState.processId,
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
      processId: appState.processId,
      linkId: link.id,
      rawXML: `<link id="${link.id}" from="${link.from}" to="${link.to}" />`
    };
    changeTracker.pendingChanges.push(change);
  } else if (operation === 'delete') {
    const change = {
      type: 'network' as const,
      elementType: 'link' as const,
      action: 'delete' as const,
      processId: appState.processId,
      linkId: link.id
    };
    changeTracker.pendingChanges.push(change);
  }
  
  console.log(`[NetworkChangeTracking] Tracked ${operation} for link ${link.id}`);
}