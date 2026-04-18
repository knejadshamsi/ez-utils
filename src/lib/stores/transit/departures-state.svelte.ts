// ============================================================
// Transit departures edit state - separate file to stay under the
// 450-line ceiling on transit/edit-state.svelte.ts.
// ============================================================

import type {
  DeparturesDraft,
  EditableDeparture,
  ListedDeparture,
  ListedVehicle,
} from './types';

function snapshotDepartures(ds: EditableDeparture[]): string {
  return JSON.stringify(ds.map((d) => [d.id, d.departureTime, d.vehicleRefId]));
}

function toEditable(d: ListedDeparture): EditableDeparture {
  return {
    key: `dep-${d.id}`,
    id: d.id,
    departureTime: d.departureTime,
    vehicleRefId: d.vehicleRefId ?? '',
  };
}

class DeparturesEditState {
  draft = $state<DeparturesDraft | null>(null);

  load(lineId: string, routeId: string, listed: ListedDeparture[]) {
    const deps = listed.map(toEditable);
    this.draft = {
      lineId,
      routeId,
      departures: deps,
      originalSnapshot: snapshotDepartures(deps),
      saveState: 'idle',
      saveError: null,
      vehicles: this.draft?.vehicles ?? [],
      vehiclesLoaded: this.draft?.vehiclesLoaded ?? false,
    };
  }

  clear() {
    this.draft = null;
  }

  get dirty(): boolean {
    const d = this.draft;
    if (!d) return false;
    return snapshotDepartures(d.departures) !== d.originalSnapshot;
  }

  updateField<K extends keyof EditableDeparture>(
    index: number,
    field: K,
    value: EditableDeparture[K],
  ) {
    const d = this.draft;
    if (!d) return;
    const next = d.departures.slice();
    if (!next[index]) return;
    next[index] = { ...next[index], [field]: value };
    this.draft = { ...d, departures: next };
  }

  add() {
    const d = this.draft;
    if (!d) return;
    let suffix = d.departures.length + 1;
    let newId = `dep_${suffix}`;
    const existing = new Set(d.departures.map((x) => x.id));
    while (existing.has(newId)) {
      suffix += 1;
      newId = `dep_${suffix}`;
    }
    const dep: EditableDeparture = {
      key: `dep-new-${Date.now()}-${suffix}`,
      id: newId,
      departureTime: '',
      vehicleRefId: '',
    };
    this.draft = { ...d, departures: [...d.departures, dep] };
  }

  remove(index: number) {
    const d = this.draft;
    if (!d) return;
    const next = d.departures.slice();
    next.splice(index, 1);
    this.draft = { ...d, departures: next };
  }

  updateDraft(patch: Partial<DeparturesDraft>) {
    if (!this.draft) return;
    this.draft = { ...this.draft, ...patch };
  }

  markSaved() {
    const d = this.draft;
    if (!d) return;
    this.draft = {
      ...d,
      originalSnapshot: snapshotDepartures(d.departures),
      saveState: 'saved',
      saveError: null,
    };
  }

  setVehicles(vehicles: ListedVehicle[]) {
    const d = this.draft;
    if (!d) return;
    this.draft = { ...d, vehicles, vehiclesLoaded: true };
  }
}

export const transitDeparturesEdit = new DeparturesEditState();
