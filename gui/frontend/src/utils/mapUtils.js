/**
 * Calculates a route path from a given plan.
 * A plan is an array of activities and legs.
 * This function extracts activities with coordinates and returns them as a path.
 *
 * @param {Array} plan - The plan containing activities.
 * @returns {{path: Array<[number, number]>} | null} - An object with the path array or null if no valid path can be created.
 */
export const calculateRouteFromPlan = (plan) => {
  if (!plan) {
    return null;
  }

  const activities = plan.filter(item => item.type === 'activity');

  if (activities.length > 1) {
    const path = activities
      .filter(a => a.x && a.y)
      .map(a => [a.x, a.y]);

    if (path.length > 1) {
      return { path };
    }
  }

  return null;
};