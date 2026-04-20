import { invoke } from '@tauri-apps/api/core';

import {
  extractMessage,
  type ListedStop,
  type StopDetail,
  type StopLineUsage,
  type StopRouteUsage,
  type StopTransfer,
  type StopsPage,
  type ViewportBounds,
} from './types';

interface StopLocation {
  page: number;
  positionInPage: number;
}

export interface StopsFilter {
  /**
   * Resolved lowercase list of raw transport_mode values to include.
   * null = no filter (all modes). Empty array = zero results.
   */
  modes: string[] | null;
}

const STOPS_BBOX_DEBOUNCE_MS = 3000;
const STOPS_SEARCH_DEBOUNCE_MS = 250;
const STOPS_PAGE_SIZE = 50;

function emptyDetail(): StopDetail {
  return {
    linesLoading: false,
    linesError: null,
    lines: [],
    routesLoading: false,
    routesError: null,
    routes: [],
    transfersLoading: false,
    transfersError: null,
    transfers: [],
  };
}

class TransitStopsStore {
  sourceName: string | null = null;
  primaryTabActive = false;
  modesFilter: string[] | null = null;

  items = $state<ListedStop[]>([]);
  total = $state<number>(0);
  page = $state<number>(0);
  readonly pageSize = STOPS_PAGE_SIZE;
  searchOpen = $state<boolean>(false);
  searchQuery = $state<string>('');
  loading = $state<boolean>(false);
  error = $state<string | null>(null);
  settling = $state<boolean>(false);
  viewport = $state<ViewportBounds | null>(null);
  expandedStopId = $state<string | null>(null);
  detail = $state<StopDetail>(emptyDetail());
  /** Increments when a stop should be scrolled into view in the drawer list. */
  scrollTick = $state<number>(0);

  private timer: ReturnType<typeof setTimeout> | null = null;
  private queryKey: string | null = null;
  private expandedKey: string | null = null;

  setSourceName(name: string | null) {
    if (this.sourceName === name) return;
    this.sourceName = name;
    if (name === null) this.clear();
  }

  setPrimaryTabActive(active: boolean) {
    this.primaryTabActive = active;
    if (active) {
      if (this.sourceName && this.viewport) {
        void this.fetch();
      }
      return;
    }
    if (this.timer) clearTimeout(this.timer);
    this.timer = null;
    this.settling = false;
  }

  setModesFilter(modes: string[] | null) {
    const next = modes === null ? null : [...modes].map((m) => m.toLowerCase()).sort();
    const prev = this.modesFilter;
    const same =
      (prev === null && next === null) ||
      (prev !== null &&
        next !== null &&
        prev.length === next.length &&
        prev.every((m, i) => m === next[i]));
    if (same) return;
    this.modesFilter = next;
    this.page = 0;
    if (this.primaryTabActive && this.sourceName && this.viewport) {
      void this.fetch();
    }
  }

  clear() {
    this.items = [];
    this.total = 0;
    this.page = 0;
    this.searchOpen = false;
    this.searchQuery = '';
    this.error = null;
    this.settling = false;
    this.queryKey = null;
    this.collapse();
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }

  scheduleFetch(bounds: ViewportBounds) {
    this.viewport = bounds;
    if (!this.sourceName) return;
    // Bbox debounce is exclusive to the Stops tab. When the tab is inactive we
    // still record the viewport so activation can fetch, but we never settle.
    if (!this.primaryTabActive) {
      if (this.timer) clearTimeout(this.timer);
      this.timer = null;
      this.settling = false;
      return;
    }
    // While a stop is expanded the user is focused on that detail + a flyTo just
    // moved the map. Don't invalidate the list out from under them.
    // The current viewport is still tracked so a fetch fires on collapse.
    if (this.expandedStopId !== null) {
      if (this.timer) clearTimeout(this.timer);
      this.timer = null;
      this.settling = false;
      return;
    }
    if (this.timer) clearTimeout(this.timer);
    this.settling = true;
    this.timer = setTimeout(() => {
      this.timer = null;
      this.page = 0;
      void this.fetch();
    }, STOPS_BBOX_DEBOUNCE_MS);
  }

  toggleSearch() {
    this.searchOpen = !this.searchOpen;
    if (!this.searchOpen && this.searchQuery !== '') {
      this.searchQuery = '';
      this.page = 0;
      void this.fetch();
    }
  }

  setSearchQuery(q: string) {
    this.searchQuery = q;
    this.page = 0;
    if (this.timer) clearTimeout(this.timer);
    this.timer = setTimeout(() => {
      this.timer = null;
      void this.fetch();
    }, STOPS_SEARCH_DEBOUNCE_MS);
  }

