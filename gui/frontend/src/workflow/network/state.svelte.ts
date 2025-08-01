// Network state management
import type { NetworkNode, NetworkLink, NetworkSelection } from './types';

// Create reactive state for network operations
export const networkState = $state<{
  nodes: NetworkNode[];
  links: NetworkLink[];
  selection: NetworkSelection;
}>({
  nodes: [],
  links: [],
  selection: {
    selectedNodeId: null,
    selectedLinkId: null,
    selectedPolygon: null
  }
});