<script lang="ts">
  import { onMount } from 'svelte';
  import { getCurrentWindow } from '@tauri-apps/api/window';
  import { locale } from 'svelte-i18n';
  import MapView from '$lib/components/MapView.svelte';
  import SettingsModal from '$lib/components/SettingsModal.svelte';
  import ExportModal from '$lib/components/ExportModal.svelte';
  import WorkspaceCloseModal from '$lib/components/WorkspaceCloseModal.svelte';
  import WorkspaceRecoveryModal from '$lib/components/WorkspaceRecoveryModal.svelte';
  import WorkspaceStartupModal from '$lib/components/WorkspaceStartupModal.svelte';
  import { drawers, mapViewport, settings, sourceStore } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { workspace } from '$lib/stores/workspace.svelte';

  // Sync locale store → svelte-i18n whenever it changes
  $effect(() => {
    locale.set(settings.config.locale);
  });

  onMount(() => {
    document.documentElement.setAttribute('data-theme', settings.config.theme);
    let unlisten: (() => void) | undefined;
    void getCurrentWindow().onCloseRequested(async (event) => {
      await workspace.handleWindowCloseRequest(event);
    }).then((dispose) => {
      unlisten = dispose;
    });
    return () => {
      unlisten?.();
    };
  });

  $effect(() => {
    if (!workspace.ready) return;

    settings.config.theme;
    settings.config.locale;
    sourceStore.popoverOpen;
    mapViewport.state.center;
    mapViewport.state.zoom;
    sources.items;
    drawers.snapshot();

    workspace.queueUiPersist();
  });

  $effect(() => {
    settings.config.autosaveSettings.enabled;
    settings.config.autosaveSettings.intervalMinutes;
    workspace.refreshAutosave();
  });
</script>

<main class="w-full h-screen">
  <MapView />
  <SettingsModal />
  <ExportModal />
  <WorkspaceCloseModal />
  <WorkspaceRecoveryModal />
  <WorkspaceStartupModal />
</main>
