<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    open: boolean;
    title: string;
    confirmLabel?: string;
    cancelLabel?: string;
    submitting?: boolean;
    confirmDisabled?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
    children?: Snippet;
  }

  let {
    open,
    title,
    confirmLabel = 'Delete',
    cancelLabel = 'Cancel',
    submitting = false,
    confirmDisabled = false,
    onConfirm,
    onCancel,
    children,
  }: Props = $props();

  function onBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget && !submitting) onCancel();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && open && !submitting) onCancel();
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if open}
  <div
    class="confirm-modal-backdrop"
    role="button"
    tabindex="-1"
    aria-label="Close dialog"
    onclick={onBackdropClick}
    onkeydown={onKeydown}
  >
    <div class="confirm-modal-panel">
      <h2 class="confirm-modal-title">{title}</h2>
      <div class="confirm-modal-body">
        {@render children?.()}
      </div>
      <div class="confirm-modal-actions">
        <button class="confirm-modal-btn" disabled={submitting} onclick={onCancel}>{cancelLabel}</button>
        <button
          class="confirm-modal-btn confirm-modal-btn-danger"
          disabled={submitting || confirmDisabled}
          onclick={onConfirm}
        >{confirmLabel}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .confirm-modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 2600;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.7);
  }

  .confirm-modal-panel {
    width: 440px;
    padding: 24px;
    border-radius: 10px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  }

  .confirm-modal-title {
    margin: 0 0 12px 0;
    font-size: 1rem;
    font-weight: 700;
  }

  .confirm-modal-body {
    margin-bottom: 20px;
    font-size: 0.85rem;
    line-height: 1.5;
    color: oklch(var(--bc) / 0.75);
  }

  .confirm-modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
  }

  .confirm-modal-btn {
    padding: 6px 14px;
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--color-base-content);
    background: none;
    border: 1px solid transparent;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.1s, color 0.1s, border-color 0.1s;
  }

  .confirm-modal-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .confirm-modal-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .confirm-modal-btn-danger {
    color: #ef4444;
  }

  .confirm-modal-btn-danger:hover:not(:disabled) {
    background: none;
    border-color: #ef4444;
  }
</style>
