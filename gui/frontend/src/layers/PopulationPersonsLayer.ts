import { ScatterplotLayer } from '@deck.gl/layers';

export function createPopulationPersonsLayer(props: {
  id: string;
  data: any[];
}) {
  return new ScatterplotLayer({
    id: props.id,
    data: props.data,
    getPosition: d => d.position,
    getFillColor: [100, 200, 100, 200],
    getRadius: 80,
    radiusMinPixels: 5,
    radiusMaxPixels: 8,
    pickable: false,
    stroked: true,
    getLineColor: [255, 255, 255],
    lineWidthMinPixels: 1,
  });
}

export function createSelectedPersonActivitiesLayer(props: {
  id: string;
  data: any[];
  selectingActivityId: string | null;
  onClick: () => void;
}) {
  return new ScatterplotLayer({
    id: props.id,
    data: props.data,
    getPosition: d => d.location,
    getFillColor: d => 
      props.selectingActivityId === d.id 
        ? [255, 255, 0, 255] 
        : [0, 100, 255, 255],
    getRadius: d => 
      props.selectingActivityId === d.id ? 250 : 200,
    radiusMinPixels: 12,
    radiusMaxPixels: 24,
    pickable: true,
    stroked: true,
    getLineColor: [255, 255, 255],
    lineWidthMinPixels: 3,
    onClick: props.onClick,
    updateTriggers: {
      getFillColor: props.selectingActivityId
    },
    // Ensure dots render on top
    parameters: {
      depthTest: false,
    }
  });
}