<script lang="ts">
  import { onMount } from 'svelte';
  import L from 'leaflet';
  import 'leaflet/dist/leaflet.css';
  import PrimaryDrawer from '$lib/components/drawers/PrimaryDrawer.svelte';
  import SecondaryDrawer from '$lib/components/drawers/SecondaryDrawer.svelte';
  import { settings, exportStore, sourceStore, THEMES, mapViewport } from '$lib/stores/ui.svelte';
  import { population } from '$lib/stores/population.svelte';
  import { network } from '$lib/stores/network.svelte';
  import { transit, transitEdit } from '$lib/stores/transit.svelte';
  import { handleTransitMapClick } from '$lib/stores/transit/profile-stops-actions';
  import { createNode as createNetworkNode, createLink as createNetworkLink } from '$lib/stores/network-actions';
  import { ensureNetworkLayers, clearNetworkLayers, type NetworkLayers } from '$lib/components/network/network-map';
  import { ensureTransitLayers, clearTransitLayers, type TransitLayers } from '$lib/components/transit/transit-map';
  import NetworkMapLayer from '$lib/components/network/NetworkMapLayer.svelte';
  import TransitMapLayer from '$lib/components/transit/TransitMapLayer.svelte';
  import { ez } from '$lib/stores/ez.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import SourcePopover from '$lib/components/SourcePopover.svelte';

  let networkLayers = $state<NetworkLayers | null>(null);
  let transitLayers = $state<TransitLayers | null>(null);
  let map = $state<L.Map | null>(null);

  let mapContainer: HTMLDivElement;
  let tileLayer: L.TileLayer;
  // Montreal island center
  // Swap map tiles when theme changes
  $effect(() => {
    const url = THEMES[settings.config.theme].mapTiles;
    if (!map || !tileLayer) return;
    tileLayer.setUrl(url);
  });

  $effect(() => {
    if (!map) return;
    const center = map.getCenter();
    const [lng, lat] = mapViewport.state.center;
    if (Math.abs(center.lng - lng) < 0.000001 && Math.abs(center.lat - lat) < 0.000001 && map.getZoom() === mapViewport.state.zoom) {
      return;
    }
    map.setView([lat, lng], mapViewport.state.zoom);
  });

  onMount(() => {
    const m = L.map(mapContainer, {
      attributionControl: false,
      zoomControl: false,
    }).setView([mapViewport.state.center[1], mapViewport.state.center[0]], mapViewport.state.zoom);

    m.on('moveend', () => {
      const center = m.getCenter();
      mapViewport.set([center.lng, center.lat], m.getZoom());
      ez.queueUiPersist();
      const bounds = m.getBounds();
      const viewportBounds = {
        west: bounds.getWest(),
        south: bounds.getSouth(),
        east: bounds.getEast(),
        north: bounds.getNorth(),
      };
      population.scheduleViewportFetch(viewportBounds);
      network.scheduleViewportFetch(viewportBounds, m.getZoom());
      transit.scheduleStopsFetch(viewportBounds);
    });

    m.on('click', (event) => {
      if (sources.activeKind === 'network') {
        if (network.mapAction === 'add_node') {
          const id = network.generateNodeId();
          void createNetworkNode({ id, lng: event.latlng.lng, lat: event.latlng.lat });
          network.mapAction = 'idle';
          return;
        }
        if (network.selection) {
          network.clearSelection();
          return;
        }
      } else if (sources.activeKind === 'population') {
        void population.handleMapClick(event.latlng.lng, event.latlng.lat);
      } else if (sources.activeKind === 'transit') {
        void handleTransitMapClick(event.latlng.lng, event.latlng.lat);
      }
    });

    // Escape cancels any active map action.
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        if (sources.activeKind === 'network' && network.mapAction !== 'idle') {
          network.cancelMapAction();
        } else if (sources.activeKind === 'population' && population.mapAction !== null) {
          population.mapAction = null;
        } else if (sources.activeKind === 'transit' && transitEdit.mapAction !== 'idle') {
          transitEdit.setMapAction('idle');
        }
      }
    };
    document.addEventListener('keydown', onKeyDown);

    // Logo control — top-left
    const LogoControl = L.Control.extend({
      options: { position: 'topleft' },
      onAdd() {
        const c = L.DomUtil.create('div', 'leaflet-control ez-logo');
        c.innerHTML = 'EZ-Utils';
        L.DomEvent.disableClickPropagation(c);
        return c;
      },
    });
    new LogoControl().addTo(m);

    // Single-button toolbar control factory.
    type Pos = 'topleft' | 'topright';
    const addToolbarBtn = (pos: Pos, svg: string, title: string, onClick: () => void, extraClass = '') => {
      const Ctl = L.Control.extend({
        options: { position: pos },
        onAdd() {
          const c = L.DomUtil.create('div', 'leaflet-control ez-toolbar');
          const b = L.DomUtil.create('a', `ez-toolbar-btn ${extraClass}`.trim(), c);
          b.innerHTML = svg; b.title = title; b.href = '#'; b.role = 'button'; b.setAttribute('aria-label', title);
          L.DomEvent.disableClickPropagation(c);
          L.DomEvent.on(b, 'click', (e) => { L.DomEvent.preventDefault(e); onClick(); });
          return c;
        },
      });
      new Ctl().addTo(m);
    };
    const SVG_ATTRS = `xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"`;

    addToolbarBtn('topleft', `<svg ${SVG_ATTRS}><path d="M4 5h16"/><path d="M4 12h16"/><path d="M4 19h16"/></svg>`, 'Sources', () => sourceStore.togglePopover(), 'ez-source-btn');

    // Zoom control: [-] [+] — single horizontal bar, top-right.
    const ZoomControl = L.Control.extend({
      options: { position: 'topright' },
      onAdd(m: L.Map) {
        const c = L.DomUtil.create('div', 'leaflet-control ez-toolbar');
        const make = (svg: string, title: string, onClick: () => void) => {
          const b = L.DomUtil.create('a', 'ez-toolbar-btn', c);
          b.innerHTML = svg; b.title = title; b.href = '#'; b.role = 'button'; b.setAttribute('aria-label', title);
          L.DomEvent.on(b, 'click', (e) => { L.DomEvent.preventDefault(e); onClick(); });
        };
        make(`<svg ${SVG_ATTRS}><path d="M5 12h14"/></svg>`, 'Zoom out', () => m.zoomOut());
        make(`<svg ${SVG_ATTRS}><path d="M5 12h14"/><path d="M12 5v14"/></svg>`, 'Zoom in', () => m.zoomIn());
        L.DomEvent.disableClickPropagation(c);
        return c;
      },
    });
    new ZoomControl().addTo(m);

    addToolbarBtn('topright', `<svg ${SVG_ATTRS}><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>`, 'Settings', () => settings.toggle());
    addToolbarBtn('topright', `<svg ${SVG_ATTRS}><path d="M15.2 3a2 2 0 0 1 1.4.6l3.8 3.8a2 2 0 0 1 .6 1.4V19a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z"/><path d="M17 21v-7a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v7"/><path d="M7 3v4a1 1 0 0 0 1 1h7"/></svg>`, 'Save', () => void ez.save());
    addToolbarBtn('topright', `<svg ${SVG_ATTRS}><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M8 12h8"/><path d="m12 8 4 4-4 4"/></svg>`, 'Export', () => exportStore.show());

    // Tile layer — uses current theme
    tileLayer = L.tileLayer(THEMES[settings.config.theme].mapTiles, {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/">CARTO</a>',
      maxZoom: 19,
    }).addTo(m);
    population.populationLayer = L.layerGroup().addTo(m);
    population.selectedPlanLayer = L.layerGroup().addTo(m);
    population.map = m;
    networkLayers = ensureNetworkLayers(m);
    transitLayers = ensureTransitLayers(m);

    m.on('zoomend', () => {
      if (population.selected) {
        population.renderActivePlan();
      }
    });

    const bounds = m.getBounds();
    const initialBounds = {
      west: bounds.getWest(),
      south: bounds.getSouth(),
      east: bounds.getEast(),
      north: bounds.getNorth(),
    };
    population.scheduleViewportFetch(initialBounds);
    network.scheduleViewportFetch(initialBounds, m.getZoom());
    transit.scheduleStopsFetch(initialBounds);

    map = m;

    return () => {
      population.populationLayer = null;
      population.selectedPlanLayer = null;
      population.map = null;
      if (networkLayers) {
        clearNetworkLayers(networkLayers);
        networkLayers = null;
      }
      if (transitLayers) {
        clearTransitLayers(transitLayers);
        transitLayers = null;
      }
      map = null;
      document.removeEventListener('keydown', onKeyDown);
      m.remove();
    };
  });

  // Wipe map state when the active source changes (governed by sourceStore.wipeOnSwitch).
  let wipeInitialized = false;
  let previousActiveId: string | null = null;
  $effect(() => {
    const _id = sources.activeId;
    if (!wipeInitialized) {
      wipeInitialized = true;
      previousActiveId = _id;
      return;
    }
    if (_id === previousActiveId) return;
    previousActiveId = _id;
    if (sourceStore.wipeOnSwitch) {
      network.wipe();
      population.wipe();
    }
  });

  // Trigger fetch when switching to a new source type.
  let previousActiveKind: string | null = null;
  $effect(() => {
    const kind = sources.activeKind;
    if (kind === previousActiveKind) return;
    previousActiveKind = kind;
    if (!map) return;
    const bounds = map.getBounds();
    const viewport = {
      west: bounds.getWest(),
      south: bounds.getSouth(),
      east: bounds.getEast(),
      north: bounds.getNorth(),
    };
    if (kind === 'network') {
      network.scheduleViewportFetch(viewport, map.getZoom());
    } else if (kind === 'population') {
      population.scheduleViewportFetch(viewport);
    } else if (kind === 'transit') {
      transit.scheduleStopsFetch(viewport);
    }
  });

</script>

<div class="relative w-full h-full overflow-hidden">
  <div bind:this={mapContainer} class="w-full h-full"></div>

  {#if map && transitLayers}
    <TransitMapLayer {map} layers={transitLayers} />
  {/if}
  {#if map && networkLayers}
    <NetworkMapLayer {map} layers={networkLayers} />
  {/if}

  <!-- Source popover -->
  <SourcePopover />

  <PrimaryDrawer />
  <SecondaryDrawer />
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

  /* ── Make top-left controls horizontal ── */
  :global(.leaflet-top.leaflet-left) {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    gap: 6px;
    padding-top: 10px;
    padding-left: 10px;
  }
  :global(.leaflet-top.leaflet-left .leaflet-control) {
    margin: 0 !important;
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

</style>
