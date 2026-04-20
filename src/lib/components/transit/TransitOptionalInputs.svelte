<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from 'svelte-i18n';
  import { open } from '@tauri-apps/plugin-dialog';
  import { invoke } from '@tauri-apps/api/core';
  import { AlertCircle, Check, ChevronDown, Loader, Network, Paperclip } from 'lucide-svelte';
  import { sources } from '$lib/stores/data.svelte';
  import { ez } from '$lib/stores/ez.svelte';
  import type { TransitVehiclesMetadata } from '$lib/components/types';
  import type { AttachState } from './types';

  let attach = $state<AttachState>({ status: 'idle', errorMessage: null });
  let networkDropdownOpen = $state(false);

  let linkedNetworkName = $derived(sources.active?.transitMetadata?.linkedNetworkSourceName ?? null);
  let networkSources = $derived(sources.items.filter((s) => s.kind === 'network'));

  $effect(() => {
    const meta = sources.active?.transitMetadata?.vehicles;
    untrack(() => {
      if (meta) {
        if (attach.status !== 'success') {
          attach = { status: 'success', errorMessage: null };
        }
      } else if (attach.status === 'success') {
        attach = { status: 'idle', errorMessage: null };
      }
    });
  });

  function closeDropdownOnOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest('.transit-network-wrapper')) {
      networkDropdownOpen = false;
    }
  }

  function toggleNetworkDropdown() {
    networkDropdownOpen = !networkDropdownOpen;
  }

  function pickNetwork(name: string) {
    const src = sources.active;
    if (!src) return;
    if (!src.transitMetadata) {
      src.transitMetadata = {
        transitScheduleAttributesBlob: null,
        transitStopsTagBlob: '',
        transitLinesTagBlob: '',
      };
    }
    if (src.transitMetadata.linkedNetworkSourceName === name) {
      src.transitMetadata.linkedNetworkSourceName = undefined;
    } else {
      src.transitMetadata.linkedNetworkSourceName = name;
    }
    networkDropdownOpen = false;
    ez.queueUiPersist();
  }

  async function clickVehicles() {
    if (attach.status === 'loading') return;
    if (attach.status === 'success') {
      await detachVehicles();
      return;
    }
    await openPickerAndAttach();
  }

  async function detachVehicles() {
    const activeName = sources.activeName;
    if (!activeName) return;
    try {
      await invoke('detach_transit_vehicles_cmd', { sourceName: activeName });
      const src = sources.active;
      if (src?.transitMetadata) {
        src.transitMetadata.vehicles = undefined;
      }
      attach = { status: 'idle', errorMessage: null };
      ez.queueUiPersist();
    } catch (err) {
      attach = { status: 'error', errorMessage: extractMessage(err) };
    }
  }

  async function openPickerAndAttach() {
    const activeName = sources.activeName;
    if (!activeName) return;
    const filePath = await open({
      title: 'Select transit vehicles file',
      multiple: false,
      directory: false,
      filters: [{ name: 'MATSim vehicles XML', extensions: ['xml'] }],
    });
    if (typeof filePath !== 'string') return;

    attach = { status: 'loading', errorMessage: null };
    try {
      const meta = await invoke<TransitVehiclesMetadata>('import_transit_vehicles_cmd', {
        sourceName: activeName,
        filePath,
      });
      const src = sources.active;
      if (src) {
        src.transitMetadata = {
          ...(src.transitMetadata ?? {
            transitScheduleAttributesBlob: null,
            transitStopsTagBlob: '',
            transitLinesTagBlob: '',
          }),
          vehicles: meta,
        };
      }
      attach = { status: 'success', errorMessage: null };
      ez.queueUiPersist();
    } catch (err) {
      attach = { status: 'error', errorMessage: extractMessage(err) };
    }
  }

  function extractMessage(err: unknown): string {
    if (typeof err === 'object' && err !== null && 'message' in err) {
      const maybe = (err as { message?: unknown }).message;
      if (typeof maybe === 'string') return maybe;
    }
    return String(err);
  }
</script>

<svelte:window onclick={closeDropdownOnOutside} />

