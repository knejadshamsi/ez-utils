import { invoke } from '@tauri-apps/api/core';

import type {
  NetworkBboxResult, NetworkLinkDetailPayload, NetworkLinkPayload, NetworkNodePayload,
  NetworkViewportBounds, NetworkMapAction, TagField, TagValues, NetworkSelection,
} from '$lib/components/network/types';
import { emptyTagValues, TAG_FIELDS } from '$lib/components/network/types';
import { parseAttributesBlob, serializeAttributes, type AttributeRow } from '$lib/xml/attributes';
import { sources } from '$lib/stores/data.svelte';
import { drawers } from '$lib/stores/ui.svelte';

const OPACITY_STEPS = [1, 0.75, 0.5, 0.25, 0];
function nextOpacityStep(current: number): number {
  const idx = OPACITY_STEPS.findIndex(v => Math.abs(v - current) < 0.01);
  const next = idx < 0 ? 0 : (idx + 1) % OPACITY_STEPS.length;
  return OPACITY_STEPS[next];
}

function parseTagBlobToValues(blob: string): TagValues {
  const out = emptyTagValues();
  const regex = /([a-zA-Z_][a-zA-Z0-9_-]*)\s*=\s*"([^"]*)"/g;
  let m: RegExpExecArray | null;
  while ((m = regex.exec(blob)) !== null) {
    if ((TAG_FIELDS as readonly string[]).includes(m[1])) {
      out[m[1] as TagField] = m[2];
    }
  }
  return out;
}

export const NETWORK_ZOOM_THRESHOLD = 14;
export const NETWORK_BBOX_NODE_CAP = 5000;
export const NETWORK_BBOX_LINK_CAP = 10000;
export const NETWORK_SEARCH_CAP = 500;
export const NETWORK_BBOX_DEBOUNCE_MS = 3000;

export interface NetworkSearchState {
  open: boolean;
  query: string;
  exact: boolean;
  loading: boolean;
  capExceeded: boolean;
  total: number;
  resultLinks: NetworkLinkPayload[];
  resultNodes: NetworkNodePayload[];
}

function emptySearchState(): NetworkSearchState {
  return {
    open: false,
    query: '',
    exact: false,
    loading: false,
    capExceeded: false,
    total: 0,
    resultLinks: [],
    resultNodes: [],
  };
}

class NetworkStore {
  // Core data from last bbox query.
  nodes = $state<NetworkNodePayload[]>([]);
  links = $state<NetworkLinkPayload[]>([]);
  nodeTotal = $state(0);
  linkTotal = $state(0);
  nodeCapExceeded = $state(false);
  linkCapExceeded = $state(false);

  // Derived id maps for quick lookup. Rebuilt as new references on every data
  // update to keep Svelte 5's fine-grained reactivity happy.
  nodesById = $state(new Map<string, NetworkNodePayload>());

  // Viewport / gating state.
  viewport = $state<NetworkViewportBounds | null>(null);
  zoom = $state<number>(0);
  settling = $state(false);
  loading = $state(false);
  error = $state<string | null>(null);

  // Layer opacity (5-step cycle: 1, 0.75, 0.5, 0.25, 0 — 0 === hidden).
  nodeOpacity = $state<number>(1);
  linkOpacity = $state<number>(1);
  nodeDragEnabled = $state(false);
  private opacityBeforeSelection: { nodes: number; links: number } | null = null;

  // Search state (separate per section, cross-linked on render).
  linkSearch = $state<NetworkSearchState>(emptySearchState());
  nodeSearch = $state<NetworkSearchState>(emptySearchState());

  // Unified selection: either a node or a link, not both.
  selection = $state<NetworkSelection>(null);
  selectedLinkDetail = $state<NetworkLinkDetailPayload | null>(null);

