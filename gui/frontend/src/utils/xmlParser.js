import { XMLParser, XMLBuilder } from 'fast-xml-parser';

const parserOptions = {
  ignoreAttributes: false,
  attributeNamePrefix: '@_',
  parseAttributeValue: true,
  trimValues: true
};

const builderOptions = {
  ignoreAttributes: false,
  attributeNamePrefix: '@_',
  format: true,
  indentBy: '  '
};

export const parsePersonXML = (xmlString) => {
  const parser = new XMLParser(parserOptions);
  return parser.parse(xmlString);
};

export const buildPersonXML = (jsonObj) => {
  const builder = new XMLBuilder(builderOptions);
  return builder.build(jsonObj);
};

export const extractPlans = (personData) => {
  const plans = personData.person?.plan || [];
  return Array.isArray(plans) ? plans : [plans];
};

export const extractActivities = (plan) => {
  const activities = plan?.activity || [];
  return Array.isArray(activities) ? activities : [activities];
};

export const extractLegs = (plan) => {
  const legs = plan?.leg || [];
  return Array.isArray(legs) ? legs : [legs];
};

// Convert plans to array structure with interleaved activities and legs
export const extractPlansAsArrays = (personData) => {
  const plans = extractPlans(personData);
  
  return plans.map(plan => {
    const activities = extractActivities(plan);
    const legs = extractLegs(plan);
    
    // Interleave activities and legs in chronological order
    const items = [];
    const maxLength = Math.max(activities.length, legs.length);
    
    for (let i = 0; i < maxLength; i++) {
      // Add activity
      if (i < activities.length) {
        items.push({
          type: 'activity',
          x: parseFloat(activities[i]['@_x'] || 0),
          y: parseFloat(activities[i]['@_y'] || 0),
          start_time: activities[i]['@_start_time'] || '',
          end_time: activities[i]['@_end_time'] || '',
          activity_type: activities[i]['@_type'] || ''
        });
      }
      
      // Add leg (except after the last activity)
      if (i < legs.length) {
        items.push({
          type: 'leg',
          mode: legs[i]['@_mode'] || '',
          dep_time: legs[i]['@_dep_time'] || '',
          duration: legs[i]['@_duration'] || ''
        });
      }
    }
    
    return items;
  });
};

// Build XML from plans array structure
export const buildXMLFromPlans = (person) => {
  if (!person.plans || person.plans === null) {
    return person.raw_xml; // Return original XML if plans not parsed
  }
  
  const xmlObj = {
    person: {
      '@_id': person.id,
      plan: person.plans.map((planArray, index) => ({
        '@_selected': index === 0 ? 'yes' : 'no',
        ...planArray.reduce((acc, item, idx) => {
          if (item.type === 'activity') {
            if (!acc.activity) acc.activity = [];
            acc.activity.push({
              '@_x': item.x,
              '@_y': item.y,
              '@_type': item.activity_type,
              '@_start_time': item.start_time,
              '@_end_time': item.end_time
            });
          } else if (item.type === 'leg') {
            if (!acc.leg) acc.leg = [];
            acc.leg.push({
              '@_mode': item.mode,
              '@_dep_time': item.dep_time,
              '@_duration': item.duration
            });
          }
          return acc;
        }, {})
      }))
    }
  };
  
  return buildPersonXML({ person: xmlObj });
};