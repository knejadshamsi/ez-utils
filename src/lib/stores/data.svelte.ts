// ============================================================
// Domain data stores - sources, population, network, PT
// ============================================================

import type { Source, SourceKind } from '$lib/components/types';

// ============================================================
// Sources
// ============================================================

class SourceDataStore {
  items = $state<Source[]>([]);
  activeId = $state<string | null>(null);

  readonly maxSources = 5;

  get active(): Source | null {
    if (!this.activeId) return null;
    return this.items.find(s => s.id === this.activeId) ?? null;
  }

  get activeKind(): SourceKind | null {
    return this.active?.kind ?? null;
  }

  get activeName(): string | null {
    return this.active?.name ?? null;
  }

  setActive(id: string | null) {
    this.activeId = id;
  }

  setItems(items: Source[]) {
    this.items = [...items];
  }

  addLocal(source: Source) {
    this.items = [...this.items, source];
  }

  get isFull(): boolean {
    return this.items.length >= this.maxSources;
  }

  remove(id: string) {
    this.items = this.items.filter(s => s.id !== id);
  }

  renameLocal(id: string, name: string) {
    const idx = this.items.findIndex(s => s.id === id);
    if (idx >= 0 && name.trim()) {
      this.items[idx].name = name.trim();
    }
  }

  setColor(id: string, color: string, opacity: number) {
    const idx = this.items.findIndex(s => s.id === id);
    if (idx >= 0) {
      this.items[idx].color = color;
      this.items[idx].opacity = Math.max(0.2, opacity);
    }
  }

  toggleVisibility(id: string) {
    const idx = this.items.findIndex(s => s.id === id);
    if (idx >= 0) this.items[idx].visible = !this.items[idx].visible;
  }

  moveUp(idx: number) {
    if (idx <= 0) return;
    [this.items[idx - 1], this.items[idx]] = [this.items[idx], this.items[idx - 1]];
  }

  moveDown(idx: number) {
    if (idx < 0 || idx >= this.items.length - 1) return;
    [this.items[idx], this.items[idx + 1]] = [this.items[idx + 1], this.items[idx]];
  }

  get(id: string): Source | undefined {
    return this.items.find(s => s.id === id);
  }

  indexOf(id: string): number {
    return this.items.findIndex(s => s.id === id);
  }
}

export const sources = new SourceDataStore();
