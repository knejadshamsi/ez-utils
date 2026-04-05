<script lang="ts">
  import { FileText, FolderOpen } from 'lucide-svelte';
  import { ez } from '$lib/stores/ez.svelte';
</script>

{#if !ez.active}
  <div class="fixed inset-0 z-[2500] flex items-center justify-center bg-black/70">
    <div class="w-[420px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      <h1 class="mb-2 text-xl font-semibold">EZ-Utils</h1>
      <p class="mb-5 text-sm text-base-content/70">
        Open an existing .ez file or start a new one from a MATSim XML file.
      </p>

      <div class="flex flex-col gap-3">
        <button class="btn btn-outline justify-start gap-3" onclick={() => void ez.promptOpen()}>
          <FolderOpen size={16} />
          Open .ez File
        </button>
        <button class="btn btn-primary justify-start gap-3" onclick={() => void ez.promptNewFromXml()}>
          <FileText size={16} />
          New from XML
        </button>
      </div>

      {#if ez.error}
        <div class="mt-4 rounded-lg bg-error/10 px-3 py-2 text-sm text-error">
          {ez.error}
        </div>
      {/if}
    </div>
  </div>
{/if}
