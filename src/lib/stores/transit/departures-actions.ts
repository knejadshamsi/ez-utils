import { invoke } from '@tauri-apps/api/core';

import { sources } from '$lib/stores/data.svelte';
import { transit, transitDeparturesEdit } from '$lib/stores/transit.svelte';
import type { ListedVehicle } from '$lib/stores/transit/types';

function activeTransitSource(): string | null {
  const active = sources.active;
  if (!active || active.kind !== 'transit') return null;
  return active.name;
}

function formatError(err: unknown): string {
  if (typeof err === 'string') return err;
  const e = err as { kind?: string; message?: string };
  return e?.message ?? 'Departure edit failed.';
}

// Module-level sentinel: remember whether the last successful vehicles load
// happened while a .vehicles.db was attached. Lets us detect mid-session
// attach (file goes from absent to present) and re-fetch so the picker
// populates without forcing the user to reopen the route.
let lastLoadedWithFile: boolean | null = null;
let lastLoadedSourceName: string | null = null;

export async function loadVehiclesForPicker(): Promise<void> {
  const sourceName = activeTransitSource();
  if (!sourceName) return;
  const hasFile = !!sources.active?.transitMetadata?.vehicles;
  const draft = transitDeparturesEdit.draft;
  if (
    draft?.vehiclesLoaded &&
    lastLoadedSourceName === sourceName &&
    lastLoadedWithFile === hasFile
  ) {
    return;
  }
  try {
    const vehicles = await invoke<ListedVehicle[]>('list_transit_vehicles_cmd', {
      sourceName,
    });
    transitDeparturesEdit.setVehicles(vehicles);
    lastLoadedWithFile = hasFile;
    lastLoadedSourceName = sourceName;
  } catch {
    transitDeparturesEdit.setVehicles([]);
    lastLoadedWithFile = hasFile;
    lastLoadedSourceName = sourceName;
  }
}

export async function saveDepartures(): Promise<boolean> {
  const sourceName = activeTransitSource();
  const draft = transitDeparturesEdit.draft;
  if (!sourceName || !draft) return false;

  // Client-side validation before round-trip.
  for (const d of draft.departures) {
    if (d.id.trim().length === 0) {
      transitDeparturesEdit.updateDraft({
        saveState: 'error',
        saveError: 'Each departure needs an id.',
      });
      return false;
    }
    if (d.departureTime.trim().length === 0) {
      transitDeparturesEdit.updateDraft({
        saveState: 'error',
        saveError: `Departure '${d.id}' needs a time.`,
      });
      return false;
    }
  }
  const ids = draft.departures.map((d) => d.id.trim());
  const dupIdx = ids.findIndex((id, i) => ids.indexOf(id) !== i);
  if (dupIdx >= 0) {
    transitDeparturesEdit.updateDraft({
      saveState: 'error',
      saveError: `Duplicate id '${ids[dupIdx]}'.`,
    });
    return false;
  }

  transitDeparturesEdit.updateDraft({ saveState: 'saving', saveError: null });
  try {
    const payload = draft.departures.map((d) => ({
      id: d.id.trim(),
      departureTime: d.departureTime.trim(),
      vehicleRefId: d.vehicleRefId.trim().length > 0 ? d.vehicleRefId.trim() : null,
    }));
    await invoke('apply_route_departures_edits_cmd', {
      sourceName,
      lineId: draft.lineId,
      routeId: draft.routeId,
      departures: payload,
    });
    transitDeparturesEdit.markSaved();
    await transit.loadDeparturesForRoute(sourceName, draft.lineId, draft.routeId);
    return true;
  } catch (err) {
    transitDeparturesEdit.updateDraft({ saveState: 'error', saveError: formatError(err) });
    return false;
  }
}

export function discardDepartures() {
  const sourceName = activeTransitSource();
  const draft = transitDeparturesEdit.draft;
  if (!sourceName || !draft) return;
  void transit
    .loadDeparturesForRoute(sourceName, draft.lineId, draft.routeId)
    .then(() => {
      transitDeparturesEdit.load(draft.lineId, draft.routeId, transit.departures);
    });
}
