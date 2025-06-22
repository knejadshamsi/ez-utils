import { Database } from 'bun:sqlite';
import { XMLParser } from 'fast-xml-parser';

// Worker thread for XML processing and database operations
export interface WorkerMessage {
  type: 'process_segment';
  segmentData: {
    segmentId: number;
    filePath: string;
    startByte: number;
    endByte: number;
    personsPerChunk: number;
    binSize: number;
    dbPath: string;
    isFirstSegment?: boolean;
    isLastSegment?: boolean;
  };
}

export interface WorkerResult {
  type: 'segment_complete';
  segmentId: number;
  processedCount: number;
  localBins: string[];
  duration: number;
  error?: string;
}

// Global origin point shared across workers
let originPoint: [number, number] | null = null;

// Listen for messages from main thread
self.onmessage = async (event: MessageEvent<WorkerMessage>) => {
  const { type, segmentData } = event.data;
  
  if (type === 'process_segment') {
    try {
      console.log(`[Worker ${segmentData.segmentId}] Starting segment processing`);
      const startTime = performance.now();
      
      // Create dedicated database connection for this worker
      const db = new Database(segmentData.dbPath);
      
      // Create tables if they don't exist
      db.exec(`
        CREATE TABLE IF NOT EXISTS persons (
          person_id TEXT PRIMARY KEY,
          coordinates TEXT NOT NULL,
          bin_row INTEGER NOT NULL,
          bin_column INTEGER NOT NULL,
          raw_xml TEXT NOT NULL,
          pct_1 BOOLEAN DEFAULT FALSE,
          pct_2 BOOLEAN DEFAULT FALSE,
          pct_3 BOOLEAN DEFAULT FALSE,
          pct_4 BOOLEAN DEFAULT FALSE,
          pct_5 BOOLEAN DEFAULT FALSE,
          pct_6 BOOLEAN DEFAULT FALSE,
          pct_7 BOOLEAN DEFAULT FALSE,
          pct_8 BOOLEAN DEFAULT FALSE,
          pct_9 BOOLEAN DEFAULT FALSE,
          pct_10 BOOLEAN DEFAULT FALSE
        )
      `);
      
      const result = await processFileSegment(segmentData, db);
      
      db.close();
      
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      console.log(`[Worker ${segmentData.segmentId}] Completed: ${result.processedCount} persons in ${duration.toFixed(0)}ms`);
      
      // Send result back to main thread
      self.postMessage({
        type: 'segment_complete',
        segmentId: segmentData.segmentId,
        processedCount: result.processedCount,
        localBins: Array.from(result.localBins),
        duration
      } as WorkerResult);
      
    } catch (error) {
      console.error(`[Worker ${segmentData.segmentId}] Error:`, error);
      
      self.postMessage({
        type: 'segment_complete',
        segmentId: segmentData.segmentId,
        processedCount: 0,
        localBins: [],
        duration: 0,
        error: error.message
      } as WorkerResult);
    }
  }
};

