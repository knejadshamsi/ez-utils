<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Check, X } from 'lucide-svelte';

  interface Props {
    idValue: string;
    nameValue: string;
    modeValue: string;
    submitting: boolean;
    error: string | null;
    onIdChange: (value: string) => void;
    onNameChange: (value: string) => void;
    onModeChange: (value: string) => void;
    onSave: () => void;
    onCancel: () => void;
  }

  let {
    idValue,
    nameValue,
    modeValue,
    submitting,
    error,
    onIdChange,
    onNameChange,
    onModeChange,
    onSave,
    onCancel,
  }: Props = $props();

  const COMMON_MODES = ['bus', 'metro', 'subway', 'rail', 'tram', 'light_rail'];

  function errorMessage(err: string | null): string | null {
    if (!err) return null;
    if (err === 'empty') return $t('transit.line_id_required');
    if (err === 'duplicate') return $t('transit.line_id_duplicate');
    if (err === 'mode_required') return $t('transit.line_mode_required');
    return err;
  }

  const canSave = $derived(
    !submitting && idValue.trim().length > 0 && modeValue.trim().length > 0,
  );
</script>

<div class="transit-edit-row">
  <label class="transit-edit-field">
    <span class="transit-edit-label">{$t('transit.line_id_label')}</span>
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      class="transit-edit-input"
      placeholder={$t('transit.line_id_placeholder')}
      value={idValue}
      oninput={(e) => onIdChange((e.target as HTMLInputElement).value)}
      disabled={submitting}
      autofocus
    />
  </label>
  <label class="transit-edit-field">
    <span class="transit-edit-label">{$t('transit.line_name_label')}</span>
    <input
      type="text"
      class="transit-edit-input"
      placeholder={$t('transit.line_name_placeholder')}
      value={nameValue}
      oninput={(e) => onNameChange((e.target as HTMLInputElement).value)}
      disabled={submitting}
    />
  </label>
  <label class="transit-edit-field">
    <span class="transit-edit-label">{$t('transit.line_mode_label')}</span>
    <select
      class="transit-edit-input"
      value={modeValue}
      onchange={(e) => onModeChange((e.target as HTMLSelectElement).value)}
      disabled={submitting}
    >
      <option value="" disabled>{$t('transit.line_mode_placeholder')}</option>
      {#each COMMON_MODES as mode}
        <option value={mode}>{mode}</option>
      {/each}
      {#if modeValue && !COMMON_MODES.includes(modeValue)}
        <option value={modeValue}>{modeValue}</option>
      {/if}
    </select>
  </label>
  <button
    class="transit-icon-btn transit-icon-btn-confirm"
    onclick={() => onSave()}
    disabled={!canSave}
    title={$t('transit.save')}
    aria-label={$t('transit.save')}
  >
    <Check size={14} />
  </button>
  <button
    class="transit-icon-btn"
    onclick={() => onCancel()}
    disabled={submitting}
    title={$t('transit.cancel')}
    aria-label={$t('transit.cancel')}
  >
    <X size={14} />
  </button>
</div>
{#if error}
  <div class="transit-edit-error">{errorMessage(error)}</div>
{/if}

<style>
  .transit-edit-row {
    display: flex;
    align-items: flex-end;
    gap: 6px;
    padding: 6px 10px;
    background: var(--color-base-100);
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-edit-field {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1 1 0;
    min-width: 0;
  }

  .transit-edit-label {
    font-size: 0.65rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    opacity: 0.55;
  }

  .transit-edit-input {
    width: 100%;
    padding: 4px 8px;
    font-size: 0.78rem;
    background: var(--color-base-200);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 4px;
    outline: none;
  }

  .transit-edit-input:focus {
    border-color: var(--color-primary);
  }

  .transit-edit-input:disabled {
    opacity: 0.5;
  }

  .transit-edit-error {
    padding: 4px 12px 6px;
    font-size: 0.72rem;
    color: #ef4444;
    background: var(--color-base-100);
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 4px;
    border: none;
    background: none;
    color: var(--color-base-content);
    opacity: 0.6;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s, color 0.1s;
  }

  .transit-icon-btn:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-300);
  }

  .transit-icon-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .transit-icon-btn-confirm:hover:not(:disabled) {
    color: var(--color-primary);
  }
</style>
