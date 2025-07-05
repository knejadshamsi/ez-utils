import { create } from 'zustand';

const useUiStore = create((set) => ({
  // Modal states
  showWelcomeModal: true,
  showLoadingModal: false,
  loadingMessage: 'Loading...',
  
  // View mode
  viewMode: 'list',

  // Actions
  setShowWelcomeModal: (show) => set({ showWelcomeModal: show }),
  
  setShowLoadingModal: (show, message = 'Loading...') => set({ 
    showLoadingModal: show, 
    loadingMessage: message 
  }),
  
  setViewMode: (mode) => set({ viewMode: mode }),
}));

export default useUiStore;