/** A single source/layer in the source panel */
export type SourceKind = 'network' | 'population' | 'transit';

export type CrsPresetKind =
  | 'montreal_mtm8'
  | 'tehran_utm39n'
  | 'paris_lambert93'
  | 'berlin_etrs89_utm33n'
  | 'london_bng'
  | 'new_york_utm18n'
  | 'tokyo_jgd2011_zone9'
  | 'sydney_mga94_zone56'
  | 'zurich_lv95';

export type CrsConfig =
  | { kind: CrsPresetKind }
  | { kind: 'custom'; projString: string; center: [number, number]; label: string };

export interface CrsInfo {
  kind: CrsConfig;
  label: string;
  center: [number, number];
  projString: string;
}

export interface NetworkMetadata {
  networkAttributesBlob: string | null;
  linksTagBlob: string;
}

export interface TransitVehiclesMetadata {
  rootTagAttributesBlob: string;
  sourceFileName: string;
}

export interface TransitMetadata {
  transitScheduleAttributesBlob: string | null;
  transitStopsTagBlob: string;
  transitMinimalTransfersTagBlob?: string;
  transitLinesTagBlob: string;
  vehicles?: TransitVehiclesMetadata;
  linkedNetworkSourceName?: string;
}

export interface Source {
  id: string;
  name: string;
  kind: SourceKind;
  color: string;
  opacity: number;
  visible: boolean;
  networkMetadata?: NetworkMetadata;
  transitMetadata?: TransitMetadata;
}
