import { invoke } from '@tauri-apps/api/core';

import type {
  CreateLinkInput,
  CreateNodeInput,
  DeleteNodeResult,
  LinkTagEditInput,
  NetworkSearchLinkResult,
  NetworkSearchNodeResult,
  TagField,
} from '$lib/components/network/types';
import { TAG_FIELDS } from '$lib/components/network/types';
import { serializeAttributes } from '$lib/xml/attributes';

import { network } from '$lib/stores/network.svelte';

const searchTimers = { link: null as ReturnType<typeof setTimeout> | null, node: null as ReturnType<typeof setTimeout> | null };

export function setLinkSearchQuery(query: string) {
  network.linkSearch = { ...network.linkSearch, query };
  if (searchTimers.link) clearTimeout(searchTimers.link);
  searchTimers.link = setTimeout(() => { void executeLinkSearch(); }, 200);
}

export function setNodeSearchQuery(query: string) {
  network.nodeSearch = { ...network.nodeSearch, query };
  if (searchTimers.node) clearTimeout(searchTimers.node);
  searchTimers.node = setTimeout(() => { void executeNodeSearch(); }, 200);
}

export async function executeLinkSearch() {
  if (!network.activeSourceName) return;
  const q = network.linkSearch.query.trim();
  if (!q) {
    network.linkSearch = { ...network.linkSearch, loading: false, capExceeded: false, total: 0, resultLinks: [], resultNodes: [] };
    return;
  }
  network.linkSearch = { ...network.linkSearch, loading: true };
  try {
    const payload = await invoke<NetworkSearchLinkResult>('search_network_links', {
      sourceName: network.activeSourceName, query: q, exact: network.linkSearch.exact,
    });
    network.linkSearch = {
      ...network.linkSearch,
      loading: false,
      capExceeded: payload.capExceeded,
      total: payload.total,
      resultLinks: payload.links,
      resultNodes: payload.connectedNodes,
    };
  } catch (err) {
    network.error = formatError(err);
    network.linkSearch = { ...network.linkSearch, loading: false };
  }
}

export async function executeNodeSearch() {
  if (!network.activeSourceName) return;
  const q = network.nodeSearch.query.trim();
  if (!q) {
    network.nodeSearch = { ...network.nodeSearch, loading: false, capExceeded: false, total: 0, resultLinks: [], resultNodes: [] };
    return;
  }
  network.nodeSearch = { ...network.nodeSearch, loading: true };
  try {
    const payload = await invoke<NetworkSearchNodeResult>('search_network_nodes', {
      sourceName: network.activeSourceName, query: q, exact: network.nodeSearch.exact,
    });
    network.nodeSearch = {
      ...network.nodeSearch,
      loading: false,
      capExceeded: payload.capExceeded,
      total: payload.total,
      resultLinks: payload.connectedLinks,
      resultNodes: payload.nodes,
    };
  } catch (err) {
    network.error = formatError(err);
    network.nodeSearch = { ...network.nodeSearch, loading: false };
  }
}

/** Persist a node drag to disk and update local state. */
export async function moveNode(nodeId: string, lng: number, lat: number) {
  if (!network.activeSourceName) return;
  try {
    await invoke('update_node_position', {
      sourceName: network.activeSourceName,
      nodeId,
      lng,
      lat,
    });
    network.nodes = network.nodes.map(n => (n.id === nodeId ? { ...n, lng, lat } : n));
    network.links = network.links.map(l => {
      if (l.fromNode === nodeId) return { ...l, fromLng: lng, fromLat: lat };
      if (l.toNode === nodeId) return { ...l, toLng: lng, toLat: lat };
      return l;
    });
  } catch (err) {
    network.error = formatError(err);
    void network.fetchCurrentViewport();
  }
}

export async function deleteLink(linkId: string) {
  if (!network.activeSourceName) return;
  try {
    await invoke('delete_network_link', { sourceName: network.activeSourceName, linkId });
    network.links = network.links.filter(l => l.id !== linkId);
    if (network.selectedLinkId === linkId) network.closeSecondary();
  } catch (err) {
    network.error = formatError(err);
  }
}

export async function deleteNode(nodeId: string) {
  if (!network.activeSourceName) return;
  try {
    const payload = await invoke<DeleteNodeResult>('delete_network_node', {
      sourceName: network.activeSourceName,
      nodeId,
    });
    const removed = new Set(payload.deletedLinkIds);
    network.links = network.links.filter(l => !removed.has(l.id));
    network.nodes = network.nodes.filter(n => n.id !== nodeId);
    if (network.selectedLinkId && removed.has(network.selectedLinkId)) network.closeSecondary();
  } catch (err) {
    network.error = formatError(err);
  }
}

