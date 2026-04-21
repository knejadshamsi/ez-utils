import L from 'leaflet';

import type { EditablePlanDraft, PopulationQueryPoint } from '$lib/components/population/types';
import { population } from '$lib/stores/population.svelte';

export function renderPopulationPoints(populationLayer: L.LayerGroup, points: PopulationQueryPoint[], sourceColor: string) {
  populationLayer.clearLayers();
  points.forEach(point => {
    const marker = L.circleMarker([point.lat, point.lng], {
      radius: 5,
      color: sourceColor,
      weight: 1,
      fillColor: sourceColor,
      fillOpacity: 0.85,
    });
    marker.on('click', () => {
      void population.selectPerson({
        sourceName: point.sourceName,
        personId: point.personId,
      });
    });
    marker.addTo(populationLayer);
  });
}

export function renderSelectedPlan(
  selectedPlanLayer: L.LayerGroup,
  activePlan: EditablePlanDraft | null,
  editLocations: boolean,
  map: L.Map | null,
) {
  selectedPlanLayer.clearLayers();
  if (!activePlan) return;

  const zoom = map?.getZoom() ?? 12;
  const offsetFactor = Math.max(0, 1 - (zoom - 12) / 3);
  const offsetRadius = 20 * offsetFactor;

  const coordGroups = new Map<string, number[]>();
  activePlan.activities.forEach((activity, index) => {
    if (activity.lng == null || activity.lat == null) return;
    const key = `${activity.lng},${activity.lat}`;
    const group = coordGroups.get(key);
    if (group) {
      group.push(index);
    } else {
      coordGroups.set(key, [index]);
    }
  });

  const offsets = new Map<number, { dlat: number; dlng: number }>();
  if (map && offsetRadius > 0) {
    coordGroups.forEach((indices) => {
      if (indices.length < 2) return;
      indices.forEach((actIndex, posInGroup) => {
        if (posInGroup === 0) return;
        const activity = activePlan.activities[actIndex];
        if (activity.lng == null || activity.lat == null) return;
        const center = map.project([activity.lat, activity.lng], zoom);
        const angle = ((posInGroup - 1) / (indices.length - 1)) * 2 * Math.PI;
        const offsetPoint = L.point(center.x + Math.cos(angle) * offsetRadius, center.y + Math.sin(angle) * offsetRadius);
        const offsetLatLng = map.unproject(offsetPoint, zoom);
        offsets.set(actIndex, {
          dlat: offsetLatLng.lat - activity.lat,
          dlng: offsetLatLng.lng - activity.lng,
        });
      });
    });
  }

  const latLngs: L.LatLngExpression[] = [];
  const iconMarkers = new Map<number, L.Marker>();
  const visibleIndexByActivity = new Map<number, number>();
  activePlan.activities.forEach((activity, index) => {
    if (activity.lng == null || activity.lat == null) return;
    latLngs.push([activity.lat, activity.lng]);
    visibleIndexByActivity.set(index, latLngs.length - 1);
    const offset = offsets.get(index);
    const displayLat = activity.lat + (offset?.dlat ?? 0);
    const displayLng = activity.lng + (offset?.dlng ?? 0);
    const markerColor = activityColor(activity.type);
    const size = 22;
    const icon = L.divIcon({
      className: '',
      iconSize: [size, size],
      iconAnchor: [size / 2, size / 2],
      html: `<div style="width:${size}px;height:${size}px;border-radius:50%;background:${markerColor};border:2px solid rgba(255,255,255,0.9);display:flex;align-items:center;justify-content:center;color:#fff;font-size:11px;font-weight:700;line-height:1;">${index + 1}</div>`,
    });
    const marker = L.marker([displayLat, displayLng], { icon, interactive: false });
    marker.addTo(selectedPlanLayer);
    iconMarkers.set(index, marker);
  });

  const polyline = latLngs.length > 1
    ? L.polyline(latLngs, { color: '#10b981', weight: 2, opacity: 0.8 }).addTo(selectedPlanLayer)
    : null;

  if (editLocations) {
    activePlan.activities.forEach((activity, index) => {
      if (activity.lng == null || activity.lat == null) return;
      const draggable = L.marker([activity.lat, activity.lng], {
        draggable: true,
        opacity: 0,
      });
      draggable.on('drag', (event) => {
        const pos = event.target.getLatLng();
        iconMarkers.get(index)?.setLatLng(pos);
        if (polyline) {
          const visibleIndex = visibleIndexByActivity.get(index);
          if (visibleIndex !== undefined) {
            const lls = polyline.getLatLngs() as L.LatLng[];
            lls[visibleIndex] = pos;
            polyline.setLatLngs(lls);
          }
        }
      });
      draggable.on('dragend', (event) => {
        const next = event.target.getLatLng();
        population.moveActivity(index, next.lng, next.lat);
      });
      draggable.addTo(selectedPlanLayer);
    });
  }
}

export function activityColor(type: string) {
  switch (type) {
    case 'home': return '#10b981';
    case 'work': return '#3b82f6';
    case 'school': return '#8b5cf6';
    case 'shop': return '#f59e0b';
    case 'eat': return '#ef4444';
    case 'recreation': return '#ec4899';
    default: return '#6b7280';
  }
}
