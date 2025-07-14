import { EditableGeoJsonLayer, DrawPolygonMode, DrawLineStringMode, ModifyMode } from '@deck.gl-community/editable-layers';
import type { FeatureCollection } from 'geojson';

export function createEditableLayer(props: {
  id: string;
  data: FeatureCollection;
  mode: any;
  selectedFeatureIndexes: number[];
  onEdit: (info: any) => void;
}) {
  return new EditableGeoJsonLayer({
    id: props.id,
    data: props.data,
    mode: props.mode,
    selectedFeatureIndexes: props.selectedFeatureIndexes,
    onEdit: props.onEdit,
    pickable: true,
    stroked: true,
    filled: true,
    getFillColor: [0, 100, 255, 50],
    getLineColor: [0, 100, 255, 255],
    getLineWidth: 4,
    getPointRadius: 8,
    getEditHandlePointColor: [255, 255, 0],
    getEditHandlePointRadius: 8,
    getTentativeLineColor: [0, 0, 255],
    getTentativeLineWidth: 3,
    getTentativeFillColor: [0, 0, 255, 80],
    autoHighlight: true,
    highlightColor: [255, 255, 0, 100],
    getCursor: () => 'crosshair',
  });
}

export { DrawPolygonMode, DrawLineStringMode, ModifyMode };