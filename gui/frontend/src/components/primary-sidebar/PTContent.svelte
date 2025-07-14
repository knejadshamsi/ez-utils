<script lang="ts">
  import { Button, Checkbox, Label, Input } from 'flowbite-svelte';
  import { PlusOutline, ChevronDownOutline, ChevronRightOutline, CogOutline, CheckOutline, CloseOutline } from 'flowbite-svelte-icons';
  import { ptState, TransportMode } from '$lib/stores/pt.svelte';
  import { appState } from '$lib/stores/app.svelte.ts';
  import { changeTracker } from '$lib/changeTracker.svelte';
  import { getCurrentProcessId } from '$lib/utils/processId';
  import { generateId } from '$lib/utils/generateId';
  import AddLineModal from '../modals/pt/AddLineModal.svelte';
  import EditLineModal from '../modals/pt/EditLineModal.svelte';

  let showAddLineModal = $state(false);
  let showEditLineModal = $state(false);
  let collapsedModes = $state<Set<TransportMode>>(new Set());
  let collapsedLines = $state<Set<string>>(new Set());
  let editingLine = $state<any>(null);
  let addingRouteToLine = $state<string | null>(null);
  let newRouteName = $state('');
  
  // Debug logging
  $effect(() => {
    console.log('[PrimaryPTSidebar] Rendering with lines:', ptState.lines.size);
    console.log('[PrimaryPTSidebar] All lines:', Array.from(ptState.lines.values()));
  });
  
  function toggleModeCollapse(mode: TransportMode) {
    const newSet = new Set(collapsedModes);
    if (newSet.has(mode)) {
      newSet.delete(mode);
    } else {
      newSet.add(mode);
    }
    collapsedModes = newSet;
  }
  
  function toggleLineCollapse(lineId: string) {
    const newSet = new Set(collapsedLines);
    if (newSet.has(lineId)) {
      newSet.delete(lineId);
    } else {
      newSet.add(lineId);
    }
    collapsedLines = newSet;
  }

  function handleLineClick(lineId: string) {
    ptState.setSelectedLine(lineId);
  }
  
  function handleRouteClick(routeId: string) {
    console.log('[PrimaryPTSidebar] Route clicked:', routeId);
    ptState.setSelectedRoute(routeId);
    appState.secondarySidebar = 'EXPANDED';
    console.log('[PrimaryPTSidebar] After selection - selectedRouteId:', ptState.selectedRouteId);
  }

  function handleAddLine() {
    showAddLineModal = true;
  }
  
  function handleEditLine(line: any) {
    editingLine = line;
    showEditLineModal = true;
  }
  
  function handleAddRoute(lineId: string) {
    addingRouteToLine = lineId;
    newRouteName = '';
  }
  
  function handleSaveRoute() {
    if (!addingRouteToLine || !newRouteName.trim()) return;
    
    const line = ptState.lines.get(addingRouteToLine);
    if (!line) return;
    
    const newRoute = {
      id: generateId(),
      direction: newRouteName.trim(),
      lineId: addingRouteToLine,
      firstDeparture: '06:00',
      stopSequence: [],
      telemetryId: '',
      raw_xml: ''
    };
    
    // Add to change tracker
    changeTracker.pendingChanges.push({
      type: 'pt',
      elementType: 'route',
      action: 'add',
      processId: getCurrentProcessId(),
      route: newRoute
    });
    
    // Add to line
    line.routes.push(newRoute);
    
    // Update state to trigger reactivity
    const newLines = new Map(ptState.lines);
    newLines.set(line.id, { ...line });
    ptState.lines = newLines;
    
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

<div class="h-full flex flex-col">
  <div class="p-4">
    <Button
      size="sm"
      color="primary"
      class="w-full mb-4"
      onclick={handleAddLine}
    >
      <PlusOutline size="sm" class="mr-1" />
      Add Line
    </Button>

    <div class="flex gap-4">
      <Label class="flex items-center gap-2">
        <Checkbox
          checked={ptState.visibility.stops}
          onchange={() => ptState.toggleVisibility('stops')}
        />
        Show Stops
      </Label>
      <Label class="flex items-center gap-2">
        <Checkbox
          checked={ptState.visibility.routes}
          onchange={() => ptState.toggleVisibility('routes')}
        />
        Show Routes
      </Label>
    </div>
  </div>

  <div class="flex-1 overflow-y-auto">
    {#each Array.from(ptState.modes) as mode (mode)}
        {@const lines = Array.from(ptState.lines.values()).filter(line => line.mode === mode)}
        <div class="border-b border-gray-200 dark:border-gray-700">
          <button
            class="w-full px-4 py-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors flex items-center justify-between"
            onclick={() => toggleModeCollapse(mode)}
          >
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
              {ptState.getModeIcon(mode)} {mode.toUpperCase()} ({lines.length})
            </h3>
            {#if collapsedModes.has(mode)}
              <ChevronRightOutline size="sm" class="text-gray-500" />
            {:else}
              <ChevronDownOutline size="sm" class="text-gray-500" />
            {/if}
          </button>
          {#if !collapsedModes.has(mode)}
            <div class="py-1">
              {#if lines.length === 0}
                <div class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 italic">
                  No {mode} lines
                </div>
              {:else}
                {#each lines as line (line.id)}
                  <div class="mx-3 mb-3">
                    <!-- Line Card -->
                    <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                      <!-- Line Header -->
                      <div class="flex items-center gap-2 p-3 border-b border-gray-100 dark:border-gray-700">
                        <div 
                          class="w-3 h-3 rounded-full border-2 border-white shadow-sm" 
                          style="background-color: {line.color}"
                        ></div>
                        
                        <button
                          class="flex items-center gap-2 flex-1 text-left hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                          onclick={() => handleLineClick(line.id)}
                        >
                          <div>
                            <div class="text-sm font-semibold text-gray-900 dark:text-white">
                              {truncateText(line.name)}
                            </div>
                            <div class="text-xs text-gray-500 dark:text-gray-400">
                              {line.routes.length} route{line.routes.length !== 1 ? 's' : ''}
                              {#if line.number}• #{line.number}{/if}
                            </div>
                          </div>
                        </button>
                        
                        <button
                          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                          onclick={() => handleEditLine(line)}
                        >
                          <CogOutline size="sm" class="text-gray-500" />
                        </button>
                        
                        <button
                          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                          onclick={() => toggleLineCollapse(line.id)}
                        >
                          {#if collapsedLines.has(line.id)}
                            <ChevronRightOutline size="sm" class="text-gray-500" />
                          {:else}
                            <ChevronDownOutline size="sm" class="text-gray-500" />
                          {/if}
                        </button>
                      </div>
                      
                      <!-- Routes Content -->
                      {#if !collapsedLines.has(line.id)}
                        <div class="p-3">
                          {#if line.routes.length === 0}
                            <div class="text-center py-3">
                              <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No routes</p>
                              {#if addingRouteToLine === line.id}
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
                                  onclick={() => handleAddRoute(line.id)}
                                >
                                  <PlusOutline size="xs" class="mr-1" />
                                  Add Route
                                </Button>
                              {/if}
                            </div>
                          {:else}
                            <div class="space-y-2">
                              {#each line.routes as route (route.id)}
                                <button
                                  class="w-full flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700 rounded border hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors text-left {ptState.selectedRouteId === route.id ? 'ring-2 ring-blue-500 border-blue-500' : ''}"
                                  onclick={() => handleRouteClick(route.id)}
                                >
                                  <div>
                                    <div class="text-sm font-medium text-gray-900 dark:text-white">
                                      {truncateText(route.direction || 'Route')}
                                    </div>
                                    <div class="text-xs text-gray-500 dark:text-gray-400">
                                      {route.stopSequence.length} stops
                                    </div>
                                  </div>
                                  <div class="text-xs text-gray-400 dark:text-gray-500">
                                    {route.firstDeparture || 'No schedule'}
                                  </div>
                                </button>
                              {/each}
                              
                              <div class="pt-2 border-t border-gray-100 dark:border-gray-600">
                                {#if addingRouteToLine === line.id}
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
                                    onclick={() => handleAddRoute(line.id)}
                                  >
                                    <PlusOutline size="xs" class="mr-1" />
                                    Add Route
                                  </Button>
                                {/if}
                              </div>
                            </div>
                          {/if}
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        </div>
    {/each}
  </div>
</div>

<AddLineModal bind:open={showAddLineModal} />

{#if editingLine}
  <EditLineModal 
    bind:open={showEditLineModal} 
    line={editingLine}
    onclose={() => editingLine = null}
  />
{/if}