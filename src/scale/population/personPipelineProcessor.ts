import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
import { XMLParser, XMLBuilder } from 'fast-xml-parser';

interface PersonPipelineResult {
  processedCount: number;
  binCoordinates: Array<{ row: number; column: number }>;
  originSet: boolean;
  localBins: Set<string>;
}

// Global origin point shared across all processors
let originPoint: [number, number] | null = null;

export async function personPipelineProcessor(
  chunkBuffer: string,
  dbConnection: Database,
  binSize: number,
  arrayMutex: Mutex
): Promise<PersonPipelineResult> {
  console.log(`PersonPipelineProcessor: Starting processing for chunk (${chunkBuffer.length} bytes)`);
  
  const processedBins: Array<{ row: number; column: number }> = [];
  const localBins = new Set<string>();
  let processedCount = 0;
  let originSet = false;

  // Prepare batch for database insertion
  const batchData: Array<{
    personId: string;
    coordinates: string;
    binRow: number;
    binColumn: number;
    rawXml: string;
  }> = [];

  // Initialize XML parser and builder
  const parser = new XMLParser({
    ignoreAttributes: false,
    attributeNamePrefix: '',
    parseAttributeValue: true,
    textNodeName: '_text'
  });
  
  const builder = new XMLBuilder({
    ignoreAttributes: false,
    attributeNamePrefix: '',
    format: false,
    suppressEmptyNode: true
  });

  // Wrap chunk in root element to make valid XML
  const wrappedXml = `<root>${chunkBuffer}</root>`;
  console.log(`PersonPipelineProcessor: Starting XML parsing...`);
  
  try {
    const parsed = parser.parse(wrappedXml);
    console.log(`PersonPipelineProcessor: XML parsing completed successfully`);
    
    // Handle case where there's only one person (not an array)
    const persons = parsed.root?.person;
    if (!persons) {
      return {
        processedCount: 0,
        binCoordinates: [],
        originSet: false,
        localBins
      };
    }
    
    // Ensure persons is always an array
    const personArray = Array.isArray(persons) ? persons : [persons];
    console.log(`PersonPipelineProcessor: Found ${personArray.length} persons in chunk`);
    
    for (const person of personArray) {
      // Extract person ID
      const personId = person.id;
      if (!personId) continue;
      
      // Find the plan (handle single plan or array of plans)
      let selectedPlan = null;
      if (person.plan) {
        if (Array.isArray(person.plan)) {
          // Multiple plans - find selected="yes" or use first
          selectedPlan = person.plan.find((p: any) => p.selected === 'yes') || person.plan[0];
        } else {
          // Single plan
          selectedPlan = person.plan;
        }
      }
      
      if (!selectedPlan || !selectedPlan.activity) continue;
      
      // Get first activity (handle single activity or array)
      const activity = Array.isArray(selectedPlan.activity) 
        ? selectedPlan.activity[0] 
        : selectedPlan.activity;
      
      if (!activity || typeof activity.x !== 'number' || typeof activity.y !== 'number') continue;
      
      const coordinates: [number, number] = [activity.x, activity.y];
      
      // Set origin point if not set (with double-check locking)
      if (originPoint === null) {
        await arrayMutex.runExclusive(async () => {
          if (originPoint === null) {
            originPoint = coordinates;
            originSet = true;
          }
        });
      }

      // Calculate bin coordinates
      const binRow = Math.floor((coordinates[1] - originPoint![1]) / binSize);
      const binColumn = Math.floor((coordinates[0] - originPoint![0]) / binSize);
      const binKey = `${binRow},${binColumn}`;

      // Track unique bins locally (no mutex needed)
      localBins.add(binKey);

      // Reconstruct raw XML for this person (for database storage)
      const rawXml = builder.build({ person });

      // Add to batch instead of immediate insert
      const pointGeometry = `POINT(${coordinates[0]} ${coordinates[1]})`;
      batchData.push({
        personId: personId.toString(),
        coordinates: pointGeometry,
        binRow,
        binColumn,
        rawXml
      });

      processedBins.push({ row: binRow, column: binColumn });
      processedCount++;
    }
  } catch (error) {
    console.error('PersonPipelineProcessor: XML parsing error:', error);
    // Return empty result on parse error
    return {
      processedCount: 0,
      binCoordinates: [],
      originSet: false,
      localBins: new Set<string>()
    };
  }

  // Perform batch insert in a single transaction
  console.log(`PersonPipelineProcessor: Starting database insertion for ${batchData.length} persons`);
  if (batchData.length > 0) {
    const insertStmt = dbConnection.prepare(`
      INSERT OR IGNORE INTO persons (person_id, coordinates, bin_row, bin_column, raw_xml, 
                          pct_1, pct_2, pct_3, pct_4, pct_5, 
                          pct_6, pct_7, pct_8, pct_9, pct_10)
      VALUES (?, ?, ?, ?, ?, FALSE, FALSE, FALSE, FALSE, FALSE, 
              FALSE, FALSE, FALSE, FALSE, FALSE)
    `);

    // Use transaction for batch insert
    const insertBatch = dbConnection.transaction((batch) => {
      for (const person of batch) {
        insertStmt.run(
          person.personId,
          person.coordinates,
          person.binRow,
          person.binColumn,
          person.rawXml
        );
      }
    });

    try {
      insertBatch(batchData);
      console.log(`PersonPipelineProcessor: Database insertion completed successfully for ${batchData.length} persons`);
    } catch (error) {
      console.error(`PersonPipelineProcessor: Database insertion failed:`, error);
      throw error;
    }
  }

  console.log(`PersonPipelineProcessor: Completed processing - ${processedCount} persons processed, ${localBins.size} unique bins`);
  return {
    processedCount,
    binCoordinates: processedBins,
    originSet,
    localBins
  };
}