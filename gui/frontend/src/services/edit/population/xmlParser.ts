import type { Person, Plan, Activity, Leg } from './populationStore.svelte';

export function parsePersonXML(rawXML: string): Partial<Person> {
  try {
    const parser = new DOMParser();
    const doc = parser.parseFromString(rawXML, 'text/xml');
    
    // Check for parser errors
    const parserError = doc.querySelector('parsererror');
    if (parserError) {
      console.error('XML parsing error:', parserError.textContent);
      return { plans: [] };
    }
    
    const personElement = doc.querySelector('person');
    if (!personElement) {
      return { plans: [] };
    }
    
    const plans: Plan[] = [];
    const planElements = personElement.querySelectorAll('plan');
    
    // If no plans exist, create a default one from activities at person level
    if (planElements.length === 0) {
      const activities = parseActivities(personElement);
      if (activities.length > 0) {
        plans.push({
          type: 'weekday',
          activities,
          legs: generateLegs(activities)
        });
      }
    } else {
      planElements.forEach(planElement => {
        const planType = (planElement.getAttribute('type') || 'weekday') as Plan['type'];
        const activities = parseActivities(planElement);
        
        plans.push({
          type: planType,
          activities,
          legs: generateLegs(activities)
        });
      });
    }
    
    return { plans };
  } catch (error) {
    console.error('Error parsing person XML:', error);
    return { plans: [] };
  }
}

function parseActivities(parentElement: Element): Activity[] {
  const activities: Activity[] = [];
  const activityElements = parentElement.querySelectorAll('activity, act');
  
  activityElements.forEach((element, index) => {
    const activity: Activity = {
      id: `activity_${Date.now()}_${index}`,
      type: normalizeActivityType(element.getAttribute('type') || 'other'),
      location: [
        parseFloat(element.getAttribute('x') || '0'),
        parseFloat(element.getAttribute('y') || '0')
      ],
      startTime: formatTime(element.getAttribute('start_time') || element.getAttribute('start') || '00:00:00'),
      endTime: formatTime(element.getAttribute('end_time') || element.getAttribute('end') || '23:59:59')
    };
    activities.push(activity);
  });
  
  return activities;
}

function generateLegs(activities: Activity[]): Leg[] {
  const legs: Leg[] = [];
  
  for (let i = 0; i < activities.length - 1; i++) {
    legs.push({
      fromActivityId: activities[i].id,
      toActivityId: activities[i + 1].id,
      mode: 'car', // Default mode
      duration: 30  // Default duration in minutes
    });
  }
  
  return legs;
}

function normalizeActivityType(type: string): Activity['type'] {
  const typeMap: Record<string, Activity['type']> = {
    'home': 'home',
    'h': 'home',
    'work': 'work',
    'w': 'work',
    'school': 'school',
    's': 'school',
    'shop': 'shop',
    'shopping': 'shop',
    'eat': 'eat',
    'dining': 'eat',
    'recreation': 'recreation',
    'leisure': 'recreation',
    'other': 'other'
  };
  
  return typeMap[type.toLowerCase()] || 'other';
}

function formatTime(timeStr: string): string {
  // Convert various time formats to HH:MM
  const parts = timeStr.split(':');
  if (parts.length >= 2) {
    const hours = parts[0].padStart(2, '0');
    const minutes = parts[1].padStart(2, '0');
    return `${hours}:${minutes}`;
  }
  return '00:00';
}

export function extractZoneFromCoords(coords: string): string {
  try {
    const [x, y] = coords.split(',').map(c => parseFloat(c.trim()));
    
    // Simple zone calculation based on coordinates
    // This is a placeholder - in a real implementation, you would
    // check against actual zone boundaries
    if (isNaN(x) || isNaN(y)) {
      return 'default';
    }
    
    // Example zone assignment based on coordinate ranges
    const zoneX = Math.floor(x / 0.01);
    const zoneY = Math.floor(y / 0.01);
    return `zone_${zoneX}_${zoneY}`;
  } catch (error) {
    return 'default';
  }
}