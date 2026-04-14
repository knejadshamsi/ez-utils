import { invoke } from '@tauri-apps/api/core';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { save, open } from '@tauri-apps/plugin-dialog';

import type { Source } from '$lib/components/types';
import { sources } from '$lib/stores/data.svelte';
import { population } from '$lib/stores/population.svelte';
import type { CloseMode, ImportedSourcePayload, NewSessionPayload, StatePayload, UiState } from '$lib/stores/types';
import {
  drawers,
  mapViewport,
  settings,
  sourceStore,
  status,
} from '$lib/stores/ui.svelte';

class EzStore {
  active = $state(false);
  dirty = $state(false);
  filePath = $state<string | null>(null);
  error = $state<string | null>(null);
  closeConfirmOpen = $state(false);

  private persistTimer: number | null = null;
  private autosaveTimer: number | null = null;
  private hydrating = false;
  private closeResolver: ((mode: CloseMode | null) => void) | null = null;
  private allowCloseOnce = false;
  private lastAutosaveKey: string | null = null;

  async promptOpen() {
    const selected = await open({
      title: 'Open .ez File',
      multiple: false,
      directory: false,
      filters: [{ name: 'EZ File', extensions: ['ez'] }],
    });
    if (typeof selected !== 'string') return;
    await this.openEz(selected);
  }

  async promptNewFromXml() {
    const xmlPath = await open({
      title: 'Select XML File',
      multiple: false,
      directory: false,
      filters: [{ name: 'MATSim XML', extensions: ['xml', 'xml.gz'] }],
    });
    if (typeof xmlPath !== 'string') return;

    this.error = null;
    try {
      await invoke('validate_xml', { filePath: xmlPath });
    } catch (err) {
      this.handleError(err);
      return;
    }

    const ezPath = await save({
      title: 'Save .ez File',
      filters: [{ name: 'EZ File', extensions: ['ez'] }],
    });
    if (typeof ezPath !== 'string') return;

    await this.newFromXml(xmlPath, ezPath);
  }

  async promptImportSource() {
    if (!this.active || sources.isFull) return;
    const selected = await open({
      title: 'Import Source',
      multiple: false,
      directory: false,
    });
    if (typeof selected !== 'string') return;
    await this.importSource(selected);
  }

  async save() {
    if (!this.active || !this.filePath) return;
    status.setOperation('saving');
    try {
      await this.flushPendingUiState();
      const payload = await invoke<StatePayload>('save_ez');
      this.applyState(payload);
      status.setOperation(null);
      status.flash('saved');
    } catch (err) {
      this.handleError(err);
      status.setOperation(null);
    }
  }

  async importSource(filePath: string) {
    this.error = null;
    status.setOperation('importing');
    try {
      const payload = await invoke<ImportedSourcePayload>('import_source', { filePath });
      const hue = Math.floor(Math.random() * 360);
      const source: Source = {
        id: crypto.randomUUID(),
        name: payload.name,
        kind: payload.kind,
        color: `hsl(${hue}, 70%, 55%)`,
        opacity: 1,
        visible: true,
        networkMetadata: payload.networkMetadata,
        transitMetadata: payload.transitMetadata,
      };
      sources.addLocal(source);
      this.dirty = true;
      await this.persistUiStateNow();
      if (payload.kind === 'population' && population.viewport) {
        void population.fetchCurrentPage();
      }
    } catch (err) {
      this.handleError(err);
    } finally {
      status.setOperation(null);
    }
  }

  async renameSource(id: string, nextName: string) {
    const source = sources.get(id);
    if (!source) return;
    if (!nextName.trim()) return;

    try {
      await invoke('rename_source', { sourceName: source.name, newName: nextName });
      sources.renameLocal(id, nextName);
      this.dirty = true;
      await this.persistUiStateNow();
    } catch (err) {
      this.handleError(err);
    }
  }

