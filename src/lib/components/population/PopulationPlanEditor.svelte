<script lang="ts">
  import { Check, Plus, Trash2 } from 'lucide-svelte';
  import type { PopulationPlanEditorProps } from '$lib/components/population/types';
  import { population } from '$lib/stores/population.svelte';

  let { plans, activePlanId }: PopulationPlanEditorProps = $props();

  let isActivePlanSelected = $derived(
    plans.find(p => p.planId === activePlanId)?.selected ?? false
  );
</script>

<section class="population-section population-plan-toolbar">
  <div class="population-plan-row">
    <select
      class="population-select flex-1"
      value={activePlanId ?? ''}
      onchange={(event) => population.setActivePlan((event.currentTarget as HTMLSelectElement).value)}
    >
      {#each plans as plan}
        <option value={plan.planId}>Plan {plan.planIndex + 1}</option>
      {/each}
    </select>
    {#if activePlanId}
      <button
        class="population-button population-icon-button"
        class:population-mark-selected-active={isActivePlanSelected}
        onclick={() => void population.setPlanSelected(activePlanId)}
        title="Mark selected"
        aria-label="Mark selected"
      >
        <Check size={14} />
      </button>
      <button
        class="population-button population-icon-button"
        disabled={plans.length <= 1}
        onclick={() => void population.deleteActivePlan()}
        title="Delete plan"
        aria-label="Delete plan"
      >
        <Trash2 size={14} />
      </button>
    {/if}
    <button
      class="population-button population-icon-button"
      onclick={() => void population.createPlan()}
      title="Add plan"
      aria-label="Add plan"
    >
      <Plus size={14} />
    </button>
  </div>
</section>
