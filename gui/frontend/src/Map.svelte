<script lang="ts">
  import { MapLibre, DeckGlLayer, NavigationControl } from 'svelte-maplibre';
  import { ScatterplotLayer } from '@deck.gl/layers';
  import { createEditableLayer, DrawPolygonMode, DrawLineStringMode, ModifyMode } from '$layers/EditableLayer';
  import { createSelectionPolygonLayer } from '$layers/SelectionPolygonLayer';
  import type { Map } from 'maplibre-gl';
  import type { FeatureCollection } from 'geojson';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import PopulationMapLayer from './components/map/PopulationMapLayer.svelte';
  import { populationState } from '$lib/stores/population.svelte';
  import { ptState } from '$lib/stores/pt.svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { commandArgs } from '$lib/stores/app.svelte.ts';
  import PTMapInteraction from './components/map/PTMapInteraction.svelte';
  import PTMapLayer from './components/map/PTMapLayer.svelte';
  import { trackStopChange } from '$lib/utils/ptChangeTracking';
  
  let map: Map | undefined = $state();
  let ptMapInteraction: any = $state();
  
  // Listen for clear editable layer event  
  $effect(() => {
    if (typeof window !== 'undefined') {
      const handler = () => {
        console.log('Clearing editable layer');
        editableData = {
          type: 'FeatureCollection',
          features: []
        };
      };
      window.addEventListener('clearEditableLayer', handler);
      return () => window.removeEventListener('clearEditableLayer', handler);
    }
  });
  
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
  let { 
    isEditingEnabled = false,
    onPolygonComplete = null,
    selectionPolygon = null
  }: { 
    isEditingEnabled?: boolean;
    onPolygonComplete?: ((polygon: number[][]) => void) | null;
    selectionPolygon?: number[][] | null;
  } = $props();
  
  // Enable drawing when creating a new zone (for population mode)
  const isEditingEnabledForZone = $derived(populationState.isDrawingZone);
  
  // Combined editing state - either from props (network mode) or from population state
  const finalEditingEnabled = $derived(isEditingEnabled || isEditingEnabledForZone);
  
  // Handle map clicks for activity location selection
  $effect(() => {
    if (map && populationState.isSelectingActivityLocation) {
      const handleMapClick = (e: any) => {
        if (populationState.isSelectingActivityLocation && populationState.selectingActivityId) {
          const lngLat = e.lngLat;
          const coordinates: [number, number] = [lngLat.lng, lngLat.lat];
          
          // Update the activity location
          const person = populationState.persons.get(populationState.selectedPersonId!);
          if (person) {
            const updatedPerson = {
              ...person,
              plans: person.plans.map(plan => ({
                ...plan,
                activities: plan.activities.map(activity =>
                  activity.id === populationState.selectingActivityId
                    ? { ...activity, location: coordinates }
                    : activity
                )
              }))
            };
            
            populationState.persons.set(person.id, updatedPerson);
          }
          
          // Reset selection state
          populationState.isSelectingActivityLocation = false;
          populationState.selectingActivityId = null;
        }
      };
      
      map.on('click', handleMapClick);
      return () => map.off('click', handleMapClick);
    }
  });
  
  // Handle map clicks for PT stop location selection
  $effect(() => {
    if (map && ptState.isSelectingStopLocation) {
      const handlePTMapClick = (e: any) => {
        if (ptState.isSelectingStopLocation && ptState.selectingStopId) {
          const lngLat = e.lngLat;
          const coordinates: [number, number] = [lngLat.lng, lngLat.lat];
          
          // Update the stop location
          const stop = ptState.stops.get(ptState.selectingStopId);
          if (stop) {
            const updatedStop = {
              ...stop,
              location: coordinates,
              x: coordinates[0],
              y: coordinates[1]
            };
            
            // Update stops map to trigger reactivity
            const newStops = new Map(ptState.stops);
            newStops.set(stop.id, updatedStop);
            ptState.stops = newStops;
            
            // Track the stop update
            trackStopChange(updatedStop, 'update');
          }
          
          // Reset selection state
          ptState.isSelectingStopLocation = false;
          ptState.selectingStopId = null;
        }
      };
      
      const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === 'Escape' && ptState.isSelectingStopLocation) {
          ptState.isSelectingStopLocation = false;
          ptState.selectingStopId = null;
        }
      };
      
      map.on('click', handlePTMapClick);
      window.addEventListener('keydown', handleKeyDown);
      return () => {
        map.off('click', handlePTMapClick);
        window.removeEventListener('keydown', handleKeyDown);
      };
    }
  });
  
  // Change cursor when selecting location
  $effect(() => {
    if (map && (populationState.isSelectingActivityLocation || ptState.isSelectingStopLocation)) {
      map.getCanvas().style.cursor = 'crosshair';
      return () => {
        map.getCanvas().style.cursor = '';
      };
    }
  });
  
  // Convert selection polygon to deck.gl format
  let selectionPolygonData = $derived(
    selectionPolygon ? [{
      polygon: selectionPolygon,
      id: 'selection'
    }] : []
  );
  
  // Debug logging
  $effect(() => {
    console.log('=== Map State Update ===');
    console.log('isEditingEnabled:', isEditingEnabled);
    console.log('Current mode:', currentMode?.name);
    console.log('Features count:', editableData.features.length);
  });
  
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
      
      // Check if a new feature was added
      if (info.editType === 'addFeature' && info.updatedData.features.length > 0) {
        const newFeature = info.updatedData.features[info.updatedData.features.length - 1];
        
        // Handle zone creation for population mode
        if (populationState.isDrawingZone) {
          console.log('Zone creation - editType:', info.editType, 'features:', info.updatedData.features);
          if (newFeature) {
            console.log('New zone geometry:', newFeature.geometry);
            const timestamp = Date.now();
            const newZoneId = `zone_${timestamp}`;
            
            // Create new zone with geometry
            const newZone = {
              id: newZoneId,
              name: `Zone ${populationState.zones.length + 1}`,
              personCount: 0,
              geometry: newFeature.geometry
            };
            
            console.log('Creating zone:', newZone);
            
            // Add to zones array - force reactivity with assignment
            populationState.zones = [...populationState.zones, newZone];
            
            // Make new zone visible by default - force reactivity
            populationState.selectedZones = new SvelteSet([...populationState.selectedZones, newZoneId]);
            
            // Exit drawing mode
            populationState.isDrawingZone = false;
            
            // Don't clear the editable layer - let the zone show in both layers
          }
        }
        
        // Handle polygon completion for network mode
        if (newFeature.geometry.type === 'Polygon' && onPolygonComplete) {
          const coordinates = newFeature.geometry.coordinates[0];
          // Convert to simple coordinate array (remove last duplicate point)
          const polygon = coordinates.slice(0, -1).map((coord: any) => [coord[0], coord[1]]);
          onPolygonComplete(polygon);
          // Don't clear the polygon - let it remain visible
        }
      }
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
  interactive={true}
  cooperativeGestures={false}
