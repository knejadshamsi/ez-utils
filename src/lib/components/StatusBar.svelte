<script lang="ts">
  import { t } from 'svelte-i18n';
  import { status } from '$lib/stores/ui.svelte';

  let displayStatus = $state('');

  $effect(() => {
    if (status.isVisible) {
      displayStatus = status.current;
    }
  });
</script>

<div class="status-wrapper" class:status-visible={status.isVisible}>
  <div class="status-pill">
    {#if status.isBusy}
      <span class="status-spinner"></span>
    {/if}
    <span class="status-text">
      {$t(`status.${displayStatus}`)}{status.isBusy ? '...' : ''}
    </span>
  </div>
</div>

<style>
  .status-wrapper {
    position: fixed;
    top: 0;
    left: 50%;
    transform: translateX(-50%) translateY(-200%);
    z-index: 2600;
    pointer-events: none;
    transition: transform 0.3s ease;
  }
  .status-visible {
    transform: translateX(-50%) translateY(10px);
  }
  .status-pill {
    display: flex;
    align-items: center;
    height: 30px;
    gap: 6px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    font-size: 0.75rem;
    padding: 0 12px;
    border-radius: var(--radius-field);
    white-space: nowrap;
    box-shadow: 0 1px 5px rgba(0, 0, 0, 0.4);
  }
  .status-spinner {
    width: 10px;
    height: 10px;
    border: 2px solid var(--color-base-300);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: status-spin 0.6s linear infinite;
  }
  @keyframes status-spin {
    to { transform: rotate(360deg); }
  }
</style>
