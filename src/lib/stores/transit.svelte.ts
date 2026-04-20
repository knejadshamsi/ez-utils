// ============================================================
// Transit store - lines/routes/route-detail state.
// Stops state lives in transit/stops-store.svelte.ts (transitStops).
// Edit/form state lives in transit/edit-state.svelte.ts (transitEdit).
// ============================================================

import { invoke } from '@tauri-apps/api/core';

import { transitStops } from './transit/stops-store.svelte';
import { transitEdit } from './transit/edit-state.svelte';
import {
  extractMessage,
  lineMatchesMode,
  resolveEnabledModeAliases,
  type LinePathPoint,
  type LinkedRoutePath,
  type ListedDeparture,
  type ListedLine,
  type ListedPathLink,
  type ListedProfileStop,
  type ListedRoute,
  type PrimaryTab,
  type SecondaryTab,
  type TransitMode,
  type TransitModeFlags,
  type ViewportBounds,
} from './transit/types';

export * from './transit/types';
export { transitStops } from './transit/stops-store.svelte';
export { transitEdit } from './transit/edit-state.svelte';
export { transitDeparturesEdit } from './transit/departures-state.svelte';

class TransitStore {
  modes = $state<TransitModeFlags>({ bus: true, metro: true, tram: true });
  selectedLineId = $state<string | null>(null);
  searchOpen = $state<boolean>(false);
  searchQuery = $state<string>('');
  lines = $state<ListedLine[]>([]);
  loadingLines = $state<boolean>(false);
  linesError = $state<string | null>(null);
  routes = $state<ListedRoute[]>([]);
  loadingRoutes = $state<boolean>(false);
  routesError = $state<string | null>(null);
  selectedRouteId = $state<string | null>(null);
  profileStops = $state<ListedProfileStop[]>([]);
  loadingProfileStops = $state<boolean>(false);
  profileStopsError = $state<string | null>(null);
  pathLinks = $state<ListedPathLink[]>([]);
  loadingPathLinks = $state<boolean>(false);
  pathLinksError = $state<string | null>(null);
  departures = $state<ListedDeparture[]>([]);
  loadingDepartures = $state<boolean>(false);
  departuresError = $state<string | null>(null);
  linePathStops = $state<LinePathPoint[]>([]);
  loadingLinePathStops = $state<boolean>(false);
  linePathStopsError = $state<string | null>(null);
  linkedRoutePath = $state<LinkedRoutePath | null>(null);
  loadingLinkedRoutePath = $state<boolean>(false);
  linkedRoutePathError = $state<string | null>(null);
  secondaryTab = $state<SecondaryTab>('stops');
  primaryTab = $state<PrimaryTab>('lines');

  private currentSourceName: string | null = null;
  private currentRoutesLineId: string | null = null;
  private currentProfileStopsKey: string | null = null;
  private currentPathLinksKey: string | null = null;
  private currentDeparturesKey: string | null = null;
  private currentLinePathStopsLineId: string | null = null;
  private currentLinkedRoutePathKey: string | null = null;

  toggleMode(mode: TransitMode) {
    this.modes[mode] = !this.modes[mode];
    transitStops.setModesFilter(resolveEnabledModeAliases(this.modes));
  }

  isModeOn(mode: TransitMode): boolean {
    return this.modes[mode];
  }

  get activeModes(): TransitMode[] {
    return (Object.keys(this.modes) as TransitMode[]).filter((m) => this.modes[m]);
  }

  selectLine(id: string | null) {
    this.selectedLineId = id;
    this.selectedRouteId = null;
    this.clearRouteDetail();
    this.clearLinkedRoutePath();
    transitEdit.closeCreateRouteForm();
    if (id === null) {
      this.routes = [];
      this.routesError = null;
      this.currentRoutesLineId = null;
      this.clearLinePathStops();
    }
  }

