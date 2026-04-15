import proj4 from 'proj4';

import type {
  ApplyPlanEditsInput,
  EditableActivity,
  EditableLeg,
  EditablePlanDraft,
  PlanLegEditInput,
  PopulationPlanPayload,
} from '$lib/components/population/types';
import { parseXml, requireRoot } from '$lib/xml/attributes';

// Re-export shared attribute helpers so existing population callers keep working.
export {
  createNewAttributeRow,
  normalizeAttributeType,
  parseAttributesBlob,
  parseXml,
  requireRoot,
  serializeAttributes,
} from '$lib/xml/attributes';

const WGS84_CRS = 'WGS84';

export function parsePlanBlob(plan: PopulationPlanPayload, projString: string): EditablePlanDraft {
  const doc = parseXml(plan.planBlob);
  const root = requireRoot(doc, 'plan');
  const activities = Array.from(root.getElementsByTagName('activity')).map((node, index) => ({
    ...(projectActivityCoords(node, projString)),
    key: `activity-${plan.planId}-${index}`,
    originalIndex: index,
    type: node.getAttribute('type') ?? 'other',
    startTime: toUiTime(node.getAttribute('start_time')),
    endTime: toUiTime(node.getAttribute('end_time')),
    dirty: false,
  }));
  const legs = Array.from(root.getElementsByTagName('leg')).map((node, index) => ({
    key: `leg-${plan.planId}-${index}`,
    originalIndex: index,
    mode: node.getAttribute('mode') ?? 'car',
    durationMinutes: toMinutes(node.getAttribute('trav_time')),
    dirty: false,
  }));

  return {
    planId: plan.planId,
    planIndex: plan.planIndex,
    selected: plan.selected === 1,
    planBlob: plan.planBlob,
    activities,
    legs,
  };
}

export function clonePlanDraft(plan: EditablePlanDraft): EditablePlanDraft {
  return {
    ...plan,
    activities: plan.activities.map(activity => ({ ...activity })),
    legs: plan.legs.map(leg => ({ ...leg })),
  };
}

export function buildPlanEditPayload(plan: EditablePlanDraft): ApplyPlanEditsInput {
  return {
    activities: plan.activities.map(activity => ({
      originalIndex: activity.originalIndex,
      typeName: activity.type,
      startTime: toXmlTime(activity.startTime),
      endTime: toXmlTime(activity.endTime),
      lng: activity.lng,
      lat: activity.lat,
      dirty: activity.dirty,
    })),
    legs: plan.legs.map<PlanLegEditInput>(leg => ({
      originalIndex: leg.originalIndex,
      mode: leg.mode,
      travelTime: fromMinutes(leg.durationMinutes),
      dirty: leg.dirty,
    })),
  };
}

export function toUiTime(value: string | null): string {
  if (!value) {
    return '';
  }
  return value.slice(0, 5);
}

export function toXmlTime(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }
  return trimmed.length === 5 ? `${trimmed}:00` : trimmed;
}

export function toMinutes(value: string | null): number {
  if (!value) {
    return 0;
  }
  const [hours, minutes, seconds] = value.split(':').map(part => Number(part || 0));
  return hours * 60 + minutes + Math.round(seconds / 60);
}

export function fromMinutes(value: number): string | null {
  if (!Number.isFinite(value) || value < 0) {
    return '00:00:00';
  }
  const hours = Math.floor(value / 60);
  const minutes = value % 60;
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:00`;
}

export function buildLocalPlanDraft(lng: number, lat: number): EditablePlanDraft {
  return {
    planId: 'p0',
    planIndex: 0,
    selected: true,
    planBlob: '',
    activities: [createNewActivity(0, lng, lat)],
    legs: [],
  };
}

export function createNewActivity(index: number, lng: number, lat: number): EditableActivity {
  return {
    key: `activity-new-${index}-${Date.now()}`,
    originalIndex: null,
    type: 'home',
    startTime: '',
    endTime: '',
    lng,
    lat,
    dirty: true,
  };
}

export function createNewLeg(index: number): EditableLeg {
  return {
    key: `leg-new-${index}-${Date.now()}`,
    originalIndex: null,
    mode: 'car',
    durationMinutes: 0,
    dirty: true,
  };
}

function projectActivityCoords(node: Element, projString: string): { lng: number | null; lat: number | null } {
  const x = Number(node.getAttribute('x'));
  const y = Number(node.getAttribute('y'));
  if (!Number.isFinite(x) || !Number.isFinite(y)) {
    return { lng: null, lat: null };
  }
  const [lng, lat] = proj4(projString, WGS84_CRS, [x, y]);
  return { lng, lat };
}
