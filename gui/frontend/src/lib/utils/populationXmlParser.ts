import type { Person, Plan, Activity, Leg, PersonAttribute } from '../stores/population.svelte';

export function parsePersonXML(rawXML: string): Partial<Person> {
  try {
    const parser = new DOMParser();
    const doc = parser.parseFromString(rawXML, 'text/xml');
    
    // Check for parser errors
    const parserError = doc.querySelector('parsererror');
    if (parserError) {
      console.error('XML parsing error:', parserError.textContent);
      return { attributes: [], plans: [] };
    }
    
    const personElement = doc.querySelector('person');
    if (!personElement) {
      return { attributes: [], plans: [] };
    }
    
    // Parse attributes
    const attributes = parseAttributes(personElement);
    
    const plans: Plan[] = [];
    const planElements = personElement.querySelectorAll('plan');
    
    // If no plans exist, create a default one from activities at person level
    if (planElements.length === 0) {
      const activities = parseActivities(personElement);
      if (activities.length > 0) {
        plans.push({
          id: 1,
          activities,
          legs: generateLegs(activities)
        });
      }
    } else {
      planElements.forEach((planElement, index) => {
        const activities = parseActivities(planElement);
        
        plans.push({
          id: index + 1,
          activities,
          legs: generateLegs(activities)
        });
      });
    }
    
    return { attributes, plans };
  } catch (error) {
    console.error('Error parsing person XML:', error);
    return { attributes: [], plans: [] };
  }
}

function parseAttributes(personElement: Element): PersonAttribute[] {
  const attributes: PersonAttribute[] = [];
  const attributesElement = personElement.querySelector('attributes');
  
  if (!attributesElement) {
    return attributes;
  }
  
  const attributeElements = attributesElement.querySelectorAll('attribute');
  attributeElements.forEach(element => {
    const name = element.getAttribute('name');
    const type = element.getAttribute('class') as PersonAttribute['type'];
    const textContent = element.textContent?.trim();
    
    if (name && type && textContent !== undefined) {
      let value: string | number | boolean = textContent;
      
      // Parse value based on type
      if (type === 'java.lang.Integer') {
        value = parseInt(textContent, 10);
      } else if (type === 'java.lang.Double') {
        value = parseFloat(textContent);
      } else if (type === 'java.lang.Boolean') {
        value = textContent.toLowerCase() === 'true';
      }
      
      attributes.push({
        name,
        type,
        value,
        included: true // All parsed attributes are included by default
      });
    }
  });
  
  return attributes;
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
      mode: "person's choice", // Default mode
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