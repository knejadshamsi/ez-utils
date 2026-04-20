import type { NetworkMetadata, Source, SourceKind, TransitMetadata } from '$lib/components/types';
import type { MapViewportState, SettingsConfig } from '$lib/stores/ui.svelte';

export interface UiState {
  version: number;
  sources: Source[];
  settings: SettingsConfig;
  sourcePopoverOpen: boolean;
  openDrawers: string[];
  mapState: MapViewportState;
}

export interface StatePayload {
  ezPath: string;
  dirty: boolean;
  uiState: UiState;
}

export interface ImportedSourcePayload {
  name: string;
  kind: SourceKind;
  networkMetadata?: NetworkMetadata;
  transitMetadata?: TransitMetadata;
}

export interface NewSessionPayload {
  state: StatePayload;
  imported: ImportedSourcePayload;
}

export type CloseMode = 'save' | 'discard';