  // Edit buffers (population save/discard pattern). Seeded on selectLink,
  // diffed against snapshots by hasUnsavedChanges, reset on save/discard.
  tagValues = $state<TagValues>(emptyTagValues());
  originalTagValues = $state<TagValues>(emptyTagValues());
  linkAttrRows = $state<AttributeRow[]>([]);
  fromAttrRows = $state<AttributeRow[]>([]);
  toAttrRows = $state<AttributeRow[]>([]);
  originalLinkAttributesBlob = $state<string | null>(null);
  originalFromNodeAttributesBlob = $state<string | null>(null);
  originalToNodeAttributesBlob = $state<string | null>(null);

  // Switch-while-dirty prompt (mirror of population's switchPromptOpen).
  switchPromptOpen = $state(false);
  pendingLinkId = $state<string | null>(null);
  pendingClose = $state(false);

  // Map action (create node / create link two-click).
  mapAction = $state<NetworkMapAction>('idle');
  createLinkFromId = $state<string | null>(null);

  private bboxTimer: ReturnType<typeof setTimeout> | null = null;
  private lastQueryKey: string | null = null;

  get activeSourceName(): string | null {
    if (sources.activeKind !== 'network') return null;
    return sources.activeName;
  }

  get activeSourceColor(): string {
    if (sources.activeKind !== 'network') return '#a3e635';
    return sources.active?.color ?? '#a3e635';
  }

  get belowZoom(): boolean {
    return this.zoom < NETWORK_ZOOM_THRESHOLD;
  }

  get primaryDrawerVisible(): boolean {
    return Boolean(this.activeSourceName) && !this.belowZoom;
  }

  /** Convenience getter for selected link ID. */
  get selectedLinkId(): string | null {
    return this.selection?.type === 'link' ? this.selection.id : null;
  }

  /** Convenience getter for selected node ID (filter). */
  get selectedNodeId(): string | null {
    return this.selection?.type === 'node' ? this.selection.id : null;
  }

  /** Links list: node-selection wins, else link-search, else cross-link from node-search, else viewport. */
  get displayedLinks(): NetworkLinkPayload[] {
    if (this.selection?.type === 'node') {
      const nodeId = this.selection.id;
      return this.links.filter(l => l.fromNode === nodeId || l.toNode === nodeId);
    }
    if (this.linkSearch.open && this.linkSearch.query.trim().length > 0) return this.linkSearch.resultLinks;
    if (this.nodeSearch.open && this.nodeSearch.query.trim().length > 0) return this.nodeSearch.resultLinks;
    return this.links;
  }

  get displayedNodes(): NetworkNodePayload[] {
    // When link selected, show only the two endpoint nodes
    if (this.selection?.type === 'link' && this.selectedLinkDetail) {
      const fromId = this.selectedLinkDetail.fromNode;
      const toId = this.selectedLinkDetail.toNode;
      return this.nodes.filter(n => n.id === fromId || n.id === toId);
    }
    // When node selected, show selected node + all connected nodes
    if (this.selection?.type === 'node') {
      const selectedId = this.selection.id;
      const connectedIds = new Set<string>([selectedId]);
      for (const link of this.links) {
        if (link.fromNode === selectedId) connectedIds.add(link.toNode);
        if (link.toNode === selectedId) connectedIds.add(link.fromNode);
      }
      return this.nodes.filter(n => connectedIds.has(n.id));
    }
    if (this.nodeSearch.open && this.nodeSearch.query.trim().length > 0) return this.nodeSearch.resultNodes;
    if (this.linkSearch.open && this.linkSearch.query.trim().length > 0) return this.linkSearch.resultNodes;
    return this.nodes.filter(n => !n.ghost);
  }

  /** Links rendered on the map: always viewport data, honoring node selection but not search. */
  get mapLinks(): NetworkLinkPayload[] {
    if (this.selection?.type === 'node') {
      const nodeId = this.selection.id;
      return this.links.filter(l => l.fromNode === nodeId || l.toNode === nodeId);
    }
    return this.links;
  }

  /** Nodes rendered on the map: always viewport data (ghosts included for link continuity). */
  get mapNodes(): NetworkNodePayload[] {
    return this.nodes;
  }

