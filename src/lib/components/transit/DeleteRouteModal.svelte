<script lang="ts">
  import { t } from 'svelte-i18n';
  import { transitEdit } from '$lib/stores/transit.svelte';
  import { confirmDeleteRoute } from '$lib/stores/transit-actions';
</script>

{#if transitEdit.deleteRouteTarget}
  <div class="fixed inset-0 z-[2600] flex items-center justify-center bg-black/70">
    <div class="w-[480px] rounded-xl bg-base-100 p-6 text-base-content shadow-2xl">
      <h2 class="mb-2 text-lg font-semibold">
        {$t('transit.delete_route_title', { values: { id: transitEdit.deleteRouteTarget.routeId } })}
      </h2>

      {#if transitEdit.deleteRouteTarget.loadingPreview}
        <p class="mb-6 text-sm text-base-content/60">…</p>
      {:else if transitEdit.deleteRouteTarget.preview}
        {@const p = transitEdit.deleteRouteTarget.preview}
        <p class="mb-3 text-sm text-base-content/70">
          {$t('transit.delete_route_cascade', {
            values: {
              profileStops: p.profileStops,
              pathLinks: p.pathLinks,
              departures: p.departures,
            },
          })}
        </p>
        <p class="mb-6 text-sm text-base-content/60">{$t('transit.delete_line_cannot_undo')}</p>
      {/if}

      {#if transitEdit.deleteRouteTarget.error}
        <p class="mb-4 text-sm text-error">{transitEdit.deleteRouteTarget.error}</p>
      {/if}

      <div class="flex justify-end gap-1">
        <button
          class="transit-modal-btn"
          disabled={transitEdit.deleteRouteTarget.submitting}
          onclick={() => transitEdit.cancelDeleteRoute()}
        >
          {$t('transit.cancel')}
        </button>
        <button
          class="transit-modal-btn transit-modal-btn-danger"
          disabled={transitEdit.deleteRouteTarget.submitting || transitEdit.deleteRouteTarget.loadingPreview}
          onclick={() => void confirmDeleteRoute()}
        >
          {$t('transit.delete_confirm')}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .transit-modal-btn {
    padding: 6px 14px;
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--color-base-content);
    background: none;
    border: 1px solid transparent;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.1s, color 0.1s, border-color 0.1s;
  }

  .transit-modal-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-modal-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .transit-modal-btn-danger {
    color: #ef4444;
  }

  .transit-modal-btn-danger:hover:not(:disabled) {
    background: none;
    border-color: #ef4444;
  }
</style>
