<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Bus, TrainFront, TramFront } from 'lucide-svelte';
  import { transit } from '$lib/stores/transit.svelte';
</script>

<div class="transit-modes-section">
  <div class="transit-section-label">{$t('transit.modes_label')}</div>
  <div class="transit-mode-row">
    <button
      class="transit-mode-btn"
      class:on={transit.modes.bus}
      onclick={() => transit.toggleMode('bus')}
      title={transit.modes.bus
        ? $t('transit.mode_hint_on', { values: { mode: $t('transit.mode_bus') } })
        : $t('transit.mode_hint_off', { values: { mode: $t('transit.mode_bus') } })}
    >
      <Bus size={14} />
      <span>{$t('transit.mode_bus')}</span>
    </button>
    <button
      class="transit-mode-btn"
      class:on={transit.modes.metro}
      onclick={() => transit.toggleMode('metro')}
      title={transit.modes.metro
        ? $t('transit.mode_hint_on', { values: { mode: $t('transit.mode_metro') } })
        : $t('transit.mode_hint_off', { values: { mode: $t('transit.mode_metro') } })}
    >
      <TrainFront size={14} />
      <span>{$t('transit.mode_metro')}</span>
    </button>
    <button
      class="transit-mode-btn"
      class:on={transit.modes.tram}
      onclick={() => transit.toggleMode('tram')}
      title={transit.modes.tram
        ? $t('transit.mode_hint_on', { values: { mode: $t('transit.mode_tram') } })
        : $t('transit.mode_hint_off', { values: { mode: $t('transit.mode_tram') } })}
    >
      <TramFront size={14} />
      <span>{$t('transit.mode_tram')}</span>
    </button>
  </div>
</div>

<style>
  .transit-modes-section {
    padding: 10px 16px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    border-bottom: 1px solid var(--color-base-300);
  }

  .transit-section-label {
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: oklch(var(--bc) / 0.55);
  }

  .transit-mode-row {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  .transit-mode-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid var(--color-base-300);
    background: var(--color-base-200);
    color: var(--color-base-content);
    font-size: 0.72rem;
    cursor: pointer;
    opacity: 0.4;
    transition: opacity 0.15s, background 0.15s;
  }

  .transit-mode-btn:hover {
    background: var(--color-base-300);
  }

  .transit-mode-btn.on {
    opacity: 1;
  }
</style>
