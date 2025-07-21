// Export progress state using Svelte 5 runes
export const exportProgress = $state<{
  isExporting: boolean;
  current: number;
  total: number;
}>({
  isExporting: false,
  current: 0,
  total: 0
});