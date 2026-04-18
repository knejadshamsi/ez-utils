// ============================================================
// Transit edit-state store - holds form/modal state for line and
// route CRUD. Separate from the main TransitStore to keep each
// file under the 450-line ceiling.
// ============================================================

import type {
  CreateLineForm,
  CreateRouteForm,
  DeleteLineTarget,
  DeleteRouteTarget,
  DeleteStopFacilityTarget,
  EditableProfileStop,
  EditStopFacilityTarget,
  ListedLine,
  ListedProfileStop,
  ListedRoute,
  ListedStop,
  ProfileStopsDraft,
  RenameLineTarget,
  RenameRouteTarget,
} from './types';

function snapshotStops(stops: EditableProfileStop[]): string {
  return JSON.stringify(
    stops.map((s) => [
      s.stopRefId,
      s.arrivalOffset,
      s.departureOffset,
      s.allowBoarding,
      s.allowAlighting,
      s.awaitDeparture,
    ]),
  );
}

function toEditable(s: ListedProfileStop): EditableProfileStop {
  return {
    key: `stop-${s.stopRefId}-${s.sequence}`,
    stopRefId: s.stopRefId,
    stopName: s.stopName,
    stopLng: s.stopLng,
    stopLat: s.stopLat,
    stopLinkRefId: s.stopLinkRefId,
    stopAreaId: s.stopAreaId,
    stopIsBlocking: s.stopIsBlocking,
    arrivalOffset: s.arrivalOffset ?? '',
    departureOffset: s.departureOffset ?? '',
    allowBoarding: s.allowBoarding,
    allowAlighting: s.allowAlighting,
    awaitDeparture: s.awaitDeparture,
  };
}

let draftKeyCounter = 0;
function nextDraftKey(prefix: string): string {
  draftKeyCounter += 1;
  return `${prefix}-${Date.now()}-${draftKeyCounter}`;
}

class TransitEditState {
  createLineForm = $state<CreateLineForm>({
    open: false,
    id: '',
    name: '',
    transportMode: '',
    submitting: false,
    error: null,
  });
  renameTarget = $state<RenameLineTarget | null>(null);
  deleteTarget = $state<DeleteLineTarget | null>(null);
  createRouteForm = $state<CreateRouteForm>({
    open: false,
    id: '',
    description: '',
    submitting: false,
    error: null,
  });
  renameRouteTarget = $state<RenameRouteTarget | null>(null);
  deleteRouteTarget = $state<DeleteRouteTarget | null>(null);
  profileStopsDraft = $state<ProfileStopsDraft | null>(null);
  deleteStopFacilityTarget = $state<DeleteStopFacilityTarget | null>(null);
  editStopFacilityTarget = $state<EditStopFacilityTarget | null>(null);
  mapAction = $state<'idle' | 'place_new_stop' | 'pick_link_for_new_stop' | 'pick_link_for_edit_stop'>('idle');

  openCreateLineForm() {
    this.createLineForm = {
      open: true,
      id: '',
      name: '',
      transportMode: '',
      submitting: false,
      error: null,
    };
    this.renameTarget = null;
  }

  closeCreateLineForm() {
    this.createLineForm = {
      open: false,
      id: '',
      name: '',
      transportMode: '',
      submitting: false,
      error: null,
    };
  }

  setCreateLineFormId(value: string) {
    this.createLineForm = { ...this.createLineForm, id: value, error: null };
  }

  setCreateLineFormName(value: string) {
    this.createLineForm = { ...this.createLineForm, name: value };
  }

  setCreateLineFormMode(value: string) {
    this.createLineForm = { ...this.createLineForm, transportMode: value, error: null };
  }

  beginRename(line: ListedLine) {
    this.renameTarget = {
      originalId: line.id,
      id: line.id,
      name: line.name ?? '',
      transportMode: line.modes[0] ?? '',
      submitting: false,
      error: null,
    };
    this.createLineForm = { ...this.createLineForm, open: false };
  }

  cancelRename() { this.renameTarget = null; }

  setRenameId(value: string) {
    if (!this.renameTarget) return;
    this.renameTarget = { ...this.renameTarget, id: value, error: null };
  }

  setRenameName(value: string) {
    if (!this.renameTarget) return;
    this.renameTarget = { ...this.renameTarget, name: value };
  }

