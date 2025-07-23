import type { gui } from "../../../../wailsjs/go/models";

export const welcomeModalState = $state({
  showTable: false,
  processes: [] as gui.Process[],
  isProcessFetching: false,
  currentPage: 1
});

export const ITEMS_PER_PAGE = 10;

export function getPaginatedProcesses() {
  const successfulProcesses = welcomeModalState.processes.filter(p => p.status === 'PROCESSING_SUCCESS');
  const start = (welcomeModalState.currentPage - 1) * ITEMS_PER_PAGE;
  const end = start + ITEMS_PER_PAGE;
  return successfulProcesses.slice(start, end);
}

export function getTotalPages() {
  const successfulProcesses = welcomeModalState.processes.filter(p => p.status === 'PROCESSING_SUCCESS');
  return Math.ceil(successfulProcesses.length / ITEMS_PER_PAGE);
}