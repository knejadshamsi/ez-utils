import { createNewLeg } from '$lib/components/population/xml';
import type {
  EditableLeg,
  EditablePlanDraft,
  PopulationPersonSummary,
  PopulationQueryPoint,
  PopulationSelection,
} from '$lib/components/population/types';

export function summarizePeople(points: PopulationQueryPoint[]): PopulationPersonSummary[] {
  const counts = new Map<string, PopulationPersonSummary>();
  points.forEach(point => {
    const key = `${point.sourceName}:${point.personId}`;
    const current = counts.get(key);
    if (current) {
      current.pointCount += 1;
      return;
    }
    counts.set(key, {
      key,
      sourceName: point.sourceName,
      personId: point.personId,
      pointCount: 1,
    });
  });
  return Array.from(counts.values());
}

export function regenerateLegs(plan: EditablePlanDraft) {
  const existing = new Map<string, EditableLeg>();
  plan.legs.forEach(leg => {
    existing.set(leg.key, leg);
  });

  const nextLegs: EditableLeg[] = [];
  for (let index = 0; index < plan.activities.length - 1; index += 1) {
    const fromKey = plan.activities[index].key;
    const toKey = plan.activities[index + 1].key;
    const preserved =
      existing.get(`${fromKey}:${toKey}`) ??
      plan.legs.find((_, legIndex) => legIndex === index && plan.activities[index] && plan.activities[index + 1]);
    nextLegs.push(
      preserved
        ? { ...preserved, dirty: true, key: `${fromKey}:${toKey}` }
        : {
            ...createNewLeg(index),
            key: `${fromKey}:${toKey}`,
          }
    );
  }
  plan.legs = nextLegs;
}

export function sortActivities(plan: EditablePlanDraft) {
  plan.activities.sort((left, right) => compareActivityTimes(left.startTime, right.startTime));
}

export function snapshotPlan(plan: EditablePlanDraft): string {
  return JSON.stringify({
    selected: plan.selected,
    activities: plan.activities.map(activity => ({
      originalIndex: activity.originalIndex,
      type: activity.type,
      startTime: activity.startTime,
      endTime: activity.endTime,
      lng: activity.lng,
      lat: activity.lat,
    })),
    legs: plan.legs.map(leg => ({
      originalIndex: leg.originalIndex,
      mode: leg.mode,
      durationMinutes: leg.durationMinutes,
    })),
  });
}

export function markPlanClean(plan: EditablePlanDraft) {
  plan.activities = plan.activities.map(activity => ({ ...activity, dirty: false }));
  plan.legs = plan.legs.map(leg => ({ ...leg, dirty: false }));
}

export function formatError(error: unknown): string {
  if (typeof error === 'string') {
    return error;
  }
  if (error && typeof error === 'object' && 'kind' in error) {
    const typed = error as {
      kind: string;
      person_id?: string;
      plan_id?: string;
      message?: string;
    };
    if (typed.kind === 'population_person_not_found' && typed.person_id) {
      return `Person ${typed.person_id} was not found.`;
    }
    if (typed.kind === 'population_plan_not_found' && typed.person_id && typed.plan_id) {
      return `Plan ${typed.plan_id} for person ${typed.person_id} was not found.`;
    }
    if (typed.kind === 'population_last_plan' && typed.person_id) {
      return `Person ${typed.person_id} must keep at least one plan.`;
    }
    if (typed.kind === 'io' && typed.message) {
      return typed.message;
    }
  }
  return 'Population operation failed.';
}

export function sameSelection(
  left: PopulationSelection | null,
  right: PopulationSelection | null,
): boolean {
  return !!left && !!right && left.personId === right.personId && left.sourceName === right.sourceName;
}

function compareActivityTimes(left: string, right: string): number {
  return timeToSortable(left) - timeToSortable(right);
}

function timeToSortable(value: string): number {
  if (!value.trim()) {
    return -1;
  }
  const [hours, minutes] = value.split(':').map(part => Number(part || 0));
  return hours * 60 + minutes;
}
