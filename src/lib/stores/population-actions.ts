import { invoke } from '@tauri-apps/api/core';

import {
  buildLocalPlanDraft,
  buildPlanEditPayload,
  parseAttributesBlob,
  parsePlanBlob,
  serializeAttributes,
} from '$lib/components/population/xml';
import { renderSelectedPlan } from '$lib/components/population/population-map';
import type { PopulationPersonPayload, PopulationSelection } from '$lib/components/population/types';
import { currentProjString, drawers, status } from '$lib/stores/ui.svelte';
import { formatError, markPlanClean, sameSelection, snapshotPlan } from '$lib/stores/population-helpers';
import type { PopulationStore } from '$lib/stores/population.svelte';

export function createPersonAt(store: PopulationStore, lng: number, lat: number) {
  const sourceName = store.activeSourceName;
  if (!sourceName) return;

  const localPlan = buildLocalPlanDraft(lng, lat);
  const selection: PopulationSelection = { sourceName, personId: '__new__' };

  store.people = [
    { key: `${sourceName}:__new__`, sourceName, personId: '__new__', pointCount: 1 },
    ...store.people,
  ];

  store.isNewPerson = true;
  store.newPersonCoords = { lng, lat };
  store.selected = selection;
  store.person = { personId: '__new__', attributesBlob: null, plans: [] };
  store.attributeRows = [];
  store.attributesExpanded = false;
  store.originalAttributesBlob = serializeAttributes([]);
  store.plans = [localPlan];
  store.originalPlanSnapshots.clear();
  store.originalPlanSnapshots.set(localPlan.planId, '');
  store.activePlanId = localPlan.planId;
  store.saveState = 'idle';

  store.populationLayer?.clearLayers();
  store.renderActivePlan();
  drawers.open('secondary');
}

export async function loadPerson(store: PopulationStore, selection: PopulationSelection) {
  store.selected = selection;
  store.personLoading = true;
  store.error = null;
  store.editLocations = false;
  drawers.open('secondary');
  try {
    const payload = await invoke<PopulationPersonPayload>('get_person', {
      sourceName: selection.sourceName,
      personId: selection.personId,
    });
    const projString = await currentProjString();
    hydratePerson(store, payload, selection, projString);
    store.populationLayer?.clearLayers();
    store.renderActivePlan();
  } catch (error) {
    store.error = formatError(error);
    store.person = null;
    store.attributeRows = [];
    store.plans = [];
    store.activePlanId = null;
  } finally {
    store.personLoading = false;
  }
}

export async function saveCurrentPerson(store: PopulationStore) {
  if (!store.selected || !store.person) return;
  store.saveState = 'saving';
  try {
    if (store.isNewPerson && store.newPersonCoords) {
      const created = await invoke<PopulationPersonPayload>('create_person', {
        sourceName: store.selected.sourceName,
        lng: store.newPersonCoords.lng,
        lat: store.newPersonCoords.lat,
      });
      const realId = created.personId;
      store.selected = { sourceName: store.selected.sourceName, personId: realId };
      store.person = { ...store.person, personId: realId };
      store.people = store.people.map(p =>
        p.personId === '__new__' ? { ...p, personId: realId, key: `${p.sourceName}:${realId}` } : p
      );

      if (created.plans.length > 0 && store.plans.length > 0) {
        const realPlanId = created.plans[0].planId;
        const localPlan = store.plans[0];
        store.originalPlanSnapshots.delete(localPlan.planId);
        localPlan.planId = realPlanId;
        localPlan.planBlob = created.plans[0].planBlob;
        store.originalPlanSnapshots.set(realPlanId, '');
        store.activePlanId = realPlanId;
        store.plans = [...store.plans];
      }

      store.isNewPerson = false;
      store.newPersonCoords = null;
    }

    await invoke('update_person_attributes', {
      sourceName: store.selected.sourceName,
      personId: store.selected.personId,
      newBlob: serializeAttributes(store.attributeRows),
    });

    for (const plan of store.plans) {
      if (snapshotPlan(plan) === store.originalPlanSnapshots.get(plan.planId)) {
        continue;
      }
      const savedPlan = await invoke<PopulationPersonPayload['plans'][number]>('apply_plan_edits', {
        sourceName: store.selected.sourceName,
        personId: store.selected.personId,
        planId: plan.planId,
        edits: buildPlanEditPayload(plan),
      });
      plan.planBlob = savedPlan.planBlob;
      markPlanClean(plan);
      store.originalPlanSnapshots.set(plan.planId, snapshotPlan(plan));
    }

    store.originalAttributesBlob = serializeAttributes(store.attributeRows);
    store.person = {
      ...store.person,
      attributesBlob: store.originalAttributesBlob,
      plans: store.plans.map(plan => ({
        planId: plan.planId,
        planIndex: plan.planIndex,
        selected: plan.selected ? 1 : 0,
        planBlob: plan.planBlob,
      })),
    };
    store.setSaved();
    status.flash('saved');
    await store.fetchCurrentPage();
    if (store.selectedPlanLayer) {
      renderSelectedPlan(store.selectedPlanLayer, store.activePlan, store.editLocations, store.map);
    }
  } catch (error) {
    store.saveState = 'idle';
    store.error = formatError(error);
  }
}

