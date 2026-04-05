<script lang="ts">
  import { ez } from '$lib/stores/ez.svelte';
</script>

{#if ez.closeConfirmOpen}
  <div class="fixed inset-0 z-[2600] flex items-center justify-center bg-black/70">
    <div class="w-[420px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      <h2 class="mb-2 text-lg font-semibold">Unsaved Changes</h2>
      <p class="mb-6 text-sm text-base-content/70">
        Save changes to your .ez file before closing, or exit without saving.
      </p>

      <div class="flex justify-end gap-2">
        <button class="btn btn-ghost" onclick={() => ez.requestCloseDecision(null)}>Cancel</button>
        <button class="btn btn-outline" onclick={() => ez.requestCloseDecision('discard')}>Exit Without Saving</button>
        <button class="btn btn-primary" onclick={() => ez.requestCloseDecision('save')}>Save and Exit</button>
      </div>
    </div>
  </div>
{/if}
