import type { FeatureCollection, Feature, Point, LineString } from 'geojson';
import type { Person } from '../../types';

export interface PopulationData {
  persons: Person[];
}

export interface NetworkData {
  nodes: any[];
  links: any[];
}

export interface PTData {
  stops: any[];
  routes: any[];
  routeStops: any[];
}

export function transformPopulationToGeoJSON(data: PopulationData): FeatureCollection {
  const features: Feature[] = [];
  
  data.persons.forEach(person => {
    const [x, y] = person.Coords.split(',').map(Number);
    
    features.push({
      type: 'Feature',
      properties: {
        id: person.ID,
        entityType: 'person',
        entityId: person.ID
      },
      geometry: {
        type: 'Point',
        coordinates: [x, y]
      } as Point
    });
    
    const activities = extractActivitiesFromXML(person.RawXML);
    if (activities.length > 1) {
      const path = activities.map(act => [act.x, act.y]);
      features.push({
        type: 'Feature',
        properties: {
          id: `${person.ID}-path`,
          entityType: 'path',
          entityId: person.ID,
          personId: person.ID
        },
        geometry: {
          type: 'LineString',
          coordinates: path
        } as LineString
      });
    }
  });
  
  return {
    type: 'FeatureCollection',
    features
  };
}

export function transformNetworkToGeoJSON(data: NetworkData): FeatureCollection {
  const features: Feature[] = [];
  
  data.nodes.forEach(node => {
    features.push({
      type: 'Feature',
      properties: {
        id: node.id,
        entityType: 'node',
        entityId: node.id
      },
      geometry: {
        type: 'Point',
        coordinates: [node.x, node.y]
      } as Point
    });
  });
  
  data.links.forEach(link => {
    const fromNode = data.nodes.find(n => n.id === link.from);
    const toNode = data.nodes.find(n => n.id === link.to);
    
    if (fromNode && toNode) {
      features.push({
        type: 'Feature',
        properties: {
          id: link.id,
          entityType: 'link',
          entityId: link.id,
          fromNodeId: link.from,
          toNodeId: link.to
        },
        geometry: {
          type: 'LineString',
          coordinates: [
            [fromNode.x, fromNode.y],
            [toNode.x, toNode.y]
          ]
        } as LineString
      });
    }
  });
  
  return {
    type: 'FeatureCollection',
    features
  };
}

export function transformPTToGeoJSON(data: PTData): FeatureCollection {
  const features: Feature[] = [];
  
  data.stops.forEach(stop => {
    features.push({
      type: 'Feature',
      properties: {
        id: stop.id,
        entityType: 'stop',
        entityId: stop.id,
        name: stop.name
      },
      geometry: {
        type: 'Point',
        coordinates: [stop.x, stop.y]
      } as Point
    });
  });
  
  data.routes.forEach(route => {
    const routeStops = data.routeStops
      .filter(rs => rs.route_id === route.id)
      .sort((a, b) => a.stop_order - b.stop_order);
    
    const stopCoords = routeStops
      .map(rs => data.stops.find(s => s.id === rs.stop_ref_id))
      .filter(Boolean)
      .map(stop => [stop.x, stop.y]);
    
    if (stopCoords.length > 1) {
      features.push({
        type: 'Feature',
        properties: {
          id: route.id,
          entityType: 'route',
          entityId: route.id,
          lineId: route.line_id,
          stopIds: routeStops.map(rs => rs.stop_ref_id)
        },
        geometry: {
          type: 'LineString',
          coordinates: stopCoords
        } as LineString
      });
    }
  });
  
  return {
    type: 'FeatureCollection',
    features
  };
}

function extractActivitiesFromXML(xml: string): Array<{x: number, y: number}> {
  const activities: Array<{x: number, y: number}> = [];
  const activityRegex = /<activity[^>]*\sx="([^"]+)"[^>]*\sy="([^"]+)"[^>]*(?:\/?>|>)/g;
  const reverseRegex = /<activity[^>]*\sy="([^"]+)"[^>]*\sx="([^"]+)"[^>]*(?:\/?>|>)/g;
  
  let match;
  while ((match = activityRegex.exec(xml)) !== null) {
    activities.push({
      x: parseFloat(match[1]),
      y: parseFloat(match[2])
    });
  }
  
  if (activities.length === 0) {
    while ((match = reverseRegex.exec(xml)) !== null) {
      activities.push({
        x: parseFloat(match[2]),
        y: parseFloat(match[1])
      });
    }
  }
  
  return activities;
}