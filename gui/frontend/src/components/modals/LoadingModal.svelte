<script lang="ts">
  import { Modal, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline, ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { appState, commandArgs } from "$lib/stores/app.svelte";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { populationState } from "$lib/stores/population.svelte";
  import { networkState } from "$lib/stores/network.svelte";
  import { ptState } from "$lib/stores/pt.svelte";
  import { LoadProcessData, GetPTStops, GetPTRoutes, GetPTRouteStops, GetPTDepartures } from "@wailsjs/go/gui/App";
  import { gui } from "@wailsjs/go/models";
  import { getCurrentViewportBounds } from "$lib/map/mapUtils";

  $effect(() => {
    if (appState.display === 'LOADING') {
      (async () => {
        try {
          const viewport = getCurrentViewportBounds();
          
          const config = {
            POPULATION: { randomFactor: 0.3, maxElements: 500 },
            NETWORK: { randomFactor: 0.8, maxElements: 2000 },
            PT: { randomFactor: 1.0, maxElements: 100 }
          };
          
          const params = new gui.LoadingParams({
            processId: appState.processId,
            fileEditMode: commandArgs.fileEditMode,
            viewport: viewport,
            randomFactor: config[commandArgs.fileEditMode].randomFactor,
            maxElements: config[commandArgs.fileEditMode].maxElements
          });
          
          const result = await LoadProcessData(params);
          
          if (commandArgs.fileEditMode === 'POPULATION') {
            populationState.persons.clear();
            if (result.data && Array.isArray(result.data)) {
              result.data.forEach((person: any) => {
                populationState.persons.set(person.id, person);
              });
            }
          } else if (commandArgs.fileEditMode === 'NETWORK') {
            networkState.nodes = result.data.nodes || [];
            networkState.links = result.data.links || [];
          } else if (commandArgs.fileEditMode === 'PT') {
            if (result.data && result.data.lines && result.data.lines.length > 0) {
              // Get all line IDs from the viewport-filtered data
              const lineIds = result.data.lines.map((line: any) => line.id);
              
              // Load full PT data for these lines
              const [stops, routes, routeStops, departures] = await Promise.all([
                GetPTStops(appState.processId),
                Promise.all(lineIds.map((lineId: string) => GetPTRoutes(appState.processId, lineId))).then(r => r.flat()),
                Promise.all(lineIds.map((lineId: string) => 
                  GetPTRoutes(appState.processId, lineId).then((routes: any) => 
                    Promise.all(routes.map((route: any) => GetPTRouteStops(appState.processId, route.id)))
                  )
                )).then(r => r.flat().flat()),
                Promise.all(lineIds.map((lineId: string) => 
                  GetPTRoutes(appState.processId, lineId).then((routes: any) => 
                    Promise.all(routes.map((route: any) => GetPTDepartures(appState.processId, route.id)))
                  )
                )).then(r => r.flat().flat())
              ]);
              
              // Load complete PT data
              ptState.loadPTData(
                result.data.lines,
                stops as any[],
                routes as any[],
                routeStops as any[],
                departures as any[],
                false
              );
              
              // Mark lines as loaded
              lineIds.forEach((lineId: string) => ptState.loadedLines.add(lineId));
              ptState.stopsLoaded = true;
              
              // Generate summaries from loaded data
              const summaries = result.data.lines.map((line: any) => {
                const lineData = ptState.lines.get(line.id);
                return {
                  id: line.id,
                  name: line.id.replace(/_/g, ' '),
                  mode: line.mode,
                  route_count: lineData?.routes.length || 0,
                  departure_count: lineData?.routes.reduce((sum, route) => sum + (route.departures?.length || 0), 0) || 0
                };
              });
              
              ptState.loadLineSummaries(summaries);
            }
          }
          
          appState.display = 'LOADING_SUCCESS';
          
        } catch (error) {
          appState.display = 'LOADING_FAILED';
        }
      })();
    }
  });
</script>

<Modal 
  open={appState.display === 'LOADING' || appState.display === 'LOADING_SUCCESS' || appState.display === 'LOADING_FAILED'} 
  permanent
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="md"
  placement="center"
>
{#snippet header()}
  <h3 class="text-xl font-semibold text-gray-900 dark:text-white text-center flex items-center justify-center gap-3">
    {#if appState.display === 'LOADING'}
      <Spinner size="6" />
    {:else if appState.display === 'LOADING_SUCCESS'}
      <CheckCircleOutline class="w-6 h-6 text-green-500" />
    {:else}
      <ExclamationCircleOutline class="w-6 h-6 text-red-500" />
    {/if}
    {appState.display === 'LOADING' ? 'Loading Process Data' : appState.display === 'LOADING_SUCCESS' ? 'Process Loaded Successfully' : 'Loading Failed'}
  </h3>
{/snippet}

<div class="text-center space-y-4">
  {appState.display === 'LOADING' ? 'Loading process data from database...' : appState.display === 'LOADING_SUCCESS' ? 'Process data loaded successfully. Ready to begin editing.' : 'Failed to load process data'}
  
  {#if appState.display === 'LOADING_SUCCESS'}
    <Button color="primary" size="lg" onclick={() => appState.display = 'EDITING'}>
      Begin Editing
    </Button>
  {:else if appState.display === 'LOADING_FAILED'}
    <Button color="alternative" onclick={() => {
      welcomeModalState.showTable = false;
      welcomeModalState.currentPage = 1;
      welcomeModalState.processes = [];
      appState.display = 'WELCOME';
    }}>
      Back to Welcome
    </Button>
  {/if}
</div>
</Modal>