  setRenameMode(value: string) {
    if (!this.renameTarget) return;
    this.renameTarget = { ...this.renameTarget, transportMode: value, error: null };
  }

  requestDeleteLine(lineId: string) {
    this.deleteTarget = {
      lineId,
      preview: null,
      loadingPreview: true,
      submitting: false,
      error: null,
    };
  }

  cancelDeleteLine() { this.deleteTarget = null; }

  openCreateRouteForm() {
    this.createRouteForm = {
      open: true,
      id: '',
      description: '',
      submitting: false,
      error: null,
    };
  }

  closeCreateRouteForm() {
    this.createRouteForm = {
      open: false,
      id: '',
      description: '',
      submitting: false,
      error: null,
    };
  }

  setCreateRouteFormId(value: string) {
    this.createRouteForm = { ...this.createRouteForm, id: value, error: null };
  }

  setCreateRouteFormDescription(value: string) {
    this.createRouteForm = { ...this.createRouteForm, description: value };
  }

  beginRenameRoute(lineId: string, route: ListedRoute) {
    this.renameRouteTarget = {
      lineId,
      originalId: route.id,
      id: route.id,
      description: route.description ?? '',
      submitting: false,
      error: null,
    };
    this.createRouteForm = { ...this.createRouteForm, open: false };
  }

  cancelRenameRoute() { this.renameRouteTarget = null; }

  setRenameRouteId(value: string) {
    if (!this.renameRouteTarget) return;
    this.renameRouteTarget = { ...this.renameRouteTarget, id: value, error: null };
  }

  setRenameRouteDescription(value: string) {
    if (!this.renameRouteTarget) return;
    this.renameRouteTarget = { ...this.renameRouteTarget, description: value };
  }

  requestDeleteRoute(lineId: string, routeId: string) {
    this.deleteRouteTarget = {
      lineId,
      routeId,
      preview: null,
      loadingPreview: true,
      submitting: false,
      error: null,
    };
  }

  cancelDeleteRoute() { this.deleteRouteTarget = null; }

  // ----- Profile stops draft -----

  loadProfileStopsDraft(lineId: string, routeId: string, listed: ListedProfileStop[]) {
    const stops = listed.map(toEditable);
    this.profileStopsDraft = {
      lineId, routeId, stops,
      originalSnapshot: snapshotStops(stops),
      saveState: 'idle', saveError: null,
      pickerOpen: false, pickerQuery: '', pickerResults: [], pickerLoading: false,
      createOpen: false, createId: '', createName: '', createLng: '', createLat: '',
      createLinkRefId: '', createStopAreaId: '', createIsBlocking: false,
      createNearbyLinks: [], createNearbyLinksLoading: false,
      createSubmitting: false, createError: null,
    };
  }

  clearProfileStopsDraft() { this.profileStopsDraft = null; }

  get profileStopsDirty(): boolean {
    const d = this.profileStopsDraft;
    return d ? snapshotStops(d.stops) !== d.originalSnapshot : false;
  }

  updateProfileStopField<K extends keyof EditableProfileStop>(
    index: number,
    field: K,
    value: EditableProfileStop[K],
  ) {
    const d = this.profileStopsDraft;
    if (!d) return;
    const next = d.stops.slice();
    if (!next[index]) return;
    next[index] = { ...next[index], [field]: value };
    this.profileStopsDraft = { ...d, stops: next };
  }

  moveProfileStop(index: number, delta: number) {
    const d = this.profileStopsDraft;
    if (!d) return;
    const target = index + delta;
    if (target < 0 || target >= d.stops.length) return;
    const next = d.stops.slice();
    [next[index], next[target]] = [next[target], next[index]];
    this.profileStopsDraft = { ...d, stops: next };
  }

  removeProfileStop(index: number) {
    const d = this.profileStopsDraft;
    if (!d) return;
    const next = d.stops.slice();
    next.splice(index, 1);
    this.profileStopsDraft = { ...d, stops: next };
  }

