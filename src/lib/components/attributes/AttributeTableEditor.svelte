<script lang="ts">
  import { Trash2 } from 'lucide-svelte';

  import type { AttributeRow, AttributeType } from '$lib/xml/attributes';

  interface Props {
    rows: AttributeRow[];
    onUpdateName: (index: number, value: string) => void;
    onUpdateType: (index: number, value: AttributeType) => void;
    onUpdateValue: (index: number, value: string) => void;
    onRemove: (index: number) => void;
  }

  let { rows, onUpdateName, onUpdateType, onUpdateValue, onRemove }: Props = $props();
</script>

<div class="attribute-list">
  {#each rows as row, index (row.key)}
    <div class="attribute-row">
      {#if row.isNew}
        <input class="attribute-input attribute-name" value={row.name} placeholder="name" oninput={(event) => onUpdateName(index, (event.currentTarget as HTMLInputElement).value)} />
      {:else}
        <span class="attribute-name-fixed">{row.name}</span>
      {/if}
      <select class="attribute-select" value={row.type} onchange={(event) => onUpdateType(index, (event.currentTarget as HTMLSelectElement).value as AttributeType)}>
        <option value="java.lang.String">String</option>
        <option value="java.lang.Integer">Integer</option>
        <option value="java.lang.Double">Double</option>
        <option value="java.lang.Boolean">Boolean</option>
      </select>
      {#if row.type === 'java.lang.Boolean'}
        <select class="attribute-select attribute-value" value={row.value} onchange={(event) => onUpdateValue(index, (event.currentTarget as HTMLSelectElement).value)}>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
      {:else}
        <input
          class="attribute-input attribute-value"
          type={row.type === 'java.lang.String' ? 'text' : 'number'}
          value={row.value}
          placeholder="value"
          oninput={(event) => onUpdateValue(index, (event.currentTarget as HTMLInputElement).value)}
        />
      {/if}
      <button class="attribute-delete" onclick={() => onRemove(index)} aria-label="Delete attribute">
        <Trash2 size={14} />
      </button>
    </div>
  {/each}
</div>

<style>
  .attribute-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 8px;
  }
  .attribute-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .attribute-name-fixed {
    min-width: 80px;
    font-size: 0.8rem;
    opacity: 0.8;
  }
  .attribute-input,
  .attribute-select {
    flex: 1;
    min-width: 0;
    padding: 6px 8px;
    border: 1px solid var(--color-base-300);
    border-radius: 6px;
    background: var(--color-base-100);
    color: var(--color-base-content);
    font-size: 0.8rem;
  }
  .attribute-name {
    flex: 0.8;
  }
  .attribute-select {
    flex: 0.6;
  }
  .attribute-value {
    flex: 1;
  }
  .attribute-delete {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    background: none;
    border: none;
    color: var(--color-base-content);
    opacity: 0.5;
    cursor: pointer;
  }
  .attribute-delete:hover {
    opacity: 1;
    color: #dc2626;
  }
</style>
