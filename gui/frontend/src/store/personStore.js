import { create } from 'zustand';
import { parsePersonXML, extractPlansAsArrays } from '../utils/xmlParser';
import { GetPopulation, GetPopulationByBbox, AddPerson } from '@wailsjs/go/gui/App';


const usePersonStore = create((set, get) => ({
  // Person management
  persons: [],
  visiblePersons: new Set(),
  selectedPerson: null,
  
  // Plan management
  selectedPlanRoute: null,
  
  // Map interaction
  mapClickHandler: null,
  bboxFilter: null,
  
  // Cache
  
  // Sync management
  changeCounter: 0,
  syncThreshold: 10,
  isSyncing: false,
  lastSyncTime: null,
  
  // Centralized data loading action
  loadPopulation: async (tableName, bboxFilter = null) => {
    const { setPersons } = get();
    try {
      let data;
      if (bboxFilter) {
        data = await GetPopulationByBbox(
          tableName,
          bboxFilter.minLat,
          bboxFilter.minLng,
          bboxFilter.maxLat,
          bboxFilter.maxLng
        );
      } else {
        data = await GetPopulation(tableName);
      }
      
      const currentPersons = get().persons;
      const localModifiedMap = new Map(currentPersons.filter(p => p.plans !== null).map(p => [p.id, p]));
      
      const mergedData = (data || []).map(person => {
        const localVersion = localModifiedMap.get(person.id);
        if (localVersion) {
          return localVersion;
        }
        return { ...person, plans: null };
      });

      setPersons(mergedData);
      return mergedData;
    } catch (error) {
      throw error;
    }
  },

  // Actions
  setPersons: (persons) => set({ persons }),
  togglePersonVisibility: (personId) => set(state => {
    const newVisible = new Set(state.visiblePersons);
    if (newVisible.has(personId)) {
      newVisible.delete(personId);
    } else {
      newVisible.add(personId);
    }
    return { visiblePersons: newVisible };
  }),
  setSelectedPerson: (person) => set(state => {
    if (!person) {
      return { selectedPerson: null };
    }
    const personFromStore = state.persons.find(p => p.id === person.id);
    return { selectedPerson: personFromStore || person };
  }),
  setMapClickHandler: (handler) => set({ mapClickHandler: handler }),
  setSelectedPlanRoute: (route) => set({ selectedPlanRoute: route }),
  setBboxFilter: (bbox) => set({ bboxFilter: bbox }),
  clearPersonCache: () => {
      set({ selectedPerson: null, selectedPlanRoute: null });
  },
  
  // Sync management actions
  incrementChangeCounter: () => set(state => {
    const newCount = state.changeCounter + 1;
    if (newCount >= state.syncThreshold && !state.isSyncing) {
      get().syncToDatabase();
    }
    return { changeCounter: newCount };
  }),
  
  resetChangeCounter: () => set({ changeCounter: 0 }),
  
  updatePersonById: (personId, updates) => {
    set(state => {
      const newPersons = state.persons.map(p =>
        p.id === personId ? { ...p, ...updates } : p
      );
      const updatedPerson = newPersons.find(p => p.id === personId);
      return {
        persons: newPersons,
        selectedPerson: state.selectedPerson?.id === personId ? updatedPerson : state.selectedPerson
      };
    });
  },
  
  // Efficient person update with change tracking
  updatePersonLocally: (personId, updates) => {
    get().updatePersonById(personId, updates);
    get().incrementChangeCounter();
  },
  
  // Parse XML to plans structure (lazy loading)
  parsePersonPlans: (personId) => {
    const person = get().persons.find(p => p.id === personId);
    if (!person || person.plans !== null) return;

    try {
      const parsedData = parsePersonXML(person.raw_xml);
      const plans = extractPlansAsArrays(parsedData);
      get().updatePersonById(personId, { plans });
    } catch (error) {
      get().updatePersonById(personId, { plans: [] });
    }
  },
  
  // Create new person with optional initial data
  createPerson: async (tableName, personData = {}) => {
    const state = get();
    
    try {
      // Get existing persons if not provided
      const existingPersons = personData.existingPersons || await GetPopulation(tableName);
      
      // Determine ID
      let id = personData.id;
      if (!id) {
        const { generateHighestPlusOneId, generateRandomId } = await import('../utils/idGenerator');
        id = generateHighestPlusOneId(existingPersons || []);
        if (!id) {
          id = generateRandomId(existingPersons || []);
        }
      }
      
      if (existingPersons.some(p => p.id === id)) {
        throw new Error('Person ID already exists');
      }
      
      let personXml;
      if (personData.withHomeActivity && personData.x && personData.y) {
        personXml = `<person id="${id}">
  <plan selected="yes">
    <activity type="home" start_time="08:00:00" x="${personData.x}" y="${personData.y}" />
  </plan>
</person>`;
      } else {
        personXml = `<person id="${id}">
  <plan selected="yes">
  </plan>
</person>`;
      }
      
      const coords = personData.x && personData.y ? `${personData.x},${personData.y}` : '0,0';
      
      const newPerson = {
        id,
        coords,
        raw_xml: personXml
      };
      
      await AddPerson(tableName, newPerson);
      
      const updatedPersons = await state.loadPopulation(tableName);
      
      return updatedPersons.find(p => p.id === id);
    } catch (error) {
      throw error;
    }
  },
  
  // Batch sync function
  syncToDatabase: async (tableName) => {
    const state = get();
    if (state.isSyncing) return;
    
    const { message } = await import('antd');
    
    const modifiedPersons = state.persons.filter(p => p.plans !== null);
    
    if (modifiedPersons.length === 0) {
      message.info('No changes to sync');
      return;
    }
    
    set({ isSyncing: true });
    
    try {
      const { buildXMLFromPlans } = await import('../utils/xmlParser');
      const { BatchUpdatePersons } = await import('@wailsjs/go/gui/App');
      
      const updates = modifiedPersons.map(person => ({
        id: person.id,
        coords: person.coords,
        raw_xml: buildXMLFromPlans(person)
      }));
      
      await BatchUpdatePersons(
        tableName,
        updates
      );
      
      set({
        changeCounter: 0,
        lastSyncTime: Date.now(),
        isSyncing: false
      });
      
      message.success(`Synced ${updates.length} persons`);
    } catch (error) {
      message.error('Failed to sync changes');
      set({ isSyncing: false });
    }
  }
}));

export default usePersonStore;