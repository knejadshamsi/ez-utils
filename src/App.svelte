<script lang="ts">
  import { onMount } from 'svelte';
  import { locale } from 'svelte-i18n';
  import MapView from '$lib/components/MapView.svelte';
  import SettingsModal from '$lib/components/SettingsModal.svelte';
  import { settings } from '$lib/stores/ui.svelte';

  // Sync locale store → svelte-i18n whenever it changes
  $effect(() => {
    locale.set(settings.config.locale);
  });

  onMount(() => {
    document.documentElement.setAttribute('data-theme', settings.config.theme);
  });
</script>

<main class="w-full h-screen">
  <MapView />
  <SettingsModal />
</main>
