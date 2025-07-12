import type { gui } from "../../../wailsjs/go/models";

// Local state for welcome modal
export const welcomeModalState = $state({
  showTable: false,
  selectedProcessId: null as number | null,
  processes: [] as gui.Process[],
  isProcessFetching: false,
  currentPage: 1
});

// Constants
export const ITEMS_PER_PAGE = 10;

// Computed values as functions
export function getPaginatedProcesses() {
  const start = (welcomeModalState.currentPage - 1) * ITEMS_PER_PAGE;
  const end = start + ITEMS_PER_PAGE;
  return welcomeModalState.processes.slice(start, end);
}

export function getTotalPages() {
  return Math.ceil(welcomeModalState.processes.length / ITEMS_PER_PAGE);
}