<script lang="ts">
  import { MapLibre, DeckGlLayer, NavigationControl } from 'svelte-maplibre';
  import { ScatterplotLayer } from '@deck.gl/layers';
  import type { Map } from 'maplibre-gl';
  import 'maplibre-gl/dist/maplibre-gl.css';
  
  let map: Map | undefined = $state();
  
  // Initial center point - Montreal
  let center = { lng: -73.7, lat: 45.55 };
  let zoom = 10.5;
  
  // Example data - will be replaced with actual data later
  let data = $state([]);
</script>

<MapLibre
  style="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
  class="w-full h-full"
  {center}
  {zoom}
  bind:map
>
  <NavigationControl position="bottom-left" />
  
  <DeckGlLayer
    type={ScatterplotLayer}
    {data}
    id="scatterplot-layer"
    pickable={true}
    opacity={0.8}
    stroked={true}
    filled={true}
    radiusScale={6}
    radiusMinPixels={1}
    radiusMaxPixels={100}
    lineWidthMinPixels={1}
    getPosition={d => d.coordinates}
    getRadius={d => Math.sqrt(d.exits)}
    getFillColor={d => [255, 140, 0]}
    getLineColor={d => [0, 0, 0]}
  />
</MapLibre>

<style>
  :global(.maplibregl-ctrl-bottom-left) {
    left: 50%;
    transform: translateX(-50%);
  }
  
  :global(.maplibregl-ctrl-bottom-left .maplibregl-ctrl) {
    display: flex;
    flex-direction: row;
  }
  
  :global(.maplibregl-ctrl-zoom-in),
  :global(.maplibregl-ctrl-zoom-out),
  :global(.maplibregl-ctrl-compass) {
    margin: 0;
    border-radius: 0;
  }
  
  :global(.maplibregl-ctrl-zoom-in) {
    border-radius: 4px 0 0 4px;
    border-right: 1px solid rgba(0,0,0,0.1);
  }
  
  :global(.maplibregl-ctrl-compass) {
    border-radius: 0 4px 4px 0;
  }
</style>