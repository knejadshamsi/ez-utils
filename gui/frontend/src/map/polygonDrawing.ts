import L from 'leaflet';
import 'leaflet-draw';
import 'leaflet-draw/dist/leaflet.draw.css';

interface DrawnZone {
  id: string;
  polygon: L.Polygon;
  editHandler: any; // L.EditToolbar.Edit
  displayPolygon?: L.Polygon; // Native polygon for display
}

let drawingLayer: L.FeatureGroup;
let currentDrawHandler: any = null; // L.Draw.Polygon
let drawnZones: Map<string, DrawnZone> = new Map();
let onCompleteCallback: ((zoneId: string, coordinates: [number, number][]) => void) | null = null;
let mapInstance: L.Map;

export function initializePolygonDrawing(map: L.Map): L.FeatureGroup {
  mapInstance = map;
  drawingLayer = new L.FeatureGroup({ pane: 'polygonPane' });
  map.addLayer(drawingLayer);
  
  // Listen for draw events
  map.on('draw:created', handlePolygonCreated);
  
  return drawingLayer;
}

export function startPolygonDrawing(map: L.Map, onComplete: (zoneId: string, coordinates: [number, number][]) => void) {
  onCompleteCallback = onComplete;
  
  // Create polygon draw handler without toolbar
  currentDrawHandler = new (L as any).Draw.Polygon(map, {
    shapeOptions: {
      color: '#3388ff',
      fillColor: '#3388ff',
      fillOpacity: 0.2,
      weight: 2,
      pane: 'polygonPane'
      // Removed canvas renderer - let it use default SVG renderer
    },
    allowIntersection: false,
    showArea: false
  });
  
  currentDrawHandler.enable();
}

export function stopPolygonDrawing() {
  if (currentDrawHandler) {
    currentDrawHandler.disable();
    currentDrawHandler = null;
  }
  onCompleteCallback = null;
}

function handlePolygonCreated(e: any) {
  const layer = e.layer as L.Polygon;
  
  // Generate zone ID
  const zoneId = `zone_${Date.now()}`;
  
  // Get coordinates
  const latlngs = layer.getLatLngs()[0] as L.LatLng[];
  const coordinates: [number, number][] = latlngs.map(latlng => [
    latlng.lng,
    latlng.lat
  ]);
  
  // Close the polygon
  coordinates.push(coordinates[0]);
  
  // Create native display polygon
  const displayPolygon = L.polygon(latlngs, {
    color: '#3388ff',
    fillColor: '#3388ff',
    fillOpacity: 0.2,
    weight: 2,
    pane: 'polygonPane'
  });
  displayPolygon.addTo(mapInstance);
  
  // Store both polygons (draw polygon not added to map)
  drawnZones.set(zoneId, {
    id: zoneId,
    polygon: layer,
    editHandler: null,
    displayPolygon: displayPolygon
  });
  
  // Call callback
  if (onCompleteCallback) {
    onCompleteCallback(zoneId, coordinates);
  }
}

export function enableZoneEditing(zoneId: string) {
  const zone = drawnZones.get(zoneId);
  if (!zone) return;
  
  // Hide display polygon
  if (zone.displayPolygon) {
    mapInstance.removeLayer(zone.displayPolygon);
  }
  
  // Add draw polygon to map for editing
  drawingLayer.addLayer(zone.polygon);
  
  // Create edit handler for this specific polygon
  zone.editHandler = new (L as any).EditToolbar.Edit(mapInstance, {
    featureGroup: drawingLayer,
    selectedPathOptions: {
      dashArray: '10, 10',
      fill: true,
      fillColor: '#fe57a1',
      fillOpacity: 0.1,
      maintainColor: false
    }
  });
  
  // Enable editing for just this polygon
  zone.polygon.editing.enable();
  
  // Update the zone when editing is done
  zone.polygon.on('edit', () => {
    const latlngs = zone.polygon.getLatLngs()[0] as L.LatLng[];
    
    // Update display polygon with new coordinates
    if (zone.displayPolygon) {
      zone.displayPolygon.setLatLngs(latlngs);
    }
    
    const coordinates: [number, number][] = latlngs.map(latlng => [
      latlng.lng,
      latlng.lat
    ]);
    coordinates.push(coordinates[0]);
    
    console.log(`Zone ${zoneId} updated:`, coordinates);
  });
}

export function disableZoneEditing(zoneId: string) {
  const zone = drawnZones.get(zoneId);
  if (!zone) return;
  
  // Disable editing for this polygon
  (zone.polygon as any).editing.disable();
  zone.polygon.off('edit');
  zone.editHandler = null;
  
  // Remove draw polygon from map
  drawingLayer.removeLayer(zone.polygon);
  
  // Show display polygon again
  if (zone.displayPolygon) {
    zone.displayPolygon.addTo(mapInstance);
  }
}

export function deleteZone(zoneId: string) {
  const zone = drawnZones.get(zoneId);
  if (!zone) return;
  
  // Disable editing if active
  if ((zone.polygon as any).editing && (zone.polygon as any).editing.enabled()) {
    (zone.polygon as any).editing.disable();
  }
  
  // Remove draw polygon if on map
  if (drawingLayer.hasLayer(zone.polygon)) {
    drawingLayer.removeLayer(zone.polygon);
  }
  
  // Remove display polygon
  if (zone.displayPolygon && mapInstance.hasLayer(zone.displayPolygon)) {
    mapInstance.removeLayer(zone.displayPolygon);
  }
  
  // Remove from storage
  drawnZones.delete(zoneId);
}

export function clearAllZones() {
  // Disable all editing and remove all polygons
  drawnZones.forEach(zone => {
    if (zone.editHandler) {
      zone.editHandler.disable();
    }
    
    // Remove display polygon
    if (zone.displayPolygon && mapInstance.hasLayer(zone.displayPolygon)) {
      mapInstance.removeLayer(zone.displayPolygon);
    }
  });
  
  // Clear drawing layer
  drawingLayer.clearLayers();
  
  // Clear storage
  drawnZones.clear();
}

export function getZonePolygon(zoneId: string): L.Polygon | null {
  const zone = drawnZones.get(zoneId);
  return zone ? zone.polygon : null;
}

export function setZoneStyle(zoneId: string, style: L.PathOptions) {
  const zone = drawnZones.get(zoneId);
  if (zone && zone.displayPolygon) {
    zone.displayPolygon.setStyle(style);
  }
}