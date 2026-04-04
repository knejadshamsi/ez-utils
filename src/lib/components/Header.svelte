<script lang="ts">
  import { toolbar } from '$lib/stores/ui.svelte';
</script>

<header class="flex items-center justify-between px-4 h-10 bg-base-200 border-b border-base-300 shrink-0">
  <!-- Logo -->
  <div class="text-base-content font-bold text-lg tracking-tight">
    EZ-Utils
  </div>

  <!-- Toolbar -->
  <div class="flex items-center gap-2">
    {#if toolbar.isBusy}
      <span class="text-xs text-base-content/50">{toolbar.status}...</span>
      <span class="loading loading-spinner loading-xs text-primary"></span>
    {/if}
  </div>
</header>
