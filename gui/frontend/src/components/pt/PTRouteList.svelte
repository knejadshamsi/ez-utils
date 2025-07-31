<script lang="ts">
  import { Button, Input } from 'flowbite-svelte';
  import { PlusOutline, CheckOutline, CloseOutline } from 'flowbite-svelte-icons';
  import { selected } from '@workflow/pt/state.svelte';
  import { createRoute } from '@workflow/pt/crud.svelte';
  import { saveCurrentRoute, loadRouteStops, loadRouteDepartures } from '@workflow/pt/loading.svelte';
  import { selectRoute, selectStop, selectLine } from '@workflow/pt/functions.svelte';
  import type { Route } from '@workflow/pt/types';
  import { appState } from '$lib/stores/app.svelte';
  import { updatePTVisualization } from '../../map/updatePTVisualization';

  const { lineId, routes, isExpanded } = $props<{ 
    lineId: string; 
    routes: Route[]; 
    isExpanded: boolean; 
  }>();

  let addingRouteToLine = $state<string | null>(null);
  let newRouteName = $state('');

  async function handleRouteClick(routeId: string) {
    // Save previous route data if needed
    if (selected.routeId && selected.routeId !== routeId) {
      try {
        await saveCurrentRoute();
      } catch (error) {
        console.error('Failed to auto-save route:', error);
      }
    }
    
    // Toggle functionality - if clicking the same route, unselect it
    if (selected.routeId === routeId) {
      selectRoute(null);
      selectStop(null);
      appState.secondarySidebar = 'HIDDEN';
      updatePTVisualization();
      return;
    }
    
    // Otherwise select the new route
    selectLine(lineId); // Ensure line is selected
    selectRoute(routeId);
    appState.secondarySidebar = 'EXPANDED';
    
    // Load route data (always fresh fetch)
    try {
      await Promise.all([
        loadRouteStops(routeId),
        loadRouteDepartures(routeId)
      ]);
    } catch (error) {
      console.error('Failed to load route data:', error);
    }
    
    // Update visualization
    updatePTVisualization();
  }

  function handleAddRoute(currentLineId: string) {
    addingRouteToLine = currentLineId;
    newRouteName = '';
  }
  
  function handleSaveRoute() {
    if (!addingRouteToLine || !newRouteName.trim()) return;
    
    createRoute(newRouteName.trim(), addingRouteToLine);
    
    // Clear adding state
    addingRouteToLine = null;
    newRouteName = '';
  }
  
  function handleCancelRoute() {
    addingRouteToLine = null;
    newRouteName = '';
  }
  
  function truncateText(text: string, maxLength: number = 20): string {
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
  }
</script>

{#if isExpanded && routes.length > 0}
  <div class="space-y-2">
    {#each routes as route (route.id)}
      <button
        class="w-full flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700 rounded border hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors text-left {selected.routeId === route.id ? 'ring-2 ring-blue-500 border-blue-500' : ''}"
        onclick={() => handleRouteClick(route.id)}
      >
        <span class="text-sm font-medium text-gray-900 dark:text-white">
          {truncateText(route.name || 'Route')}
        </span>
      </button>
    {/each}
    
    <div class="pt-2">
      {#if addingRouteToLine === lineId}
        <div class="flex gap-2 items-center">
          <Input
            bind:value={newRouteName}
            placeholder="Route name (e.g., Uptown)"
            size="sm"
            class="flex-1"
            onkeydown={(e) => {
              if (e.key === 'Enter') handleSaveRoute();
              if (e.key === 'Escape') handleCancelRoute();
            }}
            autofocus
          />
          <Button
            size="xs"
            color="primary"
            onclick={handleSaveRoute}
            disabled={!newRouteName.trim()}
          >
            <CheckOutline size="xs" />
          </Button>
          <Button
            size="xs"
            color="alternative"
            onclick={handleCancelRoute}
          >
            <CloseOutline size="xs" />
          </Button>
        </div>
      {:else}
        <Button
          size="xs"
          color="alternative"
          class="w-full"
          onclick={() => handleAddRoute(lineId)}
        >
          <PlusOutline size="xs" class="mr-1" />
          Add Route
        </Button>
      {/if}
    </div>
  </div>
{:else if isExpanded}
  <div class="text-center py-3">
    <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No routes</p>
    {#if addingRouteToLine === lineId}
      <div class="flex gap-2 items-center">
        <Input
          bind:value={newRouteName}
          placeholder="Route name (e.g., Downtown)"
          size="sm"
          class="flex-1"
          onkeydown={(e) => {
            if (e.key === 'Enter') handleSaveRoute();
            if (e.key === 'Escape') handleCancelRoute();
          }}
          autofocus
        />
        <Button
          size="xs"
          color="primary"
          onclick={handleSaveRoute}
          disabled={!newRouteName.trim()}
        >
          <CheckOutline size="xs" />
        </Button>
        <Button
          size="xs"
          color="alternative"
          onclick={handleCancelRoute}
        >
          <CloseOutline size="xs" />
        </Button>
      </div>
    {:else}
      <Button
        size="xs"
        color="primary"
        onclick={() => handleAddRoute(lineId)}
      >
        <PlusOutline size="xs" class="mr-1" />
        Add Route
      </Button>
    {/if}
  </div>
{/if}