  async removeSource(id: string) {
    const source = sources.get(id);
    if (!source) return;

    try {
      await invoke('remove_source', { sourceName: source.name });
      const wasNetwork = source.kind === 'network';
      const removedName = source.name;
      sources.remove(id);
      if (wasNetwork) {
        for (const other of sources.items) {
          if (other.kind === 'transit' && other.transitMetadata?.linkedNetworkSourceName === removedName) {
            other.transitMetadata.linkedNetworkSourceName = undefined;
          }
        }
      }
      this.dirty = true;
      await this.persistUiStateNow();
    } catch (err) {
      this.handleError(err);
    }
  }

  queueUiPersist() {
    if (!this.active || this.hydrating || !this.filePath) return;

    this.dirty = true;
    if (this.persistTimer !== null) {
      window.clearTimeout(this.persistTimer);
    }

    this.persistTimer = window.setTimeout(async () => {
      await this.persistUiStateNow();
    }, 500);
  }

  refreshAutosave() {
    this.configureAutosaveIfNeeded();
  }

  async flushPendingUiState() {
    await this.persistUiStateNow();
  }

  async handleWindowCloseRequest(event: { preventDefault: () => void }) {
    if (this.allowCloseOnce) {
      this.allowCloseOnce = false;
      return;
    }

    if (!this.active || !this.filePath) return;

    if (!this.dirty) {
      event.preventDefault();
      try {
        await this.flushPendingUiState();
        await invoke('close_ez', { mode: 'discard' as CloseMode });
      } catch { /* close anyway */ }
      this.reset();
      this.allowCloseOnce = true;
      await getCurrentWindow().close();
      return;
    }

    event.preventDefault();
    const decision = await this.waitForCloseDecision();
    if (!decision) return;

    try {
      await this.flushPendingUiState();
      await invoke('close_ez', { mode: decision });
    } catch { /* close anyway */ }
    this.reset();
    this.allowCloseOnce = true;
    await getCurrentWindow().close();
  }

  requestCloseDecision(mode: CloseMode | null) {
    this.closeConfirmOpen = false;
    this.closeResolver?.(mode);
    this.closeResolver = null;
  }

  snapshotUiState(): UiState {
    return {
      version: 1,
      sources: sources.items.map((source: Source) => ({ ...source })),
      settings: { ...settings.config, autosaveSettings: { ...settings.config.autosaveSettings } },
      sourcePopoverOpen: sourceStore.popoverOpen,
      openDrawers: drawers.snapshot(),
      mapState: { ...mapViewport.state },
    };
  }

  private async openEz(path: string) {
    this.error = null;
    this.active = true;
    status.setOperation('unpacking');
    try {
      await invoke('unpack_ez', { ezPath: path });
      status.setOperation('loading');
      const payload = await invoke<StatePayload>('load_ez', { ezPath: path });
      this.applyState(payload);
      if (population.activeSourceName && population.viewport) {
        void population.fetchCurrentPage();
      }
    } catch (err) {
      this.active = false;
      this.handleError(err);
    } finally {
      status.setOperation(null);
    }
  }

  private async newFromXml(xmlPath: string, ezPath: string) {
    this.error = null;
    this.active = true;
    status.setOperation('importing');
    try {
      const payload = await invoke<NewSessionPayload>('new_from_xml', {
        xmlPath,
        ezPath: ezPath.endsWith('.ez') ? ezPath : `${ezPath}.ez`,
        crs: settings.config.crs,
      });
      this.applyState(payload.state);
      const hue = Math.floor(Math.random() * 360);
      const source: Source = {
        id: crypto.randomUUID(),
        name: payload.imported.name,
        kind: payload.imported.kind,
        color: `hsl(${hue}, 70%, 55%)`,
        opacity: 1,
        visible: true,
        networkMetadata: payload.imported.networkMetadata,
        transitMetadata: payload.imported.transitMetadata,
      };
      sources.addLocal(source);
      await this.persistUiStateNow();
      if (payload.imported.kind === 'population' && population.viewport) {
        void population.fetchCurrentPage();
      }
    } catch (err) {
      this.active = false;
      this.handleError(err);
    } finally {
      status.setOperation(null);
    }
  }

  private async closeEz(mode: CloseMode, shouldFlashSaved: boolean) {
    try {
      await this.flushPendingUiState();
      await invoke('close_ez', { mode });
      if (shouldFlashSaved) {
        status.flash('saved');
      }
      this.reset();
      this.allowCloseOnce = true;
      await getCurrentWindow().close();
    } catch (err) {
      this.handleError(err);
    }
  }