>
  <NavigationControl position="bottom-left" />
  
  <!-- Editable GeoJSON Layer - rendered first so it's below everything else -->
  {#if finalEditingEnabled}
    <DeckGlLayer
      layer={createEditableLayer({
        id: 'editable-layer',
        data: editableData,
        mode: currentMode,
        selectedFeatureIndexes: selectedFeatureIndexes,
        onEdit: handleEdit
      })}
    />
  {/if}
  
  <!-- Polygon Layer for showing selection (below data layers) -->
  {#if selectionPolygonData.length > 0}
    <DeckGlLayer
      layer={createSelectionPolygonLayer({
        id: 'selection-polygon-layer',
        data: selectionPolygonData
      })}
    />
  {/if}
  
  <!-- ScatterplotLayer for data visualization -->
  {#if data.length > 0}
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
  {/if}
  
  <!-- Population layers -->
  {#if commandArgs.fileEditMode === 'POPULATION'}
    <PopulationMapLayer />
  {/if}
  
  <!-- PT layers -->
  {#if commandArgs.fileEditMode === 'PT'}
    <PTMapLayer />
  {/if}
  
  <!-- PT Map Interaction component (for handling add stop logic) -->
  {#if commandArgs.fileEditMode === 'PT' && map}
    <PTMapInteraction bind:this={ptMapInteraction} {map} />
  {/if}
</MapLibre>

<!-- Location Selection Indicator -->
{#if populationState.isSelectingActivityLocation}
  <div class="absolute top-4 left-1/2 -translate-x-1/2 bg-yellow-500 text-black px-4 py-2 rounded-lg shadow-lg z-20">
    <p class="text-sm font-medium">Click on the map to set activity location</p>
  </div>
{/if}

<!-- 
  Demo Controls - Uncomment to enable drawing controls
  Usage: Uncomment the block below and ensure isEditingEnabled prop is passed
  
  {#if finalEditingEnabled}
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