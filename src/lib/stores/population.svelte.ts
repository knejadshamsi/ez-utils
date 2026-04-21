import {
  clonePlanDraft,
  createNewActivity,
  createNewAttributeRow,
  serializeAttributes,
} from '$lib/components/population/xml';
import { renderPopulationPoints, renderSelectedPlan } from '$lib/components/population/population-map';
import type {
  EditablePlanDraft,
  PopulationAttributeRow,
  PopulationMapAction,
  PopulationTab,
  PopulationPersonPayload,
  PopulationPersonSummary,
  PopulationQueryPoint,
  PopulationQueryResponse,
  PopulationSaveState,
  PopulationSelection,
  PopulationViewportBounds,
} from '$lib/components/population/types';
import {
  confirmDeletePerson,
  createPersonAt,
  createPlan,
  deleteActivePlan,
  discardThenSwitch,
  loadPerson,
  saveCurrentPerson,
  saveThenSwitch,
  setPlanSelected,
} from '$lib/stores/population-actions';
import { formatError, regenerateLegs, sameSelection, snapshotPlan, sortActivities, summarizePeople } from '$lib/stores/population-helpers';
import { sources } from '$lib/stores/data.svelte';
import { drawers } from '$lib/stores/ui.svelte';
import { invoke } from '@tauri-apps/api/core';
import type L from 'leaflet';

export const POPULATION_BBOX_DEBOUNCE_MS = 3000;
export const POPULATION_PAGE_SIZE = 50;

export class PopulationStore {
  settling = $state(false);
  loading = $state(false);
  points = $state<PopulationQueryPoint[]>([]);
  people = $state<PopulationPersonSummary[]>([]);
  totalRows = $state(0);
  totalPeople = $state(0);
  currentPage = $state(0);
  error = $state<string | null>(null);
  viewport = $state<PopulationViewportBounds | null>(null);
  selected = $state<PopulationSelection | null>(null);
  person = $state<PopulationPersonPayload | null>(null);
  personLoading = $state(false);
  attributeRows = $state<PopulationAttributeRow[]>([]);
  attributesExpanded = $state(false);
  plans = $state<EditablePlanDraft[]>([]);
  activePlanId = $state<string | null>(null);
  editLocations = $state(false);
  switchPromptOpen = $state(false);
  pendingSelection = $state<PopulationSelection | null>(null);
  deletePromptOpen = $state(false);
  deleteTarget = $state<PopulationSelection | null>(null);
  saveState = $state<PopulationSaveState>('idle');
  mapAction = $state<PopulationMapAction>(null);
  activeTab = $state<PopulationTab>('plans');
  isNewPerson = $state(false);
  newPersonCoords = $state<{ lng: number; lat: number } | null>(null);
  searchOpen = $state(false);
  searchQuery = $state('');
  searchExact = $state(false);
  searchLoading = $state(false);
  populationLayer: L.LayerGroup | null = null;
  selectedPlanLayer: L.LayerGroup | null = null;
  map: L.Map | null = null;

  private timer: number | null = null;
  private saveTimer: number | null = null;
  private searchTimer: number | null = null;
  originalAttributesBlob: string | null = null;
  originalPlanSnapshots = new Map<string, string>();

  get activeSourceName(): string | null {
    if (sources.activeKind !== 'population') return null;
    return sources.activeName;
  }

  get activeSourceColor(): string {
    if (sources.activeKind !== 'population') return '#60a5fa';
    return sources.active?.color ?? '#60a5fa';
  }

  get activePlan(): EditablePlanDraft | null {
    if (!this.activePlanId) return null;
    return this.plans.find(plan => plan.planId === this.activePlanId) ?? null;
  }

  get totalPages(): number {
    return Math.max(1, Math.ceil(this.totalPeople / POPULATION_PAGE_SIZE));
  }

  get hasUnsavedChanges(): boolean {
    if (this.isNewPerson) return true;
    if (serializeAttributes(this.attributeRows) !== this.originalAttributesBlob) {
      return true;
    }
    return this.plans.some(plan => snapshotPlan(plan) !== this.originalPlanSnapshots.get(plan.planId));
  }

