import L from 'leaflet';

import type { LinePathPoint, LinkedRoutePath, ListedStop, NearbyNetworkLink } from '$lib/stores/transit.svelte';

export interface TransitLayers {
  stopsLayer: L.LayerGroup;
  markers: Map<string, { marker: L.Marker; signature: string }>;
  pathsLayer: L.LayerGroup;
  pathLines: Map<string, L.Polyline>;
  sequenceLayer: L.LayerGroup;
  sequenceMarkers: Map<string, L.Marker>;
  editLayer: L.LayerGroup;
}

/** User-facing mode tokens in outer -> inner ring priority. */
type UiMode = 'metro' | 'tram' | 'bus';
const MODE_PRIORITY: UiMode[] = ['metro', 'tram', 'bus'];

const MODE_COLORS: Record<UiMode, string> = {
  bus: '#2563eb',
  metro: '#dc2626',
  tram: '#10b981',
};

/** lucide icon paths (24x24 viewBox) - inlined to avoid component overhead per marker. */
const ICON_PATHS: Record<UiMode, string> = {
  bus: '<path d="M8 6v6"/><path d="M15 6v6"/><path d="M2 12h19.6"/><path d="M18 18h3s.5-1.7.8-2.8c.1-.4.2-.8.2-1.2 0-.4-.1-.8-.2-1.2l-1.4-5C20.1 6.8 19.1 6 18 6H4a2 2 0 0 0-2 2v10h3"/><circle cx="7" cy="18" r="2"/><circle cx="16" cy="18" r="2"/>',
  metro: '<path d="M8 3.1V7a4 4 0 0 0 8 0V3.1"/><path d="m9 15-1-1"/><path d="m15 15 1-1"/><path d="M9 19c-2.8 0-5-2.2-5-5v-4a8 8 0 0 1 16 0v4c0 2.8-2.2 5-5 5Z"/><path d="m8 19-2 3"/><path d="m16 19 2 3"/>',
  tram: '<rect x="4" y="3" width="16" height="16" rx="2"/><path d="M4 11h16"/><path d="m8 19-2 3"/><path d="m18 22-2-3"/><path d="M8 15h0"/><path d="M16 15h0"/>',
};

const MODE_ALIAS_TO_UI: Record<string, UiMode> = {
  bus: 'bus',
  metro: 'metro',
  subway: 'metro',
  rail: 'metro',
  tram: 'tram',
  light_rail: 'tram',
  lightrail: 'tram',
  streetcar: 'tram',
};

/** Zoom-responsive radius. */
function radiusForZoom(zoom: number): number {
  if (zoom <= 11) return 4;
  if (zoom <= 13) return 6;
  if (zoom <= 15) return 9;
  return 11;
}

function iconVisibleForZoom(zoom: number): boolean {
  return zoom >= 14;
}

/** Resolve raw mode strings to our three-mode UI palette, deduped, ordered outer->inner. */
function resolveUiModes(raw: string[]): UiMode[] {
  const present = new Set<UiMode>();
  for (const m of raw) {
    const ui = MODE_ALIAS_TO_UI[m.toLowerCase()];
    if (ui) present.add(ui);
  }
  return MODE_PRIORITY.filter((m) => present.has(m));
}

function buildStopIcon(
  modes: string[],
  sourceColor: string,
  sourceOpacity: number,
  zoom: number,
): { icon: L.DivIcon; signature: string } {
  const radius = radiusForZoom(zoom);
  const ui = resolveUiModes(modes);
  const size = radius * 2 + 4;
  const cx = size / 2;
  const cy = size / 2;

  const showIcon = iconVisibleForZoom(zoom) && ui.length > 0;
  const signature = `${ui.join('|')}|${radius}|${showIcon ? 'i' : '-'}|${sourceColor}|${sourceOpacity}`;

  let svg = `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">`;

  if (ui.length === 0) {
    // Fallback: filled with source color.
    svg += `<circle cx="${cx}" cy="${cy}" r="${radius}" fill="${sourceColor}" fill-opacity="${sourceOpacity}" stroke="#ffffff" stroke-width="1"/>`;
  } else {
    // Concentric rings: outer -> inner by UI priority.
    // Outer ring uses white stroke for contrast; inner rings stack over it.
    const n = ui.length;
    ui.forEach((mode, i) => {
      const r = radius * (1 - i / (n + 0.5));
      const isOuter = i === 0;
      svg += `<circle cx="${cx}" cy="${cy}" r="${r}" fill="${MODE_COLORS[mode]}"`;
      if (isOuter) {
        svg += ` stroke="#ffffff" stroke-width="1"`;
      }
      svg += `/>`;
    });

    // Subtle halo in source color so multiple transit sources can still be distinguished.
    if (sourceColor && sourceOpacity > 0) {
      svg += `<circle cx="${cx}" cy="${cy}" r="${radius + 1}" fill="none" stroke="${sourceColor}" stroke-width="1" stroke-opacity="${Math.min(1, sourceOpacity * 0.6)}"/>`;
    }
  }

  if (showIcon) {
    const primary = ui[0];
    const path = ICON_PATHS[primary];
    const iconSize = radius * 1.2;
    const iconOffset = cx - iconSize / 2;
    const scale = iconSize / 24;
    svg += `<g transform="translate(${iconOffset} ${iconOffset}) scale(${scale})" fill="none" stroke="#ffffff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">${path}</g>`;
  }

  svg += `</svg>`;

  const icon = L.divIcon({
    html: svg,
    className: 'transit-stop-marker',
    iconSize: [size, size],
    iconAnchor: [cx, cy],
  });

  return { icon, signature };
}

