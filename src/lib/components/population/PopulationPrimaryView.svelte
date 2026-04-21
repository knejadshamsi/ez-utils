<script lang="ts">
  import { ChevronLeft, ChevronRight, Search, Trash2, User, UserPlus } from 'lucide-svelte';
  import { population } from '$lib/stores/population.svelte';
  import { sources } from '$lib/stores/data.svelte';
</script>

<div class="population-panel">
  <header class="panel-header">
    <h2>{sources.activeName ?? 'No active source'}</h2>
  </header>

  {#if population.error}
    <p class="list-error">{population.error}</p>
  {/if}

  <section class="drawer-section">
    <div class="section-toolbar">
      <h3>Persons</h3>
      <div class="section-toolbar-actions">
        <button
          class="toolbar-btn"
          class:active={population.searchOpen}
          onclick={() => population.toggleSearch()}
          title={population.searchOpen ? 'Close search' : 'Search persons'}
          aria-label={population.searchOpen ? 'Close search' : 'Search persons'}
        >
          <Search size={14} />
        </button>
        <button
          class="toolbar-btn"
          class:active={population.mapAction === 'add_person'}
          disabled={population.mapAction === 'add_person'}
          onclick={() => population.beginAddPerson()}
          title={population.mapAction === 'add_person' ? 'Click map to place person...' : 'Add person'}
          aria-label={population.mapAction === 'add_person' ? 'Click map to place person' : 'Add person'}
        >
          <UserPlus size={14} />
        </button>
      </div>
    </div>

    {#if population.searchOpen}
      <div class="search-row">
        <input
          type="text"
          placeholder="Search person ID..."
          value={population.searchQuery}
          oninput={(e) => population.setSearchQuery(e.currentTarget.value)}
          onkeydown={(e) => {
            if (e.key === 'Enter') {
              void population.executeSearch();
            }
            if (e.key === 'Escape') {
              population.toggleSearch();
            }
          }}
        />
        <button
          onclick={() => population.toggleSearchExact()}
          title={population.searchExact ? 'Switch to partial match' : 'Switch to exact match'}
        >
          {population.searchExact ? 'Exact' : 'Partial'}
        </button>
      </div>
    {/if}

  <div class="list-container">
    {#if population.searchLoading}
      <p class="list-empty">loading search results...</p>
    {/if}
    {#each population.people as person}
      {@const isSelected = population.selected?.personId === person.personId && population.selected?.sourceName === person.sourceName}
      <div class="list-row group" class:list-row-selected={isSelected}>
        <button
          class="list-row-main"
          onclick={() =>
            void population.selectPerson({
              sourceName: person.sourceName,
              personId: person.personId,
            })}
        >
          <span class="list-row-label">
            {#if person.personId === '__new__'}
              <span class="population-unsaved-dot"></span>
            {/if}
            <User size={14} />
            <span class="list-row-id">{person.personId}</span>
          </span>
        </button>
        <div class="list-row-actions opacity-0 group-hover:opacity-100">
          <button
            onclick={(event) => {
              event.stopPropagation();
              population.requestDeletePerson({
                sourceName: person.sourceName,
                personId: person.personId,
              });
            }}
            title="Delete person"
            aria-label={`Delete person ${person.personId}`}
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    {/each}
  </div>

  <div class="population-pagination">
    <button
      class="population-button population-icon-button"
      disabled={population.currentPage === 0}
      onclick={() => void population.previousPage()}
      aria-label="Previous page"
    >
      <ChevronLeft size={16} />
    </button>
    <span>Page {population.currentPage + 1}/{population.totalPages}</span>
    <button
      class="population-button population-icon-button"
      disabled={population.currentPage + 1 >= population.totalPages}
      onclick={() => void population.nextPage()}
      aria-label="Next page"
    >
      <ChevronRight size={16} />
    </button>
  </div>
  </section>
</div>

<style>
  :global(.population-panel) {
    height: 100%;
    display: flex;
    flex-direction: column;
    color: var(--color-base-content);
  }

  :global(.population-detail-header),
  :global(.population-section-header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  :global(.population-detail-header) {
    padding: 16px;
    border-bottom: 1px solid var(--color-base-300);
  }

  :global(.population-detail-header h2),
  :global(.population-section-header h3) {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 700;
  }

  :global(.population-detail-header p) {
    margin: 4px 0 0;
    font-size: 0.75rem;
    opacity: 0.7;
  }

  :global(.population-plan-active) {
    border-color: var(--color-primary);
    background: color-mix(in srgb, var(--color-primary) 12%, var(--color-base-100));
  }

  :global(.population-pagination) {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 16px 16px;
    border-top: 1px solid var(--color-base-300);
  }

  :global(.population-section) {
    padding: 16px;
    border-bottom: 1px solid var(--color-base-300);
  }

  :global(.population-card-list) {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 12px;
  }

  :global(.population-card) {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
    border: 1px solid var(--color-base-300);
    border-radius: 8px;
    background: var(--color-base-200);
  }

  :global(.population-card-title) {
    font-size: 1rem;
    font-weight: 800;
  }

  :global(.population-field) {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 0.8rem;
  }

  :global(.population-field input),
  :global(.population-field select),
  :global(.population-textarea),
  :global(.population-select) {
    width: 100%;
    padding: 8px 10px;
    border: 1px solid var(--color-base-300);
    border-radius: 8px;
    background: var(--color-base-100);
    color: var(--color-base-content);
  }

  :global(.population-button) {
    border: 1px solid var(--color-base-300);
    border-radius: 8px;
    padding: 8px 10px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    font-size: 0.8rem;
    cursor: pointer;
  }

  :global(.population-button:disabled) {
    opacity: 0.5;
  }

  :global(.population-button-primary) {
    border-color: var(--color-primary);
  }

  :global(.population-icon-button) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 8px;
  }

  :global(.population-button-danger) {
    border-color: #dc2626;
    color: #991b1b;
  }

  :global(.population-inline-actions) {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  :global(.population-detail) {
    height: 100%;
    overflow: auto;
  }

  :global(.population-detail-shell) {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :global(.population-detail-fixed) {
    flex-shrink: 0;
  }

  :global(.population-plan-toolbar) {
    flex-shrink: 0;
  }

  :global(.population-detail-scroll) {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  :global(.population-action-bar) {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    border-top: 1px solid var(--color-base-300);
    background: var(--color-base-100);
  }

  :global(.population-modal-backdrop) {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 5;
  }

  :global(.population-modal) {
    width: min(360px, calc(100% - 24px));
    background: var(--color-base-100);
    border: 1px solid var(--color-base-300);
    border-radius: 10px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  :global(.population-modal h3) {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 700;
  }

  :global(.population-modal p) {
    margin: 0;
    font-size: 0.8rem;
  }

  :global(.population-table) {
    width: 100%;
    border-collapse: collapse;
    margin-top: 12px;
    font-size: 0.8rem;
  }

  :global(.population-table th),
  :global(.population-table td) {
    padding: 8px 6px;
    border-bottom: 1px solid var(--color-base-300);
    text-align: left;
    vertical-align: top;
  }

  :global(.population-coordinates) {
    display: flex;
    justify-content: space-between;
    gap: 8px;
    font-size: 0.75rem;
    opacity: 0.8;
  }

  :global(.population-header-row) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  :global(.population-tab-bar) {
    display: flex;
    gap: 0;
    flex: 1;
  }

  :global(.population-tab) {
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
    transition: opacity 0.15s ease;
  }

  :global(.population-tab-active) {
    opacity: 1;
    border-bottom-color: var(--color-primary);
  }

  :global(.population-plan-row) {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  :global(.population-activity-compact) {
    padding: 8px 10px;
    gap: 6px;
  }

  :global(.population-activity-compact .population-card-title) {
    font-size: 0.85rem;
    font-weight: 700;
  }

  :global(.population-leg-compact) {
    padding: 6px 10px;
    gap: 6px;
    background: var(--color-base-300);
  }

  :global(.population-mark-selected-active),
  :global(.population-plan-row .population-mark-selected-active) {
    border-color: var(--color-primary);
    color: var(--color-primary);
  }

  :global(.population-table-equal) {
    table-layout: fixed;
  }

  :global(.population-table-equal th:last-child),
  :global(.population-table-equal td:last-child) {
    width: 40px;
  }

  :global(.population-table-equal input),
  :global(.population-table-equal select) {
    width: 100%;
    min-width: 0;
    box-sizing: border-box;
    padding: 8px 10px;
    border: 1px solid var(--color-base-300);
    border-radius: 8px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    font-size: 0.8rem;
  }

  :global(.population-close-button) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    background: none;
    border: none;
    color: var(--color-base-content);
    opacity: 0.6;
    cursor: pointer;
  }

  :global(.population-close-button:hover) {
    opacity: 1;
  }

  :global(.population-unsaved-dot) {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-primary);
    flex-shrink: 0;
  }

  :global(.population-row-id) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
  }

  :global(.population-plan-row .population-select),
  :global(.population-plan-row .population-button),
  :global(.population-action-bar .population-button) {
    border-color: color-mix(in srgb, var(--color-base-content) 20%, transparent);
  }

  :global(.population-action-bar .population-button-primary) {
    border-color: var(--color-primary);
  }

  :global(.population-unit) {
    font-size: 0.75rem;
    opacity: 0.7;
    white-space: nowrap;
  }

  :global(.population-leg-wrap) {
    position: relative;
    margin-left: 32px;
  }

  :global(.population-leg-connector) {
    position: absolute;
    left: -18px;
    top: -6px;
    bottom: -6px;
    width: 2px;
    background: #6b7280;
  }
</style>
