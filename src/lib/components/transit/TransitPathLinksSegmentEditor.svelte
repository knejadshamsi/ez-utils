<script lang="ts">
  import { invoke } from '@tauri-apps/api/core';
  import { t } from 'svelte-i18n';
  import { Check, ChevronRight, X } from 'lucide-svelte';

  import { sources } from '$lib/stores/data.svelte';
  import { transit } from '$lib/stores/transit.svelte';
  import type { ListedPathLink } from '$lib/stores/transit/types';

  interface Props {
    fromLinkRef: string | null;
    toLinkRef: string | null;
    lineId: string;
    routeId: string;
  }

  let { fromLinkRef, toLinkRef, lineId, routeId }: Props = $props();

  let expanded = $state(false);
  let editValue = $state('');
  let saving = $state(false);
  let error = $state<string | null>(null);

  const allLinks = $derived(
    transit.pathLinks
      .slice()
      .sort((a, b) => a.sequence - b.sequence)
      .map((p) => p.linkId),
  );

  const segmentBounds = $derived.by<{ fromIdx: number; toIdx: number } | null>(() => {
    if (fromLinkRef === null || toLinkRef === null) return null;
    const fromIdx = allLinks.indexOf(fromLinkRef);
    const toIdx = allLinks.indexOf(toLinkRef);
    if (fromIdx < 0 || toIdx < 0 || fromIdx > toIdx) return null;
    return { fromIdx, toIdx };
  });

  const segment = $derived.by<string[]>(() => {
    const b = segmentBounds;
    return b ? allLinks.slice(b.fromIdx, b.toIdx + 1) : [];
  });

  function toggleExpand() {
    if (!segmentBounds) return;
    if (!expanded) {
      editValue = segment.join('\n');
      error = null;
    }
    expanded = !expanded;
  }

  async function saveSegment() {
    const b = segmentBounds;
    if (!b) return;
    const active = sources.active;
    if (!active || active.kind !== 'transit') return;

    const newSegmentLinks = editValue
      .split('\n')
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
    if (newSegmentLinks.length === 0) {
      error = $t('transit.segment_empty_error');
      return;
    }

    // Rebuild the full list: prefix + new segment + suffix. Dedupe seams where
    // the boundary link repeats (common pattern for through-link sequences).
    const prefix = allLinks.slice(0, b.fromIdx);
    const suffix = allLinks.slice(b.toIdx + 1);
    const rebuilt: string[] = [];
    for (const link of [...prefix, ...newSegmentLinks, ...suffix]) {
      if (rebuilt.length === 0 || rebuilt[rebuilt.length - 1] !== link) {
        rebuilt.push(link);
      }
    }

    saving = true;
    error = null;
    try {
      await invoke('apply_route_path_links_cmd', {
        sourceName: active.name,
        lineId,
        routeId,
        links: rebuilt,
      });
      await transit.loadPathLinksForRoute(active.name, lineId, routeId);
      await transit.loadLinkedRoutePath(
        active.name,
        active.transitMetadata?.linkedNetworkSourceName ?? '',
        lineId,
        routeId,
      ).catch(() => void 0);
      expanded = false;
    } catch (e) {
      const err = e as { kind?: string; message?: string };
      error = err?.message ?? 'Save failed';
    } finally {
      saving = false;
    }
  }

  function cancelEdit() {
    expanded = false;
    editValue = segment.join('\n');
    error = null;
  }
</script>

<div class="transit-segment">
  {#if segmentBounds === null}
    <span class="transit-segment-label transit-segment-unknown">—</span>
  {:else}
    <button
      class="transit-segment-label"
      class:transit-segment-open={expanded}
      onclick={toggleExpand}
      disabled={saving}
    >
      <span class="transit-segment-chev" class:transit-segment-chev-open={expanded}>
        <ChevronRight size={10} />
      </span>
      {$t('transit.links_count', { values: { count: segment.length } })}
    </button>
  {/if}
  {#if expanded}
    <div class="transit-segment-panel">
      <textarea
        class="transit-segment-textarea"
        rows={Math.min(8, segment.length + 1)}
        value={editValue}
        oninput={(e) => {
          editValue = (e.target as HTMLTextAreaElement).value;
          error = null;
        }}
        disabled={saving}
        placeholder={$t('transit.one_link_per_line')}
      ></textarea>
      {#if error}<div class="transit-segment-error">{error}</div>{/if}
      <div class="transit-segment-actions">
        <button class="transit-segment-btn" onclick={cancelEdit} disabled={saving}>
          <X size={11} /> {$t('transit.cancel')}
        </button>
        <button
          class="transit-segment-btn transit-segment-btn-primary"
          onclick={() => void saveSegment()}
          disabled={saving}
        >
          <Check size={11} /> {$t('transit.save')}
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .transit-segment {
    padding: 2px 24px 2px 28px;
    font-size: 0.68rem;
  }

  .transit-segment-label {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 1px 6px;
    font-size: 0.65rem;
    font-weight: 500;
    color: var(--color-base-content);
    background: none;
    border: 1px dashed var(--color-base-300);
    border-radius: 3px;
    opacity: 0.6;
    cursor: pointer;
    transition: opacity 0.1s, background 0.1s, border-color 0.1s;
  }

  .transit-segment-label:hover:not(:disabled) {
    opacity: 1;
    background: var(--color-base-200);
    border-color: var(--color-base-content);
  }

  .transit-segment-chev {
    display: inline-flex;
    align-items: center;
    transition: transform 0.15s;
  }

  .transit-segment-chev-open {
    transform: rotate(90deg);
  }

  .transit-segment-label.transit-segment-open {
    opacity: 1;
    border-style: solid;
    border-color: var(--color-primary);
    color: var(--color-primary);
  }

  .transit-segment-unknown {
    opacity: 0.35;
    cursor: default;
  }

  .transit-segment-panel {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 6px;
    background: var(--color-base-100);
    border: 1px solid var(--color-primary);
    border-radius: 4px;
  }

  .transit-segment-textarea {
    font-family: monospace;
    font-size: 0.72rem;
    padding: 4px 6px;
    background: var(--color-base-200);
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 3px;
    outline: none;
    resize: vertical;
  }

  .transit-segment-textarea:focus { border-color: var(--color-primary); }

  .transit-segment-error {
    font-size: 0.68rem;
    color: #ef4444;
  }

  .transit-segment-actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
  }

  .transit-segment-btn {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 7px;
    font-size: 0.68rem;
    font-weight: 600;
    background: none;
    color: var(--color-base-content);
    border: 1px solid var(--color-base-300);
    border-radius: 3px;
    cursor: pointer;
  }

  .transit-segment-btn:hover:not(:disabled) {
    background: var(--color-base-300);
  }

  .transit-segment-btn-primary {
    color: var(--color-primary);
    border-color: var(--color-primary);
  }
</style>
