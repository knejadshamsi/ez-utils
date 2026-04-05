// ============================================================
// Domain data stores - sources, population, network, PT
// ============================================================

import type { Source } from '$lib/components/types';

// ============================================================
// Sources
// ============================================================

class SourceDataStore {
  items = $state<Source[]>([
    { id: '1', name: 'Montreal Network', color: '#3b82f6', opacity: 1, visible: true },
    { id: '2', name: 'STM Transit', color: '#f59e0b', opacity: 1, visible: true },
    { id: '3', name: 'Census 2021', color: '#10b981', opacity: 0.6, visible: true },
  ]);

  readonly maxSources = 5;

  add(): string | null {
    if (this.items.length >= this.maxSources) return null;
    const id = crypto.randomUUID();
    const hue = Math.floor(Math.random() * 360);
    this.items = [...this.items, { id, name: `Source ${this.items.length + 1}`, color: `hsl(${hue}, 70%, 55%)`, opacity: 1, visible: true }];
    return id;
  }

  get isFull(): boolean {
    return this.items.length >= this.maxSources;
  }

  remove(id: string) {
    this.items = this.items.filter(s => s.id !== id);
  }

  rename(id: string, name: string) {
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