  selectRoute(id: string | null) {
    this.selectedRouteId = id;
    this.secondaryTab = 'stops';
    this.clearLinkedRoutePath();
    if (id === null) {
      this.clearRouteDetail();
    }
  }

  get isRoutePathLinked(): boolean {
    return this.linkedRoutePath !== null;
  }

  setSecondaryTab(tab: SecondaryTab) {
    this.secondaryTab = tab;
  }

  setPrimaryTab(tab: PrimaryTab) {
    this.primaryTab = tab;
    transitStops.setPrimaryTabActive(tab === 'stops');
  }

  toggleSearch() {
    this.searchOpen = !this.searchOpen;
    if (!this.searchOpen) this.searchQuery = '';
  }

  closeSearch() {
    this.searchOpen = false;
    this.searchQuery = '';
  }

  setSearchQuery(q: string) {
    this.searchQuery = q;
  }

  scheduleStopsFetch(bounds: ViewportBounds) {
    transitStops.scheduleFetch(bounds);
  }

  async loadLinesForSource(sourceName: string) {
    this.currentSourceName = sourceName;
    transitStops.setSourceName(sourceName);
    transitStops.setModesFilter(resolveEnabledModeAliases(this.modes));
    this.loadingLines = true;
    this.linesError = null;
    try {
      const result = await invoke<ListedLine[]>('list_transit_lines_cmd', { sourceName });
      if (this.currentSourceName === sourceName) {
        this.lines = result;
      }
    } catch (err) {
      this.linesError = extractMessage(err);
      this.lines = [];
    } finally {
      if (this.currentSourceName === sourceName) {
        this.loadingLines = false;
      }
    }
  }

  clearLines() {
    this.currentSourceName = null;
    this.lines = [];
    this.selectedLineId = null;
    this.searchQuery = '';
    this.searchOpen = false;
    this.routes = [];
    this.routesError = null;
    this.currentRoutesLineId = null;
    this.selectedRouteId = null;
    this.clearRouteDetail();
    this.clearLinePathStops();
    this.clearLinkedRoutePath();
    transitEdit.clear();
    transitStops.setSourceName(null);
  }

  async loadRoutesForLine(sourceName: string, lineId: string) {
    this.currentRoutesLineId = lineId;
    this.loadingRoutes = true;
    this.routesError = null;
    try {
      const result = await invoke<ListedRoute[]>('list_transit_routes_cmd', { sourceName, lineId });
      if (this.currentRoutesLineId === lineId) {
        this.routes = result;
      }
    } catch (err) {
      this.routesError = extractMessage(err);
      this.routes = [];
    } finally {
      if (this.currentRoutesLineId === lineId) {
        this.loadingRoutes = false;
      }
    }
  }

  async loadProfileStopsForRoute(sourceName: string, lineId: string, routeId: string) {
    const key = `${lineId}:::${routeId}`;
    this.currentProfileStopsKey = key;
    this.loadingProfileStops = true;
    this.profileStopsError = null;
    try {
      const result = await invoke<ListedProfileStop[]>('list_transit_profile_stops_cmd', {
        sourceName,
        lineId,
        routeId,
      });
      if (this.currentProfileStopsKey === key) {
        this.profileStops = result;
      }
    } catch (err) {
      this.profileStopsError = extractMessage(err);
      this.profileStops = [];
    } finally {
      if (this.currentProfileStopsKey === key) {
        this.loadingProfileStops = false;
      }
    }
  }

  async loadPathLinksForRoute(sourceName: string, lineId: string, routeId: string) {
    const key = `${lineId}:::${routeId}`;
    this.currentPathLinksKey = key;
    this.loadingPathLinks = true;
    this.pathLinksError = null;
    try {
      const result = await invoke<ListedPathLink[]>('list_transit_path_links_cmd', {
        sourceName,
        lineId,
        routeId,
      });
      if (this.currentPathLinksKey === key) {
        this.pathLinks = result;
      }
    } catch (err) {
      this.pathLinksError = extractMessage(err);
      this.pathLinks = [];
    } finally {
      if (this.currentPathLinksKey === key) {
        this.loadingPathLinks = false;
      }
    }
  }