  scheduleViewportFetch(bounds: PopulationViewportBounds) {
    this.viewport = bounds;
    this.currentPage = 0;
    this.error = null;
    if (!this.activeSourceName) {
      // Inactive source - skip fetch but don't clear data
      return;
    }
    if (this.selected || this.searchOpen) {
      this.viewport = bounds;
      return;
    }
    this.settling = true;
    if (this.timer !== null) {
      window.clearTimeout(this.timer);
    }
    this.timer = window.setTimeout(() => {
      void this.fetchCurrentPage();
    }, POPULATION_BBOX_DEBOUNCE_MS);
  }

  async fetchCurrentPage() {
    const bounds = this.viewport;
    const sourceName = this.activeSourceName;
    if (!bounds || !sourceName) {
      this.clearViewportResults();
      return;
    }

    this.loading = true;
    this.settling = false;
    this.error = null;
    try {
      const response = await invoke<PopulationQueryResponse>('query_population_bbox', {
        sourceName,
        minLng: bounds.west,
        minLat: bounds.south,
        maxLng: bounds.east,
        maxLat: bounds.north,
        limit: POPULATION_PAGE_SIZE,
        offset: this.currentPage * POPULATION_PAGE_SIZE,
      });
      this.totalRows = response.totalRows;
      this.totalPeople = response.totalPeople;
      this.points = response.rows.map(row => ({
        ...row,
        sourceName,
        markerKey: `${sourceName}:${row.personId}:${row.planId}:${row.activityIndex}`,
      }));
      this.people = summarizePeople(this.points);
      if (this.selected) {
        this.populationLayer?.clearLayers();
      } else {
        this.renderViewportPoints();
      }
      drawers.open('primary');
    } catch (error) {
      this.error = formatError(error);
      this.points = [];
      this.people = [];
      this.totalRows = 0;
      this.totalPeople = 0;
    } finally {
      this.loading = false;
    }
  }

  async nextPage() {
    if (this.currentPage + 1 >= this.totalPages) return;
    this.currentPage += 1;
    await this.fetchCurrentPage();
  }

  async previousPage() {
    if (this.currentPage === 0) return;
    this.currentPage -= 1;
    await this.fetchCurrentPage();
  }

  beginAddPerson() {
    this.mapAction = 'add_person';
  }

  beginAddActivity() {
    if (!this.activePlan) return;
    this.mapAction = 'add_activity';
  }

  toggleEditLocations() {
    this.editLocations = !this.editLocations;
    if (this.editLocations) {
      this.map?.dragging.disable();
    } else {
      this.map?.dragging.enable();
    }
    this.renderActivePlan();
  }

  async handleMapClick(lng: number, lat: number) {
    if (this.mapAction === 'add_person') {
      createPersonAt(this, lng, lat);
      this.mapAction = null;
      return;
    }
    if (this.mapAction === 'add_activity') {
      this.addActivityAt(lng, lat);
      this.mapAction = null;
    }
  }

  async selectPerson(selection: PopulationSelection) {
    if (sameSelection(this.selected, selection)) {
      if (this.hasUnsavedChanges) {
        this.pendingSelection = null;
        this.switchPromptOpen = true;
        return;
      }
      this.closeSecondary();
      return;
    }
    if (this.hasUnsavedChanges) {
      this.pendingSelection = selection;
      this.switchPromptOpen = true;
      return;
    }
    await loadPerson(this, selection);
  }

  closeSecondary() {
    if (this.isNewPerson) {
      this.people = this.people.filter(p => p.personId !== '__new__');
      this.isNewPerson = false;
      this.newPersonCoords = null;
    }
    this.selected = null;
    this.person = null;
    this.attributeRows = [];
    this.plans = [];
    this.activePlanId = null;
    this.attributesExpanded = false;
    if (this.editLocations) {
      this.map?.dragging.enable();
    }
    this.editLocations = false;
    this.activeTab = 'plans';
    this.deletePromptOpen = false;
    this.deleteTarget = null;
    this.switchPromptOpen = false;
    this.pendingSelection = null;
    this.selectedPlanLayer?.clearLayers();
    this.renderViewportPoints();
    drawers.close('secondary');
  }

