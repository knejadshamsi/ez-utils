<script lang="ts" generics="T">
  import type { Snippet } from 'svelte';
  
  interface Props {
    id: string;
    loadData?: () => Promise<T[]>;
    resetOn?: any;
    header: Snippet;
    children: Snippet<[{ items: T[], isExpanded: boolean }]>;
  }
  
  let { id, loadData, resetOn, header, children }: Props = $props();
  
  let localItems = $state<T[]>([]);
  let previousResetValue = $state(resetOn);
  
  const isExpanded = $derived(localItems.length > 0);
  
  $effect(() => {
    if (resetOn !== previousResetValue) {
      localItems = [];
      previousResetValue = resetOn;
    }
  });
  
  async function handleClick() {
    if (localItems.length === 0 && loadData) {
      localItems = await loadData();
    } else {
      localItems = [];
    }
  }
</script>

<div class="rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
  <button
    class="w-full flex items-center gap-2 p-3 border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer"
    onclick={handleClick}
  >
    {@render header()}
  </button>
  
  {#if isExpanded}
    <div class="p-3">
      {@render children({ items: localItems, isExpanded })}
    </div>
  {/if}
</div>