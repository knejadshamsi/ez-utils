import { invoke } from '@tauri-apps/api/core';

import { sources } from '$lib/stores/data.svelte';
import { transit, transitEdit } from '$lib/stores/transit.svelte';
import type { LineDeletePreview, RouteDeletePreview } from '$lib/stores/transit/types';

function activeTransitSource(): string | null {
  const active = sources.active;
  if (!active || active.kind !== 'transit') return null;
  return active.name;
}

function formatError(err: unknown): string {
  if (typeof err === 'string') return err;
  const e = err as { kind?: string; message?: string };
  if (e?.kind === 'transit_duplicate_line_id') return 'duplicate';
  if (e?.kind === 'transit_line_id_invalid') return e.message ?? 'invalid';
  if (e?.kind === 'transit_line_not_found') return 'not_found';
  if (e?.kind === 'transit_duplicate_route_id') return 'duplicate';
  if (e?.kind === 'transit_route_id_invalid') return e.message ?? 'invalid';
  if (e?.kind === 'transit_route_not_found') return 'not_found';
  if (e?.kind === 'transit_transport_mode_invalid') return 'mode_required';
  return e?.message ?? 'Transit edit failed.';
}

export async function submitCreateLine(): Promise<boolean> {
  const sourceName = activeTransitSource();
  if (!sourceName) return false;
  const id = transitEdit.createLineForm.id.trim();
  const name = transitEdit.createLineForm.name.trim();
  const mode = transitEdit.createLineForm.transportMode.trim();
  if (!id) {
    transitEdit.createLineForm = { ...transitEdit.createLineForm, error: 'empty' };
    return false;
  }
  if (!mode) {
    transitEdit.createLineForm = { ...transitEdit.createLineForm, error: 'mode_required' };
    return false;
  }
  transitEdit.createLineForm = { ...transitEdit.createLineForm, submitting: true, error: null };
  try {
    await invoke('create_line_cmd', {
      sourceName,
      lineId: id,
      name: name.length > 0 ? name : null,
      transportMode: mode,
    });
    transitEdit.closeCreateLineForm();
    transit.closeSearch();
    await transit.loadLinesForSource(sourceName);
    transit.selectLine(id);
    return true;
  } catch (err) {
    transitEdit.createLineForm = {
      ...transitEdit.createLineForm,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}

export async function submitRenameLine(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.renameTarget;
  if (!sourceName || !target) return false;
  const newId = target.id.trim();
  const newName = target.name.trim();
  const newMode = target.transportMode.trim();
  if (!newId) {
    transitEdit.renameTarget = { ...target, error: 'empty' };
    return false;
  }
  if (!newMode) {
    transitEdit.renameTarget = { ...target, error: 'mode_required' };
    return false;
  }
  const idChanged = newId !== target.originalId;
  const original = transit.lines.find((l) => l.id === target.originalId);
  const originalName = original?.name ?? '';
  const originalMode = original?.modes[0] ?? '';
  const nameChanged = newName !== originalName;
  const modeChanged = newMode !== originalMode;
  if (!idChanged && !nameChanged && !modeChanged) {
    transitEdit.renameTarget = null;
    return true;
  }
  transitEdit.renameTarget = { ...target, submitting: true, error: null };
  try {
    await invoke('update_line_cmd', {
      sourceName,
      lineId: target.originalId,
      newId: idChanged ? newId : null,
      name: nameChanged ? newName : null,
      transportMode: modeChanged ? newMode : null,
    });
    if (idChanged && transit.selectedLineId === target.originalId) {
      transit.selectedLineId = newId;
    }
    transitEdit.renameTarget = null;
    await transit.loadLinesForSource(sourceName);
    if (modeChanged) {
      await transit.loadRoutesForLine(sourceName, idChanged ? newId : target.originalId);
    }
    return true;
  } catch (err) {
    transitEdit.renameTarget = {
      ...target,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}

export async function loadDeletePreview(): Promise<void> {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteTarget;
  if (!sourceName || !target) return;
  try {
    const preview = await invoke<LineDeletePreview>('preview_delete_line_cmd', {
      sourceName,
      lineId: target.lineId,
    });
    if (transitEdit.deleteTarget?.lineId === target.lineId) {
      transitEdit.deleteTarget = {
        ...transitEdit.deleteTarget,
        preview,
        loadingPreview: false,
      };
    }
  } catch (err) {
    if (transitEdit.deleteTarget?.lineId === target.lineId) {
      transitEdit.deleteTarget = {
        ...transitEdit.deleteTarget,
        loadingPreview: false,
        error: formatError(err),
      };
    }
  }
}

export async function submitCreateRoute(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const lineId = transit.selectedLineId;
  if (!sourceName || !lineId) return false;
  const id = transitEdit.createRouteForm.id.trim();
  const description = transitEdit.createRouteForm.description.trim();
  if (!id) {
    transitEdit.createRouteForm = { ...transitEdit.createRouteForm, error: 'empty' };
    return false;
  }
  transitEdit.createRouteForm = { ...transitEdit.createRouteForm, submitting: true, error: null };
  try {
    await invoke('create_route_cmd', {
      sourceName,
      lineId,
      routeId: id,
      description: description.length > 0 ? description : null,
    });
    transitEdit.closeCreateRouteForm();
    await transit.loadRoutesForLine(sourceName, lineId);
    await transit.loadLinesForSource(sourceName);
    transit.selectRoute(id);
    return true;
  } catch (err) {
    transitEdit.createRouteForm = {
      ...transitEdit.createRouteForm,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}

export async function submitRenameRoute(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.renameRouteTarget;
  if (!sourceName || !target) return false;
  const newId = target.id.trim();
  const newDescription = target.description.trim();
  if (!newId) {
    transitEdit.renameRouteTarget = { ...target, error: 'empty' };
    return false;
  }
  const idChanged = newId !== target.originalId;
  const original = transit.routes.find((r) => r.id === target.originalId);
  const originalDesc = original?.description ?? '';
  const descChanged = newDescription !== originalDesc;
  if (!idChanged && !descChanged) {
    transitEdit.renameRouteTarget = null;
    return true;
  }
  transitEdit.renameRouteTarget = { ...target, submitting: true, error: null };
  try {
    await invoke('update_route_cmd', {
      sourceName,
      lineId: target.lineId,
      routeId: target.originalId,
      newRouteId: idChanged ? newId : null,
      description: descChanged ? newDescription : null,
    });
    if (idChanged && transit.selectedRouteId === target.originalId) {
      transit.selectedRouteId = newId;
    }
    transitEdit.renameRouteTarget = null;
    await transit.loadRoutesForLine(sourceName, target.lineId);
    return true;
  } catch (err) {
    transitEdit.renameRouteTarget = {
      ...target,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}

export async function loadRouteDeletePreview(): Promise<void> {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteRouteTarget;
  if (!sourceName || !target) return;
  try {
    const preview = await invoke<RouteDeletePreview>('preview_delete_route_cmd', {
      sourceName,
      lineId: target.lineId,
      routeId: target.routeId,
    });
    if (
      transitEdit.deleteRouteTarget?.lineId === target.lineId &&
      transitEdit.deleteRouteTarget?.routeId === target.routeId
    ) {
      transitEdit.deleteRouteTarget = {
        ...transitEdit.deleteRouteTarget,
        preview,
        loadingPreview: false,
      };
    }
  } catch (err) {
    if (
      transitEdit.deleteRouteTarget?.lineId === target.lineId &&
      transitEdit.deleteRouteTarget?.routeId === target.routeId
    ) {
      transitEdit.deleteRouteTarget = {
        ...transitEdit.deleteRouteTarget,
        loadingPreview: false,
        error: formatError(err),
      };
    }
  }
}

export async function confirmDeleteRoute(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteRouteTarget;
  if (!sourceName || !target) return false;
  transitEdit.deleteRouteTarget = { ...target, submitting: true, error: null };
  try {
    await invoke('delete_route_cmd', {
      sourceName,
      lineId: target.lineId,
      routeId: target.routeId,
    });
    if (transit.selectedRouteId === target.routeId) {
      transit.selectRoute(null);
    }
    transitEdit.deleteRouteTarget = null;
    await transit.loadRoutesForLine(sourceName, target.lineId);
    await transit.loadLinesForSource(sourceName);
    return true;
  } catch (err) {
    transitEdit.deleteRouteTarget = {
      ...target,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}

export async function confirmDeleteLine(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteTarget;
  if (!sourceName || !target) return false;
  transitEdit.deleteTarget = { ...target, submitting: true, error: null };
  try {
    await invoke('delete_line_cmd', { sourceName, lineId: target.lineId });
    if (transit.selectedLineId === target.lineId) {
      transit.selectLine(null);
    }
    transitEdit.deleteTarget = null;
    await transit.loadLinesForSource(sourceName);
    return true;
  } catch (err) {
    transitEdit.deleteTarget = {
      ...target,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}