async function processFileSegment(
  segmentData: WorkerMessage['segmentData'], 
  db: Database
): Promise<{ processedCount: number; localBins: Set<string> }> {
  
  const { filePath, startByte, endByte, personsPerChunk, binSize, segmentId, isFirstSegment, isLastSegment } = segmentData;
  
  // Use Node.js fs for offset-based reading
  const fs = require('fs');
  const stream = fs.createReadStream(filePath, {
    start: startByte,
    end: endByte,
    encoding: 'utf8'
  });
  
  let buffer = '';
  let personBuffer: string[] = [];
  let processedCount = 0;
  const localBins = new Set<string>();
  
  // XML parser setup
  const xmlParser = new XMLParser({
    ignoreAttributes: false,
    parseAttributeValue: true
  });
  
  // Prepare batch insert statement
  const insertStmt = db.prepare(`
    INSERT OR REPLACE INTO persons (
      person_id, coordinates, bin_row, bin_column, raw_xml
    ) VALUES (?, ?, ?, ?, ?)
  `);
  
  return new Promise((resolve, reject) => {
    let skipToFirstPerson = !isFirstSegment; // Skip partial XML at start unless first segment
    
    stream.on('data', (chunk: string) => {
      buffer += chunk;
      
      // Skip to first complete person if not the first segment
      if (skipToFirstPerson) {
        const firstPersonStart = buffer.indexOf('<person ');
        if (firstPersonStart !== -1) {
          buffer = buffer.substring(firstPersonStart);
          skipToFirstPerson = false;
        } else {
          buffer = ''; // No person found in this chunk, wait for more
          return;
        }
      }
      
      // Extract complete person elements
      let startPos = 0;
      while (true) {
        const personStart = buffer.indexOf('<person ', startPos);
        if (personStart === -1) break;
        
        const personEnd = buffer.indexOf('</person>', personStart);
        if (personEnd === -1) break; // Wait for complete person
        
        // Extract complete person XML
        const personXml = buffer.substring(personStart, personEnd + 9);
        
        // For non-last segments, check if this person might be cut off at segment boundary
        if (!isLastSegment && personEnd > (endByte - startByte) * 0.9) {
          // Near segment end, might be incomplete - break to avoid duplicates
          break;
        }
        
        personBuffer.push(personXml);
        
        // Process chunk when we have enough persons
        if (personBuffer.length >= personsPerChunk) {
          const processed = processPersonChunk(personBuffer, xmlParser, insertStmt, binSize, localBins, db);
          processedCount += processed;
          personBuffer = [];
          
          console.log(`[Worker ${segmentId}] Processed chunk: ${processed} persons (total: ${processedCount})`);
        }
        
        startPos = personEnd + 9;
      }
      
      // Keep remaining incomplete data in buffer
      buffer = buffer.substring(startPos);
    });
    
    stream.on('end', () => {
      // Process final chunk if there are remaining persons
      if (personBuffer.length > 0) {
        const processed = processPersonChunk(personBuffer, xmlParser, insertStmt, binSize, localBins, db);
        processedCount += processed;
        console.log(`[Worker ${segmentId}] Final chunk: ${processed} persons (total: ${processedCount})`);
      }
      
      resolve({ processedCount, localBins });
    });
    
    stream.on('error', (error) => {
      reject(error);
    });
  });
}

function processPersonChunk(
  personBuffer: string[],
  xmlParser: XMLParser,
  insertStmt: any,
  binSize: number,
  localBins: Set<string>,
  db: Database
): number {
  
  const batchData: Array<[string, string, number, number, string]> = [];
  
  for (const personXml of personBuffer) {
    try {
      // Parse person XML
      const personData = xmlParser.parse(personXml);
      const person = personData.person;
      
      if (!person || !person.plan || !person.plan.activity) {
        continue;
      }
      
      const personId = person.id || person['@_id'];
      const activity = Array.isArray(person.plan.activity) 
        ? person.plan.activity[0] 
        : person.plan.activity;
      
      const x = parseFloat(activity.x || activity['@_x']);
      const y = parseFloat(activity.y || activity['@_y']);
      
      if (isNaN(x) || isNaN(y)) {
        continue;
      }
      
      // Set origin point if not set
      if (!originPoint) {
        originPoint = [x, y];
      }
      
      // Calculate bin coordinates
      const binRow = Math.floor((y - originPoint[1]) / binSize);
      const binColumn = Math.floor((x - originPoint[0]) / binSize);
      const binKey = `${binRow},${binColumn}`;
      
      localBins.add(binKey);
      
      // Prepare batch data
      batchData.push([
        personId.toString(),
        `${x},${y}`,
        binRow,
        binColumn,
        personXml
      ]);
      
    } catch (error) {
      console.error(`Error processing person XML:`, error.message);
      continue;
    }
  }
  
  // Batch insert using transaction
  if (batchData.length > 0) {
    const transaction = db.transaction(() => {
      for (const row of batchData) {
        insertStmt.run(...row);
      }
    });
    
    transaction();
  }
  
  return batchData.length;
}