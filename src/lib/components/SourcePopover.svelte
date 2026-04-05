<script lang="ts">
  import { onDestroy } from 'svelte';
  import ColorPicker from 'svelte-awesome-color-picker';
  import { ChevronUp, ChevronDown, Plus, X, Pencil, Save } from 'lucide-svelte';
  import { sourceStore } from '$lib/stores/ui.svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { t } from 'svelte-i18n';

  let selectedId = $state<string | null>(null);
  let editingId = $state<string | null>(null);
  let editValue = $state('');
  let colorEditingId = $state<string | null>(null);
  let hasSelection = $derived(selectedId !== null);
  let selectedIndex = $derived(selectedId ? sources.indexOf(selectedId) : -1);

  function hexToRgba(hex: string, a: number) {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return { r, g, b, a };
  }

  function rgbaToHex({ r, g, b }: { r: number; g: number; b: number }) {
    return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
  }

  function onClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest('.ez-source-popover') && !target.closest('.ez-source-btn')) {
      commitRename();
      colorEditingId = null;
      sourceStore.hidePopover();
    }
  }

  function selectItem(id: string) {
    if (editingId) return;
    selectedId = selectedId === id ? null : id;
    if (colorEditingId && colorEditingId !== id) colorEditingId = null;
  }

  let swatchClickTimer: ReturnType<typeof setTimeout> | null = null;

  onDestroy(() => {
    if (swatchClickTimer) clearTimeout(swatchClickTimer);
  });

  function onSwatchClick(id: string, e: MouseEvent) {
    e.stopPropagation();
    if (swatchClickTimer) {
      clearTimeout(swatchClickTimer);
      swatchClickTimer = null;
      // Double click - toggle visibility
      if (editingId === id) commitRename();
      sources.toggleVisibility(id);
      return;
    }
    swatchClickTimer = setTimeout(() => {
      swatchClickTimer = null;
      // Single click - toggle color picker (only if visible)
      const source = sources.get(id);
      if (source && !source.visible) return;
      selectedId = id;
      colorEditingId = colorEditingId === id ? null : id;
    }, 250);
  }

  function onColorInput(id: string, event: { rgb: { r: number; g: number; b: number; a: number } | null }) {
    if (!event.rgb) return;
    sources.setColor(id, rgbaToHex(event.rgb), event.rgb.a ?? 1);
  }

  function startRename(id: string, currentName: string) {
    selectedId = id;
    editingId = id;
    editValue = currentName;
  }

  function commitRename() {
    if (editingId === null) return;
    sources.rename(editingId, editValue);
    editingId = null;
  }

  function cancelRename() {
    editingId = null;
  }

  function onRenameKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') { e.preventDefault(); commitRename(); }
    else if (e.key === 'Escape') { e.preventDefault(); cancelRename(); }
  }

  function addSource() {
    const id = sources.add();
    if (id) selectedId = id;
  }

  function deleteSource() {
    if (selectedId === null) return;
    sources.remove(selectedId);
    selectedId = null;
    colorEditingId = null;
  }

  function moveUp() {
    if (selectedIndex <= 0) return;
    sources.moveUp(selectedIndex);
  }

  function moveDown() {
    if (selectedIndex < 0 || selectedIndex >= sources.items.length - 1) return;
    sources.moveDown(selectedIndex);
  }
</script>

<svelte:window onclick={onClickOutside} />

