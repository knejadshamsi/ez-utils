<script lang="ts">
  import PopulationPersonDetail from '$lib/components/population/PopulationPersonDetail.svelte';
  import { population } from '$lib/stores/population.svelte';
</script>

<div class="population-panel relative">
  {#if population.personLoading}
    <p class="population-empty">Loading person...</p>
  {:else if population.person}
    <PopulationPersonDetail />
  {:else}
    <p class="population-empty">Select a person from the map or the primary drawer.</p>
  {/if}

  {#if population.switchPromptOpen}
    <div class="population-modal-backdrop">
      <div class="population-modal">
        <h3>Unsaved Changes</h3>
        <p>You have unsaved changes. Save before switching?</p>
        <div class="population-inline-actions">
          <button class="population-button population-button-primary" onclick={() => void population.saveThenSwitch()}>
            Save
          </button>
          <button class="population-button" onclick={() => void population.discardThenSwitch()}>
            Discard
          </button>
          <button class="population-button" onclick={() => population.cancelSwitchPrompt()}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  {/if}

</div>
