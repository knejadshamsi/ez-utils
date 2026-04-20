<script lang="ts">
  import { X } from 'lucide-svelte';
  import PopulationActivityEditor from '$lib/components/population/PopulationActivityEditor.svelte';
  import PopulationAttributeEditor from '$lib/components/population/PopulationAttributeEditor.svelte';
  import PopulationPlanEditor from '$lib/components/population/PopulationPlanEditor.svelte';
  import { population } from '$lib/stores/population.svelte';
</script>

{#if population.person}
  <div class="population-detail population-detail-shell">
    <header class="population-detail-header population-detail-fixed" style="border-bottom: none;">
      <div class="population-tab-bar">
        <button
          class="population-tab"
          class:population-tab-active={population.activeTab === 'plans'}
          onclick={() => population.setActiveTab('plans')}
        >Plans</button>
        <button
          class="population-tab"
          class:population-tab-active={population.activeTab === 'attributes'}
          onclick={() => population.setActiveTab('attributes')}
        >Attributes</button>
      </div>
      <button
        class="population-close-button"
        onclick={() => population.closeSecondary()}
        title="Close"
        aria-label="Close"
      >
        <X size={14} />
      </button>
    </header>

    <div class="population-detail-scroll">
      {#if population.activeTab === 'plans'}
        <PopulationPlanEditor plans={population.plans} activePlanId={population.activePlanId} />
        {#if population.activePlan}
          <PopulationActivityEditor plan={population.activePlan} />
        {/if}
      {:else}
        <PopulationAttributeEditor />
      {/if}
    </div>

    <footer class="population-action-bar">
      {#if population.activeTab === 'plans'}
        <div class="population-inline-actions">
          <button class="population-button flex-1" onclick={() => population.toggleEditLocations()}>
            {population.editLocations ? 'Done Editing' : 'Edit Locations'}
          </button>
          <button class="population-button flex-1" onclick={() => population.beginAddActivity()}>
            {population.mapAction === 'add_activity' ? 'Click map...' : 'Add Activity'}
          </button>
        </div>
      {:else}
        <button class="population-button w-full" onclick={() => population.addAttributeRow()}>
          Add Attribute
        </button>
      {/if}
      <button class="population-button population-button-primary w-full" onclick={() => void population.saveCurrentPerson()}>
        {population.saveState === 'saving' ? 'Saving...' : population.saveState === 'saved' ? 'Saved!' : 'Save'}
      </button>
    </footer>
  </div>
{/if}