  setPage(page: number) {
    const maxPage = Math.max(0, Math.ceil(this.total / this.pageSize) - 1);
    const clamped = Math.max(0, Math.min(maxPage, page));
    if (clamped === this.page) return;
    this.page = clamped;
    void this.fetch();
  }

  async fetch() {
    const sourceName = this.sourceName;
    const bounds = this.viewport;
    if (!sourceName || !bounds) {
      this.settling = false;
      return;
    }
    const modesKey = this.modesFilter === null ? '*' : this.modesFilter.join(',');
    const key = `${sourceName}|${bounds.west}|${bounds.south}|${bounds.east}|${bounds.north}|${this.page}|${this.searchQuery}|${modesKey}`;
    this.queryKey = key;
    this.loading = true;
    this.settling = false;
    this.error = null;
    try {
      const result = await invoke<StopsPage>('query_transit_stops_bbox_cmd', {
        sourceName,
        minLng: bounds.west,
        maxLng: bounds.east,
        minLat: bounds.south,
        maxLat: bounds.north,
        search: this.searchQuery.trim() === '' ? null : this.searchQuery,
        modes: this.modesFilter,
        page: this.page,
        pageSize: this.pageSize,
      });
      if (this.queryKey !== key) return;
      this.items = result.items;
      this.total = result.total;
    } catch (err) {
      this.error = extractMessage(err);
      this.items = [];
    } finally {
      if (this.queryKey === key) {
        this.loading = false;
      }
    }
  }

  toggleExpansion(stopId: string) {
    if (this.expandedStopId === stopId) {
      this.collapse();
    } else {
      this.expand(stopId);
      this.scrollTick += 1;
    }
  }

  /**
   * Map-click entry point: compute which page the stop is on under the current filters,
   * fetch it, then expand the stop. Triggers a scroll in the drawer list.
   */
  async locateStop(stopId: string) {
    const sourceName = this.sourceName;
    const bounds = this.viewport;
    if (!sourceName || !bounds) return;

    try {
      const result = await invoke<StopLocation>('locate_transit_stop_cmd', {
        sourceName,
        minLng: bounds.west,
        maxLng: bounds.east,
        minLat: bounds.south,
        maxLat: bounds.north,
        search: this.searchQuery.trim() === '' ? null : this.searchQuery,
        modes: this.modesFilter,
        stopId,
        pageSize: this.pageSize,
      });
      if (result.page !== this.page) {
        this.page = result.page;
        await this.fetch();
      }
    } catch (err) {
      this.error = extractMessage(err);
      return;
    }
    this.expand(stopId);
    this.scrollTick += 1;
  }

  private collapse() {
    const wasExpanded = this.expandedStopId !== null;
    this.expandedStopId = null;
    this.expandedKey = null;
    this.detail = emptyDetail();
    // Resume bbox fetching: re-query for the viewport that's been changing while paused.
    if (wasExpanded && this.sourceName && this.viewport) {
      this.page = 0;
      void this.fetch();
    }
  }

  private expand(stopId: string) {
    const sourceName = this.sourceName;
    if (!sourceName) return;
    this.expandedStopId = stopId;
    const key = `${sourceName}|${stopId}`;
    this.expandedKey = key;
    this.detail = {
      linesLoading: true,
      linesError: null,
      lines: [],
      routesLoading: true,
      routesError: null,
      routes: [],
      transfersLoading: true,
      transfersError: null,
      transfers: [],
    };

    void (async () => {
      try {
        const lines = await invoke<StopLineUsage[]>('list_stop_line_usage_cmd', {
          sourceName,
          stopId,
        });
        if (this.expandedKey === key) {
          this.detail.lines = lines;
          this.detail.linesLoading = false;
        }
      } catch (err) {
        if (this.expandedKey === key) {
          this.detail.linesError = extractMessage(err);
          this.detail.linesLoading = false;
        }
      }
    })();

    void (async () => {
      try {
        const routes = await invoke<StopRouteUsage[]>('list_stop_route_usage_cmd', {
          sourceName,
          stopId,
        });
        if (this.expandedKey === key) {
          this.detail.routes = routes;
          this.detail.routesLoading = false;
        }
      } catch (err) {
        if (this.expandedKey === key) {
          this.detail.routesError = extractMessage(err);
          this.detail.routesLoading = false;
        }
      }
    })();

    void (async () => {
      try {
        const transfers = await invoke<StopTransfer[]>('list_stop_transfers_cmd', {
          sourceName,
          stopId,
        });
        if (this.expandedKey === key) {
          this.detail.transfers = transfers;
          this.detail.transfersLoading = false;
        }
      } catch (err) {
        if (this.expandedKey === key) {
          this.detail.transfersError = extractMessage(err);
          this.detail.transfersLoading = false;
        }
      }
    })();
  }
}

export const transitStops = new TransitStopsStore();