  /** Called by MapView's moveend handler. Schedules a debounced bbox fetch. */
  scheduleViewportFetch(bounds: NetworkViewportBounds, zoom: number) {
    this.viewport = bounds;
    this.zoom = zoom;

    // Skip fetching while something is selected - user is focused on selection
    if (this.selection) {
      this.settling = false;
      if (this.bboxTimer) clearTimeout(this.bboxTimer);
      this.bboxTimer = null;
      return;
    }

    if (!this.activeSourceName) {
      // Inactive source - skip fetch but don't clear data
      this.settling = false;
      if (this.bboxTimer) clearTimeout(this.bboxTimer);
      this.bboxTimer = null;
      return;
    }

    if (this.belowZoom) {
      // Below zoom threshold - clear data
      this.clearData();
      this.settling = false;
      if (this.bboxTimer) clearTimeout(this.bboxTimer);
      this.bboxTimer = null;
      return;
    }

    if (this.bboxTimer) clearTimeout(this.bboxTimer);
    this.settling = true;
    this.bboxTimer = setTimeout(() => {
      this.bboxTimer = null;
      void this.fetchCurrentViewport();
    }, NETWORK_BBOX_DEBOUNCE_MS);
  }

  async fetchCurrentViewport() {
    if (!this.activeSourceName || !this.viewport) {
      this.settling = false;
      return;
    }
    const sourceName = this.activeSourceName;
    const b = this.viewport;
    const key = `${sourceName}|${b.west}|${b.south}|${b.east}|${b.north}`;
    this.lastQueryKey = key;
    this.loading = true;
    this.settling = false;

    try {
      const payload = await invoke<NetworkBboxResult>('query_network_bbox', {
        sourceName,
        minLng: b.west,
        minLat: b.south,
        maxLng: b.east,
        maxLat: b.north,
      });
      if (this.lastQueryKey !== key) return;
      this.nodes = payload.nodes;
      this.links = payload.links;
      this.nodeTotal = payload.nodeTotal;
      this.linkTotal = payload.linkTotal;
      this.nodeCapExceeded = payload.nodeCapExceeded;
      this.linkCapExceeded = payload.linkCapExceeded;
      this.rebuildIndexes();
      this.error = null;
    } catch (err) {
      this.error = this.formatError(err);
    } finally {
      this.loading = false;
    }
  }

  private clearData() {
    this.nodes = [];
    this.links = [];
    this.nodeTotal = 0;
    this.linkTotal = 0;
    this.nodeCapExceeded = false;
    this.linkCapExceeded = false;
    this.nodesById = new Map();
  }

  /** Public: wipe all map-state when the active source changes out from under us. */
  wipe() {
    this.clearData();
    this.selection = null;
    this.selectedLinkDetail = null;
    this.mapAction = 'idle';
    this.createLinkFromId = null;
    this.opacityBeforeSelection = null;
    this.resetEditBuffers();
  }

  private rebuildIndexes() {
    const map = new Map<string, NetworkNodePayload>();
    for (const n of this.nodes) map.set(n.id, n);
    this.nodesById = map;
  }

  /** 5-step cycle: 1 -> 0.75 -> 0.5 -> 0.25 -> 0 -> 1. Matches the toolbar button label. */
  cycleNodeOpacity() { this.nodeOpacity = nextOpacityStep(this.nodeOpacity); }
  cycleLinkOpacity() { this.linkOpacity = nextOpacityStep(this.linkOpacity); }

  /** Save current opacity and reset to 100% for clean focus highlighting. */
  private saveAndResetOpacity() {
    this.opacityBeforeSelection = { nodes: this.nodeOpacity, links: this.linkOpacity };
    this.nodeOpacity = 1;
    this.linkOpacity = 1;
  }

  /** Restore opacity saved before selection. */
  private restoreOpacity() {
    if (this.opacityBeforeSelection) {
      this.nodeOpacity = this.opacityBeforeSelection.nodes;
      this.linkOpacity = this.opacityBeforeSelection.links;
      this.opacityBeforeSelection = null;
    }
  }

