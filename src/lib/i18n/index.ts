import { register, init, getLocaleFromNavigator } from 'svelte-i18n';

// Register locales — lazy-loaded, only fetched when needed
register('en', () => import('./en.json'));
register('fr', () => import('./fr.json'));

// Initialize with English as default, fallback
init({
  fallbackLocale: 'en',
  initialLocale: getLocaleFromNavigator() ?? 'en',
});
