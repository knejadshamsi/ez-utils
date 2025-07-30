<script lang="ts">
  import { Modal, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline, ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { appState, commandArgs } from "$lib/stores/app.svelte";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { populationState, type Person } from "$lib/stores/population.svelte";
  import { networkState } from "$lib/stores/network.svelte";
  import { ptState } from "$lib/stores/pt.svelte";
  import { LoadProcessData } from "@wailsjs/go/gui/App";
  import { gui } from "@wailsjs/go/models";
  import { getCurrentViewportBounds } from "$lib/map/mapUtils";
  import { parsePersonXML } from "$lib/utils/populationXmlParser";
  import { updateNetworkVisualization } from "../../map/updateNetworkVisualization";

  $effect(() => {
    if (appState.display === 'LOADING') {
      (async () => {
        try {
          const viewport = getCurrentViewportBounds();
          
          const config = {
            POPULATION: { randomFactor: 0.3, maxElements: 500, minThreshold: 200 },
            NETWORK: { randomFactor: 0.8, maxElements: 2000, minThreshold: 0 },
            PT: { randomFactor: 1.0, maxElements: 100, minThreshold: 0 }
          };
          
          const params = new gui.LoadingParams({
            processId: appState.processId,
            fileEditMode: commandArgs.fileEditMode,
            viewport: viewport,
            randomFactor: config[commandArgs.fileEditMode].randomFactor,
            maxElements: config[commandArgs.fileEditMode].maxElements,
            minThreshold: config[commandArgs.fileEditMode].minThreshold,
            mode: commandArgs.fileEditMode === 'PT' ? ptState.selected.mode : ''
          });
          
          const result = await LoadProcessData(params);
          
          if (commandArgs.fileEditMode === 'POPULATION') {
            populationState.persons.clear();
            if (result.data && Array.isArray(result.data)) {
              result.data.forEach((person: any) => {
                const parsedData = parsePersonXML(person.raw_xml);
                const processedPerson: Person = {
                  id: person.id,
                  zoneId: '',
                  attributes: parsedData.attributes || [],
                  plans: parsedData.plans || []
                };
                populationState.persons.set(person.id, processedPerson);
              });
            }
          } else if (commandArgs.fileEditMode === 'NETWORK') {
            networkState.nodes = result.data.nodes || [];
            networkState.links = result.data.links || [];
            // Update visualization immediately after loading data
            updateNetworkVisualization();
          } else if (commandArgs.fileEditMode === 'PT') {
            if (result.data) {
              // Load PT data using the store's method
              ptState.loadPTData({
                lines: result.data.lines || [],
                routes: [],
                stops: [],
                departures: []
              });
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