  setNodeDragEnabled(enabled: boolean) {
    this.nodeDragEnabled = enabled;
  }

  toggleLinkSearch() { this.linkSearch = { ...emptySearchState(), open: !this.linkSearch.open }; }
  toggleNodeSearch() { this.nodeSearch = { ...emptySearchState(), open: !this.nodeSearch.open }; }
  toggleLinkSearchExact() { this.linkSearch = { ...this.linkSearch, exact: !this.linkSearch.exact }; }
  toggleNodeSearchExact() { this.nodeSearch = { ...this.nodeSearch, exact: !this.nodeSearch.exact }; }

  /** Select a node (filters links). Toggle off if same node clicked again. */
  selectNode(nodeId: string) {
    // Toggle off if same node
    if (this.selection?.type === 'node' && this.selection.id === nodeId) {
      this.restoreOpacity();
      this.selection = null;
      this.triggerFetchAfterDeselect();
      return;
    }
    // If link was selected with unsaved changes, prompt
    if (this.selection?.type === 'link' && this.hasUnsavedChanges) {
      // For now, just clear - could add prompt later
      this.selectedLinkDetail = null;
      this.resetEditBuffers();
      drawers.close('secondary');
    }
    // Save and reset opacity if this is a fresh selection
    if (!this.selection) {
      this.saveAndResetOpacity();
    }
    this.selection = { type: 'node', id: nodeId };
  }

  /** Clear any selection (node or link). */
  clearSelection() {
    if (this.selection?.type === 'link' && this.hasUnsavedChanges) {
      this.pendingLinkId = null;
      this.pendingClose = true;
      this.switchPromptOpen = true;
      return;
    }
    this.restoreOpacity();
    this.selection = null;
    this.selectedLinkDetail = null;
    this.resetEditBuffers();
    drawers.close('secondary');
    this.triggerFetchAfterDeselect();
  }

  /** Trigger immediate viewport fetch after deselecting. */
  private triggerFetchAfterDeselect() {
    if (this.viewport && this.activeSourceName && !this.belowZoom) {
      void this.fetchCurrentViewport();
    }
  }

  /** Public entry: select a link. Toggle off if same link. If edits are in-flight on a different link,
   * show the Save/Discard prompt before switching. */
  async selectLink(linkId: string) {
    // Toggle off if same link
    if (this.selection?.type === 'link' && this.selection.id === linkId) {
      if (this.hasUnsavedChanges) {
        this.pendingLinkId = null;
        this.pendingClose = true;
        this.switchPromptOpen = true;
        return;
      }
      this.restoreOpacity();
      this.selection = null;
      this.selectedLinkDetail = null;
      this.resetEditBuffers();
      drawers.close('secondary');
      this.triggerFetchAfterDeselect();
      return;
    }
    // Switching to different link with unsaved changes
    if (this.selection?.type === 'link' && this.hasUnsavedChanges) {
      this.pendingLinkId = linkId;
      this.pendingClose = false;
      this.switchPromptOpen = true;
      return;
    }
    await this.loadLink(linkId);
  }

  /** Internal: load detail from DB and seed edit buffers + snapshots. */
  async loadLink(linkId: string) {
    // Save and reset opacity if this is a fresh selection
    if (!this.selection) {
      this.saveAndResetOpacity();
    }
    this.selection = { type: 'link', id: linkId };
    this.selectedLinkDetail = null;
    this.resetEditBuffers();
    drawers.open('secondary');
    if (!this.activeSourceName) return;
    try {
      const payload = await invoke<NetworkLinkDetailPayload>('get_network_link', {
        sourceName: this.activeSourceName,
        linkId,
      });
      if (this.selectedLinkId !== linkId) return;
      this.selectedLinkDetail = payload;
      this.seedEditBuffers(payload);
    } catch (err) {
      this.error = this.formatError(err);
    }
  }

