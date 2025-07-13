<script lang="ts">
  import { DeckGlLayer } from 'svelte-maplibre';
  import { ScatterplotLayer, PathLayer, PolygonLayer, TextLayer } from '@deck.gl/layers';
  import { populationState } from './populationStore.svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { appState } from '../../../store.svelte';
  
  // Debug zones
  $effect(() => {
    console.log('PopulationMapLayer - zones updated:', populationState.zones.length, 
      populationState.zones.map(z => ({ id: z.id, hasGeometry: !!z.geometry })));
    
    const zonesWithGeometry = populationState.zones.filter(zone => zone.geometry);
    console.log('Zones with geometry:', zonesWithGeometry);
    
    if (zonesWithGeometry.length > 0) {
      const firstZone = zonesWithGeometry[0];
      console.log('First zone polygon data:', {
        id: firstZone.id,
        geometryType: firstZone.geometry.type,
        coordinates: firstZone.geometry.coordinates
      });
    }
  });
  
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
      // Force update
      populationState.selectedZones = new SvelteSet(populationState.selectedZones);
    }
  }}
  updateTriggers={{
    data: populationState.zones.length,
    getFillColor: populationState.selectedZones.size,
    getLineColor: populationState.selectedZones.size
  }}
  />
{/if}

<!-- All Persons Layer - Small circles for all visible persons (hidden when someone is selected) -->
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
      updateTriggers={{
        data: populationState.visiblePersons.size
      }}
    />
  {/if}
{/if}

<!-- Background Plans Layer - Show other people's plans subtly -->
{#if populationState.visibility.plans}
  {@const backgroundPaths = []}
  {#each Array.from(populationState.persons.values()).filter(person => 
    populationState.visiblePersons.has(person.id) && person.id !== populationState.selectedPersonId
  ) as person}
    {#each person.plans as plan}
      {#if plan.activities.length > 1}
        {@const validActivities = plan.activities.filter(a => a.location[0] !== 0 || a.location[1] !== 0)}
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

<!-- Selected Person's Plan - Bold visualization -->
{#if populationState.selectedPersonId && populationState.visibility.plans}
  {@const selectedPerson = populationState.persons.get(populationState.selectedPersonId)}
  {#if selectedPerson}
    {@const currentPlan = selectedPerson.plans[populationState.currentPlanIndex]}
    {#if currentPlan}
      {@const validActivities = currentPlan.activities
        .map((activity, index) => ({ ...activity, index }))
        .filter(a => a.location[0] !== 0 || a.location[1] !== 0)}
      
      <!-- Plan Path - Bold line (rendered first so it's below dots) -->
      {#if validActivities.length > 1}
        <DeckGlLayer
          type={PathLayer}
          id="selected-plan-path"
          data={[{ path: validActivities.map(a => a.location) }]}
          getPath={d => d.path}
          getColor={[0, 100, 255, 255]}
          getWidth={5}
          widthMinPixels={5}
          pickable={false}
          updateTriggers={{
            data: populationState.currentPlanIndex
          }}
        />
      {/if}
      
      <!-- Activity Dots - Large and numbered (rendered on top of lines) -->
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
            // Clicking on any activity dot of the selected person should toggle off
            populationState.selectedPersonId = null;
            appState.secondarySidebar = 'HIDDEN';
          }}
          updateTriggers={{
            data: populationState.currentPlanIndex,
            getFillColor: populationState.selectingActivityId
          }}
        />
        
        <!-- Activity Numbers (rendered on top of dots) -->
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
          updateTriggers={{
            data: populationState.currentPlanIndex
          }}
        />
      {/if}
    {/if}
  {/if}
{/if}