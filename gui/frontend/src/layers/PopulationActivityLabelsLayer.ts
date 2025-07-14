import { TextLayer } from '@deck.gl/layers';

export function createActivityLabelsLayer(props: {
  id: string;
  data: any[];
}) {
  return new TextLayer({
    id: props.id,
    data: props.data,
    getPosition: d => d.location,
    getText: d => String(d.index + 1),
    getSize: 18,
    getColor: [255, 255, 255, 255],
    getTextAnchor: 'middle',
    getAlignmentBaseline: 'center',
    fontWeight: 700,
    fontFamily: 'Arial',
    billboard: false,
    sizeScale: 1,
    sizeMinPixels: 16,
    sizeMaxPixels: 24,
  });
}