// ===== POPULATION INTERFACES =====
interface AddPersonAction {
  type: 'population';
  elementType: 'person';
  action: 'add';
  tableName: string;
  data: {
    id: string;
    lng: number;
    lat: number;
    rawXML: string;
  };
}

interface UpdatePersonAction {
  type: 'population';
  elementType: 'person';
  action: 'update';
  tableName: string;
  personId: string;
  planXML?: string;
  lng?: number;
  lat?: number;
}

interface DeletePersonAction {
  type: 'population';
  elementType: 'person';
  action: 'delete';
  tableName: string;
  personId: string;
}

interface BatchUpdatePersonsAction {
  type: 'population';
  elementType: 'person';
  action: 'batchUpdate';
  tableName: string;
  updates: Array<{
    personId: string;
    lng?: number;
    lat?: number;
    planXML?: string;
  }>;
}

// ===== NETWORK INTERFACES =====
interface UpdateNodeAction {
  type: 'network';
  elementType: 'node';
  action: 'update';
  processId: number;
  nodeId: string;
  lng: number;
  lat: number;
  rawXML: string;
}

interface CreateNodeAction {
  type: 'network';
  elementType: 'node';
  action: 'create';
  processId: number;
  nodeId: string;
  lng: number;
  lat: number;
  rawXML: string;
}

interface DeleteNodeAction {
  type: 'network';
  elementType: 'node';
  action: 'delete';
  processId: number;
  nodeId: string;
  deleteConnectedLinks: boolean;
}

interface BatchUpdateNodesAction {
  type: 'network';
  elementType: 'node';
  action: 'batchUpdate';
  processId: number;
  updates: Record<string, { lng: number; lat: number }>;
}

interface BatchDeleteNodesAction {
  type: 'network';
  elementType: 'node';
  action: 'batchDelete';
  processId: number;
  nodeIds: string[];
  deleteConnectedLinks: boolean;
}

interface CreateLinkAction {
  type: 'network';
  elementType: 'link';
  action: 'create';
  processId: number;
  linkId: string;
  fromNode: string;
  toNode: string;
  rawXML: string;
}

interface UpdateLinkAction {
  type: 'network';
  elementType: 'link';
  action: 'update';
  processId: number;
  linkId: string;
  rawXML: string;
}

interface DeleteLinkAction {
  type: 'network';
  elementType: 'link';
  action: 'delete';
  processId: number;
  linkId: string;
}

// ===== PT INTERFACES =====
interface AddStopAction {
  type: 'pt';
  elementType: 'stop';
  action: 'add';
  processId: number;
  stop: any; // PTStop type
}

interface UpdateStopAction {
  type: 'pt';
  elementType: 'stop';
  action: 'update';
  processId: number;
  stopId: string;
  update: any; // PTStopUpdate type
}

interface DeleteStopAction {
  type: 'pt';
  elementType: 'stop';
  action: 'delete';
  processId: number;
  stopId: string;
  routeId: string;
}

interface BatchUpdateStopsAction {
  type: 'pt';
  elementType: 'stop';
  action: 'batchUpdate';
  processId: number;
  updates: Record<string, any>; // Record<string, PTStopUpdate>
}

interface AddLineAction {
  type: 'pt';
  elementType: 'line';
  action: 'add';
  processId: number;
  line: any; // PTLine type
}

interface UpdateLineAction {
  type: 'pt';
  elementType: 'line';
  action: 'update';
  processId: number;
  lineId: string;
  update: any; // PTLineUpdate type
}

interface DeleteLineAction {
  type: 'pt';
  elementType: 'line';
  action: 'delete';
  processId: number;
  lineId: string;
}

interface AddRouteAction {
  type: 'pt';
  elementType: 'route';
  action: 'add';
  processId: number;
  route: any; // PTRoute type
}

interface UpdateRouteAction {
  type: 'pt';
  elementType: 'route';
  action: 'update';
  processId: number;
  routeId: string;
  update: any; // PTRouteUpdate type
}

interface DeleteRouteAction {
  type: 'pt';
  elementType: 'route';
  action: 'delete';
  processId: number;
  routeId: string;
}


interface AddDepartureAction {
  type: 'pt';
  elementType: 'departure';
  action: 'add';
  processId: number;
  departure: any; // Departure type
}

interface UpdateDepartureAction {
  type: 'pt';
  elementType: 'departure';
  action: 'update';
  processId: number;
  departureId: string;
  update: any; // Departure update type
}

interface DeleteDepartureAction {
  type: 'pt';
  elementType: 'departure';
  action: 'delete';
  processId: number;
  departureId: string;
}

// ===== PROCESS INTERFACES =====
interface DeleteProcessAction {
  type: 'process';
  elementType: 'process';
  action: 'delete';
  processId: number;
}

// ===== UNION TYPES =====
type PopulationAction = AddPersonAction | UpdatePersonAction | DeletePersonAction | BatchUpdatePersonsAction;
type NetworkAction = CreateNodeAction | UpdateNodeAction | DeleteNodeAction | BatchUpdateNodesAction | BatchDeleteNodesAction | CreateLinkAction | UpdateLinkAction | DeleteLinkAction;
type PTAction = AddStopAction | UpdateStopAction | DeleteStopAction | BatchUpdateStopsAction | AddLineAction | UpdateLineAction | DeleteLineAction | AddRouteAction | UpdateRouteAction | DeleteRouteAction | AddDepartureAction | UpdateDepartureAction | DeleteDepartureAction;
type ProcessAction = DeleteProcessAction;
export type SyncAction = PopulationAction | NetworkAction | PTAction | ProcessAction;

// ===== CHANGE TRACKER STATE =====
export const changeTracker = $state<{
  pendingChanges: SyncAction[];
  isSyncing: boolean;
}>({
  pendingChanges: [],
  isSyncing: false
});