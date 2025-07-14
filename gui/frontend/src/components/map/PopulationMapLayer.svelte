<script lang="ts">
  import { DeckGlLayer } from 'svelte-maplibre';
  import { PolygonLayer, ScatterplotLayer, PathLayer, TextLayer } from '@deck.gl/layers';
  import { populationState } from '$lib/stores/population.svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { appState } from '$lib/stores/app.svelte.ts';

  // Function to trim path segments by radius from start/end points  
  function trimPathSegments(coordinates: [number, number][], trimRadiusPixels: number): Array<{path: [number, number][]}> {
    if (coordinates.length < 2) return [];
    
    const segments: Array<{path: [number, number][]}>  = [];
    
    for (let i = 0; i < coordinates.length - 1; i++) {
      const start = coordinates[i];
      const end = coordinates[i + 1];
      
      // Calculate the distance between points in geographic coordinates
      const dx = end[0] - start[0];
      const dy = end[1] - start[1];
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      // Convert pixel radius to geographic coordinates (rough approximation)
      // At zoom level ~11 (Montreal scale), 1 pixel ≈ 0.0001 degrees
      // This is an approximation that works reasonably well for mid-latitude locations
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

  // Prepare zones data
  const zonesData = $derived(
    populationState.zones.filter(zone => zone.geometry).map(zone => ({
      ...zone,
      polygon: zone.geometry.type === 'Polygon' ? zone.geometry.coordinates[0] : zone.geometry.coordinates[0][0]
    }))
  );
</script>

<!-- Zones Layer -->
{#if zonesData.length > 0}
  <DeckGlLayer
    type={PolygonLayer}
    id="zones-layer"
    data={zonesData}
    getPolygon={d => d.polygon}
    getFillColor={d => populationState.selectedZones.has(d.id) ? [0, 100, 255, 100] : [100, 100, 100, 30]}
    getLineColor={d => populationState.selectedZones.has(d.id) ? [0, 100, 255] : [100, 100, 100]}
    getLineWidth={2}
    lineWidthMinPixels={2}
    filled={true}
    stroked={true}
    pickable={true}
    onClick={(info) => {
      if (info.object) {
        if (populationState.selectedZones.has(info.object.id)) {
          populationState.selectedZones.delete(info.object.id);
        } else {
          populationState.selectedZones.add(info.object.id);
        }
        populationState.selectedZones = new SvelteSet(populationState.selectedZones);
      }
    }}
  />
{/if}

<!-- All Persons Layer -->
{#if populationState.visibility.persons && !populationState.selectedPersonId}
  {@const visiblePersons = Array.from(populationState.persons.values())
    .filter(person => populationState.visiblePersons.has(person.id))}
  {@const personsData = visiblePersons
    .map(person => {
      const firstActivity = person.plans[0]?.activities[0];
      return {
        id: person.id,
        position: firstActivity?.location || [0, 0]
      };
    })
    .filter(p => p.position[0] !== 0 || p.position[1] !== 0)}
  
  {#if personsData.length > 0}
    <DeckGlLayer
      type={ScatterplotLayer}
      id="all-persons-layer"
      data={personsData}
      getPosition={d => d.position}
      getFillColor={[100, 200, 100, 200]}
      getRadius={80}
      radiusMinPixels={5}
      radiusMaxPixels={8}
      pickable={false}
      stroked={true}
      getLineColor={[255, 255, 255]}
      lineWidthMinPixels={1}
    />
  {/if}
{/if}

<!-- Background Plans Layer -->
{#if populationState.visibility.plans}
  {@const backgroundPaths = []}
  {#each Array.from(populationState.persons.values()).filter(person => 
    populationState.visiblePersons.has(person.id) && person.id !== populationState.selectedPersonId
  ) as person}
    {#each person.plans as plan}
      {#if plan.activities.length > 1}
        {@const validActivities = plan.activities.filter(a => a.location && (a.location[0] !== 0 || a.location[1] !== 0))}
        {#if validActivities.length > 1}
          {@const path = validActivities.map(a => a.location)}
          {@const _ = backgroundPaths.push({ path, color: [200, 200, 200, 50] })}
        {/if}
      {/if}
    {/each}
  {/each}
  
  {#if backgroundPaths.length > 0}
    <DeckGlLayer
      type={PathLayer}
      id="background-plans-layer"
      data={backgroundPaths}
      getPath={d => d.path}
      getColor={d => d.color}
      getWidth={1}
      widthMinPixels={1}
      pickable={false}
    />
  {/if}
{/if}

<!-- Selected Person Layers -->
{#if populationState.selectedPersonId && populationState.visibility.plans}
  {@const selectedPerson = populationState.persons.get(populationState.selectedPersonId)}
  {#if selectedPerson}
    {@const currentPlan = selectedPerson.plans[populationState.currentPlanIndex]}
    {#if currentPlan}
      {@const validActivities = currentPlan.activities
        .map((activity, index) => ({ ...activity, index }))
        .filter(a => a.location && (a.location[0] !== 0 || a.location[1] !== 0))}
      
      <!-- Path Layer -->
      {#if validActivities.length > 1}
        {@const trimmedPaths = trimPathSegments(validActivities.map(a => a.location), 24)}
        <DeckGlLayer
          type={PathLayer}
          id="selected-plan-path"
          data={trimmedPaths}
          getPath={d => d.path}
          getColor={[0, 100, 255, 255]}
          getWidth={2}
          widthMinPixels={2}
          pickable={false}
        />
      {/if}
      
      <!-- Activity Dots -->
      {#if validActivities.length > 0}
        <DeckGlLayer
          type={ScatterplotLayer}
          id="selected-activities"
          data={validActivities}
          getPosition={d => d.location}
          getFillColor={d => 
            populationState.selectingActivityId === d.id 
              ? [255, 255, 0, 255] 
              : [0, 100, 255, 255]
          }
          getRadius={d => 
            populationState.selectingActivityId === d.id ? 250 : 200
          }
          radiusMinPixels={12}
          radiusMaxPixels={24}
          pickable={true}
          stroked={true}
          getLineColor={[255, 255, 255]}
          lineWidthMinPixels={3}
          onClick={() => {
            populationState.selectedPersonId = null;
            appState.secondarySidebar = 'HIDDEN';
          }}
        />
        
        <!-- Activity Numbers -->
        <DeckGlLayer
          type={TextLayer}
          id="activity-numbers"
          data={validActivities}
          getPosition={d => d.location}
          getText={d => String(d.index + 1)}
          getSize={18}
          getColor={[255, 255, 255, 255]}
          getTextAnchor={'middle'}
          getAlignmentBaseline={'center'}
          fontWeight={700}
          fontFamily={'Arial'}
          billboard={false}
          sizeScale={1}
          sizeMinPixels={16}
          sizeMaxPixels={24}
        />
      {/if}
    {/if}
  {/if}
{/if}