  setActiveTab(tab: PopulationTab) {
    this.activeTab = tab;
  }

  toggleAttributesExpanded() {
    this.attributesExpanded = !this.attributesExpanded;
  }

  addAttributeRow() {
    this.attributeRows = [...this.attributeRows, createNewAttributeRow(this.attributeRows.length)];
  }

  removeAttributeRow(index: number) {
    this.attributeRows = this.attributeRows.filter((_, currentIndex) => currentIndex !== index);
  }

  updateAttributeName(index: number, name: string) {
    this.attributeRows[index].name = name;
    this.attributeRows = [...this.attributeRows];
  }

  updateAttributeType(index: number, type: PopulationAttributeRow['type']) {
    this.attributeRows[index].type = type;
    this.attributeRows = [...this.attributeRows];
  }

  updateAttributeValue(index: number, value: string) {
    this.attributeRows[index].value = value;
    this.attributeRows = [...this.attributeRows];
  }

  setActivePlan(planId: string) {
    this.activePlanId = planId;
    this.editLocations = false;
    this.renderActivePlan();
  }

  updateActivityField(activityIndex: number, field: 'type' | 'startTime' | 'endTime', value: string) {
    const plan = this.activePlan;
    if (!plan) return;
    const next = clonePlanDraft(plan);
    next.activities[activityIndex][field] = value;
    next.activities[activityIndex].dirty = true;
    if (field === 'startTime') {
      sortActivities(next);
    }
    regenerateLegs(next);
    this.replacePlan(next);
  }

  updateLegField(legIndex: number, field: 'mode' | 'durationMinutes', value: string | number) {
    const plan = this.activePlan;
    if (!plan) return;
    const next = clonePlanDraft(plan);
    if (field === 'durationMinutes') {
      next.legs[legIndex].durationMinutes = Math.max(0, Number(value) || 0);
    } else {
      next.legs[legIndex].mode = String(value);
    }
    next.legs[legIndex].dirty = true;
    this.replacePlan(next);
  }

  moveActivity(activityIndex: number, lng: number, lat: number) {
    const plan = this.activePlan;
    if (!plan) return;
    const next = clonePlanDraft(plan);
    next.activities[activityIndex].lng = lng;
    next.activities[activityIndex].lat = lat;
    next.activities[activityIndex].dirty = true;
    this.replacePlan(next);
  }

  deleteActivity(activityIndex: number) {
    const plan = this.activePlan;
    if (!plan || plan.activities.length <= 1) return;
    const next = clonePlanDraft(plan);
    next.activities.splice(activityIndex, 1);
    regenerateLegs(next);
    this.replacePlan(next);
  }

  requestDeletePerson(selection?: PopulationSelection) {
    this.deleteTarget = selection ?? this.selected;
    this.deletePromptOpen = true;
  }

  cancelDeletePerson() {
    this.deletePromptOpen = false;
    this.deleteTarget = null;
  }

  async createPlan() {
    await createPlan(this);
  }

  async deleteActivePlan() {
    await deleteActivePlan(this);
  }

  async saveCurrentPerson() {
    await saveCurrentPerson(this);
  }

  async setPlanSelected(planId: string) {
    await setPlanSelected(this, planId);
  }

  async saveThenSwitch() {
    await saveThenSwitch(this);
  }

  async discardThenSwitch() {
    await discardThenSwitch(this);
  }

  cancelSwitchPrompt() {
    this.switchPromptOpen = false;
    this.pendingSelection = null;
  }

  async confirmDeletePerson() {
    await confirmDeletePerson(this);
  }

  addActivityAt(lng: number, lat: number) {
    const plan = this.activePlan;
    if (!plan) return;
    const next = clonePlanDraft(plan);
    next.activities.push(createNewActivity(next.activities.length, lng, lat));
    sortActivities(next);
    regenerateLegs(next);
    this.replacePlan(next);
  }

