<script lang="ts">
  import { drawers } from '$lib/stores/ui.svelte';
  import type { Snippet } from 'svelte';

  interface Props {
    id: string;
    side?: 'left' | 'right';
    width?: number;
    children: Snippet;
  }

  let {
    id,
    side = 'left',
    width = 360,
    children,
  }: Props = $props();

  let isOpen = $derived(drawers.isOpen(id));
</script>

<div
  class="ez-drawer ez-drawer-{side}"
  class:open={isOpen}
  style:width="{width}px"
>
  <div class="ez-drawer-inner">
    {@render children()}
  </div>
</div>

<style>
  .ez-drawer {
    position: absolute;
    top: 50px;
    bottom: 10px;
    z-index: 900;
    transition: transform 0.3s ease;
    pointer-events: none;
  }

  .ez-drawer.open {
    pointer-events: auto;
    transform: translateX(0);
  }

  .ez-drawer-left {
    left: 10px;
    transform: translateX(calc(-100% - 10px));
  }

  .ez-drawer-right {
    right: 10px;
    transform: translateX(calc(100% + 10px));
  }

  .ez-drawer-inner {
    width: 100%;
    height: 100%;
    background: var(--color-base-100);
    border-radius: var(--radius-box, 8px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  }
</style>
