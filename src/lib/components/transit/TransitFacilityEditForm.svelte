<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Check, X } from 'lucide-svelte';

  import { transitEdit } from '$lib/stores/transit.svelte';
  import { submitEditStopFacility } from '$lib/stores/transit/profile-stops-actions';

  function errorMessage(err: string | null): string | null {
    if (!err) return null;
    if (err === 'empty_id') return $t('transit.stop_id_required');
    if (err === 'invalid_coords') return $t('transit.stop_coords_required');
    if (err === 'duplicate_id') return $t('transit.stop_id_duplicate');
    return err;
  }
</script>

{#if transitEdit.editStopFacilityTarget}
  {@const t0 = transitEdit.editStopFacilityTarget}
  <div class="fac-edit">
    <div class="fac-row">
      <label class="fac-field">
        <span>{$t('transit.stop_id_label')}</span>
        <input
          type="text"
          value={t0.id}
          oninput={(e) => transitEdit.updateEditStopFacility({ id: (e.target as HTMLInputElement).value, error: null })}
          disabled={t0.submitting}
        />
      </label>
      <label class="fac-field">
        <span>{$t('transit.stop_name_label')}</span>
        <input
          type="text"
          value={t0.name}
          oninput={(e) => transitEdit.updateEditStopFacility({ name: (e.target as HTMLInputElement).value })}
          disabled={t0.submitting}
        />
      </label>
    </div>
    <div class="fac-row">
      <label class="fac-field">
        <span>Lng</span>
        <input
          type="text"
          inputmode="decimal"
          value={t0.lng}
          oninput={(e) => transitEdit.updateEditStopFacility({ lng: (e.target as HTMLInputElement).value, error: null })}
          disabled={t0.submitting}
        />
      </label>
      <label class="fac-field">
        <span>Lat</span>
        <input
          type="text"
          inputmode="decimal"
          value={t0.lat}
          oninput={(e) => transitEdit.updateEditStopFacility({ lat: (e.target as HTMLInputElement).value, error: null })}
          disabled={t0.submitting}
        />
      </label>
    </div>
    <div class="fac-row">
      <label class="fac-field fac-link">
        <span>{$t('transit.stop_link_ref_label')}</span>
        <input
          type="text"
          placeholder={$t('transit.optional')}
          value={t0.linkRefId}
          oninput={(e) => transitEdit.updateEditStopFacility({ linkRefId: (e.target as HTMLInputElement).value, error: null })}
          disabled={t0.submitting}
        />
      </label>
    </div>
    <div class="fac-row">
      <label class="fac-field">
        <span>{$t('transit.stop_area_id_label')}</span>
        <input
          type="text"
          placeholder={$t('transit.optional')}
          value={t0.stopAreaId}
          oninput={(e) => transitEdit.updateEditStopFacility({ stopAreaId: (e.target as HTMLInputElement).value })}
          disabled={t0.submitting}
        />
      </label>
      <label class="fac-blocking">
        <input
          type="checkbox"
          checked={t0.isBlocking}
          onchange={(e) => transitEdit.updateEditStopFacility({ isBlocking: (e.target as HTMLInputElement).checked })}
          disabled={t0.submitting}
        />
        <span>{$t('transit.is_blocking_label')}</span>
      </label>
    </div>
    {#if t0.error}
      <div class="fac-error">{errorMessage(t0.error)}</div>
    {/if}
    <div class="fac-actions">
      <button class="fac-btn" onclick={() => transitEdit.cancelEditStopFacility()} disabled={t0.submitting}>
        <X size={12} /> {$t('transit.cancel')}
      </button>
      <button class="fac-btn fac-btn-primary" onclick={() => void submitEditStopFacility()} disabled={t0.submitting || t0.id.trim().length === 0}>
        <Check size={12} /> {$t('transit.save')}
      </button>
    </div>
  </div>
{/if}

<style>
  .fac-edit {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 6px 8px;
    background: var(--color-base-100);
    border: 1px solid var(--color-primary);
    border-radius: 5px;
  }

  .fac-row { display: flex; gap: 8px; align-items: flex-end; }

  .fac-field {
    display: flex;
    flex-direction: column;
    gap: 3px;
    flex: 1 1 0;
    min-width: 0;
    font-size: 0.65rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.7;
  }

  .fac-field.fac-link { flex: 1 1 100%; }

  .fac-field input {
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

  .fac-field input:focus { border-color: var(--color-primary); }

  .fac-blocking {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 0.72rem;
    padding-bottom: 4px;
    cursor: pointer;
    white-space: nowrap;
    opacity: 0.85;
  }

  .fac-error { font-size: 0.7rem; color: #ef4444; }

  .fac-actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
  }

  .fac-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    font-size: 0.72rem;
    font-weight: 600;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    cursor: pointer;
  }

  .fac-btn:hover:not(:disabled) { background: var(--color-base-300); }
  .fac-btn:disabled { opacity: 0.4; cursor: not-allowed; }

  .fac-btn-primary { color: var(--color-primary); border-color: var(--color-primary); }
</style>