  async loadDeparturesForRoute(sourceName: string, lineId: string, routeId: string) {
    const key = `${lineId}:::${routeId}`;
    this.currentDeparturesKey = key;
    this.loadingDepartures = true;
    this.departuresError = null;
    try {
      const result = await invoke<ListedDeparture[]>('list_transit_departures_cmd', {
        sourceName,
        lineId,
        routeId,
      });
      if (this.currentDeparturesKey === key) {
        this.departures = result;
      }
    } catch (err) {
      this.departuresError = extractMessage(err);
      this.departures = [];
    } finally {
      if (this.currentDeparturesKey === key) {
        this.loadingDepartures = false;
      }
    }
  }

  async loadLinePathStopsForLine(sourceName: string, lineId: string) {
    this.currentLinePathStopsLineId = lineId;
    this.loadingLinePathStops = true;
    this.linePathStopsError = null;
    try {
      const result = await invoke<LinePathPoint[]>('list_line_profile_stops_cmd', {
        sourceName,
        lineId,
      });
      if (this.currentLinePathStopsLineId === lineId) {
        this.linePathStops = result;
      }
    } catch (err) {
      this.linePathStopsError = extractMessage(err);
      this.linePathStops = [];
    } finally {
      if (this.currentLinePathStopsLineId === lineId) {
        this.loadingLinePathStops = false;
      }
    }
  }

  private clearLinePathStops() {
    this.linePathStops = [];
    this.linePathStopsError = null;
    this.currentLinePathStopsLineId = null;
  }

  async loadLinkedRoutePath(
    transitSource: string,
    networkSource: string,
    lineId: string,
    routeId: string,
  ) {
    const key = `${transitSource}|${networkSource}|${lineId}|${routeId}`;
    this.currentLinkedRoutePathKey = key;
    this.loadingLinkedRoutePath = true;
    this.linkedRoutePathError = null;
    try {
      const result = await invoke<LinkedRoutePath | null>('resolve_route_path_geometry_cmd', {
        transitSource,
        networkSource,
        lineId,
        routeId,
      });
      if (this.currentLinkedRoutePathKey === key) {
        this.linkedRoutePath = result;
      }
    } catch (err) {
      if (this.currentLinkedRoutePathKey === key) {
        this.linkedRoutePathError = extractMessage(err);
        this.linkedRoutePath = null;
      }
    } finally {
      if (this.currentLinkedRoutePathKey === key) {
        this.loadingLinkedRoutePath = false;
      }
    }
  }

  clearLinkedRoutePath() {
    this.linkedRoutePath = null;
    this.linkedRoutePathError = null;
    this.currentLinkedRoutePathKey = null;
  }

  private clearRouteDetail() {
    this.profileStops = [];
    this.profileStopsError = null;
    this.currentProfileStopsKey = null;
    this.pathLinks = [];
    this.pathLinksError = null;
    this.currentPathLinksKey = null;
    this.departures = [];
    this.departuresError = null;
    this.currentDeparturesKey = null;
  }

  get selectedRoute(): ListedRoute | null {
    if (!this.selectedRouteId) return null;
    return this.routes.find((r) => r.id === this.selectedRouteId) ?? null;
  }

  get filteredLines(): ListedLine[] {
    const q = this.searchQuery.trim().toLowerCase();
    return this.lines.filter((line) => {
      if (!lineMatchesMode(line, this.modes)) return false;
      if (q.length === 0) return true;
      if (line.id.toLowerCase().includes(q)) return true;
      if (line.name && line.name.toLowerCase().includes(q)) return true;
      return false;
    });
  }
}

export const transit = new TransitStore();
