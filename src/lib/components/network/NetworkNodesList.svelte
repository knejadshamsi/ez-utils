<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Disc, Hand, PlusCircle, Search, Trash2 } from 'lucide-svelte';
  import { network, NETWORK_SEARCH_CAP, NETWORK_BBOX_NODE_CAP } from '$lib/stores/network.svelte';
  import { setNodeSearchQuery, requestDeleteNode } from '$lib/stores/network-actions';
</script>

<section class="network-section">
  <div class="section-toolbar">
    <h3>{$t('network.nodes')}</h3>
    <div class="section-toolbar-actions">
      <button
        class="toolbar-btn network-opacity-button"
        class:active={network.nodeOpacity > 0}
        disabled={!!network.selection}
        onclick={() => network.cycleNodeOpacity()}
        title={$t('network.cycle_node_opacity_hint')}
        aria-label={$t('network.cycle_node_opacity')}
      >
        {Math.round(network.nodeOpacity * 100)}%
      </button>
      <button
        class="toolbar-btn"
        class:active={network.nodeDragEnabled}
        onclick={() => network.setNodeDragEnabled(!network.nodeDragEnabled)}
        title={network.nodeDragEnabled ? $t('network.disable_free_move') : $t('network.enable_free_move')}
        aria-label={$t('network.free_move')}
      >
        <Hand size={14} />
      </button>
      <button
        class="toolbar-btn"
        class:active={network.nodeSearch.open}
        onclick={() => network.toggleNodeSearch()}
        title={network.nodeSearch.open ? $t('network.close_node_search') : $t('network.search_nodes')}
        aria-label={$t('network.search_nodes')}
      >
        <Search size={14} />
      </button>
      <button
        class="toolbar-btn"
        class:active={network.mapAction === 'add_node'}
        onclick={() => network.mapAction === 'idle' ? network.beginAddNode() : network.cancelMapAction()}
        title={$t('network.add_node_hint')}
        aria-label={$t('network.add_node')}
      >
        <PlusCircle size={14} />
      </button>
    </div>
  </div>

  {#if network.nodeCapExceeded && !network.nodeSearch.open}
    <p class="network-cap-warning">
      {$t('network.cap_nodes', { values: { cap: NETWORK_BBOX_NODE_CAP, total: network.nodeTotal } })}
    </p>
  {/if}

  {#if network.nodeSearch.open}
    <div class="search-row">
      <input
        type="text"
        placeholder={$t('network.search_node_placeholder')}
        value={network.nodeSearch.query}
        oninput={(e) => setNodeSearchQuery(e.currentTarget.value)}
      />
      <button onclick={() => network.toggleNodeSearchExact()}>
        {network.nodeSearch.exact ? $t('network.exact') : $t('network.partial')}
      </button>
    </div>
    {#if network.nodeSearch.capExceeded}
      <p class="network-cap-warning">
        {$t('network.search_cap', { values: { cap: NETWORK_SEARCH_CAP, total: network.nodeSearch.total } })}
      </p>
    {/if}
  {/if}

  <div class="list-container">
    {#each network.displayedNodes as node (node.id)}
      {@const selected = network.selection?.type === 'node' && network.selection.id === node.id}
      <div class="list-row group" class:list-row-selected={selected}>
        <button
          class="list-row-main"
          onclick={() => network.selectNode(node.id)}
        >
          <span class="list-row-label">
            <Disc size={14} />
            <span class="list-row-id">{node.id}</span>
          </span>
        </button>
        <div class="list-row-actions opacity-0 group-hover:opacity-100">
          <button
            onclick={(event) => {
              event.stopPropagation();
              requestDeleteNode(node.id);
            }}
            title={$t('network.delete_node')}
            aria-label={$t('network.delete_node')}
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    {/each}
    {#if network.displayedNodes.length === 0}
      <p class="list-empty">{$t('network.no_nodes')}</p>
    {/if}
  </div>
</section>

<style>
  .network-section {
    display: flex;
    flex-direction: column;
    min-height: 0;
    flex: 1;
  }
  .network-cap-warning {
    margin: 4px 12px 0;
    font-size: 0.75rem;
    color: #ef4444;
  }
</style>