<div class="transit-optional-section">
  <div class="transit-section-label">
    {$t('transit.optional_inputs_label')}
  </div>
  <div class="transit-optional-inputs">
    <!-- Network button -->
    <div class="transit-network-wrapper">
      <button
        class="transit-optional-btn"
        class:linked={linkedNetworkName !== null}
        onclick={(e) => {
          e.stopPropagation();
          toggleNetworkDropdown();
        }}
        title={linkedNetworkName ? $t('transit.network_unlink_hint') : $t('transit.network_label')}
      >
        <span class="transit-optional-icon">
          {#if linkedNetworkName}
            <Check size={14} />
          {:else}
            <Network size={14} />
          {/if}
        </span>
        <span class="transit-optional-label">{linkedNetworkName ?? $t('transit.network_label')}</span>
        <span class="transit-optional-chev"><ChevronDown size={12} /></span>
      </button>

      {#if networkDropdownOpen}
        <div class="transit-network-dropdown" onclick={(e) => e.stopPropagation()} role="listbox" tabindex="-1">
          {#if networkSources.length === 0}
            <div class="transit-network-empty">{$t('transit.network_empty_hint')}</div>
          {:else}
            {#each networkSources as net (net.id)}
              <button
                class="transit-network-item"
                class:current={linkedNetworkName === net.name}
                onclick={() => pickNetwork(net.name)}
              >
                <span class="transit-network-swatch" style:background={net.color}></span>
                <span class="transit-network-item-name">{net.name}</span>
                {#if linkedNetworkName === net.name}
                  <Check size={12} />
                {/if}
              </button>
            {/each}
          {/if}
        </div>
      {/if}
    </div>

    <!-- Vehicles button -->
    <button
      class="transit-optional-btn"
      class:loading={attach.status === 'loading'}
      class:success={attach.status === 'success'}
      class:error={attach.status === 'error'}
      disabled={attach.status === 'loading'}
      onclick={clickVehicles}
      title={attach.status === 'success'
        ? $t('transit.detach_vehicles')
        : attach.status === 'error'
          ? (attach.errorMessage ?? $t('transit.attach_failed'))
          : $t('transit.attach_vehicles')}
    >
      <span class="transit-optional-icon" class:spin={attach.status === 'loading'}>
        {#if attach.status === 'idle'}
          <Paperclip size={14} />
        {:else if attach.status === 'loading'}
          <Loader size={14} />
        {:else if attach.status === 'success'}
          <Check size={14} />
        {:else}
          <AlertCircle size={14} />
        {/if}
      </span>
      <span class="transit-optional-label">
        {#if attach.status === 'loading'}
          {$t('transit.attaching')}
        {:else if attach.status === 'error'}
          {$t('transit.attach_failed')}
        {:else}
          {$t('transit.attach_vehicles')}
        {/if}
      </span>
    </button>
  </div>
</div>

<style>
  .transit-optional-section {
    padding: 12px 16px 8px;
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

  .transit-optional-inputs {
    display: flex;
    gap: 8px;
  }

  .transit-optional-btn {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    border-radius: 6px;
    border: 1px solid var(--color-base-300);
    background: var(--color-base-200);
    color: var(--color-base-content);
    font-size: 0.75rem;
    cursor: pointer;
    text-align: left;
    position: relative;
  }

  .transit-optional-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-optional-btn:disabled {
    cursor: default;
    opacity: 0.8;
  }

  .transit-optional-btn.linked {
    border-color: #15803d;
    color: #15803d;
  }

  .transit-optional-btn.success {
    border-color: #15803d;
    color: #15803d;
  }

  .transit-optional-btn.error {
    border-color: #b91c1c;
    color: #b91c1c;
  }

  .transit-optional-icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .transit-optional-icon.spin {
    animation: transit-spin 1s linear infinite;
  }

  .transit-optional-label {
    flex: 1;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transit-optional-chev {
    flex-shrink: 0;
    opacity: 0.6;
  }

  .transit-network-wrapper {
    flex: 1;
    min-width: 0;
    position: relative;
  }

  .transit-network-wrapper > .transit-optional-btn {
    width: 100%;
  }

  .transit-network-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    z-index: 1200;
    background: var(--color-base-100);
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    max-height: 240px;
    overflow-y: auto;
  }

  .transit-network-empty {
    padding: 10px 12px;
    font-size: 0.72rem;
    color: oklch(var(--bc) / 0.6);
    line-height: 1.35;
  }

  .transit-network-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 10px;
    background: none;
    border: none;
    color: var(--color-base-content);
    cursor: pointer;
    text-align: left;
    font-size: 0.75rem;
  }

  .transit-network-item:hover {
    background: var(--color-base-200);
  }

  .transit-network-item.current {
    background: var(--color-base-300);
    color: #15803d;
  }

  .transit-network-swatch {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    flex-shrink: 0;
  }

  .transit-network-item-name {
    flex: 1;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  @keyframes transit-spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