  private async persistUiStateNow() {
    if (!this.active || this.hydrating || !this.filePath) return;
    if (this.persistTimer !== null) {
      window.clearTimeout(this.persistTimer);
      this.persistTimer = null;
    }

    try {
      const payload = await invoke<StatePayload>('save_ui_state', {
        uiState: this.snapshotUiState(),
      });
      this.applyState(payload, false);
    } catch (err) {
      this.handleError(err);
    }
  }

  private applyState(payload: StatePayload, hydrateStores = true) {
    const wasActive = this.active;
    this.filePath = payload.ezPath;
    this.dirty = payload.dirty;
    this.active = true;

    if (hydrateStores) {
      this.hydrating = true;
      settings.hydrate(payload.uiState.settings);
      sourceStore.hydrate(payload.uiState.sourcePopoverOpen);
      drawers.hydrate(payload.uiState.openDrawers);
      mapViewport.hydrate(payload.uiState.mapState);
      sources.setItems(payload.uiState.sources);
      this.hydrating = false;
    }

    this.configureAutosaveIfNeeded(!wasActive);
  }

  private reset() {
    this.active = false;
    this.dirty = false;
    this.filePath = null;
    this.error = null;
    this.lastAutosaveKey = null;
    sources.setItems([]);
    sourceStore.hydrate(false);
    drawers.hydrate([]);
    this.stopAutosave();
  }

  private autosaveKey() {
    const autosave = settings.config.autosaveSettings;
    return `${this.active}:${autosave.enabled}:${autosave.intervalMinutes}`;
  }

  private configureAutosaveIfNeeded(force = false) {
    const nextKey = this.autosaveKey();
    if (!force && nextKey === this.lastAutosaveKey) return;
    this.lastAutosaveKey = nextKey;
    this.configureAutosave();
  }

  private configureAutosave() {
    this.stopAutosave();
    const autosave = settings.config.autosaveSettings;
    if (!this.active || !autosave.enabled) return;

    this.autosaveTimer = window.setInterval(async () => {
      if (!this.dirty) return;
      try {
        await this.flushPendingUiState();
        const payload = await invoke<StatePayload>('save_ez');
        this.applyState(payload);
        status.flash('auto_saved');
      } catch (err) {
        this.handleError(err);
      }
    }, autosave.intervalMinutes * 60 * 1000);
  }

  private stopAutosave() {
    if (this.autosaveTimer !== null) {
      window.clearInterval(this.autosaveTimer);
      this.autosaveTimer = null;
    }
  }

  private waitForCloseDecision(): Promise<CloseMode | null> {
    this.closeConfirmOpen = true;
    return new Promise((resolve) => {
      this.closeResolver = resolve;
    });
  }

  private handleError(rawError: unknown) {
    if (typeof rawError === 'string') {
      this.error = rawError;
      return;
    }
    const err = rawError as { kind?: string; message?: string; source_kind?: string; limit?: number; path?: string; version?: number };
    switch (err.kind) {
      case 'session_busy':
        this.error = 'Another operation is in progress.';
        break;
      case 'archive_missing':
        this.error = 'The .ez file could not be found.';
        break;
      case 'archive_corrupt':
        this.error = 'The .ez file is corrupt and could not be opened.';
        break;
      case 'unsupported_ui_version':
        this.error = `UI state version ${err.version} is not supported.`;
        break;
      case 'unwritable_location':
        this.error = `Location must be writable: ${err.path}`;
        break;
      case 'source_name_invalid':
        this.error = err.message ?? 'Invalid source name.';
        break;
      case 'import_not_implemented':
        this.error = `Import for ${err.source_kind} is not implemented yet.`;
        break;
      case 'source_limit_reached':
        this.error = `At most ${err.limit} sources are supported.`;
        break;
      case 'unsupported_source':
        this.error = err.message ?? 'Unsupported source file.';
        break;
      case 'io':
        this.error = err.message ?? 'An I/O error occurred.';
        break;
      default:
        this.error = 'Operation failed.';
        break;
    }
  }
}

export const ez = new EzStore();
