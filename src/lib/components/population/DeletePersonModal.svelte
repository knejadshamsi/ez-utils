<script lang="ts">
  import { population } from '$lib/stores/population.svelte';
</script>

{#if population.deletePromptOpen}
  <div class="fixed inset-0 z-[2600] flex items-center justify-center bg-black/70">
    <div class="w-[420px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      <h2 class="mb-2 text-lg font-semibold">Delete Person</h2>
      <p class="mb-6 text-sm text-base-content/70">
        Delete person {population.deleteTarget?.personId}? This cannot be undone.
      </p>

      <div class="flex justify-end gap-2">
        <button class="btn btn-ghost" onclick={() => population.cancelDeletePerson()}>Cancel</button>
        <button class="btn btn-error" onclick={() => void population.confirmDeletePerson()}>Delete</button>
      </div>
    </div>
  </div>
{/if}
