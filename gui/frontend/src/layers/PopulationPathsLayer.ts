import { PathLayer } from '@deck.gl/layers';

export function createBackgroundPathsLayer(props: {
  id: string;
  data: any[];
}) {
  return new PathLayer({
    id: props.id,
    data: props.data,
    getPath: d => d.path,
    getColor: d => d.color,
    getWidth: 1,
    widthMinPixels: 1,
    pickable: false,
  });
}

export function createSelectedPersonPathLayer(props: {
  id: string;
  data: any[];
}) {
  return new PathLayer({
    id: props.id,
    data: props.data,
    getPath: d => d.path,
    getColor: [0, 100, 255, 255],
    getWidth: 5,
    widthMinPixels: 5,
    pickable: false,
  });
}