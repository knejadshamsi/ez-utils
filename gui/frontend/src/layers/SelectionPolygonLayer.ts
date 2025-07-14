import { PolygonLayer } from '@deck.gl/layers';

export function createSelectionPolygonLayer(props: {
  id: string;
  data: any[];
}) {
  return new PolygonLayer({
    id: props.id,
    data: props.data,
    pickable: false,
    stroked: true,
    filled: true,
    getFillColor: [0, 100, 255, 30],
    getLineColor: [0, 100, 255, 255],
    getLineWidth: 3,
    getPolygon: d => d.polygon,
  });
}