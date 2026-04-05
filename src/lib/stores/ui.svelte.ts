// ============================================================
// Types
// ============================================================

/** What the app is currently doing */
export type StatusState = 'idle' | 'importing' | 'exporting' | 'fetching' | 'saved' | 'auto_saved';

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
  current = $state<StatusState>('idle');
  private timer: ReturnType<typeof setTimeout> | null = null;

  set(status: StatusState) {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.current = status;
    if (status === 'saved' || status === 'auto_saved') {
      this.timer = setTimeout(() => {
        this.current = 'idle';
        this.timer = null;
      }, 2000);
    }
  }

  get isBusy(): boolean {
    return this.current === 'importing' || this.current === 'exporting' || this.current === 'fetching';
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

  togglePopover() {
    this.popoverOpen = !this.popoverOpen;
  }

  hidePopover() {
    this.popoverOpen = false;
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
