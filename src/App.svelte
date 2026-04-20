<script lang="ts">
  import { onMount } from 'svelte';
  import { getCurrentWindow } from '@tauri-apps/api/window';
  import { locale } from 'svelte-i18n';
  import MapView from '$lib/components/MapView.svelte';
  import SettingsModal from '$lib/components/SettingsModal.svelte';
  import ExportModal from '$lib/components/ExportModal.svelte';
  import CloseModal from '$lib/components/CloseModal.svelte';
  import DeletePersonModal from '$lib/components/population/DeletePersonModal.svelte';
  import DeleteNetworkItemModal from '$lib/components/network/DeleteNetworkItemModal.svelte';
  import DeleteLineModal from '$lib/components/transit/DeleteLineModal.svelte';
  import DeleteRouteModal from '$lib/components/transit/DeleteRouteModal.svelte';
  import DeleteStopFacilityModal from '$lib/components/transit/DeleteStopFacilityModal.svelte';
  import StartupModal from '$lib/components/StartupModal.svelte';
  import StatusBar from '$lib/components/StatusBar.svelte';
  import { drawers, mapViewport, settings, sourceStore } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { ez } from '$lib/stores/ez.svelte';

  $effect(() => {
    locale.set(settings.config.locale);
  });

  onMount(() => {
    document.documentElement.setAttribute('data-theme', settings.config.theme);
    let unlisten: (() => void) | undefined;
    void getCurrentWindow().onCloseRequested(async (event) => {
      await ez.handleWindowCloseRequest(event);
    }).then((dispose) => {
      unlisten = dispose;
    });
    return () => {
      unlisten?.();
    };
  });

  $effect(() => {
    if (!ez.active) return;

    settings.config.theme;
    settings.config.locale;
    settings.config.networkFocusMode;
    sourceStore.popoverOpen;
    mapViewport.state.center;
    mapViewport.state.zoom;
    sources.items;
    drawers.snapshot();

    ez.queueUiPersist();
  });

  $effect(() => {
    settings.config.autosaveSettings.enabled;
    settings.config.autosaveSettings.intervalMinutes;
    ez.refreshAutosave();
  });
</script>

<main class="w-full h-screen">
  <MapView />
  <StatusBar />
  <SettingsModal />
  <ExportModal />
  <CloseModal />
  <DeletePersonModal />
  <DeleteNetworkItemModal />
  <DeleteLineModal />
  <DeleteRouteModal />
  <DeleteStopFacilityModal />
  <StartupModal />
</main>
