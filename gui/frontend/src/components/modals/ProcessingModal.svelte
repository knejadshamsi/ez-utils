<script lang="ts">
  import { Modal, P, Spinner, Button } from "flowbite-svelte";
  import { CheckCircleOutline, ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { appState } from "$lib/stores/app.svelte";
  import { welcomeModalState } from "./welcome-modal/welcome.svelte";
  import { EventsOn, EventsOff } from "@wailsjs/runtime/runtime";

  import { onMount, onDestroy } from "svelte";
  
  let telemetryData = $state<Record<string, any>>({});
  
  onMount(() => {
    EventsOn("process:status", (data: any) => {
      appState.display = data.status;
    });
    
    EventsOn("process:telemetry", (data: any) => {
      telemetryData = data;
    });
  });
  
  onDestroy(() => {
    EventsOff("process:status");
    EventsOff("process:telemetry");
  });

</script>

<Modal 
  open={appState.display === 'PROCESSING' || appState.display === 'PROCESSING_SUCCESS' || appState.display === 'PROCESSING_FAILED'} 
  permanent
  dismissable={false}
  outsideclose={false}
  autoclose={false}
  size="md"
  placement="center"
>
  {#snippet header()}
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white text-center flex items-center justify-center gap-3">
      {#if appState.display === 'PROCESSING'}
        <Spinner size="6" />
        Processing
      {:else if appState.display === 'PROCESSING_SUCCESS'}
        <CheckCircleOutline class="w-6 h-6 text-green-500" />
        Processing Complete
      {:else if appState.display === 'PROCESSING_FAILED'}
        <ExclamationCircleOutline class="w-6 h-6 text-red-500" />
        Processing Failed
      {/if}
    </h3>
  {/snippet}
  
  <div class="space-y-6">
    {#if appState.display === 'PROCESSING_FAILED'}
      <div class="text-center space-y-4">
        <P color="red" class="text-red-600 dark:text-red-400">
          Processing failed. Please try again.
        </P>
        <Button 
          color="alternative" 
          onclick={() => {
            welcomeModalState.showTable = false;
            appState.processId = 0;
            welcomeModalState.currentPage = 1;
            welcomeModalState.processes = [];
            appState.display = 'WELCOME';
          }}
        >
          Back to Welcome
        </Button>
      </div>
    {:else}
      <div class="space-y-4" class:text-center={appState.display === 'PROCESSING_SUCCESS'}>
        <P class="text-gray-600 dark:text-gray-400">
          {appState.display === 'PROCESSING' ? 'Processing data...' : 'Processing completed successfully.'}
        </P>
        
        <div class="space-y-3" class:space-y-2={appState.display === 'PROCESSING_SUCCESS'} class:text-left={appState.display === 'PROCESSING_SUCCESS'}>
          {#each Object.entries(telemetryData) as [key, value]}
            <div class="text-sm text-gray-600 dark:text-gray-400">
              {key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}: {typeof value === 'number' ? value.toLocaleString() : String(value)}
            </div>
          {/each}
        </div>
        
        {#if appState.display === 'PROCESSING_SUCCESS'}
          <Button
            color="primary"
            size="lg"
            onclick={() => appState.display = 'LOADING'}
          >
            Load Data
          </Button>
        {/if}
      </div>
    {/if}
  </div>
</Modal>