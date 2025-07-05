// ID generation utilities for person creation

/**
 * Generate a new person ID using the "Highest + 1" method
 * @param {Array} existingPersons - Array of existing persons with IDs
 * @returns {string|null} New ID or null if non-numeric IDs found
 */
export const generateHighestPlusOneId = (existingPersons) => {
  let highestNumericId = 0;
  
  for (const person of existingPersons) {
    const id = person.id;
    const numericId = parseInt(id, 10);
    
    // Check if ID is numeric
    if (isNaN(numericId) || numericId.toString() !== id.trim()) {
      return null; // Non-numeric ID found
    }
    
    if (numericId > highestNumericId) {
      highestNumericId = numericId;
    }
  }
  
  return (highestNumericId + 1).toString();
};

/**
 * Generate a random alphanumeric ID
 * @param {Array} existingPersons - Array of existing persons with IDs
 * @param {number} length - Length of the ID (default: 8)
 * @returns {string} Random unique ID
 */
export const generateRandomId = (existingPersons, length = 8) => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  const existingIds = new Set(existingPersons.map(p => p.id));
  
  let newId;
  let attempts = 0;
  const maxAttempts = 1000;
  
  do {
    newId = '';
    for (let i = 0; i < length; i++) {
      newId += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    attempts++;
    
    if (attempts > maxAttempts) {
      throw new Error('Unable to generate unique ID after maximum attempts');
    }
  } while (existingIds.has(newId));
  
  return newId;
};

/**
 * Detect the pattern of existing IDs
 * @param {Array} existingPersons - Array of existing persons with IDs
 * @returns {Object} Pattern information { type: 'numeric'|'alphanumeric', canUseHighestPlusOne: boolean }
 */
export const detectIdPattern = (existingPersons) => {
  if (existingPersons.length === 0) {
    return { type: 'numeric', canUseHighestPlusOne: true };
  }
  
  let allNumeric = true;
  
  for (const person of existingPersons) {
    const id = person.id;
    const numericId = parseInt(id, 10);
    
    if (isNaN(numericId) || numericId.toString() !== id.trim()) {
      allNumeric = false;
      break;
    }
  }
  
  return {
    type: allNumeric ? 'numeric' : 'alphanumeric',
    canUseHighestPlusOne: allNumeric
  };
};