import { EditableGeoJsonLayer, DrawPointMode, DrawLineStringMode, DrawPolygonMode, ModifyMode } from '@deck.gl-community/editable-layers';
import { PathLayer, ScatterplotLayer } from '@deck.gl/layers';
import type { FeatureCollection, Feature, Point, LineString, Polygon } from 'geojson';
import type { FileEditMode } from '../store.svelte';

export type DrawingMode = 'POINT' | 'LINE' | 'POLYGON' | 'MODIFY' | 'VIEW';

export interface LayerConfig {
  fileEditMode: FileEditMode;
  drawingMode: DrawingMode;
  data: FeatureCollection;
  selectedFeatureIndexes: number[];
  onEdit: (info: any) => void;
}

export function getEditModeClass(drawingMode: DrawingMode) {
  switch (drawingMode) {
    case 'POINT':
      return DrawPointMode;
    case 'LINE':
      return DrawLineStringMode;
    case 'POLYGON':
      return DrawPolygonMode;
    case 'MODIFY':
      return ModifyMode;
    case 'VIEW':
    default:
      return null;
  }
}

export function createEditableLayer(config: LayerConfig) {
  const { fileEditMode, drawingMode, data, selectedFeatureIndexes, onEdit } = config;
  const modeClass = getEditModeClass(drawingMode);
  
  if (!modeClass) {
    return null;
  }

  return new EditableGeoJsonLayer({
    id: `editable-${fileEditMode.toLowerCase()}-layer`,
    data,
    mode: modeClass,
    selectedFeatureIndexes,
    onEdit,
    pickable: true,
    autoHighlight: true,
    highlightColor: [255, 255, 0, 100],
    getTentativeLineColor: [0, 0, 255, 200],
    getTentativeLineWidth: 3,
    getTentativeFillColor: [0, 0, 255, 50],
    getEditHandlePointColor: [255, 255, 0],
    getEditHandlePointRadius: 8,
    _subLayerProps: {
      geojson: {
        getFillColor: (feature: Feature) => {
          const type = feature.properties?.entityType;
          return getFeatureFillColor(fileEditMode, type);
        },
        getLineColor: (feature: Feature) => {
          const type = feature.properties?.entityType;
          return getFeatureLineColor(fileEditMode, type);
        },
        getPointRadius: (feature: Feature) => {
          const type = feature.properties?.entityType;
          return getFeatureRadius(fileEditMode, type);
        },
        getLineWidth: 3
      }
    }
  });
}

function getFeatureFillColor(fileEditMode: FileEditMode, entityType?: string): [number, number, number, number] {
  if (fileEditMode === 'POPULATION') {
    return entityType === 'person' ? [255, 0, 0, 180] : [255, 0, 0, 50];
  } else if (fileEditMode === 'NETWORK') {
    return entityType === 'node' ? [0, 255, 0, 180] : [0, 255, 0, 50];
  } else if (fileEditMode === 'PT') {
    return entityType === 'stop' ? [0, 0, 255, 180] : [0, 0, 255, 50];
  }
  return [128, 128, 128, 100];
}

function getFeatureLineColor(fileEditMode: FileEditMode, entityType?: string): [number, number, number, number] {
  if (fileEditMode === 'POPULATION') {
    return [255, 0, 0, 255];
  } else if (fileEditMode === 'NETWORK') {
    return [0, 255, 0, 255];
  } else if (fileEditMode === 'PT') {
    return [0, 0, 255, 255];
  }
  return [128, 128, 128, 255];
}

function getFeatureRadius(fileEditMode: FileEditMode, entityType?: string): number {
  if (entityType === 'person' || entityType === 'node' || entityType === 'stop') {
    return 8;
  }
  return 5;
}