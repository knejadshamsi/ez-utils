// ============================================================
// Types
// ============================================================

/** What the app is currently doing */
export type StatusState = 'idle' | 'importing' | 'exporting' | 'fetching';

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
};

// ============================================================
// Stores
// ============================================================

/** Controls what action the app is performing */
class StatusStore {
  current = $state<StatusState>('idle');

  set(status: StatusState) {
    this.current = status;
  }

  get isBusy(): boolean {
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
}

// ============================================================
// Singletons
// ============================================================

export const status = new StatusStore();
export const settings = new SettingsStore();
export const drawers = new DrawerStore();
export const exportStore = new ExportStore();
export const sourceStore = new SourceStore();
