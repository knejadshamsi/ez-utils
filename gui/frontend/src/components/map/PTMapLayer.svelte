<script lang="ts">
  import { DeckGlLayer } from 'svelte-maplibre';
  import { PathLayer, ScatterplotLayer, TextLayer } from '@deck.gl/layers';
  import { ptState } from '$lib/stores/pt.svelte';
  import { appState } from '$lib/stores/app.svelte';

  // Function to trim path segments by radius from start/end points (same as population)
  function trimPathSegments(coordinates: [number, number][], trimRadiusPixels: number): Array<{path: [number, number][]}> {
    if (coordinates.length < 2) return [];
    
    const segments: Array<{path: [number, number][]}> = [];
    
    for (let i = 0; i < coordinates.length - 1; i++) {
      const start = coordinates[i];
      const end = coordinates[i + 1];
      
      // Calculate the distance between points in geographic coordinates
      const dx = end[0] - start[0];
      const dy = end[1] - start[1];
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      // Convert pixel radius to geographic coordinates (rough approximation)
      const trimRadiusInDegrees = trimRadiusPixels * 0.0001;
      
      // If the segment is too short to trim, skip it
      if (distance <= trimRadiusInDegrees * 2) continue;
      
      // Calculate unit vector
      const unitX = dx / distance;
      const unitY = dy / distance;
      
      // Calculate trimmed start and end points
      const trimmedStart: [number, number] = [
        start[0] + unitX * trimRadiusInDegrees,
        start[1] + unitY * trimRadiusInDegrees
      ];
      
      const trimmedEnd: [number, number] = [
        end[0] - unitX * trimRadiusInDegrees,
        end[1] - unitY * trimRadiusInDegrees
      ];
      
      segments.push({ path: [trimmedStart, trimmedEnd] });
    }
    
    return segments;
  }

  // Get all routes as paths for background layer
  const backgroundRoutes = $derived(() => {
    const routes: Array<{path: [number, number][], color: number[]}> = [];
    
    if (!ptState.visibility.routes) return routes;
    
    for (const line of ptState.lines.values()) {
      for (const route of line.routes) {
        // Skip selected route (will be drawn separately)
        if (route.id === ptState.selectedRouteId) continue;
        
        // Get stop coordinates for this route
        const stopCoords = route.stopSequence
          .map(stop => {
            const stopData = ptState.stops.get(stop.stopId);
            return stopData ? [stopData.x, stopData.y] as [number, number] : null;
          })
          .filter(coord => coord !== null && (coord[0] !== 0 || coord[1] !== 0)) as [number, number][];
        
        if (stopCoords.length > 1) {
          routes.push({
            path: stopCoords,
            color: [200, 200, 200, 50]
          });
        }
      }
    }
    
    return routes;
  });

  // Get selected route data
  const selectedRouteData = $derived(() => {
    if (!ptState.selectedRouteId) return null;
    
    const routeData = ptState.selectedRoute();
    if (!routeData) return null;
    
    const { route } = routeData;
    
    // Get stop data with coordinates
    const stopsWithCoords = route.stopSequence
      .map((stop, index) => {
        const stopData = ptState.stops.get(stop.stopId);
        if (!stopData) return null;
        
        return {
          ...stop,
          ...stopData,
          index,
          position: [stopData.x, stopData.y] as [number, number]
        };
      })
      .filter(stop => stop !== null && (stop.position[0] !== 0 || stop.position[1] !== 0));
    
    return {
      route,
      stops: stopsWithCoords
    };
  });
</script>

<!-- Background Routes Layer -->
{#if backgroundRoutes().length > 0}
  <DeckGlLayer
    type={PathLayer}
    id="pt-background-routes"
    data={backgroundRoutes()}
    getPath={d => d.path}
    getColor={d => d.color}
    getWidth={1}
    widthMinPixels={1}
    pickable={false}
  />
{/if}

<!-- Selected Route Layers -->
{#if selectedRouteData()}
  {@const data = selectedRouteData()}
  
  <!-- Selected Route Path -->
  {#if data.stops.length > 1}
    {@const positions = data.stops.map(s => s.position)}
    {@const trimmedPaths = trimPathSegments(positions, 24)}
    
    <DeckGlLayer
      type={PathLayer}
      id="pt-selected-route-path"
      data={trimmedPaths}
      getPath={d => d.path}
      getColor={[0, 100, 255, 255]}
      getWidth={3}
      widthMinPixels={3}
      pickable={false}
    />
  {/if}
  
  <!-- Stop Dots -->
  <DeckGlLayer
    type={ScatterplotLayer}
    id="pt-selected-stops"
    data={data.stops}
    getPosition={d => d.position}
    getFillColor={d => 
      ptState.selectingStopId === d.id 
        ? [255, 255, 0, 255]
        : [0, 100, 255, 255]
    }
    getRadius={d => 
      ptState.selectingStopId === d.id ? 300 : 250
    }
    radiusMinPixels={12}
    radiusMaxPixels={24}
    pickable={true}
    stroked={true}
    getLineColor={[255, 255, 255]}
    lineWidthMinPixels={3}
    onClick={(info) => {
      if (info.object && ptState.isSelectingStopLocation) {
        ptState.selectingStopId = info.object.id;
      }
    }}
  />
  
  <!-- Stop Numbers -->
  <DeckGlLayer
    type={TextLayer}
    id="pt-stop-numbers"
    data={data.stops}
    getPosition={d => d.position}
    getText={d => String(d.index + 1)}
    getSize={14}
    getColor={[255, 255, 255]}
    getTextAnchor="middle"
    getAlignmentBaseline="center"
    fontFamily="Arial"
    fontWeight={700}
    pickable={false}
  />
{/if}

<!-- All Stops Layer (when no route selected) -->
{#if !ptState.selectedRouteId && ptState.visibility.stops}
  {@const allStopsData = Array.from(ptState.stops.values())
    .filter(stop => stop.x !== 0 || stop.y !== 0)
    .map(stop => {
      const lines = ptState.getLinesForStop(stop.id);
      // Only include stops that belong to at least one line
      if (lines.length === 0) return null;
      
      const primaryLine = lines[0];
      const color = primaryLine.color;
      const abbreviation = ptState.getModeAbbreviation(primaryLine.mode as TransportMode);
      
      // Parse hex color to RGB
      const hex = color.replace('#', '');
      const r = parseInt(hex.substring(0, 2), 16);
      const g = parseInt(hex.substring(2, 4), 16);
      const b = parseInt(hex.substring(4, 6), 16);
      
      return {
        ...stop,
        lines,
        primaryLine,
        color: [r, g, b, 255],
        abbreviation
      };
    })
    .filter(stop => stop !== null)}
  
  <!-- Stop circles with line colors -->
  <DeckGlLayer
    type={ScatterplotLayer}
    id="pt-all-stops"
    data={allStopsData}
    getPosition={d => [d.x, d.y]}
    getFillColor={d => d.color}
    getRadius={150}
    radiusMinPixels={8}
    radiusMaxPixels={12}
    pickable={true}
    stroked={true}
    getLineColor={[255, 255, 255]}
    lineWidthMinPixels={2}
    opacity={0.8}
  />
  
  <!-- Stop emojis -->
  <DeckGlLayer
    type={TextLayer}
    id="pt-stop-icons"
    data={allStopsData}
    getPosition={d => [d.x, d.y]}
    getText={d => d.abbreviation}
    getSize={18}
    getColor={[255, 255, 255]}
    getTextAnchor="middle"
    getAlignmentBaseline="center"
    fontFamily="Arial"
    fontWeight={700}
    pickable={false}
  />
{/if}