import { Database } from 'bun:sqlite';

interface BinCountResult {
  binCounts: Record<string, number>;
  retentionRatios: Record<number, Record<string, number>>;
  densityStats: {
    min: number;
    max: number;
    average: number;
  };
}

export async function binCountCalculator(
  uniqueBins: Set<string>,
  requestedScales: number[],
  targetDensity: number,
  dbConnection: Database
): Promise<BinCountResult> {
  const binCounts: Record<string, number> = {};
  const retentionRatios: Record<number, Record<string, number>> = {};
  
  // Initialize retention ratios structure for each scale
  requestedScales.forEach(scale => {
    retentionRatios[scale] = {};
  });

  // Query count for each unique bin
  const countStmt = dbConnection.prepare(
    'SELECT COUNT(*) as count FROM persons WHERE bin_row = ? AND bin_column = ?'
  );

  for (const binKey of uniqueBins) {
    // Parse row and column from bin key
    const [rowStr, columnStr] = binKey.split(',');
    const rowIndex = parseInt(rowStr);
    const columnIndex = parseInt(columnStr);

    // Get population count for this bin
    const result = countStmt.get(rowIndex, columnIndex) as { count: number };
    const actualDensity = result.count;
    
    // Store bin count
    binCounts[binKey] = actualDensity;

    // Calculate retention ratio for each scale
    for (const scaleValue of requestedScales) {
      // Convert scale percentage to decimal (e.g., 10% -> 0.1)
      const scaleFactor = scaleValue / 100;
      
      // Calculate retention ratio using formula: min(1.0, (targetDensity * scale) / actualDensity)
      let retentionRatio: number;
      if (actualDensity === 0) {
        retentionRatio = 0;
      } else {
        retentionRatio = Math.min(1.0, (targetDensity * scaleFactor) / actualDensity);
      }
      
      retentionRatios[scaleValue][binKey] = retentionRatio;
    }
  }

  // Calculate density statistics
  const densityValues = Object.values(binCounts);
  const densityStats = {
    min: Math.min(...densityValues),
    max: Math.max(...densityValues),
    average: densityValues.reduce((sum, val) => sum + val, 0) / densityValues.length
  };

  return {
    binCounts,
    retentionRatios,
    densityStats
  };
}