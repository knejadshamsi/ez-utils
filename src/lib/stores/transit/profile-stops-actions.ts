import { invoke } from '@tauri-apps/api/core';

import { sources } from '$lib/stores/data.svelte';
import { transit, transitEdit } from '$lib/stores/transit.svelte';
import type { ListedStop, NearbyNetworkLink, StopFacilityDeletePreview } from '$lib/stores/transit/types';

const NEARBY_LINK_RADIUS_M = 300;

function linkedNetworkSourceName(): string | null {
  const active = sources.active;
  if (!active || active.kind !== 'transit') return null;
  return active.transitMetadata?.linkedNetworkSourceName ?? null;
}

function activeTransitSource(): string | null {
  const active = sources.active;
  if (!active || active.kind !== 'transit') return null;
  return active.name;
}

function formatError(err: unknown): string {
  if (typeof err === 'string') return err;
  const e = err as { kind?: string; message?: string };
  if (e?.kind === 'transit_duplicate_stop_facility_id') return 'duplicate_id';
  if (e?.kind === 'transit_stop_facility_id_invalid') return e.message ?? 'invalid_id';
  if (e?.kind === 'transit_stop_facility_not_found') return 'not_found';
  if (e?.kind === 'transit_profile_stop_invalid') return e.message ?? 'invalid';
  return e?.message ?? 'Transit edit failed.';
}

// ----- Profile stops -----

let pickerSearchTimer: ReturnType<typeof setTimeout> | null = null;

export function setProfileStopPickerQuery(q: string) {
  transitEdit.setPickerQuery(q);
  if (pickerSearchTimer) clearTimeout(pickerSearchTimer);
  pickerSearchTimer = setTimeout(() => {
    void runProfileStopPickerSearch();
  }, 150);
}

export async function runProfileStopPickerSearch() {
  const sourceName = activeTransitSource();
  const draft = transitEdit.profileStopsDraft;
  if (!sourceName || !draft) return;
  transitEdit.setPickerResults(draft.pickerResults, true);
  try {
    const query = draft.pickerQuery.trim();
    const results = await invoke<ListedStop[]>('search_transit_stops_cmd', {
      sourceName,
      query: query.length > 0 ? query : null,
      limit: 50,
    });
    // Only keep results if draft is still the same route.
    if (
      transitEdit.profileStopsDraft?.lineId === draft.lineId &&
      transitEdit.profileStopsDraft?.routeId === draft.routeId
    ) {
      transitEdit.setPickerResults(results, false);
    }
  } catch (err) {
    if (transitEdit.profileStopsDraft) {
      transitEdit.setPickerResults([], false);
      transitEdit.setSaveState('error', formatError(err));
    }
  }
}

export async function saveProfileStops(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const draft = transitEdit.profileStopsDraft;
  if (!sourceName || !draft) return false;
  transitEdit.setSaveState('saving');
  try {
    const payload = draft.stops.map((s) => ({
      stopRefId: s.stopRefId,
      arrivalOffset: s.arrivalOffset.trim().length > 0 ? s.arrivalOffset.trim() : null,
      departureOffset: s.departureOffset.trim().length > 0 ? s.departureOffset.trim() : null,
      allowBoarding: s.allowBoarding,
      allowAlighting: s.allowAlighting,
      awaitDeparture: s.awaitDeparture,
    }));
    await invoke('apply_route_profile_edits_cmd', {
      sourceName,
      lineId: draft.lineId,
      routeId: draft.routeId,
      stops: payload,
    });
    transitEdit.markProfileStopsSaved();
    await transit.loadProfileStopsForRoute(sourceName, draft.lineId, draft.routeId);
    await transit.loadLinePathStopsForLine(sourceName, draft.lineId);
    // If the route was linked to a network, the backend may have cleared its
    // path_links (the link sequence is no longer trustworthy once stops change).
    // Refresh the linked route path so the polyline reflects the new reality.
    transit.clearLinkedRoutePath();
    return true;
  } catch (err) {
    transitEdit.setSaveState('error', formatError(err));
    return false;
  }
}

export function discardProfileStops() {
  const sourceName = activeTransitSource();
  const draft = transitEdit.profileStopsDraft;
  if (!sourceName || !draft) return;
  // Re-seed the draft from the authoritative DB list by reloading.
  void transit.loadProfileStopsForRoute(sourceName, draft.lineId, draft.routeId).then(() => {
    transitEdit.loadProfileStopsDraft(draft.lineId, draft.routeId, transit.profileStops);
  });
}

// ----- Create-new stop facility (inline form within editor) -----

