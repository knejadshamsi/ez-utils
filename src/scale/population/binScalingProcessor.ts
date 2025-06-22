import { Database } from 'bun:sqlite';
import { writeFileSync, existsSync, mkdirSync } from 'fs';
import { join } from 'path';

interface BinScalingResult {
  updatedRecords: number;
  survivalStats: Record<number, { original: number; survived: number }>;
  processingTime: number;
}

export async function binScalingProcessor(
  binKey: string,
  retentionRatios: Record<number, Record<string, number>>,
  requestedScales: number[],
  dbConnection: Database,
  outputPath: string
): Promise<BinScalingResult> {
  const startTime = performance.now();
  
  // Parse row and column from bin key
  const [rowStr, columnStr] = binKey.split(',');
  const rowIndex = parseInt(rowStr);
  const columnIndex = parseInt(columnStr);
  
  // Query all persons in this bin
  const personQuery = dbConnection.prepare(
    'SELECT person_id, raw_xml FROM persons WHERE bin_row = ? AND bin_column = ?'
  );
  const personRecords = personQuery.all(rowIndex, columnIndex) as Array<{
    person_id: string;
    raw_xml: string;
  }>;
  
  const survivalStats: Record<number, { original: number; survived: number }> = {};
  let totalUpdated = 0;
  
  // Process each requested scale
  for (const scaleValue of requestedScales) {
    const retentionProbability = retentionRatios[scaleValue]?.[binKey] || 0;
    const survivingPersons: string[] = [];
    
    // Initialize survival stats
    survivalStats[scaleValue] = {
      original: personRecords.length,
      survived: 0
    };
    
    // Prepare the update statement for this scale's column
    const columnName = `pct_${scaleValue}`;
    const updateStmt = dbConnection.prepare(
      `UPDATE persons SET ${columnName} = ? WHERE person_id = ?`
    );
    
    // Process each person with coin toss
    for (const person of personRecords) {
      const randomValue = Math.random();
      const survives = randomValue < retentionProbability;
      
      // Update the boolean column
      updateStmt.run(survives ? 1 : 0, person.person_id);
      totalUpdated++;
      
      if (survives) {
        survivingPersons.push(person.raw_xml);
        survivalStats[scaleValue].survived++;
      }
    }
    
    // Create scale chunk file for surviving persons
    if (survivingPersons.length > 0) {
      // Ensure output directory exists
      const scaleDir = join(outputPath, `scale_${scaleValue}`);
      if (!existsSync(scaleDir)) {
        mkdirSync(scaleDir, { recursive: true });
      }
      
      // Generate chunk filename with format: scale-chunk-XX-ROWxCOLUMN
      const chunkFileName = `scale-chunk-${scaleValue.toString().padStart(2, '0')}-${rowIndex}x${columnIndex}.xml`;
      const chunkFilePath = join(scaleDir, chunkFileName);
      
      // Write surviving persons to chunk file (no population tags - added during merge)
      const chunkContent = survivingPersons.join('\n');
      
      writeFileSync(chunkFilePath, chunkContent, 'utf-8');
    }
  }
  
  const processingTime = performance.now() - startTime;
  
  return {
    updatedRecords: totalUpdated,
    survivalStats,
    processingTime
  };
}