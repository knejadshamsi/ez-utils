<script lang="ts">
  import { Button, Checkbox, Label, Input } from 'flowbite-svelte';
  import { PlusOutline, CheckOutline, CloseOutline, EditOutline, TrashBinOutline } from 'flowbite-svelte-icons';
  import { ptState } from '$lib/stores/pt.svelte';
  import { appState } from '$lib/stores/app.svelte';
  import { generateId } from '$lib/utils/generateId';
  import { trackRouteChange, trackLineChange } from '$lib/utils/ptChangeTracking';
  import type { LineWithRoutes } from '$lib/stores/pt.svelte';
  import { updatePTVisualization } from '../../map/updatePTVisualization';
  import { PTService } from '$lib/api/pt';

  let isAddingLine = $state(false);
  let newLineName = $state('');
  let newLineNumber = $state('');
  let collapsedLines = $state<Set<string>>(new Set());
  let addingRouteToLine = $state<string | null>(null);
  let newRouteName = $state('');
  let editingLineId = $state<string | null>(null);
  let editingLineName = $state('');
  let editingLineNumber = $state('');
  let deletingLineId = $state<string | null>(null);
  
  // Initialize collapsed lines when summaries load or mode changes
  $effect(() => {
    const summaryIds = Array.from(ptState.lineSummaries.keys());
    if (summaryIds.length > 0) {
      // Always reinitialize collapsed state when summaries change (mode switch)
      collapsedLines = new Set(summaryIds);
    }
  });
  
  
  async function toggleLineCollapse(lineId: string) {
    const newSet = new Set(collapsedLines);
    if (newSet.has(lineId)) {
      // Expanding - remove from collapsed set
      newSet.delete(lineId);
      // Lazy load line data when expanding
      if (!ptState.isLineLoaded(lineId)) {
        try {
          await PTService.loadLineData(lineId);
        } catch (error) {
          console.error('Failed to load line data:', error);
        }
      }
    } else {
      // Collapsing - add to collapsed set
      newSet.add(lineId);
    }
    collapsedLines = newSet;
  }

  function handleLineClick(lineId: string) {
    ptState.setSelectedLine(lineId);
  }
  
  function startEditingLine(summary: any) {
    editingLineId = summary.id;
    editingLineName = summary.name;
    editingLineNumber = '';
  }
  
  function saveLineEdit() {
    if (editingLineId && editingLineName.trim()) {
      // TODO: Update line name
      editingLineId = null;
    }
  }
  
  function cancelLineEdit() {
    editingLineId = null;
    editingLineName = '';
    editingLineNumber = '';
  }
  
  function confirmDeleteLine(lineId: string) {
    deletingLineId = lineId;
  }
  
  function cancelDeleteLine() {
    deletingLineId = null;
  }
  
  function deleteLine(lineId: string) {
    // TODO: Add delete logic
    deletingLineId = null;
  }
  
  function handleRouteClick(routeId: string) {
    console.log('[PrimaryPTSidebar] Route clicked:', routeId);
    ptState.setSelectedRoute(routeId);
    appState.secondarySidebar = 'EXPANDED';
    console.log('[PrimaryPTSidebar] After selection - selectedRouteId:', ptState.selectedRouteId);
    // Update visualization
    updatePTVisualization();
  }

  
  async function handleAddRoute(lineId: string) {
    // Ensure line is loaded before adding route
    if (!ptState.isLineLoaded(lineId)) {
      try {
        await PTService.loadLineData(lineId);
      } catch (error) {
        console.error('Failed to load line data:', error);
        return;
      }
    }
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
    
    // Add to line
    line.routes.push(newRoute);
    
    // Track the new route
    trackRouteChange(newRoute, 'add');
    
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
  
  function startAddingLine() {
    isAddingLine = true;
    newLineName = '';
    newLineNumber = '';
  }
  
  function cancelAddingLine() {
    isAddingLine = false;
    newLineName = '';
    newLineNumber = '';
  }
  
  function saveNewLine() {
    const trimmedName = newLineName.trim();
    if (!trimmedName) return;
    
    // Check if line with same name exists for this mode
    const existingLine = Array.from(ptState.lines.values()).find(
      line => line.name.toLowerCase() === trimmedName.toLowerCase() && line.mode === ptState.selectedMode
    );
    
    if (existingLine) {
      alert('A line with this name already exists for this mode');
      return;
    }
    
    const newLine: LineWithRoutes = {
      id: generateId(),
      name: trimmedName,
      number: newLineNumber.trim(),
      mode: ptState.selectedMode,
      agencyId: '',
      telemetryId: '',
      raw_xml: '',
      routes: []
    };
    
    // Create new Map to trigger Svelte 5 reactivity
    const newLines = new Map(ptState.lines);
    newLines.set(newLine.id, newLine);
    ptState.lines = newLines;
    
    // Also add to line summaries for immediate display
    const newSummary = {
      id: newLine.id,
      name: newLine.name,
      mode: newLine.mode,
      route_count: 0,
      departure_count: 0
    };
    const newSummaries = new Map(ptState.lineSummaries);
    newSummaries.set(newLine.id, newSummary);
    ptState.lineSummaries = newSummaries;
    
    // Track the new line
    trackLineChange(newLine, 'add');
    ptState.setSelectedLine(newLine.id);
    
    // Reset form
    cancelAddingLine();
  }
</script>

<div class="h-full flex flex-col">
  <div class="p-4">
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
    <div class="px-4 py-3">
      <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
        {ptState.getModeIcon(ptState.selectedMode)} {ptState.selectedMode} Lines ({ptState.selectedModeSummaries().length})
      </h3>
    </div>
    <div class="py-1">
      {#if ptState.selectedModeSummaries().length === 0}
        <div class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 italic text-center">
          No {ptState.selectedMode.toLowerCase()} lines
        </div>
      {:else}
        {#each ptState.selectedModeSummaries() as summary (summary.id)}
                  <div class="mx-3 mb-3">
                    <!-- Line Card -->
                    <div class="rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                      <!-- Line Header -->
                      <div
                        class="w-full flex items-center gap-2 p-3 border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer"
                        onclick={async () => await toggleLineCollapse(summary.id)}
                        onkeydown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            toggleLineCollapse(summary.id);
                          }
                        }}
                        role="button"
                        tabindex="0"
                      >
                        <div class="flex-1">
                          {#if editingLineId === summary.id}
                            <div class="flex items-center gap-2" onclick={(e) => e.stopPropagation()} role="group">
                              <Input
                                type="text"
                                bind:value={editingLineNumber}
                                placeholder="#"
                                size="sm"
                                class="w-12 text-xs"
                                onkeydown={(e) => {
                                  if (e.key === 'Enter') saveLineEdit();
                                  if (e.key === 'Escape') cancelLineEdit();
                                }}
                              />
                              <Input
                                type="text"
                                bind:value={editingLineName}
                                size="sm"
                                class="flex-1"
                                onkeydown={(e) => {
                                  if (e.key === 'Enter') saveLineEdit();
                                  if (e.key === 'Escape') cancelLineEdit();
                                }}
                                autofocus
                              />
                              <Button size="xs" color="primary" onclick={saveLineEdit}>
                                <CheckOutline size="xs" />
                              </Button>
                              <Button size="xs" color="alternative" onclick={cancelLineEdit}>
                                <CloseOutline size="xs" />
                              </Button>
                            </div>
                          {:else if deletingLineId === summary.id}
                            <div class="flex items-center justify-between w-full" onclick={(e) => e.stopPropagation()} role="group">
                              <span class="text-sm text-gray-600 dark:text-gray-400">Delete this line?</span>
                              <div class="flex items-center gap-1">
                                <button
                                  class="p-1 hover:bg-red-100 dark:hover:bg-red-900 rounded transition-colors"
                                  onclick={() => deleteLine(summary.id)}
                                >
                                  <CheckOutline size="xs" class="text-red-500" />
                                </button>
                                <button
                                  class="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
                                  onclick={cancelDeleteLine}
                                >
                                  <CloseOutline size="xs" class="text-gray-500" />
                                </button>
                              </div>
                            </div>
                          {:else}
                            <div class="flex items-center justify-between w-full group">
                              <div 
                                class="text-sm font-semibold text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
                                onclick={(e) => {
                                  e.stopPropagation();
                                  handleLineClick(summary.id);
                                }}
                                onkeydown={(e) => {
                                  if (e.key === 'Enter' || e.key === ' ') {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    handleLineClick(summary.id);
                                  }
                                }}
                                role="link"
                                tabindex="0"
                              >
                                {truncateText(summary.name)} <span class="text-xs font-normal text-gray-500 dark:text-gray-400">• {summary.route_count} route{summary.route_count !== 1 ? 's' : ''}</span>
                              </div>
                              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                <button
                                  class="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
                                  onclick={(e) => {
                                    e.stopPropagation();
                                    startEditingLine(summary);
                                  }}
                                >
                                  <EditOutline size="xs" class="text-gray-500" />
                                </button>
                                <button
                                  class="p-1 hover:bg-red-100 dark:hover:bg-red-900 rounded transition-colors"
                                  onclick={(e) => {
                                    e.stopPropagation();
                                    confirmDeleteLine(summary.id);
                                  }}
                                >
                                  <TrashBinOutline size="xs" class="text-red-500" />
                                </button>
                              </div>
                            </div>
                          {/if}
                        </div>
                      </div>
                      
                      <!-- Routes Content -->
                      {#if !collapsedLines.has(summary.id)}
                        {@const line = ptState.lines.get(summary.id)}
                        {@const isLoading = ptState.loading.lines.get(summary.id) || false}
                        <div class="p-3">
                          {#if isLoading}
                            <div class="text-center py-3">
                              <p class="text-sm text-gray-500 dark:text-gray-400">Loading routes...</p>
                            </div>
                          {:else if !line}
                            <div class="text-center py-3">
                              <p class="text-sm text-gray-500 dark:text-gray-400">No route data loaded</p>
                            </div>
                          {:else if line.routes.length === 0}
                            <div class="text-center py-3">
                              <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">No routes</p>
                              {#if addingRouteToLine === summary.id}
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
                                  onclick={async () => await handleAddRoute(summary.id)}
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
                                  <div class="flex items-center justify-between w-full">
                                    <span class="text-sm font-medium text-gray-900 dark:text-white">
                                      {truncateText(route.direction || 'Route')}
                                    </span>
                                    <span class="text-xs text-gray-500 dark:text-gray-400">
                                      {route.stopSequence?.length || 0} stops
                                    </span>
                                  </div>
                                </button>
                              {/each}
                              
                              <div class="pt-2">
                                {#if addingRouteToLine === summary.id}
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
                                    onclick={async () => await handleAddRoute(summary.id)}
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
      
      <!-- Add Line inline creation -->
      <div class="mx-3 mt-3 mb-3">
        {#if isAddingLine}
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm p-3">
            <div class="flex items-center gap-2">
              <Input
                type="text"
                bind:value={newLineNumber}
                placeholder="#"
                size="sm"
                class="w-12 text-xs"
                onkeydown={(e) => {
                  if (e.key === 'Enter' && newLineName.trim()) saveNewLine();
                  if (e.key === 'Escape') cancelAddingLine();
                }}
              />
              <Input
                type="text"
                bind:value={newLineName}
                placeholder="Line name (required)"
                size="sm"
                class="flex-1"
                onkeydown={(e) => {
                  if (e.key === 'Enter' && newLineName.trim()) saveNewLine();
                  if (e.key === 'Escape') cancelAddingLine();
                }}
                autofocus
              />
              <Button size="xs" color="primary" onclick={saveNewLine} disabled={!newLineName.trim()}>
                <CheckOutline size="xs" />
              </Button>
              <Button size="xs" color="alternative" onclick={cancelAddingLine}>
                <CloseOutline size="xs" />
              </Button>
            </div>
          </div>
        {:else}
          <Button
            size="sm"
            color="primary"
            class="w-full"
            onclick={startAddingLine}
          >
            <PlusOutline size="sm" class="mr-1" />
            Add Line
          </Button>
        {/if}
      </div>
    </div>
  </div>
</div>


