<script lang="ts">
  import { t } from 'svelte-i18n';
  import { MapPin } from 'lucide-svelte';

  import { transitEdit } from '$lib/stores/transit.svelte';
  import { submitCreateStopFacilityInline } from '$lib/stores/transit/profile-stops-actions';

  function errorMessage(err: string | null): string | null {
    if (!err) return null;
    if (err === 'empty_id') return $t('transit.stop_id_required');
    if (err === 'invalid_coords') return $t('transit.stop_coords_required');
    if (err === 'duplicate_id') return $t('transit.stop_id_duplicate');
    return err;
  }
</script>

{#if transitEdit.profileStopsDraft?.createOpen}
  {@const d = transitEdit.profileStopsDraft}
  <div class="transit-ps-create">
    <div class="transit-ps-create-row">
      <label class="transit-ps-field">
        <span>{$t('transit.stop_id_label')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          type="text"
          placeholder={$t('transit.stop_id_placeholder')}
          value={d.createId}
          oninput={(e) => transitEdit.setCreateField('createId', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
          autofocus
        />
      </label>
      <label class="transit-ps-field">
        <span>{$t('transit.stop_name_label')}</span>
        <input
          type="text"
          placeholder={$t('transit.optional')}
          value={d.createName}
          oninput={(e) => transitEdit.setCreateField('createName', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
        />
      </label>
    </div>
    <div class="transit-ps-create-row">
      <label class="transit-ps-field">
        <span>Lng</span>
        <input
          type="text"
          inputmode="decimal"
          placeholder="-73.567"
          value={d.createLng}
          oninput={(e) => transitEdit.setCreateField('createLng', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
        />
      </label>
      <label class="transit-ps-field">
        <span>Lat</span>
        <input
          type="text"
          inputmode="decimal"
          placeholder="45.501"
          value={d.createLat}
          oninput={(e) => transitEdit.setCreateField('createLat', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
        />
      </label>
    </div>
    <div class="transit-ps-create-row">
      <label class="transit-ps-field transit-ps-link-field">
        <span>{$t('transit.stop_link_ref_label')}</span>
        <input
          type="text"
          placeholder={$t('transit.optional')}
          value={d.createLinkRefId}
          oninput={(e) => transitEdit.setCreateField('createLinkRefId', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
        />
      </label>
    </div>
    <div class="transit-ps-create-row">
      <label class="transit-ps-field">
        <span>{$t('transit.stop_area_id_label')}</span>
        <input
          type="text"
          placeholder={$t('transit.optional')}
          value={d.createStopAreaId}
          oninput={(e) => transitEdit.setCreateField('createStopAreaId', (e.target as HTMLInputElement).value)}
          disabled={d.createSubmitting}
        />
      </label>
      <label class="transit-ps-create-blocking-check">
        <input
          type="checkbox"
          checked={d.createIsBlocking}
          onchange={(e) => transitEdit.updateProfileStopsDraft({ createIsBlocking: (e.target as HTMLInputElement).checked })}
          disabled={d.createSubmitting}
        />
        <span>{$t('transit.is_blocking_label')}</span>
      </label>
    </div>
    {#if d.createError}
      <div class="transit-ps-create-error">{errorMessage(d.createError)}</div>
    {/if}
    <div class="transit-ps-create-actions">
      <button
        class="transit-ps-btn"
        class:transit-ps-btn-active={transitEdit.mapAction === 'place_new_stop'}
        onclick={() => transitEdit.setMapAction(transitEdit.mapAction === 'place_new_stop' ? 'idle' : 'place_new_stop')}
        disabled={d.createSubmitting}
        title={$t('transit.pick_on_map')}
      >
        <MapPin size={12} />
        {transitEdit.mapAction === 'place_new_stop' ? $t('transit.click_map') : $t('transit.pick_on_map')}
      </button>
      <span class="transit-ps-create-spacer"></span>
      <button
        class="transit-ps-btn"
        onclick={() => transitEdit.toggleCreate()}
        disabled={d.createSubmitting}
      >{$t('transit.cancel')}</button>
      <button
        class="transit-ps-btn transit-ps-btn-primary"
        onclick={() => void submitCreateStopFacilityInline()}
        disabled={d.createSubmitting || d.createId.trim().length === 0}
      >{$t('transit.create')}</button>
    </div>
  </div>
{/if}

<style>
  .transit-ps-create {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 12px 12px;
    border-top: 1px solid var(--color-base-300);
    background: var(--color-base-100);
  }

  .transit-ps-create-row {
    display: flex;
    gap: 8px;
  }

  .transit-ps-field {
    display: flex;
    flex-direction: column;
    gap: 3px;
    flex: 1 1 0;
    min-width: 0;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.7;
  }

  .transit-ps-link-field { flex: 1 1 100%; }

  .transit-ps-create-blocking-check {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 0.72rem;
    padding-bottom: 4px;
    cursor: pointer;
    text-transform: none;
    letter-spacing: normal;
    white-space: nowrap;
    opacity: 0.85;
  }

  .transit-ps-field input {
    padding: 4px 8px;
    font-size: 0.78rem;
    font-weight: 400;
    text-transform: none;
    letter-spacing: normal;
    background: var(--color-base-200);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    outline: none;
    opacity: 1;
  }

  .transit-ps-field input:focus {
    border-color: var(--color-primary);
  }

  .transit-ps-create-error {
    font-size: 0.72rem;
    color: #ef4444;
  }

  .transit-ps-create-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .transit-ps-create-spacer { flex: 1; }

  .transit-ps-btn-active {
    background: var(--color-base-300);
  }

  .transit-ps-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 5px 10px;
    font-size: 0.75rem;
    font-weight: 600;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
    cursor: pointer;
  }

  .transit-ps-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-ps-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .transit-ps-btn-primary:not(:disabled) {
    border-color: var(--color-primary);
    color: var(--color-primary);
  }
</style>
