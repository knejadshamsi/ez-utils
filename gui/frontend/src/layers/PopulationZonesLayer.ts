import { PolygonLayer } from '@deck.gl/layers';

export function createPopulationZonesLayer(props: {
  id: string;
  data: any[];
  selectedZones: Set<string>;
  onClick: (info: any) => void;
}) {
  return new PolygonLayer({
    id: props.id,
    data: props.data,
    getPolygon: d => d.polygon,
    getFillColor: d => props.selectedZones.has(d.id) ? [0, 100, 255, 100] : [100, 100, 100, 30],
    getLineColor: d => props.selectedZones.has(d.id) ? [0, 100, 255] : [100, 100, 100],
    getLineWidth: 2,
    lineWidthMinPixels: 2,
    filled: true,
    stroked: true,
    pickable: true,
    onClick: props.onClick,
    updateTriggers: {
      getFillColor: props.selectedZones.size,
      getLineColor: props.selectedZones.size
    }
  });
}