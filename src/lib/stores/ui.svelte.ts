import { invoke } from '@tauri-apps/api/core';

import type { CrsConfig, CrsInfo } from '$lib/components/types';
import { sources } from '$lib/stores/data.svelte';
import { network } from '$lib/stores/network.svelte';
import { population } from '$lib/stores/population.svelte';
import { transitStops } from '$lib/stores/transit.svelte';

let crsPresetCache: Record<string, string> | null = null;

async function ensureCrsPresetCache(): Promise<Record<string, string>> {
  if (crsPresetCache !== null) return crsPresetCache;
  const list = await invoke<CrsInfo[]>('list_crs_presets');
  const cache: Record<string, string> = {};
  for (const info of list) {
    if (info.kind.kind !== 'custom') {
      cache[info.kind.kind] = info.projString;
    }
  }
  crsPresetCache = cache;
  return cache;
}

export async function currentProjString(): Promise<string> {
  const crs = settings.config.crs;
  if (crs.kind === 'custom') return crs.projString;
  const cache = await ensureCrsPresetCache();
  return cache[crs.kind] ?? '';
}

// ============================================================
// Types
// ============================================================

/** What the app is currently doing */
export type StatusState =
  | 'idle'
  | 'unpacking'
  | 'loading'
  | 'importing'
  | 'exporting'
  | 'fetching'
  | 'saving'
  | 'saved'
  | 'auto_saved'
  | 'deleted'
  | 'editing'
  | 'network_below_zoom'
  | 'network_settling'
  | 'network_add_node'
  | 'network_add_link_from'
  | 'network_add_link_to'
  | 'population_settling'
  | 'population_add_person'
  | 'population_add_activity'
  | 'transit_settling';

/** Available themes */
export type ThemeName = 'dark' | 'light';

/** Supported locales */
export type LocaleName = 'en' | 'fr';

/** Theme definition — extensible for future colors */
export interface ThemeDefinition {
  mapTiles: string;
  // Add future theme-specific values here (e.g. marker colors, layer tints)
}

/** Settings configuration */
export interface SettingsConfig {
  theme: ThemeName;
  locale: LocaleName;
  autosaveSettings: AutosaveSettings;
  crs: CrsConfig;
}

export interface MapViewportState {
  center: [number, number];
  zoom: number;
}

export interface AutosaveSettings {
  enabled: boolean;
  intervalMinutes: number;
}

// ============================================================
// Constants
// ============================================================

export const THEMES: Record<ThemeName, ThemeDefinition> = {
  dark: {
    mapTiles: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',
  },
  light: {
    mapTiles: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
  },
};

export const LOCALES: Record<LocaleName, string> = {
  en: 'English',
  fr: 'Français',
};

const DEFAULT_SETTINGS: SettingsConfig = {
  theme: 'dark',
  locale: 'en',
  autosaveSettings: {
    enabled: false,
    intervalMinutes: 5,
  },
  crs: { kind: 'montreal_mtm8' },
};

const DEFAULT_MAP_VIEWPORT: MapViewportState = {
  center: [-73.5673, 45.5017],
  zoom: 12,
};

// ============================================================
// Stores
// ============================================================

/** Controls what action the app is performing */
class StatusStore {
  operationStatus = $state<StatusState | null>(null);
  transientStatus = $state<StatusState | null>(null);
  private transientTimer: ReturnType<typeof setTimeout> | null = null;

  get derivedStatus(): StatusState {
    if (sources.activeKind === 'network') {
      if (network.mapAction === 'add_node') return 'network_add_node';
      if (network.mapAction === 'add_link_from') return 'network_add_link_from';
      if (network.mapAction === 'add_link_to') return 'network_add_link_to';
      if (network.belowZoom) return 'network_below_zoom';
      if (network.settling) return 'network_settling';
      if (network.loading) return 'fetching';
    }
    if (sources.activeKind === 'population') {
      if (population.mapAction === 'add_person') return 'population_add_person';
      if (population.mapAction === 'add_activity') return 'population_add_activity';
      if (population.editLocations) return 'editing';
      if (population.settling) return 'population_settling';
      if (population.loading) return 'fetching';
    }
    if (sources.activeKind === 'transit') {
      if (transitStops.settling) return 'transit_settling';
      if (transitStops.loading) return 'fetching';
    }
    return 'idle';
  }

  get current(): StatusState {
    return this.operationStatus ?? this.transientStatus ?? this.derivedStatus;
  }

  setOperation(status: StatusState | null) {
    this.operationStatus = status;
  }

