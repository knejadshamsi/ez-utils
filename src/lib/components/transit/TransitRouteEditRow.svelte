<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Check, X } from 'lucide-svelte';

  interface Props {
    idValue: string;
    descriptionValue: string;
    submitting: boolean;
    error: string | null;
    onIdChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onSave: () => void;
    onCancel: () => void;
  }

  let {
    idValue,
    descriptionValue,
    submitting,
    error,
    onIdChange,
    onDescriptionChange,
    onSave,
    onCancel,
  }: Props = $props();

  function errorMessage(err: string | null): string | null {
    if (!err) return null;
    if (err === 'empty') return $t('transit.route_id_required');
    if (err === 'duplicate') return $t('transit.route_id_duplicate');
    return err;
  }
</script>

<div class="transit-edit-row">
  <label class="transit-edit-field transit-edit-field-id">
    <span class="transit-edit-label">{$t('transit.route_id_label')}</span>
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      class="transit-edit-input"
      placeholder={$t('transit.route_id_placeholder')}
      value={idValue}
      oninput={(e) => onIdChange((e.target as HTMLInputElement).value)}
      disabled={submitting}
      autofocus
    />
  </label>
  <label class="transit-edit-field transit-edit-field-desc">
    <span class="transit-edit-label">{$t('transit.route_description_label')}</span>
    <input
      type="text"
      class="transit-edit-input"
      placeholder={$t('transit.route_description_placeholder')}
      value={descriptionValue}
      oninput={(e) => onDescriptionChange((e.target as HTMLInputElement).value)}
      disabled={submitting}
    />
  </label>
  <button
    class="transit-icon-btn transit-icon-btn-confirm"
    onclick={() => onSave()}
    disabled={submitting || idValue.trim().length === 0}
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
    min-width: 0;
  }

  .transit-edit-field-id { flex: 1 1 0; }
  .transit-edit-field-desc { flex: 2 1 0; }

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
