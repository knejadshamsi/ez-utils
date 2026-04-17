<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Bus, Pencil, Plus, TrainFront, TramFront, Trash2 } from 'lucide-svelte';
  import { transit, transitEdit } from '$lib/stores/transit.svelte';
  import {
    loadRouteDeletePreview,
    submitCreateRoute,
    submitRenameRoute,
  } from '$lib/stores/transit-actions';
  import TransitRouteEditRow from './TransitRouteEditRow.svelte';

  $effect(() => {
    if (transitEdit.deleteRouteTarget && transitEdit.deleteRouteTarget.loadingPreview) {
      void loadRouteDeletePreview();
    }
  });

  function modeIconFor(raw: string) {
    const m = raw.toLowerCase();
    if (m === 'bus') return Bus;
    if (m === 'subway' || m === 'metro' || m === 'rail') return TrainFront;
    if (m === 'tram' || m === 'light_rail' || m === 'lightrail' || m === 'streetcar') return TramFront;
    return null;
  }
</script>

<section class="transit-list-section transit-routes-section">
  <header class="transit-list-header">
    <h3>{$t('transit.routes_heading')}</h3>
    <div class="transit-header-actions">
      <button
        class="transit-icon-btn"
        class:active={transitEdit.createRouteForm.open}
        onclick={() => {
          if (transitEdit.createRouteForm.open) {
            transitEdit.closeCreateRouteForm();
          } else {
            transitEdit.openCreateRouteForm();
          }
        }}
        title={$t('transit.add_route')}
        aria-label={$t('transit.add_route')}
      >
        <Plus size={14} />
      </button>
    </div>
  </header>

  {#if transitEdit.createRouteForm.open}
    <TransitRouteEditRow
      idValue={transitEdit.createRouteForm.id}
      descriptionValue={transitEdit.createRouteForm.description}
      submitting={transitEdit.createRouteForm.submitting}
      error={transitEdit.createRouteForm.error}
      onIdChange={(v) => transitEdit.setCreateRouteFormId(v)}
      onDescriptionChange={(v) => transitEdit.setCreateRouteFormDescription(v)}
      onSave={() => void submitCreateRoute()}
      onCancel={() => transitEdit.closeCreateRouteForm()}
    />
  {/if}

  <div class="transit-list">
    {#if transit.loadingRoutes && transit.routes.length === 0}
      <div class="transit-list-empty">…</div>
    {:else if transit.routesError}
      <div class="transit-list-empty transit-error">{transit.routesError}</div>
    {:else if transit.routes.length === 0}
      <div class="transit-list-empty">{$t('transit.no_routes_on_line')}</div>
    {:else}
      {#each transit.routes as route (route.id)}
        {#if transitEdit.renameRouteTarget && transitEdit.renameRouteTarget.originalId === route.id}
          <TransitRouteEditRow
            idValue={transitEdit.renameRouteTarget.id}
            descriptionValue={transitEdit.renameRouteTarget.description}
            submitting={transitEdit.renameRouteTarget.submitting}
            error={transitEdit.renameRouteTarget.error}
            onIdChange={(v) => transitEdit.setRenameRouteId(v)}
            onDescriptionChange={(v) => transitEdit.setRenameRouteDescription(v)}
            onSave={() => void submitRenameRoute()}
            onCancel={() => transitEdit.cancelRenameRoute()}
          />
        {:else}
          {@const Icon = modeIconFor(route.transportMode)}
          <div class="transit-line-shell" class:selected={transit.selectedRouteId === route.id}>
            <button
              class="transit-line-item"
              onclick={() => transit.selectRoute(transit.selectedRouteId === route.id ? null : route.id)}
            >
              <span class="transit-line-modes">
                {#if Icon}
                  <Icon size={16} />
                {/if}
              </span>
              <span class="transit-line-name">{route.description ?? route.id}</span>
            </button>
            <div class="transit-line-actions">
              <button
                class="transit-icon-btn"
                onclick={(e) => {
                  e.stopPropagation();
                  if (transit.selectedLineId) transitEdit.beginRenameRoute(transit.selectedLineId, route);
                }}
                title={$t('transit.rename_route')}
                aria-label={$t('transit.rename_route')}
              >
                <Pencil size={13} />
              </button>
              <button
                class="transit-icon-btn transit-icon-btn-danger"
                onclick={(e) => {
                  e.stopPropagation();
                  if (transit.selectedLineId) transitEdit.requestDeleteRoute(transit.selectedLineId, route.id);
                }}
                title={$t('transit.delete_route')}
                aria-label={$t('transit.delete_route')}
              >
                <Trash2 size={13} />
              </button>
            </div>
            <span class="transit-line-count">{route.stopCount}</span>
          </div>
        {/if}
      {/each}
    {/if}
  </div>
</section>

<style>
  .transit-list-section {
    display: flex;
    flex-direction: column;
    flex: 1 1 0;
    min-height: 0;
    overflow: hidden;
  }

  .transit-list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    background: var(--color-base-200);
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-list-header h3 {
    margin: 0;
    font-size: 0.78rem;
    font-weight: 700;
    letter-spacing: 0.02em;
  }

  .transit-header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .transit-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 4px;
    border: none;
    background: none;
    color: var(--color-base-content);
    opacity: 0.6;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s, color 0.1s;
  }

  .transit-icon-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-300);
  }

  .transit-icon-btn.active {
    opacity: 1;
    background: var(--color-base-300);
    color: var(--color-primary);
  }

  .transit-icon-btn-danger:hover:not(:disabled) {
    color: #ef4444;
  }

  .transit-list {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding: 4px 0;
  }

  .transit-list-empty {
    padding: 12px 16px;
    text-align: center;
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.5);
  }

  .transit-list-empty.transit-error {
    color: #b91c1c;
  }

  .transit-line-shell {
    display: flex;
    align-items: center;
    width: 100%;
    transition: background 0.1s;
  }

  .transit-line-shell:hover {
    background: var(--color-base-200);
  }

  .transit-line-shell.selected {
    background: var(--color-base-300);
    box-shadow: inset 2px 0 0 var(--color-primary);
  }

  .transit-line-item {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1 1 auto;
    min-width: 0;
    padding: 9px 14px;
    background: none;
    border: none;
    color: var(--color-base-content);
    text-align: left;
    cursor: pointer;
  }

  .transit-line-actions {
    display: none;
    align-items: center;
    gap: 2px;
    flex-shrink: 0;
  }

  .transit-line-shell:hover .transit-line-actions,
  .transit-line-shell.selected .transit-line-actions {
    display: flex;
  }

  .transit-line-modes {
    display: flex;
    align-items: center;
    gap: 3px;
    opacity: 0.75;
    flex-shrink: 0;
  }

  .transit-line-name {
    flex: 1;
    min-width: 0;
    font-size: 0.88rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-line-count {
    flex-shrink: 0;
    margin-left: 6px;
    margin-right: 14px;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--color-base-200);
    font-size: 0.75rem;
    font-weight: 600;
    color: oklch(var(--bc) / 0.7);
  }
</style>
