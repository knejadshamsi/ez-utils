// ============================================================
// Types
// ============================================================

/** The main view the user is on */
export type ViewState = 'welcome' | 'population' | 'network' | 'pt';

/** What the app is currently doing */
export type StatusState = 'idle' | 'importing' | 'exporting' | 'fetching';

/** Settings tabs */
export type SettingsTab = 'general' | 'sources' | 'display';

/** Available themes */
export type ThemeName = 'dark' | 'light';

/** Supported locales */
export type LocaleName = 'en' | 'fr';

/** Theme definition — extensible for future colors */
export interface ThemeDefinition {
  label: string;
  mapTiles: string;
  // Add future theme-specific values here (e.g. marker colors, layer tints)
}

/** Settings configuration */
export interface SettingsConfig {
  activeTab: SettingsTab;
  theme: ThemeName;
  locale: LocaleName;
}

// ============================================================
// Constants
// ============================================================

export const THEMES: Record<ThemeName, ThemeDefinition> = {
  dark: {
    label: 'Dark',
    mapTiles: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',
  },
  light: {
    label: 'Light',
    mapTiles: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
  },
};

export const LOCALES: Record<LocaleName, string> = {
  en: 'English',
  fr: 'Français',
};

const DEFAULT_SETTINGS: SettingsConfig = {
  activeTab: 'general',
  theme: 'dark',
  locale: 'en',
};

// ============================================================
// Stores
// ============================================================

/** Controls which view is rendered */
class UIState {
  view = $state<ViewState>('welcome');

  setView(view: ViewState) {
    this.view = view;
  }
}

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

/** Settings — includes theme, tabs, future config */
class SettingsStore {
  open = $state<boolean>(false);
  config = $state<SettingsConfig>({ ...DEFAULT_SETTINGS });

  toggle() {
    this.open = !this.open;
  }

  show() {
    this.open = true;
  }

  hide() {
    this.open = false;
  }

  setTab(tab: SettingsTab) {
    this.config.activeTab = tab;
  }

  setTheme(theme: ThemeName) {
    this.config.theme = theme;
    document.documentElement.setAttribute('data-theme', theme);
  }

  setLocale(locale: LocaleName) {
    this.config.locale = locale;
  }

  /** Current theme definition */
  get theme(): ThemeDefinition {
    return THEMES[this.config.theme];
  }

  reset() {
    this.config = { ...DEFAULT_SETTINGS };
    document.documentElement.setAttribute('data-theme', DEFAULT_SETTINGS.theme);
  }
}

// ============================================================
// Singletons
// ============================================================

export const ui = new UIState();
export const status = new StatusStore();
export const settings = new SettingsStore();
