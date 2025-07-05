import React, { useState } from 'react';
import Map from 'react-map-gl/maplibre';
import DeckGL from '@deck.gl/react';
import { ScatterplotLayer, PathLayer } from '@deck.gl/layers';
import useAppStore from '../store/appStore';
import 'maplibre-gl/dist/maplibre-gl.css';

const MapView = () => {
  const { 
    visiblePersons, 
    persons, 
    selectedPlanRoute,
    mapClickHandler
  } = useAppStore();
  
  const [viewState, setViewState] = useState({
    longitude: -73.7,
    latitude: 45.5,
    zoom: 10
  });

  const layers = [
    new ScatterplotLayer({
      id: 'persons',
      data: persons.filter(p => visiblePersons.has(p.id)),
      getPosition: d => {
        if (!d.coords || typeof d.coords !== 'string') return [0, 0];
        const parts = d.coords.split(',');
        if (parts.length !== 2) return [0, 0];
        const [x, y] = parts.map(parseFloat);
        return [isNaN(x) || isNaN(y) ? 0 : x, isNaN(y) ? 0 : y];
      },
      getRadius: 100,
      getFillColor: [0, 128, 255],
      pickable: true,
      onClick: info => {
        if (info.object && !mapClickMode) {
          useAppStore.getState().setSelectedPerson(info.object);
        }
      }
    }),
    selectedPlanRoute && new PathLayer({
      id: 'route',
      data: [selectedPlanRoute],
      getPath: d => d.path,
      getWidth: 3,
      getColor: [0, 128, 255]
    })
  ].filter(Boolean);

  const handleClick = (event) => {
    if (mapClickHandler && event.coordinate) {
      const [lng, lat] = event.coordinate;
      mapClickHandler({ lng, lat });
    }
  };

  return (
    <DeckGL
      viewState={viewState}
      onViewStateChange={({viewState}) => setViewState(viewState)}
      controller={true}
      layers={layers}
      onClick={handleClick}
    >
      <Map
        mapStyle="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
      />
    </DeckGL>
  );
};

export default MapView;