  appendProfileStopFromFacility(facility: ListedStop) {
    const d = this.profileStopsDraft;
    if (!d) return;
    const stop: EditableProfileStop = {
      key: nextDraftKey(facility.id),
      stopRefId: facility.id,
      stopName: facility.name,
      stopLng: facility.lng,
      stopLat: facility.lat,
      stopLinkRefId: facility.linkRefId,
      stopAreaId: facility.stopAreaId,
      stopIsBlocking: facility.isBlocking,
      arrivalOffset: '',
      departureOffset: '',
      allowBoarding: true,
      allowAlighting: true,
      awaitDeparture: false,
    };
    this.profileStopsDraft = { ...d, stops: [...d.stops, stop], pickerOpen: false };
  }

  // Picker helpers.
  togglePicker() {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = {
      ...d,
      pickerOpen: !d.pickerOpen,
      pickerQuery: '',
      pickerResults: [],
      pickerLoading: false,
      createOpen: false,
    };
  }

  setPickerQuery(q: string) {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = { ...d, pickerQuery: q };
  }

  setPickerResults(results: ListedStop[], loading: boolean) {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = { ...d, pickerResults: results, pickerLoading: loading };
  }

  // Create-new-stop helpers.
  toggleCreate() {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = {
      ...d,
      createOpen: !d.createOpen,
      createId: '',
      createName: '',
      createLng: '',
      createLat: '',
      createLinkRefId: '',
      createNearbyLinks: [],
      createNearbyLinksLoading: false,
      createSubmitting: false,
      createError: null,
      pickerOpen: false,
    };
  }

  setCreateField(
    field: 'createId' | 'createName' | 'createLng' | 'createLat' | 'createLinkRefId' | 'createStopAreaId',
    value: string,
  ) {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = { ...d, [field]: value, createError: null };
  }

  updateProfileStopsDraft(patch: Partial<ProfileStopsDraft>) {
    if (!this.profileStopsDraft) return;
    this.profileStopsDraft = { ...this.profileStopsDraft, ...patch };
  }

  setCreateState(submitting: boolean, error: string | null) {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = { ...d, createSubmitting: submitting, createError: error };
  }

  setSaveState(state: 'idle' | 'saving' | 'saved' | 'error', error: string | null = null) {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = { ...d, saveState: state, saveError: error };
  }

  markProfileStopsSaved() {
    const d = this.profileStopsDraft;
    if (!d) return;
    this.profileStopsDraft = {
      ...d,
      originalSnapshot: snapshotStops(d.stops),
      saveState: 'saved',
      saveError: null,
    };
  }

  // ----- Stop-facility delete target -----

  requestDeleteStopFacility(stopId: string) {
    this.deleteStopFacilityTarget = {
      stopId,
      preview: null,
      loadingPreview: true,
      submitting: false,
      error: null,
    };
  }

  cancelDeleteStopFacility() {
    this.deleteStopFacilityTarget = null;
  }

  // ----- Stop-facility inline edit -----

  beginEditStopFacility(a: {
    id: string; name: string | null; lng: number; lat: number;
    linkRefId: string | null; stopAreaId: string | null; isBlocking: boolean;
  }) {
    this.editStopFacilityTarget = {
      originalId: a.id, id: a.id, name: a.name ?? '',
      lng: a.lng.toString(), lat: a.lat.toString(),
      linkRefId: a.linkRefId ?? '', stopAreaId: a.stopAreaId ?? '', isBlocking: a.isBlocking,
      nearbyLinks: [], nearbyLinksLoading: false,
      submitting: false, error: null,
    };
  }

  cancelEditStopFacility() {
    this.editStopFacilityTarget = null;
  }

  updateEditStopFacility(patch: Partial<EditStopFacilityTarget>) {
    if (!this.editStopFacilityTarget) return;
    this.editStopFacilityTarget = { ...this.editStopFacilityTarget, ...patch };
  }

  setMapAction(action: 'idle' | 'place_new_stop' | 'pick_link_for_new_stop' | 'pick_link_for_edit_stop') {
    this.mapAction = action;
  }

  clear() {
    this.closeCreateLineForm();
    this.renameTarget = null;
    this.deleteTarget = null;
    this.closeCreateRouteForm();
    this.renameRouteTarget = null;
    this.deleteRouteTarget = null;
    this.profileStopsDraft = null;
    this.deleteStopFacilityTarget = null;
    this.editStopFacilityTarget = null;
  }
}

export const transitEdit = new TransitEditState();