export async function submitCreateStopFacilityInline(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const draft = transitEdit.profileStopsDraft;
  if (!sourceName || !draft) return false;

  const id = draft.createId.trim();
  const name = draft.createName.trim();
  const lngStr = draft.createLng.trim();
  const latStr = draft.createLat.trim();
  const linkRef = draft.createLinkRefId.trim();
  const areaId = draft.createStopAreaId.trim();
  const blocking = draft.createIsBlocking;

  if (!id) {
    transitEdit.setCreateState(false, 'empty_id');
    return false;
  }
  const lng = Number(lngStr);
  const lat = Number(latStr);
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) {
    transitEdit.setCreateState(false, 'invalid_coords');
    return false;
  }

  transitEdit.setCreateState(true, null);
  try {
    await invoke('create_stop_facility_cmd', {
      sourceName,
      stopId: id,
      lng,
      lat,
      name: name.length > 0 ? name : null,
      linkRefId: linkRef.length > 0 ? linkRef : null,
      stopAreaId: areaId.length > 0 ? areaId : null,
      isBlocking: blocking,
    });
    transitEdit.appendProfileStopFromFacility({
      id,
      name: name.length > 0 ? name : null,
      lng,
      lat,
      linkRefId: linkRef.length > 0 ? linkRef : null,
      stopAreaId: areaId.length > 0 ? areaId : null,
      isBlocking: blocking,
      modes: [],
    });
    transitEdit.toggleCreate();
    return true;
  } catch (err) {
    transitEdit.setCreateState(false, formatError(err));
    return false;
  }
}

// ----- Stop facility delete (cascade) -----

export async function handleTransitMapClick(lng: number, lat: number): Promise<void> {
  if (transitEdit.mapAction !== 'place_new_stop') return;
  const draft = transitEdit.profileStopsDraft;
  if (!draft || !draft.createOpen) {
    transitEdit.setMapAction('idle');
    return;
  }
  transitEdit.setCreateField('createLng', lng.toFixed(7));
  transitEdit.setCreateField('createLat', lat.toFixed(7));
  const netSrc = linkedNetworkSourceName();
  if (netSrc) {
    transitEdit.setMapAction('pick_link_for_new_stop');
    await fetchNearbyLinksForCreate();
  } else {
    transitEdit.setMapAction('idle');
  }
}

/** Query nearby network links around the currently-edited stop's lng/lat and
 * stash them in `editStopFacilityTarget.nearbyLinks` for the map layer to render. */
export async function fetchNearbyLinksForEdit(): Promise<void> {
  const netSrc = linkedNetworkSourceName();
  const target = transitEdit.editStopFacilityTarget;
  if (!target || !netSrc) return;
  const lng = Number(target.lng);
  const lat = Number(target.lat);
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) return;
  transitEdit.updateEditStopFacility({ nearbyLinksLoading: true });
  try {
    const links = await invoke<NearbyNetworkLink[]>('find_nearby_network_links_cmd', {
      networkSource: netSrc,
      lng,
      lat,
      radiusM: NEARBY_LINK_RADIUS_M,
    });
    if (transitEdit.editStopFacilityTarget?.originalId === target.originalId) {
      transitEdit.updateEditStopFacility({ nearbyLinks: links, nearbyLinksLoading: false });
    }
  } catch {
    if (transitEdit.editStopFacilityTarget?.originalId === target.originalId) {
      transitEdit.updateEditStopFacility({ nearbyLinks: [], nearbyLinksLoading: false });
    }
  }
}

/** Same idea but for the in-progress create-new-stop form. */
export async function fetchNearbyLinksForCreate(): Promise<void> {
  const netSrc = linkedNetworkSourceName();
  const draft = transitEdit.profileStopsDraft;
  if (!draft || !netSrc) return;
  const lng = Number(draft.createLng);
  const lat = Number(draft.createLat);
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) return;
  transitEdit.updateProfileStopsDraft({ createNearbyLinksLoading: true });
  try {
    const links = await invoke<NearbyNetworkLink[]>('find_nearby_network_links_cmd', {
      networkSource: netSrc,
      lng,
      lat,
      radiusM: NEARBY_LINK_RADIUS_M,
    });
    transitEdit.updateProfileStopsDraft({
      createNearbyLinks: links,
      createNearbyLinksLoading: false,
    });
  } catch {
    transitEdit.updateProfileStopsDraft({
      createNearbyLinks: [],
      createNearbyLinksLoading: false,
    });
  }
}

/** Click handler for a nearby-link polyline on the map. Fills the currently-open
 * edit form's link_ref_id, writes the id to the clipboard, and if we're in the
 * two-step create flow, auto-generates an id (if the user didn't type one) and
 * submits the create so the stop appears in the profile immediately. */