  closeSecondary(force = false) {
    if (!force && this.hasUnsavedChanges) {
      this.pendingLinkId = null;
      this.pendingClose = true;
      this.switchPromptOpen = true;
      return;
    }
    this.restoreOpacity();
    this.selection = null;
    this.selectedLinkDetail = null;
    this.resetEditBuffers();
    drawers.close('secondary');
    this.triggerFetchAfterDeselect();
  }

  cancelSwitchPrompt() {
    this.switchPromptOpen = false;
    this.pendingLinkId = null;
    this.pendingClose = false;
  }

  cancelMapAction() { this.setMapAction('idle'); }
  beginAddNode() { this.setMapAction('add_node'); }
  beginAddLink() { this.setMapAction('add_link_from'); }

  private setMapAction(action: NetworkMapAction) {
    this.mapAction = action;
    this.createLinkFromId = null;
  }

  /** Called when user picks first or second node during create-link two-click. */
  advanceAddLink(nodeId: string): { from: string; to: string } | null {
    if (this.mapAction === 'add_link_from') {
      this.createLinkFromId = nodeId;
      this.mapAction = 'add_link_to';
      return null;
    }
    if (this.mapAction === 'add_link_to' && this.createLinkFromId && this.createLinkFromId !== nodeId) {
      const pair = { from: this.createLinkFromId, to: nodeId };
      this.mapAction = 'idle';
      this.createLinkFromId = null;
      return pair;
    }
    return null;
  }

  generateNodeId(): string { return `new_${Date.now().toString(36)}_${Math.floor(Math.random() * 0x10000).toString(16)}`; }
  generateLinkId(): string { return this.generateNodeId(); }

  // Delete modal state; confirmDelete lives in network-actions.
  deleteTarget = $state<{ kind: 'link' | 'node'; id: string; connectedLinks: number } | null>(null);
  cancelDelete() { this.deleteTarget = null; }

  /** Dirty check = current buffer serialization vs snapshot captured at load time. */
  get hasUnsavedChanges(): boolean {
    if (!this.selectedLinkDetail) return false;
    for (const f of Object.keys(this.tagValues) as (keyof TagValues)[]) {
      if (this.tagValues[f] !== this.originalTagValues[f]) return true;
    }
    if (serializeAttributes(this.linkAttrRows) !== this.originalLinkAttributesBlob) return true;
    if (serializeAttributes(this.fromAttrRows) !== this.originalFromNodeAttributesBlob) return true;
    if (serializeAttributes(this.toAttrRows) !== this.originalToNodeAttributesBlob) return true;
    return false;
  }

  seedEditBuffers(detail: NetworkLinkDetailPayload) {
    const tag = parseTagBlobToValues(detail.tagBlob);
    this.tagValues = { ...tag };
    this.originalTagValues = { ...tag };
    this.linkAttrRows = parseAttributesBlob(detail.attributesBlob);
    this.fromAttrRows = parseAttributesBlob(detail.fromNodeAttributesBlob);
    this.toAttrRows = parseAttributesBlob(detail.toNodeAttributesBlob);
    this.originalLinkAttributesBlob = serializeAttributes(this.linkAttrRows);
    this.originalFromNodeAttributesBlob = serializeAttributes(this.fromAttrRows);
    this.originalToNodeAttributesBlob = serializeAttributes(this.toAttrRows);
  }

  resetEditBuffers() {
    this.tagValues = emptyTagValues();
    this.originalTagValues = emptyTagValues();
    this.linkAttrRows = [];
    this.fromAttrRows = [];
    this.toAttrRows = [];
    this.originalLinkAttributesBlob = null;
    this.originalFromNodeAttributesBlob = null;
    this.originalToNodeAttributesBlob = null;
  }

  private formatError(err: unknown): string {
    return typeof err === 'string' ? err : (err as { message?: string })?.message ?? 'Network operation failed.';
  }
}

export const network = new NetworkStore();
