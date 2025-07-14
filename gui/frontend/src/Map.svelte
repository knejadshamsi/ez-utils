<script lang="ts">
  import { MapLibre, DeckGlLayer, NavigationControl } from 'svelte-maplibre';
  import { ScatterplotLayer } from '@deck.gl/layers';
  import { EditableGeoJsonLayer, DrawPolygonMode, DrawLineStringMode, ModifyMode } from '@deck.gl-community/editable-layers';
  import type { Map } from 'maplibre-gl';
  import type { FeatureCollection } from 'geojson';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import { commandArgs } from './store.svelte';
  import { PTMapInteraction } from './services/edit/pt';
  
  let map: Map | undefined = $state();
  let ptMapInteraction: any = $state();
  
  // Initial center point - Montreal
  let center = { lng: -73.7, lat: 45.55 };
  let zoom = 10.5;
  
  // Example data - will be replaced with actual data later
  let data = $state([]);
  
  // Editable layer state
  let editableData = $state<FeatureCollection>({
    type: 'FeatureCollection',
    features: []
  });
  
  let selectedFeatureIndexes = $state<number[]>([]);
  let currentMode = $state<any>(DrawPolygonMode);
  
  // Props for parent component access
  // Demo: Uncomment to enable drawing functionality
  // let { isEditingEnabled = false }: { isEditingEnabled: boolean } = $props();
  let isEditingEnabled = false; // Set to true to enable drawing, or make it a prop
  
  // Debug logging - uncomment for troubleshooting
  // $effect(() => {
  //   console.log('=== State Update ===');
  //   console.log('isEditingEnabled:', isEditingEnabled);
  //   console.log('Current mode:', currentMode?.name);
  //   console.log('Features count:', editableData.features.length);
  //   console.log('editableData:', $state.snapshot(editableData));
  // });
  
  // Debug map events - uncomment for troubleshooting
  // $effect(() => {
  //   if (map) {
  //     console.log('Map instance available:', map);
  //     console.log('Map loaded:', map.loaded());
  //     
  //     map.on('click', (e) => {
  //       console.log('Map click event:', e);
  //       if (isEditingEnabled) {
  //         console.log('Preventing map click event - editing is enabled');
  //         e.preventDefault();
  //       }
  //     });
  //     
  //     map.on('load', () => {
  //       console.log('Map loaded!');
  //     });
  //   }
  // });
  
  // Handle all edit events - update state immediately for better UX
  const handleEdit = (info: any) => {
    // Debug logging - uncomment for troubleshooting
    // console.log('Edit event received:', info);
    if (info && info.updatedData) {
      // console.log('Updating editableData with:', info.updatedData);
      editableData = info.updatedData;
    }
  };
  
  // Mode switching functions - for drawing controls
  const switchToDrawPolygon = () => {
    // console.log('Switching to DrawPolygonMode');
    currentMode = DrawPolygonMode;
    selectedFeatureIndexes = [];
  };
  
  const switchToDrawLine = () => {
    // console.log('Switching to DrawLineStringMode');
    currentMode = DrawLineStringMode;
    selectedFeatureIndexes = [];
  };
  
  const switchToEdit = () => {
    // console.log('Switching to ModifyMode');
    currentMode = ModifyMode;
  };
  
  const clearAll = () => {
    // console.log('Clearing all features');
    editableData = {
      type: 'FeatureCollection',
      features: []
    };
    selectedFeatureIndexes = [];
  };
  
  // Export function for PT map interaction
  export function startAddingStop() {
    if (ptMapInteraction) {
      ptMapInteraction.startAddingStop?.();
      console.log('[Map] Starting add stop mode');
    } else {
      console.warn('[Map] PTMapInteraction not available');
    }
  }
  
  // Add test feature - for testing layer functionality
  const addTestFeature = () => {
    // console.log('Adding test feature');
    editableData = {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          geometry: {
            type: 'Polygon',
            coordinates: [[
              [-73.7, 45.55],
              [-73.69, 45.55],
              [-73.69, 45.56],
              [-73.7, 45.56],
              [-73.7, 45.55]
            ]]
          },
          properties: {}
        }
      ]
    };
  };
</script>

<MapLibre
  style="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
  class="w-full h-full"
  {center}
  {zoom}
  bind:map
>
  <NavigationControl position="bottom-left" />
  
  <!-- Editable GeoJSON Layer - Must be on top to receive events -->
  {#if isEditingEnabled}
    <DeckGlLayer
      type={EditableGeoJsonLayer}
      id="editable-layer"
      data={editableData}
      mode={currentMode}
      selectedFeatureIndexes={selectedFeatureIndexes}
      onEdit={handleEdit}
      pickable={true}
      stroked={true}
      filled={true}
      getFillColor={[255, 0, 0, 100]}
      getLineColor={[255, 0, 0]}
      getLineWidth={3}
      getPointRadius={8}
      getEditHandlePointColor={[255, 255, 0]}
      getEditHandlePointRadius={8}
      getTentativeLineColor={[0, 0, 255]}
      getTentativeLineWidth={3}
      getTentativeFillColor={[0, 0, 255, 80]}
      autoHighlight={true}
      highlightColor={[255, 255, 0, 100]}
      onClick={(info) => {/* console.log('Layer clicked:', info) */}}
      onHover={(info) => {/* console.log('Layer hover:', info) */}}
      getCursor={() => 'crosshair'}
      interleaved={true}
    />
  {/if}
  
  <!-- ScatterplotLayer for data visualization -->
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
  
  <!-- PT Map Interaction component -->
  {#if commandArgs.fileEditMode === 'PT' && map}
    <PTMapInteraction bind:this={ptMapInteraction} {map} />
  {/if}
</MapLibre>

<!-- 
  Demo Controls - Uncomment to enable drawing controls
  Usage: Uncomment the block below and ensure isEditingEnabled prop is passed
  
  {#if isEditingEnabled}
    <div class="absolute top-20 left-4 z-10 bg-white p-4 rounded shadow-lg">
      <div class="flex flex-col gap-2">
        <button 
          class="px-3 py-1 bg-blue-500 text-white rounded text-sm hover:bg-blue-600"
          onclick={switchToDrawPolygon}
        >
          Draw Polygon
        </button>
        <button 
          class="px-3 py-1 bg-green-500 text-white rounded text-sm hover:bg-green-600"
          onclick={switchToDrawLine}
        >
          Draw Line
        </button>
        <button 
          class="px-3 py-1 bg-yellow-500 text-white rounded text-sm hover:bg-yellow-600"
          onclick={switchToEdit}
        >
          Edit Mode
        </button>
        <button 
          class="px-3 py-1 bg-red-500 text-white rounded text-sm hover:bg-red-600"
          onclick={clearAll}
        >
          Clear All
        </button>
        <button 
          class="px-3 py-1 bg-purple-500 text-white rounded text-sm hover:bg-purple-600"
          onclick={addTestFeature}
        >
          Add Test
        </button>
      </div>
      <div class="mt-2 text-sm text-gray-600">
        Features: {editableData.features.length}<br>
        Mode: {currentMode?.name || 'Unknown'}<br>
        Editing: {isEditingEnabled ? 'ON' : 'OFF'}
      </div>
    </div>
  {/if}
-->

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