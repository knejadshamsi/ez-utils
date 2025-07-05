import { create } from 'zustand';
import LRUCache from '../utils/lruCache';
import { parsePersonXML, extractPlansAsArrays } from '../utils/xmlParser';

const xmlCache = new LRUCache(500);

const useAppStore = create((set, get) => ({
  // Startup configuration
  filePath: '',
  editMode: '',
  startupError: null,
  
  // Modal states
  showWelcomeModal: true,
  showLoadingModal: false,
  loadingMessage: 'Loading...',
  
  // Process data
  selectedProcess: null,
  currentTableName: null,
  
  // Telemetry data
  currentTelemetry: null,
  isPollingTelemetry: false,
  telemetryInterval: null,
  
  // Actions
  setStartupConfig: (config) => set({
    filePath: config.filePath || '',
    editMode: config.editMode || '',
    startupError: config.error || null
  }),
  
  setShowWelcomeModal: (show) => set({ showWelcomeModal: show }),
  
  setEditMode: (mode) => set({ editMode: mode }),
  
  setShowLoadingModal: (show, message = 'Loading...') => set({ 
    showLoadingModal: show, 
    loadingMessage: message 
  }),
  
  setSelectedProcess: (process) => set({ 
    selectedProcess: process,
    currentTableName: process ? `population_data_${process.id}` : null
  }),
  
  clearSelectedProcess: () => set({ 
    selectedProcess: null,
    currentTableName: null
  }),
  
  // Telemetry actions
  startTelemetryPolling: (processId) => {
    const { GetProcessTelemetry } = window.go.gui.App;
    
    // Clear any existing interval
    const currentInterval = useAppStore.getState().telemetryInterval;
    if (currentInterval) {
      clearInterval(currentInterval);
    }
    
    // Start polling
    const interval = setInterval(async () => {
      try {
        const telemetry = await GetProcessTelemetry(processId);
        set({ currentTelemetry: telemetry });
      } catch (error) {
        console.error('Failed to fetch telemetry:', error);
      }
    }, 1000);
    
    set({ isPollingTelemetry: true, telemetryInterval: interval });
  },
  
  stopTelemetryPolling: () => {
    const interval = useAppStore.getState().telemetryInterval;
    if (interval) {
      clearInterval(interval);
    }
    set({ 
      isPollingTelemetry: false, 
      telemetryInterval: null,
      currentTelemetry: null
    });
  },
  
  updateTelemetry: (telemetry) => set({ currentTelemetry: telemetry }),
  
  // Person management
  persons: [],
  visiblePersons: new Set(),
  selectedPerson: null,
  viewMode: 'list',
  
  // Plan management
  selectedPlanRoute: null,
  
  // Map interaction
  mapClickHandler: null,
  bboxFilter: null,
  
  // Cache
  xmlCache,
  
  // Sync management
  changeCounter: 0,
  syncThreshold: 10,
  isSyncing: false,
  lastSyncTime: null,
  
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
    // Always get the person from the store to ensure it has the plans property
    const personFromStore = state.persons.find(p => p.id === person.id);
    return { selectedPerson: personFromStore || person };
  }),
  setViewMode: (mode) => set({ viewMode: mode }),
  setMapClickHandler: (handler) => set({ mapClickHandler: handler }),
  setSelectedPlanRoute: (route) => set({ selectedPlanRoute: route }),
  setBboxFilter: (bbox) => set({ bboxFilter: bbox }),
  clearPersonCache: () => {
    xmlCache.clear();
    set({ selectedPerson: null, selectedPlanRoute: null });
  },
  
  // Sync management actions
  incrementChangeCounter: () => set(state => {
    const newCount = state.changeCounter + 1;
    // Auto-sync check
    if (newCount >= state.syncThreshold && !state.isSyncing) {
      get().syncToDatabase();
    }
    return { changeCounter: newCount };
  }),
  
  resetChangeCounter: () => set({ changeCounter: 0 }),
  
  // Efficient person update with change tracking
  updatePersonLocally: (personId, updates) => set(state => {
    const persons = state.persons.map(p => 
      p.id === personId 
        ? { ...p, ...updates }
        : p
    );
    
    // Increment counter for any change
    get().incrementChangeCounter();
    
    const updatedPerson = persons.find(p => p.id === personId);
    return { 
      persons,
      // Update selectedPerson if it's the one being updated
      selectedPerson: state.selectedPerson?.id === personId ? updatedPerson : state.selectedPerson
    };
  }),
  
  // Parse XML to plans structure (lazy loading)
  // Only called when user opens a person for editing
  parsePersonPlans: (personId) => {
    try {
      const person = get().persons.find(p => p.id === personId);
      if (!person || person.plans !== null) return; // Already parsed or doesn't exist
      
      const parsedData = parsePersonXML(person.raw_xml);
      const plans = extractPlansAsArrays(parsedData);
      
      set(state => {
        const updatedPersons = state.persons.map(p =>
          p.id === personId ? { ...p, plans } : p
        );
        const updatedPerson = updatedPersons.find(p => p.id === personId);
        return {
          persons: updatedPersons,
          // Update selectedPerson if it's the one being parsed
          selectedPerson: state.selectedPerson?.id === personId ? updatedPerson : state.selectedPerson
        };
      });
    } catch (error) {
      console.error('Failed to parse person plans:', error);
      // Set empty plans array on error
      set(state => {
        const updatedPersons = state.persons.map(p =>
          p.id === personId ? { ...p, plans: [] } : p
        );
        const updatedPerson = updatedPersons.find(p => p.id === personId);
        return {
          persons: updatedPersons,
          // Update selectedPerson if it's the one being parsed
          selectedPerson: state.selectedPerson?.id === personId ? updatedPerson : state.selectedPerson
        };
      });
    }
  },
  
  // Create new person with optional initial data
  createPerson: async (personData = {}) => {
    const state = get();
    if (!state.selectedProcess) {
      throw new Error('No process selected');
    }
    
    const { GetPopulation, AddPerson } = window.go.gui.App;
    
    try {
      // Get existing persons if not provided
      const existingPersons = personData.existingPersons || await GetPopulation(state.selectedProcess.table_name);
      
      // Determine ID
      let id = personData.id;
      if (!id) {
        // Try highest + 1 first
        const { generateHighestPlusOneId, generateRandomId } = await import('../utils/idGenerator');
        id = generateHighestPlusOneId(existingPersons || []);
        if (!id) {
          id = generateRandomId(existingPersons || []);
        }
      }
      
      // Check if ID already exists
      if (existingPersons.some(p => p.id === id)) {
        throw new Error('Person ID already exists');
      }
      
      // Create XML structure
      let personXml;
      if (personData.withHomeActivity && personData.x && personData.y) {
        personXml = `<person id="${id}">
  <plan selected="yes">
    <activity type="home" start_time="08:00:00" x="${personData.x}" y="${personData.y}" />
  </plan>
</person>`;
      } else {
        // Minimal person with empty plan
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
      
      // Add to database
      await AddPerson(state.selectedProcess.table_name, newPerson);
      
      // Refresh person list
      const updatedPersons = await GetPopulation(state.selectedProcess.table_name);
      const personsWithPlans = (updatedPersons || []).map(p => ({ ...p, plans: null }));
      set({ persons: personsWithPlans });
      
      // Return the created person
      return personsWithPlans.find(p => p.id === id);
    } catch (error) {
      console.error('Failed to create person:', error);
      throw error;
    }
  },
  
  // Batch sync function
  syncToDatabase: async () => {
    const state = get();
    if (state.isSyncing) return;
    
    // Import message early
    const { message } = await import('antd');
    
    // Filter persons that have been opened/modified (plans !== null)
    const modifiedPersons = state.persons.filter(p => p.plans !== null);
    
    if (modifiedPersons.length === 0) {
      message.info('No changes to sync');
      return;
    }
    
    set({ isSyncing: true });
    
    try {
      // Import necessary functions dynamically to avoid circular dependencies
      const { buildXMLFromPlans } = await import('../utils/xmlParser');
      const { BatchUpdatePersons } = await import('../../wailsjs/go/gui/App');
      
      // Build batch update payload
      const updates = modifiedPersons.map(person => ({
        id: person.id,
        coords: person.coords,
        raw_xml: buildXMLFromPlans(person)
      }));
      
      // Call batch update endpoint
      await BatchUpdatePersons(
        state.selectedProcess.table_name,
        updates
      );
      
      // Reset counter and update sync time
      // NOTE: We keep plans in memory - don't clear them!
      set({
        changeCounter: 0,
        lastSyncTime: Date.now(),
        isSyncing: false
      });
      
      message.success(`Synced ${updates.length} persons`);
    } catch (error) {
      console.error('Sync failed:', error);
      message.error('Failed to sync changes');
      set({ isSyncing: false });
    }
  }
}));

export default useAppStore;