/** Diff store buffers vs snapshots and persist only changed pieces, then reload. */
export async function saveLinkEdits(): Promise<boolean> {
  if (!network.activeSourceName || !network.selectedLinkId || !network.selectedLinkDetail) return false;
  const sourceName = network.activeSourceName;
  const linkId = network.selectedLinkId;
  const detail = network.selectedLinkDetail;

  const tagEdits: LinkTagEditInput = { length: null, freespeed: null, capacity: null, permlanes: null, oneway: null, modes: null };
  let tagChanged = false;
  for (const f of TAG_FIELDS as TagField[]) {
    if (network.tagValues[f] !== network.originalTagValues[f]) {
      tagEdits[f] = network.tagValues[f];
      tagChanged = true;
    }
  }
  const linkBlob = serializeAttributes(network.linkAttrRows);
  const fromBlob = serializeAttributes(network.fromAttrRows);
  const toBlob = serializeAttributes(network.toAttrRows);
  const linkAttrsChanged = linkBlob !== network.originalLinkAttributesBlob;
  const fromAttrsChanged = fromBlob !== network.originalFromNodeAttributesBlob;
  const toAttrsChanged = toBlob !== network.originalToNodeAttributesBlob;

  try {
    if (tagChanged) {
      await invoke('apply_link_tag_edits_cmd', { sourceName, linkId, edits: tagEdits });
    }
    if (linkAttrsChanged) {
      await invoke('update_link_attributes', { sourceName, linkId, attributesBlob: linkBlob || null });
    }
    if (fromAttrsChanged) {
      await invoke('update_node_attributes', { sourceName, nodeId: detail.fromNode, attributesBlob: fromBlob || null });
    }
    if (toAttrsChanged) {
      await invoke('update_node_attributes', { sourceName, nodeId: detail.toNode, attributesBlob: toBlob || null });
    }
    await network.loadLink(linkId);
    return true;
  } catch (err) {
    network.error = formatError(err);
    return false;
  }
}

/** Drop in-flight edits by re-seeding buffers from the currently loaded detail. */
export function discardLinkEdits() {
  if (!network.selectedLinkDetail) return;
  network.seedEditBuffers(network.selectedLinkDetail);
}

/** Switch-prompt resolutions: save/discard then honor pending intent. */
export async function saveThenSwitch() {
  const ok = await saveLinkEdits();
  if (!ok) return;
  await resolveSwitch();
}

export async function discardThenSwitch() {
  discardLinkEdits();
  await resolveSwitch();
}

async function resolveSwitch() {
  const pendingId = network.pendingLinkId;
  const pendingClose = network.pendingClose;
  network.switchPromptOpen = false;
  network.pendingLinkId = null;
  network.pendingClose = false;
  if (pendingClose) network.closeSecondary(true);
  else if (pendingId) await network.loadLink(pendingId);
}

export async function createNode(input: CreateNodeInput) {
  if (!network.activeSourceName) return;
  try {
    await invoke('create_node', { sourceName: network.activeSourceName, input });
    await network.fetchCurrentViewport();
  } catch (err) {
    network.error = formatError(err);
  }
}

export async function createLink(input: CreateLinkInput) {
  if (!network.activeSourceName) return;
  try {
    await invoke('create_link', { sourceName: network.activeSourceName, input });
    await network.fetchCurrentViewport();
    await network.selectLink(input.id);
  } catch (err) {
    network.error = formatError(err);
  }
}

export function requestDeleteLink(linkId: string) {
  network.deleteTarget = { kind: 'link', id: linkId, connectedLinks: 0 };
}

export function requestDeleteNode(nodeId: string) {
  const connected = network.links.filter(l => l.fromNode === nodeId || l.toNode === nodeId).length;
  network.deleteTarget = { kind: 'node', id: nodeId, connectedLinks: connected };
}

export async function confirmDelete() {
  const target = network.deleteTarget;
  if (!target) return;
  network.deleteTarget = null;
  if (target.kind === 'link') await deleteLink(target.id);
  else await deleteNode(target.id);
}

function formatError(err: unknown): string {
  if (typeof err === 'string') return err;
  const e = err as { message?: string };
  return e?.message ?? 'Network operation failed.';
}