  flash(status: 'saved' | 'auto_saved' | 'deleted') {
    if (this.transientTimer) clearTimeout(this.transientTimer);
    this.transientStatus = status;
    this.transientTimer = setTimeout(() => {
      this.transientStatus = null;
      this.transientTimer = null;
    }, 2000);
  }

  get isBusy(): boolean {
    const s = this.current;
    return s === 'unpacking' || s === 'loading' || s === 'importing' || s === 'exporting' || s === 'fetching' || s === 'saving';
  }

  get isVisible(): boolean {
    return this.current !== 'idle';
  }
}

/** Export modal state */
class ExportStore {
  open = $state<boolean>(false);

  show() {
    this.open = true;
  }

  hide() {
    this.open = false;
  }
}

/** Source popover state */
class SourceStore {
  popoverOpen = $state<boolean>(false);
  /** When true, switching the active source first wipes the previous source's map data. */
  wipeOnSwitch = $state<boolean>(true);

  togglePopover() {
    this.popoverOpen = !this.popoverOpen;
  }

  hidePopover() {
    this.popoverOpen = false;
  }

  toggleWipeOnSwitch() {
    this.wipeOnSwitch = !this.wipeOnSwitch;
  }

  hydrate(open: boolean) {
    this.popoverOpen = open;
  }
}

/** Settings — includes theme, tabs, future config */
class SettingsStore {
  open = $state<boolean>(false);
  config = $state<SettingsConfig>({ ...DEFAULT_SETTINGS });

  toggle() {
    this.open = !this.open;
  }

  hide() {
    this.open = false;
  }

  setTheme(theme: ThemeName) {
    this.config.theme = theme;
    document.documentElement.setAttribute('data-theme', theme);
  }

  setLocale(locale: LocaleName) {
    this.config.locale = locale;
  }

  setAutosaveEnabled(enabled: boolean) {
    this.config.autosaveSettings.enabled = enabled;
  }

  setAutosaveInterval(intervalMinutes: number) {
    this.config.autosaveSettings.intervalMinutes = Math.min(24 * 60, Math.max(5, Math.round(intervalMinutes)));
  }

  async setCrs(config: CrsConfig): Promise<void> {
    this.config.crs = config;
    if (sources.items.length > 0) return;
    try {
      const center = await invoke<[number, number]>('set_crs', { config });
      mapViewport.set(center, 12);
    } catch {
      // Backend rejects when no session (welcome screen) or sources exist.
      // Frontend keeps the value; new_from_xml picks it up on session creation.
    }
  }

  hydrate(config: SettingsConfig) {
    this.config = {
      ...config,
      autosaveSettings: {
        enabled: config.autosaveSettings?.enabled ?? DEFAULT_SETTINGS.autosaveSettings.enabled,
        intervalMinutes: Math.min(
          24 * 60,
          Math.max(5, config.autosaveSettings?.intervalMinutes ?? DEFAULT_SETTINGS.autosaveSettings.intervalMinutes)
        ),
      },
      crs: config.crs ?? DEFAULT_SETTINGS.crs,
    };
    document.documentElement.setAttribute('data-theme', this.config.theme);
  }
}

/** Manages drawer open/close state */
class DrawerStore {
  private openSet = $state<Set<string>>(new Set());

  isOpen(id: string): boolean {
    return this.openSet.has(id);
  }

  toggle(id: string) {
    if (this.openSet.has(id)) {
      this.openSet.delete(id);
    } else {
      this.openSet.add(id);
    }
    this.openSet = new Set(this.openSet);
  }

  open(id: string) {
    if (this.openSet.has(id)) return;
    this.openSet.add(id);
    this.openSet = new Set(this.openSet);
  }

  close(id: string) {
    if (!this.openSet.has(id)) return;
    this.openSet.delete(id);
    this.openSet = new Set(this.openSet);
  }

  snapshot(): string[] {
    return Array.from(this.openSet);
  }

  hydrate(ids: string[]) {
    this.openSet = new Set(ids);
  }
}

class MapViewportStore {
  state = $state<MapViewportState>({ ...DEFAULT_MAP_VIEWPORT });

  set(center: [number, number], zoom: number) {
    this.state = { center, zoom };
  }

  hydrate(next: MapViewportState) {
    this.state = { ...next };
  }
}

// ============================================================
// Singletons
// ============================================================

export const status = new StatusStore();
export const settings = new SettingsStore();
export const drawers = new DrawerStore();
export const exportStore = new ExportStore();
export const sourceStore = new SourceStore();
export const mapViewport = new MapViewportStore();