export function ensureTransitLayers(map: L.Map): TransitLayers {
  const pathsLayer = L.layerGroup().addTo(map);
  const sequenceLayer = L.layerGroup().addTo(map);
  const stopsLayer = L.layerGroup().addTo(map);
  const editLayer = L.layerGroup().addTo(map);
  return {
    stopsLayer,
    markers: new Map(),
    pathsLayer,
    pathLines: new Map(),
    sequenceLayer,
    sequenceMarkers: new Map(),
    editLayer,
  };
}

export function clearTransitLayers(layers: TransitLayers) {
  clearTransitStops(layers);
  clearLinePaths(layers);
  clearRouteSequence(layers);
  layers.editLayer.clearLayers();
}

/** Render nearby network links as clickable polylines. The one matching
 * `currentLinkRefId` (if any) renders in primary color to stand out from the
 * picker candidates. Click forwards the id to the callback. */
export function renderNearbyLinks(
  layer: L.LayerGroup,
  existing: L.Polyline[],
  items: NearbyNetworkLink[],
  currentLinkRefId: string | null,
  onClick: (linkId: string) => void,
): L.Polyline[] {
  for (const p of existing) layer.removeLayer(p);
  const next: L.Polyline[] = [];
  for (const link of items) {
    const isCurrent = currentLinkRefId !== null && link.id === currentLinkRefId;
    const baseColor = isCurrent ? '#2563eb' : '#fbbf24';
    const baseWeight = isCurrent ? 8 : 6;
    const hoverColor = isCurrent ? '#1d4ed8' : '#ef4444';
    const hoverWeight = isCurrent ? 10 : 8;
    const line = L.polyline(
      [
        [link.fromLat, link.fromLng],
        [link.toLat, link.toLng],
      ],
      {
        color: baseColor,
        weight: baseWeight,
        opacity: 0.9,
        interactive: true,
        className: 'transit-nearby-link',
      },
    );
    line.bindTooltip(link.id, { direction: 'top', sticky: true });
    line.on('click', () => onClick(link.id));
    line.on('mouseover', () => line.setStyle({ color: hoverColor, weight: hoverWeight }));
    line.on('mouseout', () => line.setStyle({ color: baseColor, weight: baseWeight }));
    line.addTo(layer);
    next.push(line);
  }
  return next;
}

export function clearNearbyLinks(layer: L.LayerGroup, existing: L.Polyline[]): L.Polyline[] {
  for (const p of existing) layer.removeLayer(p);
  return [];
}

export function clearTransitStops(layers: TransitLayers) {
  layers.stopsLayer.clearLayers();
  layers.markers.clear();
}

export function clearLinePaths(layers: TransitLayers) {
  layers.pathsLayer.clearLayers();
  layers.pathLines.clear();
}

export function clearRouteSequence(layers: TransitLayers) {
  layers.sequenceLayer.clearLayers();
  layers.sequenceMarkers.clear();
}

/** 10-color palette for per-route distinction within a single line. */
const ROUTE_COLORS = [
  '#2563eb',
  '#dc2626',
  '#10b981',
  '#f59e0b',
  '#7c3aed',
  '#db2777',
  '#0891b2',
  '#84cc16',
  '#f97316',
  '#6366f1',
];

/**
 * Renders one polyline per route on the line. Coordinates source per route:
 *   linkedRoutePath (if it applies to that route) -> network-linked geometry,
 *   else profile-stop straight-line points.
 * One layer, one polyline per route, updated in place via setLatLngs when data changes.
 */
