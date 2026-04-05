/** A single source/layer in the source panel */
export type SourceKind = 'network' | 'population' | 'transit';

export interface Source {
  id: string;
  name: string;
  kind: SourceKind;
  color: string;
  opacity: number;
  visible: boolean;
}