  toggleSearch() {
    this.searchOpen = !this.searchOpen;
    if (!this.searchOpen) {
      this.searchQuery = '';
      this.searchLoading = false;
      if (this.searchTimer !== null) {
        window.clearTimeout(this.searchTimer);
        this.searchTimer = null;
      }
    }
  }

  setSearchQuery(query: string) {
    this.searchQuery = query;
    if (this.searchTimer !== null) {
      window.clearTimeout(this.searchTimer);
    }
    this.points = [];
    this.people = [];
    this.totalRows = 0;
    this.totalPeople = 0;
    this.populationLayer?.clearLayers();
    if (!query.trim()) {
      this.searchLoading = false;
      this.searchTimer = window.setTimeout(() => {
        if (this.viewport) {
          void this.fetchCurrentPage();
        }
      }, 1000);
      return;
    }
    this.searchLoading = true;
    this.searchTimer = window.setTimeout(() => {
      void this.executeSearch();
    }, POPULATION_BBOX_DEBOUNCE_MS);
  }

  toggleSearchExact() {
    this.searchExact = !this.searchExact;
    if (this.searchQuery.trim()) {
      if (this.searchTimer !== null) {
        window.clearTimeout(this.searchTimer);
      }
      this.searchLoading = true;
      this.searchTimer = window.setTimeout(() => {
        void this.executeSearch();
      }, POPULATION_BBOX_DEBOUNCE_MS);
    }
  }

  async executeSearch() {
    const sourceName = this.activeSourceName;
    const query = this.searchQuery.trim();
    if (!sourceName || !query) {
      this.searchLoading = false;
      return;
    }
    this.loading = true;
    this.searchLoading = false;
    this.error = null;
    try {
      const response = await invoke<PopulationQueryResponse>('search_population', {
        sourceName,
        query,
        exact: this.searchExact,
        limit: POPULATION_PAGE_SIZE,
        offset: this.currentPage * POPULATION_PAGE_SIZE,
      });
      this.totalRows = response.totalRows;
      this.totalPeople = response.totalPeople;
      this.points = response.rows.map(row => ({
        ...row,
        sourceName,
        markerKey: `${sourceName}:${row.personId}:${row.planId}:${row.activityIndex}`,
      }));
      this.people = summarizePeople(this.points);
      this.renderViewportPoints();
    } catch (error) {
      this.error = formatError(error);
      this.points = [];
      this.people = [];
      this.totalRows = 0;
      this.totalPeople = 0;
    } finally {
      this.loading = false;
    }
  }

  replacePlan(nextPlan: EditablePlanDraft) {
    this.plans = this.plans.map(plan => (plan.planId === nextPlan.planId ? nextPlan : plan));
    this.renderActivePlan();
  }

  setSaved() {
    this.saveState = 'saved';
    if (this.saveTimer !== null) {
      window.clearTimeout(this.saveTimer);
    }
    this.saveTimer = window.setTimeout(() => {
      this.saveState = 'idle';
    }, 1500);
  }

  wipe() {
    this.clearViewportResults();
    this.selected = null;
    this.person = null;
    this.attributeRows = [];
    this.plans = [];
    this.activePlanId = null;
    this.isNewPerson = false;
    this.newPersonCoords = null;
    this.searchOpen = false;
    this.searchQuery = '';
  }

  private clearViewportResults() {
    this.settling = false;
    this.loading = false;
    this.points = [];
    this.people = [];
    this.totalRows = 0;
    this.totalPeople = 0;
    this.populationLayer?.clearLayers();
    this.selectedPlanLayer?.clearLayers();
    drawers.close('secondary');
  }

  renderViewportPoints() {
    if (!this.populationLayer) return;
    renderPopulationPoints(this.populationLayer, this.points, this.activeSourceColor);
  }

  renderActivePlan() {
    if (!this.selectedPlanLayer) return;
    renderSelectedPlan(this.selectedPlanLayer, this.activePlan, this.editLocations, this.map);
  }
}

export const population = new PopulationStore();