export async function setPlanSelected(store: PopulationStore, planId: string) {
  if (!store.selected) return;
  try {
    await invoke('set_plan_selected', {
      sourceName: store.selected.sourceName,
      personId: store.selected.personId,
      planId,
    });
    store.plans = store.plans.map(plan => ({ ...plan, selected: plan.planId === planId }));
    store.activePlanId = planId;
    store.person = store.person
      ? {
          ...store.person,
          plans: store.person.plans.map(plan => ({
            ...plan,
            selected: plan.planId === planId ? 1 : 0,
          })),
        }
      : null;
    store.renderActivePlan();
  } catch (error) {
    store.error = formatError(error);
  }
}

export async function confirmDeletePerson(store: PopulationStore) {
  const target = store.deleteTarget ?? store.selected;
  if (!target) return;
  try {
    await invoke('delete_person', {
      sourceName: target.sourceName,
      personId: target.personId,
    });
    store.deletePromptOpen = false;
    store.deleteTarget = null;
    if (sameSelection(store.selected, target)) {
      store.closeSecondary();
    }
    await store.fetchCurrentPage();
    status.flash('deleted');
    if (store.selectedPlanLayer) {
      renderSelectedPlan(store.selectedPlanLayer, store.activePlan, store.editLocations, store.map);
    }
  } catch (error) {
    store.error = formatError(error);
  }
}

export async function createPlan(store: PopulationStore) {
  if (!store.selected) return;
  const seed = store.activePlan?.activities[0];
  const lng = seed?.lng ?? 0;
  const lat = seed?.lat ?? 0;
  try {
    const plan = await invoke<PopulationPersonPayload['plans'][number]>('create_plan', {
      sourceName: store.selected.sourceName,
      personId: store.selected.personId,
      lng,
      lat,
    });
    const projString = await currentProjString();
    const parsed = parsePlanBlob(plan, projString);
    store.plans = [...store.plans, parsed];
    store.originalPlanSnapshots.set(parsed.planId, snapshotPlan(parsed));
    store.activePlanId = parsed.planId;
    if (store.person) {
      store.person = {
        ...store.person,
        plans: [...store.person.plans, plan],
      };
    }
    if (store.selectedPlanLayer) {
      renderSelectedPlan(store.selectedPlanLayer, store.activePlan, store.editLocations, store.map);
    }
  } catch (error) {
    store.error = formatError(error);
  }
}

export async function deleteActivePlan(store: PopulationStore) {
  if (!store.selected || !store.activePlan || store.plans.length <= 1) return;
  const deletingPlanId = store.activePlan.planId;
  const deletedWasSelected = store.activePlan.selected;
  try {
    await invoke('delete_plan', {
      sourceName: store.selected.sourceName,
      personId: store.selected.personId,
      planId: deletingPlanId,
    });
    const remaining = store.plans.filter(plan => plan.planId !== deletingPlanId);
    const nextActive = remaining[0] ?? null;
    if (deletedWasSelected && nextActive) {
      nextActive.selected = true;
    }
    store.plans = remaining;
    store.activePlanId = nextActive?.planId ?? null;
    store.originalPlanSnapshots.delete(deletingPlanId);
    if (store.person) {
      store.person = {
        ...store.person,
        plans: store.person.plans
          .filter(plan => plan.planId !== deletingPlanId)
          .map((plan, index) => ({
            ...plan,
            selected: deletedWasSelected && index === 0 ? 1 : plan.selected,
          })),
      };
    }
    if (store.selectedPlanLayer) {
      renderSelectedPlan(store.selectedPlanLayer, store.activePlan, store.editLocations, store.map);
    }
  } catch (error) {
    store.error = formatError(error);
  }
}

export async function saveThenSwitch(store: PopulationStore) {
  const target = store.pendingSelection;
  store.switchPromptOpen = false;
  store.pendingSelection = null;
  await saveCurrentPerson(store);
  if (target) {
    await loadPerson(store, target);
  } else {
    store.closeSecondary();
  }
}

export async function discardThenSwitch(store: PopulationStore) {
  const target = store.pendingSelection;
  store.switchPromptOpen = false;
  store.pendingSelection = null;
  if (target) {
    await loadPerson(store, target);
  } else {
    store.closeSecondary();
  }
}

function hydratePerson(store: PopulationStore, payload: PopulationPersonPayload, selection: PopulationSelection, projString: string) {
  store.selected = selection;
  store.person = payload;
  store.attributeRows = parseAttributesBlob(payload.attributesBlob);
  store.attributesExpanded = false;
  store.originalAttributesBlob = serializeAttributes(store.attributeRows);
  store.plans = payload.plans.map(plan => parsePlanBlob(plan, projString));
  store.originalPlanSnapshots.clear();
  store.plans.forEach(plan => {
    store.originalPlanSnapshots.set(plan.planId, snapshotPlan(plan));
  });
  store.activePlanId = store.plans.find(plan => plan.selected)?.planId ?? store.plans[0]?.planId ?? null;
  store.saveState = 'idle';
}
