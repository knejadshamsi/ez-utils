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
  <header class="section-toolbar">
    <h3>{$t('transit.routes_heading')}</h3>
    <div class="section-toolbar-actions">
      <button
        class="toolbar-btn"
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

  <div class="list-container">
    {#if transit.loadingRoutes && transit.routes.length === 0}
      <div class="list-empty">…</div>
    {:else if transit.routesError}
      <div class="list-error">{transit.routesError}</div>
    {:else if transit.routes.length === 0}
      <div class="list-empty">{$t('transit.no_routes_on_line')}</div>
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
          <div class="list-row group" class:list-row-selected={transit.selectedRouteId === route.id}>
            <button
              class="list-row-main"
              onclick={() => transit.selectRoute(transit.selectedRouteId === route.id ? null : route.id)}
            >
              <span class="list-row-label">
                <span class="list-row-modes">
                  {#if Icon}
                    <Icon size={14} />
                  {/if}
                </span>
                <span class="list-row-id">{route.description ?? route.id}</span>
              </span>
              <span class="list-row-meta opacity-0 group-hover:opacity-100">
                {route.stopCount === 1 ? '1 stop' : `${route.stopCount} stops`}
              </span>
            </button>
            <div class="list-row-actions opacity-0 group-hover:opacity-100">
              <button
                onclick={(e) => {
                  e.stopPropagation();
                  if (transit.selectedLineId) transitEdit.beginRenameRoute(transit.selectedLineId, route);
                }}
                title={$t('transit.rename_route')}
                aria-label={$t('transit.rename_route')}
              >
                <Pencil size={14} />
              </button>
              <button
                onclick={(e) => {
                  e.stopPropagation();
                  if (transit.selectedLineId) transitEdit.requestDeleteRoute(transit.selectedLineId, route.id);
                }}
                title={$t('transit.delete_route')}
                aria-label={$t('transit.delete_route')}
              >
                <Trash2 size={14} />
              </button>
            </div>
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

</style>
