<script lang="ts">
  import L from 'leaflet';

  import { mapViewport } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { transit, transitEdit, transitStops } from '$lib/stores/transit.svelte';
  import {
    renderTransitStops,
    clearTransitLayers,
    clearTransitStops,
    renderLinePaths,
    clearLinePaths,
    renderRouteSequence,
    clearRouteSequence,
    renderNearbyLinks,
    clearNearbyLinks,
    type TransitLayers,
  } from '$lib/components/transit/transit-map';
  import {
    assignLinkRefFromMap,
    fetchNearbyLinksForEdit,
  } from '$lib/stores/transit/profile-stops-actions';

  interface Props {
    map: L.Map;
    layers: TransitLayers;
  }

  let { map, layers }: Props = $props();

  // Non-reactive list of transparent drag overlay markers. Kept OFF the $state
  // TransitLayers so mutations don't retrigger the effect that owns them
  // (that was causing an infinite update loop).
  let dragMarkers: L.Marker[] = [];
  function clearDragMarkers() {
    for (const m of dragMarkers) {
      layers.editLayer.removeLayer(m);
    }
    dragMarkers = [];
  }

  // Nearby network-link polylines; also kept non-reactive.
  let nearbyLinkLines: L.Polyline[] = [];

  // Render nearby network links on the map whenever the edit or create flow
  // has a set. Use create-flow links first (newer selection) else edit-flow.
  // The currently-assigned link renders in primary color for easy spotting.
  $effect(() => {
    const editLinks = transitEdit.editStopFacilityTarget?.nearbyLinks ?? [];
    const createLinks = transitEdit.profileStopsDraft?.createNearbyLinks ?? [];
    const items = createLinks.length > 0 ? createLinks : editLinks;
    const currentRef = createLinks.length > 0
      ? transitEdit.profileStopsDraft?.createLinkRefId ?? null
      : transitEdit.editStopFacilityTarget?.linkRefId ?? null;
    const currentNonEmpty = currentRef && currentRef.length > 0 ? currentRef : null;
    if (items.length === 0) {
      nearbyLinkLines = clearNearbyLinks(layers.editLayer, nearbyLinkLines);
      return;
    }
    nearbyLinkLines = renderNearbyLinks(
      layers.editLayer,
      nearbyLinkLines,
      items,
      currentNonEmpty,
      (id) => void assignLinkRefFromMap(id),
    );
  });

  // Fly to the stop + fetch nearby network links whenever a new facility edit
  // begins (originalId changes). Fires once per edit session; drag doesn't refire.
  let previousEditId: string | null = null;
  $effect(() => {
    const target = transitEdit.editStopFacilityTarget;
    const id = target?.originalId ?? null;
    if (id === previousEditId) return;
    previousEditId = id;
    if (!target) return;
    const lng = Number(target.lng);
    const lat = Number(target.lat);
    if (!Number.isFinite(lng) || !Number.isFinite(lat)) return;
    const targetZoom = Math.max(map.getZoom(), 16);
    map.flyTo([lat, lng], targetZoom, { duration: 0.6 });
    void fetchNearbyLinksForEdit();
  });

  // Re-render transit stop markers on data / color / zoom / source-kind / tab changes.
  $effect(() => {
    if (sources.activeKind !== 'transit') {
      clearTransitLayers(layers);
      return;
    }
    if (transit.primaryTab !== 'stops') {
      clearTransitStops(layers);
      return;
    }
    const color = sources.active?.color ?? '#0066ff';
    const opacity = sources.active?.opacity ?? 1;
    renderTransitStops(
      layers,
      transitStops.items,
      color,
      opacity,
      mapViewport.state.zoom,
      (stopId) => void transitStops.locateStop(stopId),
    );
  });

  // When a facility is being edited, drop a transparent draggable marker on top
  // of every sequence marker that references it (via profileStops). Dragging fires
  // the 'drag' event continuously: we move the visible sequence marker(s) + the
  // polyline vertex(es) + the bbox stop marker in real time, and update the edit
  // target's lng/lat so the form mirrors the dragged position. dragend is implicit
  // - state is already current. User still clicks Save to persist.
  $effect(() => {
    const target = transitEdit.editStopFacilityTarget;
    const refId = target?.originalId ?? null;
    const routeId = transit.selectedRouteId;
    const profileStops = transit.profileStops;

    // Tear down any existing drag overlays on every change. Additionally,
    // when an edit session ends (target goes null, save or cancel), force a
    // re-render of relevant layers from authoritative data - otherwise any
    // drag mutations to sequence markers / polyline / bbox marker would
    // persist on cancel because they were made directly via Leaflet calls
    // bypassing the source-of-truth arrays.
    clearDragMarkers();
    if (!target || !refId || !routeId) {
      if (sources.activeKind === 'transit') {
        const color = sources.active?.color ?? '#0066ff';
        const opacity = sources.active?.opacity ?? 1;
        if (transit.primaryTab === 'stops') {
          renderTransitStops(
            layers,
            transitStops.items,
            color,
            opacity,
            mapViewport.state.zoom,
            (stopId) => void transitStops.locateStop(stopId),
          );
        }
        if (transit.primaryTab === 'lines' && transit.selectedLineId !== null) {
          renderLinePaths(
            layers,
            transit.linePathStops,
            transit.selectedRouteId,
            transit.linkedRoutePath,
          );
          if (transit.selectedRouteId !== null) {
            renderRouteSequence(layers, transit.linePathStops, transit.selectedRouteId);
          }
        }
      }
      return;
    }

    // Sequence indices in the current route that reference this facility.
    const matches = profileStops.filter((p) => p.stopRefId === refId);
    if (matches.length === 0) return;

    const lng0 = Number(target.lng);
    const lat0 = Number(target.lat);
    if (!Number.isFinite(lng0) || !Number.isFinite(lat0)) return;

    // Population pattern: during drag, only mutate visuals. Do NOT touch store
    // state - that would retrigger this $effect and destroy the marker mid-drag.
    // Form state syncs on dragend.
    const onDragTo = (newLat: number, newLng: number) => {
      for (const pfs of matches) {
        const key = `${pfs.sequence}`;
        const marker = layers.sequenceMarkers.get(key);
        marker?.setLatLng([newLat, newLng]);
      }
      const line = layers.pathLines.get(routeId);
      if (line) {
        const idxs: number[] = transit.linePathStops
          .filter((p) => p.routeId === routeId)
          .slice()
          .sort((a, b) => a.sequence - b.sequence)
          .map((p, i) => (matches.some((m) => m.sequence === p.sequence) ? i : -1))
          .filter((i) => i >= 0);
        const latlngs = line.getLatLngs() as L.LatLng[];
        for (const i of idxs) {
          latlngs[i] = L.latLng(newLat, newLng);
        }
        line.setLatLngs(latlngs);
      }
      const bbox = layers.markers.get(refId);
      bbox?.marker.setLatLng([newLat, newLng]);
    };

    for (const _pfs of matches) {
      const icon = L.divIcon({
        className: 'transit-edit-drag',
        html: '',
        iconSize: [28, 28],
        iconAnchor: [14, 14],
      });
      const dragMarker = L.marker([lat0, lng0], {
        icon,
        draggable: true,
        opacity: 0,
        zIndexOffset: 3000,
      });
      dragMarker.on('drag', (ev) => {
        const ll = (ev.target as L.Marker).getLatLng();
        onDragTo(ll.lat, ll.lng);
      });
      dragMarker.on('dragend', (ev) => {
        const ll = (ev.target as L.Marker).getLatLng();
        transitEdit.updateEditStopFacility({
          lng: ll.lng.toString(),
          lat: ll.lat.toString(),
          error: null,
        });
        // Re-query nearby network links around the new drop position so the
        // user can re-pick if they dragged the stop off its original link.
        void fetchNearbyLinksForEdit();
      });
      dragMarker.addTo(layers.editLayer);
      dragMarkers.push(dragMarker);
    }
  });

  // Hydrate line path stops whenever the selected line changes (transit only).
  $effect(() => {
    const lineId = transit.selectedLineId;
    const sourceName = sources.active?.name;
    if (sources.activeKind !== 'transit' || !lineId || !sourceName) return;
    void transit.loadLinePathStopsForLine(sourceName, lineId);
  });

  // Hydrate linked route geometry only when the transit source references a network
  // via OPTIONAL INPUTS. No reference -> clear so renderer falls back to straight-line.
  $effect(() => {
    const routeId = transit.selectedRouteId;
    const lineId = transit.selectedLineId;
    const transitSource = sources.active?.name;
    const linkedName = sources.active?.transitMetadata?.linkedNetworkSourceName ?? null;
    const networkSourceName =
      linkedName && sources.items.some((s) => s.kind === 'network' && s.name === linkedName)
        ? linkedName
        : null;
    if (
      sources.activeKind !== 'transit' ||
      !transitSource ||
      !lineId ||
      !routeId ||
      !networkSourceName
    ) {
      transit.clearLinkedRoutePath();
      return;
    }
    void transit.loadLinkedRoutePath(transitSource, networkSourceName, lineId, routeId);
  });

  // Render / clear line-path polylines (Lines tab only).
  // One polyline per route, coords sourced from linkedRoutePath when it applies
  // to that route, otherwise from the profile-stop straight-line. Single render
  // call updates any existing polyline in place via setLatLngs.
  $effect(() => {
    if (
      sources.activeKind !== 'transit' ||
      transit.primaryTab !== 'lines' ||
      transit.selectedLineId === null
    ) {
      clearLinePaths(layers);
      return;
    }
    renderLinePaths(
      layers,
      transit.linePathStops,
      transit.selectedRouteId,
      transit.linkedRoutePath,
    );
  });

  // Render / clear numbered stop markers for the selected route (Lines tab only).
  $effect(() => {
    const routeId = transit.selectedRouteId;
    if (
      sources.activeKind !== 'transit' ||
      transit.primaryTab !== 'lines' ||
      routeId === null
    ) {
      clearRouteSequence(layers);
      return;
    }
    renderRouteSequence(layers, transit.linePathStops, routeId);
  });

  // FlyTo the currently-expanded transit stop.
  $effect(() => {
    const id = transitStops.expandedStopId;
    if (!id) return;
    const stop = transitStops.items.find((s) => s.id === id);
    if (!stop) return;
    const targetZoom = Math.max(map.getZoom(), 15);
    map.flyTo([stop.lat, stop.lng], targetZoom, { duration: 0.6 });
  });

  // FlyTo the bounds of the currently-selected line (once its stops load).
  // Skipped while a route is selected so the route fly-to stays authoritative.
  let previousLineFlyKey: string | null = null;
  $effect(() => {
    const lineId = transit.selectedLineId;
    const routeId = transit.selectedRouteId;
    const pts = transit.linePathStops;
    const key = lineId && pts.length > 0 && routeId === null ? `${lineId}:${pts.length}` : null;
    if (key === previousLineFlyKey) return;
    previousLineFlyKey = key;
    if (!key) return;
    const latlngs: L.LatLngExpression[] = pts.map((p) => [p.lat, p.lng]);
    const bounds = L.latLngBounds(latlngs);
    map.flyToBounds(bounds, { padding: [60, 60], maxZoom: 15, duration: 0.6 });
  });

  // FlyTo the bounds of the currently-selected route.
  let previousRouteFlyKey: string | null = null;
  $effect(() => {
    const routeId = transit.selectedRouteId;
    const pts = transit.linePathStops;
    if (routeId === null) {
      previousRouteFlyKey = null;
      return;
    }
    const routePts = pts.filter((p) => p.routeId === routeId);
    const key = routePts.length > 0 ? `${routeId}:${routePts.length}` : null;
    if (key === previousRouteFlyKey) return;
    previousRouteFlyKey = key;
    if (!key) return;
    const latlngs: L.LatLngExpression[] = routePts.map((p) => [p.lat, p.lng]);
    const bounds = L.latLngBounds(latlngs);
    map.flyToBounds(bounds, { padding: [60, 60], maxZoom: 16, duration: 0.6 });
  });
</script>

<style>
  :global(.transit-edit-drag) {
    cursor: grab;
    background: none;
    border: none;
  }

  :global(.transit-edit-drag:active),
  :global(.leaflet-dragging .transit-edit-drag) {
    cursor: grabbing;
  }

  :global(.transit-nearby-link) {
    cursor: pointer;
  }
</style>
