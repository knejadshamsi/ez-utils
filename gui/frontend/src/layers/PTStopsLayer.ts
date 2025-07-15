import type { LayerSpecification } from 'maplibre-gl';

export function createPTStopsLayer(): LayerSpecification {
  return {
    id: 'pt-stops-layer',
    type: 'circle',
    source: 'pt-stops',
    paint: {
      'circle-radius': 6,
      'circle-color': '#ffffff',
      'circle-stroke-color': '#333333',
      'circle-stroke-width': 2
    }
  };
}

export function createPTStopsSelectedLayer(): LayerSpecification {
  return {
    id: 'pt-stops-selected',
    type: 'circle',
    source: 'pt-stops',
    filter: ['==', ['get', 'id'], ''],
    paint: {
      'circle-radius': 8,
      'circle-color': '#3B82F6',
      'circle-stroke-color': '#ffffff',
      'circle-stroke-width': 3
    }
  };
}

export function createPTStopLocationSelectLayer(): LayerSpecification {
  return {
    id: 'pt-stop-location-select',
    type: 'circle',
    source: 'pt-stops',
    filter: ['==', ['get', 'id'], ''],
    paint: {
      'circle-radius': 10,
      'circle-color': '#FFFF00', // Yellow like population activity selection
      'circle-stroke-color': '#333333',
      'circle-stroke-width': 2
    }
  };
}