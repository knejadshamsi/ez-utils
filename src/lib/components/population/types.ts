export interface PopulationActivityCoord {
  personId: string;
  planId: string;
  planIndex: number;
  activityIndex: number;
  lng: number;
  lat: number;
}

export interface PopulationQueryResponse {
  rows: PopulationActivityCoord[];
  totalRows: number;
  totalPeople: number;
}

export interface PopulationPlanPayload {
  planId: string;
  planIndex: number;
  selected: number;
  planBlob: string;
}

export interface PopulationPersonPayload {
  personId: string;
  attributesBlob: string | null;
  plans: PopulationPlanPayload[];
}

export interface PopulationViewportBounds {
  west: number;
  south: number;
  east: number;
  north: number;
}

export interface PopulationQueryPoint extends PopulationActivityCoord {
  sourceName: string;
  markerKey: string;
}

export interface PopulationPersonSummary {
  key: string;
  sourceName: string;
  personId: string;
  pointCount: number;
}

export interface PopulationSelection {
  sourceName: string;
  personId: string;
}

export interface EditableActivity {
  key: string;
  originalIndex: number | null;
  type: string;
  startTime: string;
  endTime: string;
  lng: number | null;
  lat: number | null;
  dirty: boolean;
}

export interface EditableLeg {
  key: string;
  originalIndex: number | null;
  mode: string;
  durationMinutes: number;
  dirty: boolean;
}

export interface EditablePlanDraft {
  planId: string;
  planIndex: number;
  selected: boolean;
  planBlob: string;
  activities: EditableActivity[];
  legs: EditableLeg[];
}

export interface PlanActivityEditInput {
  originalIndex: number | null;
  typeName: string;
  startTime: string | null;
  endTime: string | null;
  lng: number | null;
  lat: number | null;
  dirty: boolean;
}

export interface PlanLegEditInput {
  originalIndex: number | null;
  mode: string;
  travelTime: string | null;
  dirty: boolean;
}

export interface ApplyPlanEditsInput {
  activities: PlanActivityEditInput[];
  legs: PlanLegEditInput[];
}

export interface PopulationActivityEditorProps {
  plan: EditablePlanDraft;
}

export interface PopulationPlanEditorProps {
  plans: EditablePlanDraft[];
  activePlanId: string | null;
}

import type { AttributeRow, AttributeType } from '$lib/xml/attributes';

/** @deprecated Use `AttributeType` from `$lib/xml/attributes` directly. */
export type PopulationAttributeType = AttributeType;

/** @deprecated Use `AttributeRow` from `$lib/xml/attributes` directly. */
export type PopulationAttributeRow = AttributeRow;

export interface PopulationPromptState {
  open: boolean;
  target: PopulationSelection | null;
}

export type PopulationSaveState = 'idle' | 'saving' | 'saved';

export type PopulationMapAction = 'add_person' | 'add_activity' | null;

export type PopulationTab = 'plans' | 'attributes';