export async function assignLinkRefFromMap(linkId: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(linkId);
  } catch {
    // Clipboard may fail in some contexts; field assignment still proceeds.
  }
  if (transitEdit.editStopFacilityTarget) {
    transitEdit.updateEditStopFacility({ linkRefId: linkId, error: null });
    return;
  }
  if (transitEdit.mapAction === 'pick_link_for_new_stop' && transitEdit.profileStopsDraft?.createOpen) {
    transitEdit.setCreateField('createLinkRefId', linkId);
    transitEdit.setMapAction('idle');
    const draft = transitEdit.profileStopsDraft;
    if (!draft) return;
    if (draft.createId.trim().length === 0) {
      // Auto-generate a unique id so the user doesn't have to type one just
      // to complete the 2-step flow. They can rename via Edit stop later.
      transitEdit.setCreateField('createId', `stop_${Date.now()}`);
    }
    await submitCreateStopFacilityInline();
  } else if (transitEdit.profileStopsDraft?.createOpen) {
    transitEdit.setCreateField('createLinkRefId', linkId);
  }
}

export async function loadDeleteStopFacilityPreview() {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteStopFacilityTarget;
  if (!sourceName || !target) return;
  try {
    const preview = await invoke<StopFacilityDeletePreview>('preview_delete_stop_facility_cmd', {
      sourceName,
      stopId: target.stopId,
    });
    if (transitEdit.deleteStopFacilityTarget?.stopId === target.stopId) {
      transitEdit.deleteStopFacilityTarget = {
        ...transitEdit.deleteStopFacilityTarget,
        preview,
        loadingPreview: false,
      };
    }
  } catch (err) {
    if (transitEdit.deleteStopFacilityTarget?.stopId === target.stopId) {
      transitEdit.deleteStopFacilityTarget = {
        ...transitEdit.deleteStopFacilityTarget,
        loadingPreview: false,
        error: formatError(err),
      };
    }
  }
}

export async function submitEditStopFacility(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.editStopFacilityTarget;
  if (!sourceName || !target) return false;

  const newId = target.id.trim();
  const newName = target.name.trim();
  const lngStr = target.lng.trim();
  const latStr = target.lat.trim();
  const linkRef = target.linkRefId.trim();

  if (!newId) {
    transitEdit.updateEditStopFacility({ submitting: false, error: 'empty_id' });
    return false;
  }
  const lng = Number(lngStr);
  const lat = Number(latStr);
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) {
    transitEdit.updateEditStopFacility({ submitting: false, error: 'invalid_coords' });
    return false;
  }

  const idChanged = newId !== target.originalId;

  transitEdit.updateEditStopFacility({ submitting: true, error: null });
  try {
    await invoke('update_stop_facility_cmd', {
      sourceName,
      stopId: target.originalId,
      newId: idChanged ? newId : null,
      name: newName.length > 0 ? newName : null,
      lng,
      lat,
      linkRefId: linkRef,
      stopAreaId: target.stopAreaId,
      isBlocking: target.isBlocking,
    });
    transitEdit.cancelEditStopFacility();
    // Refetch anything that references this stop.
    const stopsStore = transitEdit;
    void stopsStore; // silence unused
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    if (lineId && routeId) {
      await transit.loadProfileStopsForRoute(sourceName, lineId, routeId);
      if (transitEdit.profileStopsDraft) {
        transitEdit.loadProfileStopsDraft(lineId, routeId, transit.profileStops);
      }
      await transit.loadLinePathStopsForLine(sourceName, lineId);
    }
    // Refresh the stops list so the renamed/moved stop shows up.
    await transitStopsRefresh();
    return true;
  } catch (err) {
    transitEdit.updateEditStopFacility({ submitting: false, error: formatError(err) });
    return false;
  }
}

async function transitStopsRefresh() {
  // Bump the stops store's viewport fetch so the list reflects the edit.
  // Importing lazily to avoid a circular dependency.
  const { transitStops } = await import('$lib/stores/transit/stops-store.svelte');
  await transitStops.fetch();
}

export async function confirmDeleteStopFacility(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const target = transitEdit.deleteStopFacilityTarget;
  if (!sourceName || !target) return false;
  transitEdit.deleteStopFacilityTarget = { ...target, submitting: true, error: null };
  try {
    await invoke('delete_stop_facility_cmd', {
      sourceName,
      stopId: target.stopId,
    });
    transitEdit.deleteStopFacilityTarget = null;
    // Reload anything that might have referenced this stop.
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    if (lineId && routeId) {
      await transit.loadProfileStopsForRoute(sourceName, lineId, routeId);
      if (transitEdit.profileStopsDraft) {
        transitEdit.loadProfileStopsDraft(lineId, routeId, transit.profileStops);
      }
      await transit.loadLinePathStopsForLine(sourceName, lineId);
    }
    return true;
  } catch (err) {
    transitEdit.deleteStopFacilityTarget = {
      ...target,
      submitting: false,
      error: formatError(err),
    };
    return false;
  }
}
