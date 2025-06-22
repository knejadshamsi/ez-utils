import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
import { resolve } from 'path';
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
    new winston.transports.File({ filename: 'multithread-debug.log' })
  ]
});

interface MultiThreadProcessorParams {
  inputFilePath: string;
  personsPerChunk: number;
  maxConcurrency: number;
  dbConnection: Database;
  binSize: number;
  arrayMutex: Mutex;
}

interface WorkerResult {
  type: 'segment_complete';
  segmentId: number;
  processedCount: number;
  localBins: string[];
  duration: number;
  error?: string;
}

export async function multiThreadProcessor(params: MultiThreadProcessorParams): Promise<{
  totalProcessed: number;
  uniqueBins: Set<string>;
  chunkCount: number;
}> {
  const { inputFilePath, maxConcurrency, personsPerChunk, dbConnection, binSize } = params;
  
  // Get file size
  const file = Bun.file(inputFilePath);
  const fileSize = file.size;
  const dbPath = resolve(process.cwd(), 'population.db');
  
  logger.info(`=== STARTING MULTITHREADED PROCESSOR ===`);
  logger.info(`File size: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  logger.info(`Max concurrency: ${maxConcurrency} worker threads`);
  logger.info(`Persons per chunk: ${personsPerChunk}`);
  logger.info(`Database path: ${dbPath}`);
  
  console.log(`Multithreaded Processing - File size: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  console.log(`Creating ${maxConcurrency} worker threads`);
  
  // Calculate segment sizes with overlap for XML boundary handling
  const segmentSize = Math.floor(fileSize / maxConcurrency);
  const overlapSize = 50000; // 50KB overlap buffer for XML boundaries
  const segments: Array<{start: number, end: number, segmentId: number}> = [];
  
  for (let i = 0; i < maxConcurrency; i++) {
    const start = i === 0 ? 0 : (i * segmentSize) - overlapSize;
    const end = i === maxConcurrency - 1 ? fileSize - 1 : (start + segmentSize + overlapSize - 1);
    segments.push({ start, end, segmentId: i + 1 });
  }
  
  console.log(`Created ${segments.length} file segments of ~${(segmentSize / 1024 / 1024).toFixed(2)} MB each`);
  logger.info(`Created ${segments.length} segments for worker threads`);
  
  const globalUniqueBins = new Set<string>();
  let totalProcessed = 0;
  const startTime = performance.now();
  
  // Create worker threads
  const workers: Worker[] = [];
  const workerResults: Promise<WorkerResult>[] = [];
  
  for (let i = 0; i < segments.length; i++) {
    const segment = segments[i];
    const workerPath = resolve(import.meta.dir, 'xmlWorker.ts');
    
    logger.info(`WORKER_START: Creating worker ${segment.segmentId} for bytes ${segment.start}-${segment.end}`);
    console.log(`[${new Date().toISOString()}] Worker ${segment.segmentId} THREAD START (bytes ${segment.start}-${segment.end})`);
    
    // Create worker thread
    const worker = new Worker(workerPath);
    workers.push(worker);
    
    // Create promise to handle worker result
    const workerPromise = new Promise<WorkerResult>((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error(`Worker ${segment.segmentId} timeout after 10 minutes`));
      }, 10 * 60 * 1000); // 10 minute timeout
      
      worker.onmessage = (event: MessageEvent<WorkerResult>) => {
        clearTimeout(timeout);
        const result = event.data;
        
        if (result.type === 'segment_complete') {
          if (result.error) {
            reject(new Error(`Worker ${result.segmentId} error: ${result.error}`));
          } else {
            resolve(result);
          }
        }
      };
      
      worker.onerror = (error) => {
        clearTimeout(timeout);
        reject(error);
      };
    });
    
    workerResults.push(workerPromise);
    
    // Send work to worker thread
    worker.postMessage({
      type: 'process_segment',
      segmentData: {
        segmentId: segment.segmentId,
        filePath: inputFilePath,
        startByte: segment.start,
        endByte: segment.end,
        personsPerChunk,
        binSize,
        dbPath,
        isFirstSegment: i === 0,
        isLastSegment: i === maxConcurrency - 1
      }
    });
  }
  
  console.log(`[${new Date().toISOString()}] MULTITHREADED EXECUTION: Processing ${workerResults.length} segments with worker threads`);
  logger.info(`MULTITHREAD_EXECUTION: Started ${workerResults.length} worker threads`);
  
  // Wait for all workers to complete
  try {
    const results = await Promise.all(workerResults);
    
    const endTime = performance.now();
    const totalDuration = endTime - startTime;
    
    // Merge results from all workers
    results.forEach(result => {
      totalProcessed += result.processedCount;
      result.localBins.forEach(bin => globalUniqueBins.add(bin));
      
      logger.info(`WORKER_COMPLETE: Worker ${result.segmentId} processed ${result.processedCount} persons in ${result.duration.toFixed(0)}ms`);
      console.log(`[${new Date().toISOString()}] Worker ${result.segmentId} THREAD COMPLETE: ${result.processedCount} persons in ${result.duration.toFixed(0)}ms`);
    });
    
    logger.info(`MULTITHREAD_COMPLETE: All workers finished in ${totalDuration.toFixed(0)}ms`);
    logger.info(`MULTITHREAD_STATS: Total processed ${totalProcessed} persons, ${globalUniqueBins.size} unique bins`);
    
    console.log(`[${new Date().toISOString()}] MULTITHREADED EXECUTION COMPLETE: ${totalProcessed} persons in ${totalDuration.toFixed(0)}ms`);
    console.log(`Multithreaded processing performance: ${(totalProcessed / (totalDuration / 1000)).toFixed(0)} persons/second`);
    
    // Terminate all workers
    workers.forEach(worker => worker.terminate());
    
    return {
      totalProcessed,
      uniqueBins: globalUniqueBins,
      chunkCount: results.length
    };
    
  } catch (error) {
    // Terminate all workers on error
    workers.forEach(worker => worker.terminate());
    
    logger.error(`MULTITHREAD_ERROR: ${error.message}`);
    console.error(`Multithreaded processing failed:`, error);
    throw error;
  }
}