export function renderLinePaths(
  layers: TransitLayers,
  points: LinePathPoint[],
  selectedRouteId: string | null,
  linkedRoutePath: LinkedRoutePath | null,
) {
  const byRoute = new Map<string, LinePathPoint[]>();
  for (const p of points) {
    let list = byRoute.get(p.routeId);
    if (!list) {
      list = [];
      byRoute.set(p.routeId, list);
    }
    list.push(p);
  }

  const present = new Set<string>();
  let routeIndex = 0;
  for (const [routeId, pts] of byRoute) {
    const color = ROUTE_COLORS[routeIndex % ROUTE_COLORS.length];
    routeIndex += 1;
    if (selectedRouteId !== null && routeId !== selectedRouteId) continue;
    present.add(routeId);

    let latlngs: L.LatLngExpression[];
    if (linkedRoutePath !== null && linkedRoutePath.routeId === routeId) {
      latlngs = linkedRoutePath.points.map((p) => [p.lat, p.lng]);
    } else {
      pts.sort((a, b) => a.sequence - b.sequence);
      latlngs = pts.map((p) => [p.lat, p.lng]);
    }
    const weight = selectedRouteId !== null ? 4 : 3;

    const existing = layers.pathLines.get(routeId);
    if (existing) {
      existing.setLatLngs(latlngs);
      existing.setStyle({ color, weight });
    } else {
      const line = L.polyline(latlngs, {
        color,
        weight,
        opacity: 0.85,
        interactive: false,
      });
      line.addTo(layers.pathsLayer);
      layers.pathLines.set(routeId, line);
    }
  }

  for (const [routeId, line] of layers.pathLines) {
    if (!present.has(routeId)) {
      layers.pathsLayer.removeLayer(line);
      layers.pathLines.delete(routeId);
    }
  }
}

/** Get the stable color for a given routeId within the set of all routes on the line. */
function routeColor(allPoints: LinePathPoint[], routeId: string): string {
  const seen = new Set<string>();
  let idx = 0;
  for (const p of allPoints) {
    if (seen.has(p.routeId)) continue;
    seen.add(p.routeId);
    if (p.routeId === routeId) return ROUTE_COLORS[idx % ROUTE_COLORS.length];
    idx += 1;
  }
  return ROUTE_COLORS[0];
}

function buildSequenceIcon(seq: number, color: string): L.DivIcon {
  const label = String(seq);
  const width = label.length <= 2 ? 20 : 24;
  const height = 20;
  const svg =
    `<svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">` +
    `<rect x="0.5" y="0.5" width="${width - 1}" height="${height - 1}" rx="4" ry="4" fill="${color}" stroke="#ffffff" stroke-width="1"/>` +
    `<text x="${width / 2}" y="${height / 2 + 3.5}" font-family="system-ui,sans-serif" font-size="11" font-weight="700" fill="#ffffff" text-anchor="middle">${label}</text>` +
    `</svg>`;
  return L.divIcon({
    html: svg,
    className: 'transit-route-sequence-marker',
    iconSize: [width, height],
    iconAnchor: [width / 2, height / 2],
  });
}

export function renderRouteSequence(
  layers: TransitLayers,
  points: LinePathPoint[],
  routeId: string,
) {
  const color = routeColor(points, routeId);
  const routePts = points.filter((p) => p.routeId === routeId).slice().sort((a, b) => a.sequence - b.sequence);

  const present = new Set<string>();
  routePts.forEach((pt, i) => {
    const seq = i + 1;
    const key = `${pt.sequence}`;
    present.add(key);
    const icon = buildSequenceIcon(seq, color);
    const existing = layers.sequenceMarkers.get(key);
    if (existing) {
      existing.setLatLng([pt.lat, pt.lng]);
      existing.setIcon(icon);
    } else {
      const marker = L.marker([pt.lat, pt.lng], { icon, keyboard: false, interactive: false });
      marker.addTo(layers.sequenceLayer);
      layers.sequenceMarkers.set(key, marker);
    }
  });

  for (const [key, marker] of layers.sequenceMarkers) {
    if (!present.has(key)) {
      layers.sequenceLayer.removeLayer(marker);
      layers.sequenceMarkers.delete(key);
    }
  }
}

export function renderTransitStops(
  layers: TransitLayers,
  items: ListedStop[],
  sourceColor: string,
  sourceOpacity: number,
  zoom: number,
  onStopClick: (stopId: string) => void,
) {
  const present = new Set<string>();

  for (const stop of items) {
    present.add(stop.id);
    const { icon, signature } = buildStopIcon(stop.modes, sourceColor, sourceOpacity, zoom);
    const existing = layers.markers.get(stop.id);
    const tooltip = stop.name ?? stop.id;

    if (existing) {
      existing.marker.setLatLng([stop.lat, stop.lng]);
      if (existing.signature !== signature) {
        existing.marker.setIcon(icon);
        existing.signature = signature;
      }
      const existingTooltip = existing.marker.getTooltip();
      if (existingTooltip && existingTooltip.getContent() !== tooltip) {
        existing.marker.setTooltipContent(tooltip);
      }
    } else {
      const marker = L.marker([stop.lat, stop.lng], { icon, keyboard: false });
      marker.bindTooltip(tooltip, { direction: 'top', offset: [0, -8] });
      marker.on('click', () => onStopClick(stop.id));
      marker.addTo(layers.stopsLayer);
      layers.markers.set(stop.id, { marker, signature });
    }
  }

  for (const [id, entry] of layers.markers) {
    if (!present.has(id)) {
      layers.stopsLayer.removeLayer(entry.marker);
      layers.markers.delete(id);
    }
  }
}
