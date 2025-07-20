<script lang="ts">
  import { onMount } from 'svelte';
  import L from 'leaflet';
  import 'leaflet/dist/leaflet.css';
  import 'leaflet-edgebuffer';
  import { initializePolygonDrawing } from './polygonDrawing';
  import { mapState, setupCentralEventHandlers } from './mapState.svelte';
  
  let mapContainer: HTMLDivElement;
  
  onMount(() => {
    // Fix leaflet default markers
    delete L.Icon.Default.prototype._getIconUrl;
    L.Icon.Default.mergeOptions({
      iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
      iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
      shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
    });
    
    // Initialize map
    const map = L.map(mapContainer, {
      preferCanvas: true,
      renderer: L.canvas(),
      zoomControl: false  // Disable zoom controls
    }).setView([45.55, -73.7], 11);
    
    // Store map instance
    mapState.map = map;
    
    // Add tile layer
    L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
      attribution: '© OpenStreetMap contributors © CARTO',
      edgeBufferTiles: 3,  // Preload 3 extra rows/columns of tiles beyond viewport
      updateWhenZooming: false,  // Only update tiles after zoom completes
      updateWhenIdle: true,  // Only update tiles after movement stops
      fadeAnimation: false  // No fade-in animation for new tiles
    }).addTo(map);
    
    // Create a custom pane for polygons with higher z-index to prevent flickering
    map.createPane('polygonPane');
    map.getPane('polygonPane').style.zIndex = '450';
    
    // Initialize polygon drawing
    initializePolygonDrawing(map);
    
    // Initialize layer groups
    mapState.connectedDotsLayer = L.featureGroup().addTo(map);
    mapState.networkLayer = L.featureGroup().addTo(map);
    
    // Create custom pane for connected dots
    map.createPane('connectedDotsPane');
    map.getPane('connectedDotsPane').style.zIndex = '460';
    
    // Setup central event handlers
    setupCentralEventHandlers();
    
    console.log('Leaflet map initialized');
    
    return () => {
      // Cleanup on unmount
      map.remove();
    };
  });
</script>

<div bind:this={mapContainer} class="map"></div>

<style>
  .map {
    width: 100%;
    height: 100%;
  }
  
  /* Make Leaflet.Draw edit markers circular */
  :global(.leaflet-edit-marker-selected) {
    border-radius: 50% !important;
    background-color: #3388ff !important;
    opacity: 1 !important;
  }
  
  /* Style the edit vertices */
  :global(.leaflet-marker-icon) {
    border-radius: 50% !important;
  }
  
  /* Override the square shape of edit markers */
  :global(.leaflet-editing-icon) {
    border-radius: 50% !important;
    width: 12px !important;
    height: 12px !important;
    border: 2px solid #fff !important;
    background-color: #3388ff !important;
    box-shadow: 0 2px 4px rgba(0,0,0,0.4) !important;
  }
  
  /* Style for connected dots markers */
  :global(.custom-circle-icon .circle-marker) {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 14px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  }
  
  /* Square marker for metro */
  :global(.custom-circle-icon .square-marker) {
    width: 30px;
    height: 30px;
    border-radius: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 14px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  }
  
  /* Rectangle marker for tram */
  :global(.custom-circle-icon .rectangle-marker) {
    width: 40px;
    height: 30px;
    border-radius: 20%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 14px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  }
  
  /* Square network node */
  :global(.network-node-icon .network-node-square) {
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 12px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
    cursor: pointer;
    white-space: nowrap;
    padding: 0 8px;
  }
</style>