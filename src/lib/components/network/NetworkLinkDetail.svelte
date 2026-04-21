<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Save, X } from 'lucide-svelte';
  import AttributeTableEditor from '$lib/components/attributes/AttributeTableEditor.svelte';
  import { network } from '$lib/stores/network.svelte';
  import {
    discardThenSwitch,
    saveLinkEdits,
    saveThenSwitch,
  } from '$lib/stores/network-actions';
  import { TAG_FIELDS, type TagField } from '$lib/components/network/types';
  import { createNewAttributeRow, type AttributeRow, type AttributeType } from '$lib/xml/attributes';

  let detail = $derived(network.selectedLinkDetail);
  let newCounter = 0;

  function setTag(field: TagField, value: string) {
    network.tagValues = { ...network.tagValues, [field]: value };
  }

  type Which = 'link' | 'from' | 'to';

  function getRows(which: Which): AttributeRow[] {
    if (which === 'link') return network.linkAttrRows;
    if (which === 'from') return network.fromAttrRows;
    return network.toAttrRows;
  }

  function setRows(which: Which, rows: AttributeRow[]) {
    if (which === 'link') network.linkAttrRows = rows;
    else if (which === 'from') network.fromAttrRows = rows;
    else network.toAttrRows = rows;
  }

  function addRow(which: Which) {
    setRows(which, [...getRows(which), createNewAttributeRow(newCounter++)]);
  }

  function onUpdateName(which: Which) {
    return (index: number, value: string) =>
      setRows(which, getRows(which).map((r, i) => i === index ? { ...r, name: value } : r));
  }
  function onUpdateType(which: Which) {
    return (index: number, value: AttributeType) =>
      setRows(which, getRows(which).map((r, i) => i === index ? { ...r, type: value } : r));
  }
  function onUpdateValue(which: Which) {
    return (index: number, value: string) =>
      setRows(which, getRows(which).map((r, i) => i === index ? { ...r, value } : r));
  }
  function onRemove(which: Which) {
    return (index: number) =>
      setRows(which, getRows(which).filter((_, i) => i !== index));
  }
</script>

{#if network.selectedLinkId}
  <div class="population-detail population-detail-shell">
    <header class="population-detail-header population-detail-fixed">
      <h2>{$t('network.link_heading', { values: { id: network.selectedLinkId } })}</h2>
      <div class="network-header-actions">
        <button
          class="population-close-button"
          class:network-save-active={network.hasUnsavedChanges}
          disabled={!network.hasUnsavedChanges}
          onclick={() => void saveLinkEdits()}
          aria-label={$t('network.save')}
        >
          <Save size={14} />
        </button>
        <button class="population-close-button" onclick={() => network.closeSecondary()} aria-label={$t('network.close_link_detail')}>
          <X size={14} />
        </button>
      </div>
    </header>

    <div class="population-detail-scroll">
      {#if !detail}
        <p class="population-empty">{$t('network.loading')}</p>
      {:else}
        <section class="population-section">
          <h3>{$t('network.link_tag')}</h3>
          <div class="network-tag-grid">
            {#each TAG_FIELDS as field}
              <label class="population-field">
                <span>{field}</span>
                {#if field === 'oneway'}
                  <select class="population-select" value={network.tagValues[field]} onchange={(e) => setTag(field, e.currentTarget.value)}>
                    <option value="">{$t('network.unset')}</option>
                    <option value="0">0</option>
                    <option value="1">1</option>
                  </select>
                {:else if field === 'modes'}
                  <input type="text" value={network.tagValues[field]} oninput={(e) => setTag(field, e.currentTarget.value)} />
                {:else}
                  <input type="number" step="any" value={network.tagValues[field]} oninput={(e) => setTag(field, e.currentTarget.value)} />
                {/if}
              </label>
            {/each}
          </div>
        </section>

        <section class="population-section">
          <div class="network-section-header">
            <h3>{$t('network.link_attributes')}</h3>
            <button class="population-button" onclick={() => addRow('link')}>{$t('network.add_attribute')}</button>
          </div>
          <AttributeTableEditor
            rows={network.linkAttrRows}
            onUpdateName={onUpdateName('link')}
            onUpdateType={onUpdateType('link')}
            onUpdateValue={onUpdateValue('link')}
            onRemove={onRemove('link')}
          />
        </section>

        <section class="population-section">
          <h3>{$t('network.from_node_heading', { values: { id: detail.fromNode } })}</h3>
          <div class="population-coordinates">
            <span>{$t('network.lng')} {detail.fromLng.toFixed(6)}</span>
            <span>{$t('network.lat')} {detail.fromLat.toFixed(6)}</span>
          </div>
        </section>

        <section class="population-section">
          <h3>{$t('network.to_node_heading', { values: { id: detail.toNode } })}</h3>
          <div class="population-coordinates">
            <span>{$t('network.lng')} {detail.toLng.toFixed(6)}</span>
            <span>{$t('network.lat')} {detail.toLat.toFixed(6)}</span>
          </div>
        </section>
      {/if}
    </div>
  </div>
{/if}

{#if network.switchPromptOpen}
  <div class="population-modal-backdrop">
    <div class="population-modal">
      <h3>{$t('network.unsaved_title')}</h3>
      <p>{$t('network.unsaved_message')}</p>
      <div class="population-modal-actions">
        <button class="population-button population-button-primary" onclick={() => void saveThenSwitch()}>{$t('network.save')}</button>
        <button class="population-button" onclick={() => void discardThenSwitch()}>{$t('network.discard')}</button>
        <button class="population-button" onclick={() => network.cancelSwitchPrompt()}>{$t('network.cancel')}</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .network-header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .network-save-active {
    opacity: 1;
    color: var(--color-primary);
  }
  .network-tag-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .network-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 8px;
  }
  .network-section-header h3 {
    margin: 0;
    font-size: 0.9rem;
  }
</style>
