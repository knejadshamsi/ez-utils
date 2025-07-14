import type { LayerSpecification } from 'maplibre-gl';

export function createPTRoutesLayer(): LayerSpecification {
  return {
    id: 'pt-routes-layer',
    type: 'line',
    source: 'pt-routes',
    paint: {
      'line-color': ['get', 'color'],
      'line-width': 2,
      'line-opacity': 0.6
    }
  };
}

export function getPTRoutesLayerSelectedPaint(selectedRouteId: string | null) {
  return {
    'line-color': ['get', 'color'],
    'line-width': [
      'case',
      ['==', ['get', 'id'], selectedRouteId || ''],
      4,
      2
    ],
    'line-opacity': [
      'case',
      ['==', ['get', 'id'], selectedRouteId || ''],
      1,
      0.6
    ]
  };
}