<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Minus, PlusCircle, Search, Trash2 } from 'lucide-svelte';
  import { network, NETWORK_SEARCH_CAP, NETWORK_BBOX_LINK_CAP } from '$lib/stores/network.svelte';
  import { setLinkSearchQuery, requestDeleteLink } from '$lib/stores/network-actions';
</script>

<section class="network-section">
  <div class="network-section-toolbar">
    <h3>{$t('network.links')}</h3>
    <span class="network-toolbar-spacer"></span>
    <button
      class="population-button network-opacity-button"
      class:population-button-primary={network.linkOpacity > 0}
      disabled={!!network.selection}
      onclick={() => network.cycleLinkOpacity()}
      title={$t('network.cycle_link_opacity_hint')}
      aria-label={$t('network.cycle_link_opacity')}
    >
      {Math.round(network.linkOpacity * 100)}%
    </button>
    <button
      class="population-button population-icon-button"
      class:population-button-primary={network.linkSearch.open}
      onclick={() => network.toggleLinkSearch()}
      title={network.linkSearch.open ? $t('network.close_link_search') : $t('network.search_links')}
      aria-label={$t('network.search_links')}
    >
      <Search size={14} />
    </button>
    <button
      class="population-button population-icon-button"
      class:population-button-primary={network.mapAction === 'add_link_from' || network.mapAction === 'add_link_to'}
      onclick={() => network.mapAction === 'idle' ? network.beginAddLink() : network.cancelMapAction()}
      title={$t('network.add_link_hint')}
      aria-label={$t('network.add_link')}
    >
      <PlusCircle size={14} />
    </button>
  </div>

  {#if network.linkCapExceeded && !network.linkSearch.open && !network.selectedNodeId}
    <p class="network-cap-warning">
      {$t('network.cap_links', { values: { cap: NETWORK_BBOX_LINK_CAP, total: network.linkTotal } })}
    </p>
  {/if}

  {#if network.linkSearch.open}
    <div class="population-search-row">
      <input
        class="population-search-input"
        type="text"
        placeholder={$t('network.search_link_placeholder')}
        value={network.linkSearch.query}
        oninput={(e) => setLinkSearchQuery(e.currentTarget.value)}
      />
      <button
        class="population-button population-search-toggle"
        onclick={() => network.toggleLinkSearchExact()}
      >
        {network.linkSearch.exact ? $t('network.exact') : $t('network.partial')}
      </button>
    </div>
    {#if network.linkSearch.capExceeded}
      <p class="network-cap-warning">
        {$t('network.search_cap', { values: { cap: NETWORK_SEARCH_CAP, total: network.linkSearch.total } })}
      </p>
    {/if}
  {/if}

  <div class="network-list">
    {#each network.displayedLinks as link (link.id)}
      {@const selected = network.selectedLinkId === link.id}
      <div class="population-row-shell group" class:population-row-selected={selected}>
        <button
          class="population-list-row population-row-main"
          onclick={() => void network.selectLink(link.id)}
        >
          <span class="network-row-label">
            <Minus size={14} />
            <span class="population-row-id">{link.id}</span>
          </span>
        </button>
        <button
          class="population-row-delete population-button population-icon-button opacity-0 group-hover:opacity-100"
          onclick={(event) => {
            event.stopPropagation();
            requestDeleteLink(link.id);
          }}
          title={$t('network.delete_link')}
          aria-label={$t('network.delete_link')}
        >
          <Trash2 size={14} />
        </button>
      </div>
    {/each}
    {#if network.displayedLinks.length === 0}
      <p class="population-empty">{$t('network.no_links')}</p>
    {/if}
  </div>
</section>

<style>
  .network-section {
    display: flex;
    flex-direction: column;
    min-height: 0;
    flex: 1;
    border-bottom: 1px solid var(--color-base-300);
  }
  .network-section-toolbar {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 12px;
    border-bottom: 1px solid var(--color-base-300);
  }
  .network-section-toolbar h3 {
    margin: 0;
    font-size: 0.85rem;
    font-weight: 700;
  }
  .network-toolbar-spacer { flex: 1; }
  .network-list {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .network-cap-warning {
    margin: 4px 12px 0;
    font-size: 0.75rem;
    color: #ef4444;
  }
  .network-opacity-button {
    font-size: 0.7rem;
    font-weight: 600;
    padding: 8px;
    min-width: 38px;
  }
  .network-row-label {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>
