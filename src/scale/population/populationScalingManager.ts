import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
import { existsSync, mkdirSync } from 'fs';
import { resolve } from 'path';

interface PopulationScalingParams {
  inputFilePath: string;
  requestedScales: number[];
  maxConcurrency: number;
  dbConfig: {
    filename: string;
    timeout: number;
  };
  binSize: number;
  targetDensity: number;
  outputPath: string;
  cleanDatabase?: boolean;
}

export async function populationScalingManager(params: PopulationScalingParams): Promise<void> {
  console.log('Starting population scaling workflow...');
  
  // Initialize mutex for origin point
  const arrayMutex = new Mutex();
  
  // Initialize database
  const dbPath = resolve(process.cwd(), params.dbConfig.filename);
  if (params.cleanDatabase && existsSync(dbPath)) {
    console.log('Cleaning up existing database...');
    // Remove the existing database file
    const { unlinkSync } = await import('fs');
    unlinkSync(dbPath);
  }
  
  const db = new Database(dbPath);
  
  // Create person table schema
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
  
  // Create output directory if it doesn't exist
  if (!existsSync(params.outputPath)) {
    mkdirSync(params.outputPath, { recursive: true });
  }
  
  console.log('Database and output directory initialized');
  
  // Import processors
  const { multiThreadProcessor } = await import('./multiThreadProcessor');
  const { binCountCalculator } = await import('./binCountCalculator');
  const { binScalingProcessor } = await import('./binScalingProcessor');
  const { scaleFileMerger } = await import('./scaleFileMerger');
  const { loadConfig } = await import('../../configLoader');
  const config = loadConfig();
  
  // 1. Multithreaded processing with worker threads - TRUE PARALLELISM
  console.log(`Phase 1: Multithreaded processing with ${params.maxConcurrency} worker threads...`);
  
  const streamingResult = await multiThreadProcessor({
    inputFilePath: params.inputFilePath,
    personsPerChunk: config.personsPerChunk,
    maxConcurrency: params.maxConcurrency,
    dbConnection: db,
    binSize: params.binSize,
    arrayMutex
  });
  
  console.log(`Phase 1 complete: ${streamingResult.totalProcessed} persons processed`);
  console.log(`Total chunks: ${streamingResult.chunkCount}`);
  console.log(`Unique bins discovered: ${streamingResult.uniqueBins.size}`);
  
  // 2. Call BinCountCalculator
  console.log('Phase 2: Calculating bin counts and retention ratios...');
  const binCountResult = await binCountCalculator(
    streamingResult.uniqueBins,
    params.requestedScales,
    params.targetDensity,
    db
  );
  
  console.log(`Bin density statistics:
  - Min: ${binCountResult.densityStats.min} persons/bin
  - Max: ${binCountResult.densityStats.max} persons/bin
  - Average: ${binCountResult.densityStats.average.toFixed(2)} persons/bin`);
  
  console.log(`Retention ratios calculated for ${params.requestedScales.length} scales`);
  
  // 3. Parallel BinScalingProcessor operations
  console.log(`Phase 3: Processing ${streamingResult.uniqueBins.size} bins for scaling...`);
  
  const scalingPromises: Promise<any>[] = [];
  const scalingSemaphore = new Mutex();
  let activeScalingCount = 0;
  
  const binsArray = Array.from(streamingResult.uniqueBins);
  for (let i = 0; i < binsArray.length; i++) {
    // Wait if we've reached max concurrency
    while (activeScalingCount >= params.maxConcurrency) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    await scalingSemaphore.runExclusive(async () => {
      activeScalingCount++;
    });
    
    const binKey = binsArray[i];
    if (!binKey) continue;
    
    const scalingPromise = binScalingProcessor(
      binKey,
      binCountResult.retentionRatios,
      params.requestedScales,
      db,
      params.outputPath
    ).then(result => {
      console.log(`Bin ${i + 1}/${binsArray.length} processed: ${result.updatedRecords} records updated`);
      return result;
    }).finally(async () => {
      await scalingSemaphore.runExclusive(async () => {
        activeScalingCount--;
      });
    });
    
    scalingPromises.push(scalingPromise);
  }
  
  // Wait for all bins to complete
  const scalingResults = await Promise.all(scalingPromises);
  
  const totalUpdated = scalingResults.reduce((sum, result) => sum + result.updatedRecords, 0);
  console.log(`Phase 3 complete: ${totalUpdated} total database updates`);
  
  // 4. Call ScaleFileMerger
  console.log('Phase 4: Merging chunk files into final population files...');
  const mergerResult = await scaleFileMerger(
    params.requestedScales,
    params.outputPath,
    params.inputFilePath
  );
  
  console.log(`Phase 4 complete: ${mergerResult.finalFiles.length} final files created`);
  for (const scale of params.requestedScales) {
    const stats = mergerResult.mergeStats[scale];
    if (stats) {
      console.log(`  Scale ${scale}%: ${stats.personCount} persons from ${stats.chunkCount} chunks (${(stats.fileSize / 1024 / 1024).toFixed(2)} MB)`);
    }
  }
  
  if (mergerResult.cleanupStatus.errors.length > 0) {
    console.log(`Warning: ${mergerResult.cleanupStatus.errors.length} cleanup errors occurred`);
  }
  
  db.close();
  console.log('Population scaling workflow complete');
}