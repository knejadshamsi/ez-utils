<script lang="ts">
  import { Trash2 } from 'lucide-svelte';

  import type { PopulationActivityEditorProps } from '$lib/components/population/types';
  import { population } from '$lib/stores/population.svelte';

  let { plan }: PopulationActivityEditorProps = $props();
</script>

<section class="population-section">
  <div class="population-section-header">
    <h3>Activities</h3>
  </div>

  <div class="population-card-list">
    {#each plan.activities as activity, index}
      <article class="population-card population-activity-compact">
        <div class="population-header-row">
          <span class="population-card-title">Activity {index + 1}</span>
          <button class="population-button population-icon-button" disabled={plan.activities.length <= 1} onclick={() => population.deleteActivity(index)} aria-label="Delete activity">
            <Trash2 size={14} />
          </button>
        </div>
        <div class="population-inline-actions">
          <label class="population-field flex-1">
            <span>Type</span>
            <select class="population-select" value={activity.type} onchange={(event) => population.updateActivityField(index, 'type', (event.currentTarget as HTMLSelectElement).value)}>
              <option value="home">Home</option>
              <option value="work">Work</option>
              <option value="school">School</option>
              <option value="shop">Shop</option>
              <option value="eat">Eat</option>
              <option value="recreation">Recreation</option>
              <option value="other">Other</option>
            </select>
          </label>
        </div>
        <div class="population-inline-actions">
          <label class="population-field flex-1">
            <span>Start</span>
            <input
              type="time"
              value={activity.startTime}
              oninput={(event) => population.updateActivityField(index, 'startTime', (event.currentTarget as HTMLInputElement).value)}
            />
          </label>
          <label class="population-field flex-1">
            <span>End</span>
            <input
              type="time"
              value={activity.endTime}
              oninput={(event) => population.updateActivityField(index, 'endTime', (event.currentTarget as HTMLInputElement).value)}
            />
          </label>
        </div>
        <div class="population-inline-actions">
          <label class="population-field flex-1">
            <span>Lng</span>
            <input
              type="number"
              step="0.000001"
              value={activity.lng?.toFixed(6) ?? ''}
              onchange={(event) => {
                const lng = parseFloat((event.currentTarget as HTMLInputElement).value);
                if (!isNaN(lng)) population.moveActivity(index, lng, activity.lat ?? 0);
              }}
            />
          </label>
          <label class="population-field flex-1">
            <span>Lat</span>
            <input
              type="number"
              step="0.000001"
              value={activity.lat?.toFixed(6) ?? ''}
              onchange={(event) => {
                const lat = parseFloat((event.currentTarget as HTMLInputElement).value);
                if (!isNaN(lat)) population.moveActivity(index, activity.lng ?? 0, lat);
              }}
            />
          </label>
        </div>
      </article>
      {#if plan.legs[index]}
        {@const leg = plan.legs[index]}
        <div class="population-leg-wrap">
          <div class="population-leg-connector"></div>
          <article class="population-card population-leg-compact">
            <div class="population-inline-actions">
              <label class="population-field flex-1">
                <span>Mode</span>
                <select class="population-select" value={leg.mode} onchange={(event) => population.updateLegField(index, 'mode', (event.currentTarget as HTMLSelectElement).value)}>
                  <option value="car">Car</option>
                  <option value="walk">Walk</option>
                  <option value="bus">Bus</option>
                  <option value="metro">Metro</option>
                  <option value="tram">Tram</option>
                  <option value="bike">Bike</option>
                  <option value="pt">Public Transit</option>
                </select>
              </label>
              <label class="population-field" style="width: 90px;">
                <span>Duration</span>
                <div class="population-inline-actions">
                  <input
                    type="number"
                    min="0"
                    value={leg.durationMinutes}
                    oninput={(event) => population.updateLegField(index, 'durationMinutes', (event.currentTarget as HTMLInputElement).value)}
                  />
                  <span class="population-unit">min</span>
                </div>
              </label>
            </div>
          </article>
        </div>
      {/if}
    {/each}
  </div>
</section>
