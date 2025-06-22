import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
import { createReadStream } from 'fs';
import { XMLParser } from 'fast-xml-parser';
import winston from 'winston';

// Logger setup
const logger = winston.createLogger({
  level: 'debug',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.printf(({ timestamp, level, message }) => {
      return `${timestamp} [${level.toUpperCase()}] ${message}`;
    })
  ),
  transports: [
    new winston.transports.File({ filename: 'parallel-debug.log' })
  ]
});

interface ParallelFileProcessorParams {
  inputFilePath: string;
  personsPerChunk: number;
  maxConcurrency: number;
  dbConnection: Database;
  binSize: number;
  arrayMutex: Mutex;
}

interface ProcessingResult {
  processedCount: number;
  localBins: Set<string>;
  originSet: boolean;
}

export async function parallelFileProcessor(params: ParallelFileProcessorParams): Promise<{
  totalProcessed: number;
  uniqueBins: Set<string>;
  chunkCount: number;
}> {
  const { inputFilePath, maxConcurrency, personsPerChunk, dbConnection, binSize, arrayMutex } = params;
  
  // Get file size
  const file = Bun.file(inputFilePath);
  const fileSize = file.size;
  
  logger.info(`=== STARTING PARALLEL FILE PROCESSOR ===`);
  logger.info(`File size: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  logger.info(`Max concurrency: ${maxConcurrency}`);
  logger.info(`Persons per chunk: ${personsPerChunk}`);
  
  console.log(`Parallel File Processing - File size: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  console.log(`Creating ${maxConcurrency} parallel file segments`);
  
  // Calculate segment sizes
  const segmentSize = Math.floor(fileSize / maxConcurrency);
  const segments: Array<{start: number, end: number, segmentId: number}> = [];
  
  for (let i = 0; i < maxConcurrency; i++) {
    const start = i * segmentSize;
    const end = i === maxConcurrency - 1 ? fileSize - 1 : (start + segmentSize - 1);
    segments.push({ start, end, segmentId: i + 1 });
  }
  
  console.log(`Created ${segments.length} file segments of ~${(segmentSize / 1024 / 1024).toFixed(2)} MB each`);
  logger.info(`Created ${segments.length} segments`);
  
  const globalUniqueBins = new Set<string>();
  let totalProcessed = 0;
  
  // Import the processor
  const { personPipelineProcessor } = await import('./personPipelineProcessor');
  
  // Process segments in parallel using Promise.all
  const startTime = performance.now();
  
  const promises = segments.map(async (segment, index) => {
    const segmentStartTime = performance.now();
    
    logger.info(`SEGMENT_START: Segment ${segment.segmentId} starting (bytes ${segment.start}-${segment.end})`);
    console.log(`[${new Date().toISOString()}] Segment ${segment.segmentId} PARALLEL START (bytes ${segment.start}-${segment.end})`);
    
    return new Promise<{segmentId: number, processedCount: number, localBins: Set<string>}>((resolve, reject) => {
      // Create read stream with offset
      const stream = createReadStream(inputFilePath, {
        start: segment.start,
        end: segment.end,
        encoding: 'utf8'
      });
      
      let buffer = '';
      let personBuffer: string[] = [];
      let segmentProcessedCount = 0;
      let segmentChunkCount = 0;
      const segmentBins = new Set<string>();
      
      stream.on('data', (chunk: string) => {
        buffer += chunk;
        
        // Extract complete person elements
        let startPos = 0;
        while (true) {
          const personStart = buffer.indexOf('<person ', startPos);
          if (personStart === -1) break;
          
          const personEnd = buffer.indexOf('</person>', personStart);
          if (personEnd === -1) break; // Wait for complete person
          
          // Extract complete person XML
          const personXml = buffer.substring(personStart, personEnd + 9);
          personBuffer.push(personXml);
          
          // Check if we have enough persons for a chunk
          if (personBuffer.length >= personsPerChunk) {
            const chunkXml = personBuffer.join('\n');
            segmentChunkCount++;
            
            logger.info(`SEGMENT_CHUNK: Segment ${segment.segmentId} chunk ${segmentChunkCount} ready with ${personBuffer.length} persons`);
            
            // Process chunk
            personPipelineProcessor(chunkXml, dbConnection, binSize, arrayMutex)
              .then(result => {
                segmentProcessedCount += result.processedCount;
                result.localBins.forEach(bin => segmentBins.add(bin));
                logger.info(`SEGMENT_CHUNK_DONE: Segment ${segment.segmentId} chunk ${segmentChunkCount} processed ${result.processedCount} persons`);
              })
              .catch(error => {
                logger.error(`SEGMENT_CHUNK_ERROR: Segment ${segment.segmentId} chunk ${segmentChunkCount} failed: ${error.message}`);
              });
            
            // Reset for next chunk
            personBuffer = [];
          }
          
          // Move start position past this person
          startPos = personEnd + 9;
        }
        
        // Keep remaining incomplete data in buffer
        buffer = buffer.substring(startPos);
      });
      
      stream.on('end', async () => {
        // Process final chunk if there are remaining persons
        if (personBuffer.length > 0) {
          const chunkXml = personBuffer.join('\n');
          segmentChunkCount++;
          
          try {
            const result = await personPipelineProcessor(chunkXml, dbConnection, binSize, arrayMutex);
            segmentProcessedCount += result.processedCount;
            result.localBins.forEach(bin => segmentBins.add(bin));
            logger.info(`SEGMENT_FINAL_CHUNK: Segment ${segment.segmentId} final chunk processed ${result.processedCount} persons`);
          } catch (error) {
            logger.error(`SEGMENT_FINAL_ERROR: Segment ${segment.segmentId} final chunk failed: ${error.message}`);
          }
        }
        
        const segmentEndTime = performance.now();
        const segmentDuration = segmentEndTime - segmentStartTime;
        
        logger.info(`SEGMENT_COMPLETE: Segment ${segment.segmentId} finished in ${segmentDuration.toFixed(0)}ms - ${segmentProcessedCount} persons, ${segmentBins.size} bins`);
        console.log(`[${new Date().toISOString()}] Segment ${segment.segmentId} PARALLEL COMPLETE: ${segmentProcessedCount} persons in ${segmentDuration.toFixed(0)}ms`);
        
        resolve({
          segmentId: segment.segmentId,
          processedCount: segmentProcessedCount,
          localBins: segmentBins
        });
      });
      
      stream.on('error', (error) => {
        logger.error(`SEGMENT_STREAM_ERROR: Segment ${segment.segmentId} stream error: ${error.message}`);
        reject(error);
      });
    });
  });
  
  // Wait for all segments to complete
  logger.info(`PARALLEL_EXECUTION: Starting ${promises.length} parallel segments`);
  console.log(`[${new Date().toISOString()}] PARALLEL EXECUTION: Processing ${promises.length} segments simultaneously`);
  
  const results = await Promise.all(promises);
  
  const endTime = performance.now();
  const totalDuration = endTime - startTime;
  
  // Merge results
  results.forEach(result => {
    totalProcessed += result.processedCount;
    result.localBins.forEach(bin => globalUniqueBins.add(bin));
  });
  
  logger.info(`PARALLEL_COMPLETE: All segments finished in ${totalDuration.toFixed(0)}ms`);
  logger.info(`PARALLEL_STATS: Total processed ${totalProcessed} persons, ${globalUniqueBins.size} unique bins`);
  
  console.log(`[${new Date().toISOString()}] PARALLEL EXECUTION COMPLETE: ${totalProcessed} persons in ${totalDuration.toFixed(0)}ms`);
  console.log(`Parallel processing performance: ${(totalProcessed / (totalDuration / 1000)).toFixed(0)} persons/second`);
  
  return {
    totalProcessed,
    uniqueBins: globalUniqueBins,
    chunkCount: results.length
  };
}