<script lang="ts">
  import { onMount } from 'svelte';
  import L from 'leaflet';
  import 'leaflet/dist/leaflet.css';
  import { t } from 'svelte-i18n';
  import { settings, status, THEMES } from '$lib/stores/ui.svelte';

  let mapContainer: HTMLDivElement;
  let map: L.Map;
  let tileLayer: L.TileLayer;
  let displayStatus = $state('');

  // Keep the last busy status visible during slide-out animation
  $effect(() => {
    if (status.isBusy) {
      displayStatus = status.current;
    }
  });

  // Montreal island center
  const MONTREAL_CENTER: L.LatLngExpression = [45.5017, -73.5673];
  const DEFAULT_ZOOM = 12;


  // Swap map tiles when theme changes
  $effect(() => {
    const url = THEMES[settings.config.theme].mapTiles;
    if (!map || !tileLayer) return;
    tileLayer.setUrl(url);
  });

  onMount(() => {
    map = L.map(mapContainer, {
      attributionControl: false,
      zoomControl: false,
    }).setView(MONTREAL_CENTER, DEFAULT_ZOOM);

    // Logo control — top-left
    const LogoControl = L.Control.extend({
      options: { position: 'topleft' },
      onAdd() {
        const container = L.DomUtil.create('div', 'leaflet-control ez-logo');
        container.innerHTML = 'EZ-Utils';
        L.DomEvent.disableClickPropagation(container);
        return container;
      },
    });
    new LogoControl().addTo(map);

    // Zoom control: [-] [+] — horizontal bar, top-right
    const ZoomControl = L.Control.extend({
      options: { position: 'topright' },
      onAdd(mapInstance: L.Map) {
        const container = L.DomUtil.create('div', 'leaflet-control ez-toolbar');

        const zoomOutBtn = L.DomUtil.create('a', 'ez-toolbar-btn', container);
        zoomOutBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/></svg>`;
        zoomOutBtn.title = 'Zoom out';
        zoomOutBtn.href = '#';
        zoomOutBtn.role = 'button';
        zoomOutBtn.setAttribute('aria-label', 'Zoom out');

        const zoomInBtn = L.DomUtil.create('a', 'ez-toolbar-btn', container);
        zoomInBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>`;
        zoomInBtn.title = 'Zoom in';
        zoomInBtn.href = '#';
        zoomInBtn.role = 'button';
        zoomInBtn.setAttribute('aria-label', 'Zoom in');

        L.DomEvent.disableClickPropagation(container);
        L.DomEvent.on(zoomOutBtn, 'click', (e) => {
          L.DomEvent.preventDefault(e);
          mapInstance.zoomOut();
        });
        L.DomEvent.on(zoomInBtn, 'click', (e) => {
          L.DomEvent.preventDefault(e);
          mapInstance.zoomIn();
        });

        return container;
      },
    });
    new ZoomControl().addTo(map);

    // Settings button — separate control, top-right (renders after zoom = right of zoom)
    const SettingsControl = L.Control.extend({
      options: { position: 'topright' },
      onAdd() {
        const container = L.DomUtil.create('div', 'leaflet-control ez-toolbar');
        const btn = L.DomUtil.create('a', 'ez-toolbar-btn', container);
        btn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>`;
        btn.title = 'Settings';
        btn.href = '#';
        btn.role = 'button';
        btn.setAttribute('aria-label', 'Settings');
        L.DomEvent.disableClickPropagation(container);
        L.DomEvent.on(btn, 'click', (e) => {
          L.DomEvent.preventDefault(e);
          settings.toggle();
        });
        return container;
      },
    });
    new SettingsControl().addTo(map);

    // Tile layer — uses current theme
    tileLayer = L.tileLayer(THEMES[settings.config.theme].mapTiles, {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/">CARTO</a>',
      maxZoom: 19,
    }).addTo(map);

    return () => {
      map.remove();
    };
  });
</script>

<div class="relative w-full h-full">
  <div bind:this={mapContainer} class="w-full h-full"></div>

  <!-- Status bar — top center, slides in/out -->
  <div class="ez-status-wrapper" class:ez-status-visible={status.isBusy}>
    <div class="ez-status">
      <span class="ez-status-spinner"></span>
      <span class="ez-status-text">{$t(`status.${displayStatus}`)}...</span>
    </div>
  </div>

</div>

<style>
  /* ── Logo ── */
  :global(.ez-logo) {
    display: flex;
    align-items: center;
    height: 30px;
    font-weight: 700;
    font-size: 0.85rem;
    letter-spacing: -0.025em;
    color: var(--color-base-content);
    background: var(--color-base-100);
    padding: 0 10px;
    border-radius: var(--radius-field);
    pointer-events: none;
    user-select: none;
    box-shadow: 0 1px 5px rgba(0, 0, 0, 0.4);
  }

  /* ── Make top-right controls horizontal ── */
  :global(.leaflet-top.leaflet-right) {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    gap: 6px;
    padding-top: 10px;
    padding-right: 10px;
  }
  :global(.leaflet-top.leaflet-right .leaflet-control) {
    margin: 0 !important;
  }

  /* ── Toolbar group (zoom, settings) ── */
  :global(.ez-toolbar) {
    display: flex;
    flex-direction: row;
    background: var(--color-base-100);
    border-radius: var(--radius-field);
    box-shadow: 0 1px 5px rgba(0, 0, 0, 0.4);
    overflow: hidden;
  }
  :global(.ez-toolbar-btn) {
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    width: 30px !important;
    height: 30px !important;
    line-height: normal !important;
    font-size: 0 !important;
    padding: 0 !important;
    background: var(--color-base-100) !important;
    color: var(--color-base-content) !important;
    cursor: pointer;
    text-decoration: none !important;
    border: none !important;
    border-right: 1px solid var(--color-base-300) !important;
  }
  :global(.ez-toolbar-btn:last-child) {
    border-right: none !important;
  }
  :global(.ez-toolbar-btn:hover) {
    background: var(--color-base-200) !important;
  }

  /* ── Status bar ── */
  .ez-status-wrapper {
    position: absolute;
    top: 0;
    left: 50%;
    transform: translateX(-50%) translateY(-200%);
    z-index: 1000;
    pointer-events: none;
    transition: transform 0.3s ease;
  }
  .ez-status-visible {
    transform: translateX(-50%) translateY(10px);
  }

  .ez-status {
    display: flex;
    align-items: center;
    height: 30px;
    gap: 6px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    font-size: 0.75rem;
    padding: 0 12px;
    border-radius: var(--radius-field);
    white-space: nowrap;
    box-shadow: 0 1px 5px rgba(0, 0, 0, 0.4);
  }
  .ez-status-spinner {
    width: 10px;
    height: 10px;
    border: 2px solid var(--color-base-300);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: ez-spin 0.6s linear infinite;
  }
  @keyframes ez-spin {
    to { transform: rotate(360deg); }
  }

</style>
