<script lang="ts">
  import { t } from 'svelte-i18n';
  import { sources } from '$lib/stores/data.svelte';
  import { transit } from '$lib/stores/transit.svelte';
  import TransitOptionalInputs from './TransitOptionalInputs.svelte';
  import TransitModeToggles from './TransitModeToggles.svelte';
  import TransitLinesTab from './TransitLinesTab.svelte';
  import TransitStopsTab from './TransitStopsTab.svelte';
</script>

<div class="transit-panel">
  <header class="transit-panel-header">
    <h2>{sources.activeName ?? $t('transit.no_active_source')}</h2>
  </header>

  <TransitOptionalInputs />
  <TransitModeToggles />

  <div class="transit-primary-tab-bar">
    <button
      class="transit-primary-tab"
      class:active={transit.primaryTab === 'lines'}
      onclick={() => transit.setPrimaryTab('lines')}
    >
      {$t('transit.tab_lines')}
    </button>
    <button
      class="transit-primary-tab"
      class:active={transit.primaryTab === 'stops'}
      onclick={() => transit.setPrimaryTab('stops')}
    >
      {$t('transit.tab_stops_primary')}
    </button>
  </div>

  {#if transit.primaryTab === 'lines'}
    <TransitLinesTab />
  {:else}
    <TransitStopsTab />
  {/if}
</div>

<style>
  .transit-panel {
    height: 100%;
    display: flex;
    flex-direction: column;
    color: var(--color-base-content);
  }

  .transit-panel-header {
    padding: 16px;
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-panel-header h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 700;
  }

  .transit-primary-tab-bar {
    display: flex;
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-primary-tab {
    flex: 1;
    padding: 8px 12px;
    border: none;
    background: none;
    color: var(--color-base-content);
    font-size: 0.8rem;
    font-weight: 600;
    opacity: 0.5;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: opacity 0.1s, border-color 0.1s, background 0.1s;
  }

  .transit-primary-tab:hover {
    opacity: 0.85;
    background: var(--color-base-200);
  }

  .transit-primary-tab.active {
    opacity: 1;
    border-bottom-color: var(--color-primary);
  }
</style>