{#if sourceStore.popoverOpen}
  <div class="ez-source-popover">
    <!-- Toolbar -->
    <div class="ez-source-toolbar">
      <button class="ez-source-toolbar-btn" style="margin-left: auto;" onclick={moveUp} disabled={!hasSelection || selectedIndex <= 0} title="Move up">
        <ChevronUp size={13} />
      </button>
      <button class="ez-source-toolbar-btn" onclick={moveDown} disabled={!hasSelection || selectedIndex >= sources.items.length - 1} title="Move down">
        <ChevronDown size={13} />
      </button>
      <button class="ez-source-toolbar-btn" onclick={addSource} disabled={sources.isFull} title="Add source">
        <Plus size={13} />
      </button>
    </div>

    <!-- Source list -->
    <div class="ez-source-list">
      {#each sources.items as source (source.id)}
        <button
          class="ez-source-item"
          class:selected={source.id === selectedId}
          class:hidden-source={!source.visible}
          style:opacity={source.opacity}
          onclick={() => selectItem(source.id)}
        >
          <span
            class="ez-source-color"
            style:background={source.color}
            onclick={(e) => onSwatchClick(source.id, e)}
            role="button"
            tabindex="-1"
            title="Click: color picker / Double-click: toggle visibility"
          ></span>
          {#if editingId === source.id}
            <!-- svelte-ignore a11y_autofocus -->
            <input
              class="ez-source-rename-input"
              type="text"
              bind:value={editValue}
              onkeydown={onRenameKeydown}
              onblur={commitRename}
              autofocus
              onclick={(e) => e.stopPropagation()}
            />
          {:else}
            <span class="ez-source-name">{source.name}</span>
          {/if}
          {#if source.id === selectedId}
            <span class="ez-source-actions">
              {#if editingId === source.id}
                <button
                  class="ez-source-item-action"
                  onclick={(e) => { e.stopPropagation(); commitRename(); }}
                  title="Save name"
                >
                  <Save size={10} />
                </button>
              {:else}
                <button
                  class="ez-source-item-action"
                  onclick={(e) => { e.stopPropagation(); startRename(source.id, source.name); }}
                  disabled={!source.visible}
                  title="Rename source"
                >
                  <Pencil size={10} />
                </button>
              {/if}
              <button
                class="ez-source-item-action"
                onclick={(e) => { e.stopPropagation(); deleteSource(); }}
                title="Delete source"
              >
                <X size={10} />
              </button>
            </span>
          {/if}
        </button>

        {#if colorEditingId === source.id}
          <div class="ez-color-picker-popover" onclick={(e) => e.stopPropagation()}>
            <ColorPicker
              isDialog={false}
              isAlpha={true}
              isTextInput={false}
              sliderDirection="horizontal"
              rgb={hexToRgba(source.color, source.opacity)}
              onInput={(c) => onColorInput(source.id, c as any)}
            />
          </div>
        {/if}
      {/each}

      {#if sources.items.length === 0}
        <div class="ez-source-empty">
          <p>{$t('sources.no_sources')}</p>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .ez-source-popover {
    position: absolute;
    top: 50px;
    left: 10px;
    z-index: 1100;
    width: 220px;
    background: var(--color-base-100);
    border-radius: var(--radius-box, 8px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
    overflow: visible;
  }

  .ez-source-toolbar {
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 4px 6px;
    border-bottom: 1px solid oklch(var(--bc) / 0.08);
    border-radius: var(--radius-box, 8px) var(--radius-box, 8px) 0 0;
    background: var(--color-base-100);
  }

  .ez-source-toolbar-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 4px;
    color: var(--color-base-content);
    opacity: 0.6;
    cursor: pointer;
    background: none;
    border: none;
    transition: opacity 0.1s, background 0.1s;
  }

  .ez-source-toolbar-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-200);
  }

  .ez-source-toolbar-btn:disabled {
    opacity: 0.12;
    cursor: default;
  }

  .ez-source-list {
    max-height: 260px;
    overflow-y: auto;
    overflow-x: visible;
    border-radius: 0 0 var(--radius-box, 8px) var(--radius-box, 8px);
    background: var(--color-base-100);
  }

  .ez-source-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 10px;
    background: none;
    border: none;
    color: var(--color-base-content);
    cursor: pointer;
    text-align: left;
    transition: background 0.1s;
  }

  .ez-source-item:hover {
    background: var(--color-base-200);
  }

  .ez-source-item.selected {
    background: var(--color-base-300);
    box-shadow: inset 2px 0 0 var(--color-primary);
  }

  .ez-source-item.hidden-source {
    text-decoration: line-through;
    opacity: 0.4 !important;
  }

  .ez-source-color {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 3px;
    flex-shrink: 0;
    cursor: pointer;
    transition: outline 0.1s;
  }

  .ez-source-color:hover {
    outline: 1px solid oklch(var(--bc) / 0.5);
    outline-offset: 1px;
  }

  .ez-source-name {
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }

  .ez-source-rename-input {
    flex: 1;
    min-width: 0;
    font-size: 12px;
    padding: 1px 4px;
    background: var(--color-base-200);
    color: var(--color-base-content);
    border: 1px solid var(--color-primary);
    border-radius: 3px;
    outline: none;
  }

  .ez-color-picker-popover {
    position: absolute;
    left: calc(100% + 6px);
    z-index: 1200;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);

    /* Theme the picker to match DaisyUI dark */
    --cp-border-color: oklch(var(--bc) / 0.15);
    --cp-text-color: var(--color-base-content);
    --picker-height: 160px;
    --slider-width: 12px;
    --picker-width: 160px;
  }

  /* Reorder: alpha on top, hue below, square at bottom */
  .ez-color-picker-popover :global(.wrapper) {
    display: flex;
    flex-direction: column;
  }
  .ez-color-picker-popover :global(.wrapper .a) {
    order: -2;
  }
  .ez-color-picker-popover :global(.wrapper .h) {
    order: -1;
  }

  .ez-source-actions {
    display: flex;
    align-items: center;
    gap: 2px;
    margin-left: auto;
    flex-shrink: 0;
  }

  .ez-source-item-action {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    flex-shrink: 0;
    border-radius: 3px;
    background: none;
    border: none;
    color: var(--color-base-content);
    opacity: 0.4;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s;
  }

  .ez-source-item-action:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-200);
  }

  .ez-source-item-action:disabled {
    opacity: 0.15;
    cursor: default;
  }

  .ez-source-empty {
    padding: 16px;
    text-align: center;
    font-size: 11px;
    color: oklch(var(--bc) / 0.4);
  }
</style>
