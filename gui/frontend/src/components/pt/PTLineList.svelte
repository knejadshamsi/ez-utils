<script lang="ts">
  import { Button, Checkbox, Label, Input } from 'flowbite-svelte';
  import { PlusOutline, CheckOutline, CloseOutline, EditOutline, TrashBinOutline } from 'flowbite-svelte-icons';
  import { lines, routes, selected } from '@workflow/pt/state.svelte';
  import { updateLine, deleteLine, createLine, createRoute } from '@workflow/pt/crud.svelte';
  import { selectLine, getRoutesForLine } from '@workflow/pt/functions.svelte';
  import { TRANSPORT_MODES, type Line } from '@workflow/pt/types';
  import PTRouteList from './PTRouteList.svelte';

  let isAddingLine = $state(false);
  let newLineName = $state('');
  let editingLineId = $state<string | null>(null);
  let editingLineName = $state('');
  let deletingLineId = $state<string | null>(null);
  let expandedLines = $state<Record<string, boolean>>({});
  
  // Get lines for current mode
  const currentModeLines = $derived(
    () => Object.values(lines).filter((line: Line) => line.mode === selected.mode)
  );

  function handleLineClick(lineId: string) {
    selectLine(lineId);
  }
  
  function startEditingLine(line: Line) {
    editingLineId = line.id;
    editingLineName = line.name;
  }
  
  function saveLineEdit() {
    if (editingLineId && editingLineName.trim()) {
      updateLine(editingLineId, { name: editingLineName.trim() });
      editingLineId = null;
    }
  }
  
  function cancelLineEdit() {
    editingLineId = null;
    editingLineName = '';
  }
  
  function confirmDeleteLine(lineId: string) {
    deletingLineId = lineId;
  }
  
  function cancelDeleteLine() {
    deletingLineId = null;
  }
  
  function handleDeleteLine(lineId: string) {
    deleteLine(lineId);
    deletingLineId = null;
  }
  
  function truncateText(text: string, maxLength: number = 20): string {
    return text.length > maxLength ? text.substring(0, maxLength) + '...' : text;
  }
  
  function startAddingLine() {
    isAddingLine = true;
    newLineName = '';
  }
  
  function cancelAddingLine() {
    isAddingLine = false;
    newLineName = '';
  }
  
  function saveNewLine() {
    const trimmedName = newLineName.trim();
    if (!trimmedName) return;
    
    // Check if line with same name exists for this mode
    const existingLine = Object.values(lines).find(
      (line: Line) => line.mode === selected.mode && line.name.toLowerCase() === trimmedName.toLowerCase()
    );
    
    if (existingLine) {
      alert('A line with this name already exists for this mode');
      return;
    }
    
    createLine(trimmedName, selected.mode);
    
    // Reset form
    cancelAddingLine();
  }
</script>

<div class="p-4 space-y-3">
  <!-- Add Line button at the top -->
  <div>
    {#if isAddingLine}
      <div class="rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm p-3">
        <div class="flex items-center gap-2">
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

<div class="flex-1 overflow-y-auto">
  <div class="px-4 py-3">
    <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
      {TRANSPORT_MODES[selected.mode]?.icon || '🚌'} {TRANSPORT_MODES[selected.mode]?.name} Lines ({currentModeLines.length})
    </h3>
  </div>
  <div class="py-1">
    {#if currentModeLines.length === 0}
      <div class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 italic text-center">
        No {selected.mode.toLowerCase()} lines
      </div>
    {:else}
{#each currentModeLines() as line (line.id)}
        {@const lineRoutes = getRoutesForLine(line.id)}
        {@const isExpanded = expandedLines[line.id] || false}
        <div class="mx-3 mb-3">
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
            <button
              class="w-full px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors text-left"
              onclick={() => {
                expandedLines[line.id] = !expandedLines[line.id];
              }}
            >
              <div class="flex items-center justify-between">
                <div class="flex-1">
                  {#if editingLineId === line.id}
                    <div class="flex items-center gap-2" onclick={(e) => e.stopPropagation()} role="group">
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
                  {:else if deletingLineId === line.id}
                    <div class="flex items-center justify-between w-full" onclick={(e) => e.stopPropagation()} role="group">
                      <span class="text-sm text-gray-600 dark:text-gray-400">Delete this line?</span>
                      <div class="flex items-center gap-1">
                        <button
                          class="p-1 hover:bg-red-100 dark:hover:bg-red-900 rounded transition-colors"
                          onclick={() => handleDeleteLine(line.id)}
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
                          handleLineClick(line.id);
                        }}
                        onkeydown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            e.stopPropagation();
                            handleLineClick(line.id);
                          }
                        }}
                        role="link"
                        tabindex="0"
                      >
                        {truncateText(line.name)} 
                        {#if line.unsaved}
                          <span class="text-xs font-normal text-orange-500" title="Unsaved changes">●</span>
                        {/if}
                      </div>
                      <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity" onclick={(e) => e.stopPropagation()}>
                        <button
                          class="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
                          onclick={(e) => {
                            e.stopPropagation();
                            startEditingLine(line);
                          }}
                        >
                          <EditOutline size="xs" class="text-gray-500" />
                        </button>
                        <button
                          class="p-1 hover:bg-red-100 dark:hover:bg-red-900 rounded transition-colors"
                          onclick={(e) => {
                            e.stopPropagation();
                            confirmDeleteLine(line.id);
                          }}
                        >
                          <TrashBinOutline size="xs" class="text-red-500" />
                        </button>
                      </div>
                    </div>
                  {/if}
                </div>
                <div class="text-gray-500 ml-2">
                  {isExpanded ? '▼' : '▶'}
                </div>
              </div>
            </button>
            
            {#if isExpanded}
              <div class="px-3 pb-3">
                <PTRouteList lineId={line.id} routes={lineRoutes} isExpanded={isExpanded} />
              </div>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>