export interface NetworkNodePayload {
  id: string;
  lng: number;
  lat: number;
  ghost: boolean;
}

export interface NetworkLinkPayload {
  id: string;
  fromNode: string;
  toNode: string;
  fromLng: number;
  fromLat: number;
  toLng: number;
  toLat: number;
}

export interface NetworkBboxResult {
  nodes: NetworkNodePayload[];
  links: NetworkLinkPayload[];
  nodeTotal: number;
  linkTotal: number;
  nodeCapExceeded: boolean;
  linkCapExceeded: boolean;
}

export interface NetworkSearchLinkResult {
  links: NetworkLinkPayload[];
  connectedNodes: NetworkNodePayload[];
  total: number;
  capExceeded: boolean;
}

export interface NetworkSearchNodeResult {
  nodes: NetworkNodePayload[];
  connectedLinks: NetworkLinkPayload[];
  total: number;
  capExceeded: boolean;
}

export interface NetworkLinkDetailPayload {
  id: string;
  fromNode: string;
  toNode: string;
  fromLng: number;
  fromLat: number;
  toLng: number;
  toLat: number;
  tagBlob: string;
  attributesBlob: string | null;
  fromNodeAttributesBlob: string | null;
  toNodeAttributesBlob: string | null;
}

export type TagField = 'length' | 'freespeed' | 'capacity' | 'permlanes' | 'oneway' | 'modes';
export const TAG_FIELDS: TagField[] = ['length', 'freespeed', 'capacity', 'permlanes', 'oneway', 'modes'];
export type TagValues = Record<TagField, string>;
export function emptyTagValues(): TagValues {
  return { length: '', freespeed: '', capacity: '', permlanes: '', oneway: '', modes: '' };
}

export interface LinkTagEditInput {
  length: string | null;
  freespeed: string | null;
  capacity: string | null;
  permlanes: string | null;
  oneway: string | null;
  modes: string | null;
}

export interface CreateNodeInput {
  id: string;
  lng: number;
  lat: number;
}

export interface CreateLinkInput {
  id: string;
  fromNode: string;
  toNode: string;
}

export interface DeleteNodeResult {
  deletedLinkIds: string[];
}

export type NetworkMapAction = 'idle' | 'add_node' | 'add_link_from' | 'add_link_to';

export type NetworkSelection =
  | { type: 'node'; id: string }
  | { type: 'link'; id: string }
  | null;

export interface NetworkViewportBounds {
  west: number;
  south: number;
  east: number;
  north: number;
}
