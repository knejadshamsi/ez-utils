import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
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
    new winston.transports.File({ filename: 'xml-streaming-debug.log' })
  ]
});

interface XMLStreamingProcessorParams {
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

export async function xmlStreamingProcessor(params: XMLStreamingProcessorParams): Promise<{
  totalProcessed: number;
  uniqueBins: Set<string>;
  chunkCount: number;
}> {
  const { inputFilePath, personsPerChunk, maxConcurrency, dbConnection, binSize, arrayMutex } = params;
  
  // Use Bun's streaming API for large files
  const file = Bun.file(inputFilePath);
  const fileSize = file.size;
  
  logger.info(`=== STARTING XML STREAMING PROCESSOR ===`);
  logger.info(`Chunk size: ${personsPerChunk} persons per chunk`);
  logger.info(`Max concurrency: ${maxConcurrency}`);
  logger.info(`Input file: ${inputFilePath}`);
  logger.info(`File size: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  
  console.log(`XML Streaming - Chunk size: ${personsPerChunk} persons per chunk`);
  console.log(`XML Streaming - Max concurrency: ${maxConcurrency}`);
  
  logger.info(`Processing file: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  console.log(`Processing file: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  
  const stream = file.stream();
  const reader = stream.getReader();
  const decoder = new TextDecoder();

  let buffer = '';
  let activePipelines = 0;
  let chunkIndex = 0;
  let totalProcessed = 0;
  let personBuffer: string[] = [];
  
  const globalUniqueBins = new Set<string>();

  // Import the processor
  const { personPipelineProcessor } = await import('./personPipelineProcessor');

  // XML parser setup
  const xmlParser = new XMLParser({
    ignoreAttributes: false,
    parseAttributeValue: true
  });

  // Process stream chunks
  let bytesRead = 0;
  let ioStartTime = Date.now();
  let ioOperationCount = 0;
  let chunkExtractionTimes: number[] = [];
  let bufferSizes: number[] = [];
  
  logger.info(`Starting XML streaming loop`);
  logger.info(`IO_TIMING: Stream processing started at ${new Date(ioStartTime).toISOString()}`);
  
  while (true) {
    const ioOperationStart = performance.now();
    const { done, value } = await reader.read();
    const ioOperationEnd = performance.now();
    const ioLatency = ioOperationEnd - ioOperationStart;
    
    ioOperationCount++;
    
    if (done) {
      logger.info(`File reading complete`);
      logger.info(`IO_STATS: Total IO operations: ${ioOperationCount}`);
      logger.info(`IO_STATS: Average IO latency: ${(chunkExtractionTimes.reduce((a, b) => a + b, 0) / chunkExtractionTimes.length).toFixed(2)}ms`);
      break;
    }
    
    logger.debug(`IO_TIMING: Operation ${ioOperationCount} took ${ioLatency.toFixed(2)}ms`);
    chunkExtractionTimes.push(ioLatency);
    
    if (value) {
      bytesRead += value.byteLength;
      const progressPct = ((bytesRead / fileSize) * 100).toFixed(1);
      
      if (bytesRead % (500 * 1024 * 1024) < value.byteLength) { // Log every 500MB
        console.log(`Reading: ${progressPct}% (${(bytesRead / 1024 / 1024).toFixed(0)} MB / ${(fileSize / 1024 / 1024).toFixed(0)} MB)`);
      }
    }
    
    // Decode chunk and add to buffer
    const decodingStart = performance.now();
    buffer += decoder.decode(value, { stream: true });
    const decodingEnd = performance.now();
    const decodingLatency = decodingEnd - decodingStart;
    
    bufferSizes.push(buffer.length);
    logger.debug(`BUFFER_STATE: Size=${buffer.length} chars, Decoding took ${decodingLatency.toFixed(2)}ms`);
    
    // Extract complete person elements
    const extractionStart = performance.now();
    let startPos = 0;
    let personsExtractedThisIteration = 0;
    while (true) {
      const personStart = buffer.indexOf('<person ', startPos);
      if (personStart === -1) break;
      
      const personEnd = buffer.indexOf('</person>', personStart);
      if (personEnd === -1) break; // Wait for complete person
      
      // Extract complete person XML
      const personXml = buffer.substring(personStart, personEnd + 9); // +9 for '</person>'
      personBuffer.push(personXml);
      personsExtractedThisIteration++;
      
      logger.debug(`EXTRACTION: Person ${personBuffer.length} extracted, XML length: ${personXml.length}`);
      
      // Check if we have enough persons for a chunk
      if (personBuffer.length >= personsPerChunk) {
        const chunkReadyTime = performance.now();
        const timeSinceStart = chunkReadyTime - ioStartTime;
        
        logger.info(`CHUNK_READY: ${personBuffer.length} persons accumulated after ${timeSinceStart.toFixed(0)}ms`);
        logger.info(`CHUNK_READY: Buffer state - size: ${buffer.length} chars`);
        
        // Check if we can start a new chunk
        if (activePipelines < maxConcurrency) {
          const chunkStartTime = performance.now();
          // Create chunk XML that matches PersonPipelineProcessor expectations
          const chunkXml = personBuffer.join('\n');
          
          const currentChunkIndex = chunkIndex;
          activePipelines++;
          chunkIndex++;
          
          const chunkFireTime = performance.now();
          const chunkCreationLatency = chunkFireTime - chunkStartTime;
          
          logger.info(`CHUNK_FIRING: chunkIndex=${currentChunkIndex + 1}, activePipelines=${activePipelines}/${maxConcurrency}, personsInChunk=${personBuffer.length}`);
          logger.info(`CHUNK_TIMING: Chunk creation took ${chunkCreationLatency.toFixed(2)}ms`);
          
          console.log(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} FIRE (Active: ${activePipelines}/${maxConcurrency})`);
          
          // Fire and forget
          personPipelineProcessor(
            chunkXml,
            dbConnection,
            binSize,
            arrayMutex
          ).then(result => {
            logger.info(`CHUNK SUCCESS: chunk=${currentChunkIndex + 1}, processedCount=${result.processedCount}, localBins=${result.localBins.size}`);
            
            // Merge local bins into global set
            result.localBins.forEach(bin => globalUniqueBins.add(bin));
            totalProcessed += result.processedCount;
            
            console.log(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} DONE: ${result.processedCount} persons`);
          }).catch(error => {
            logger.error(`CHUNK ERROR: chunk=${currentChunkIndex + 1}, error=${error.message}`);
            console.error(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} ERROR:`, error);
          }).finally(() => {
            activePipelines--;
            logger.info(`SLOT FREED: chunk=${currentChunkIndex + 1}, activePipelines now=${activePipelines}/${maxConcurrency}`);
            console.log(`[${new Date().toISOString()}] SLOT FREED. Active: ${activePipelines}/${maxConcurrency}`);
          });

          // Reset for next chunk
          personBuffer = [];
          logger.info(`Person buffer reset for next chunk`);
        } else {
          const waitTime = performance.now();
          const timeSinceStart = waitTime - ioStartTime;
          
          logger.warn(`AT_MAX_CAPACITY: activePipelines=${activePipelines}/${maxConcurrency} at ${timeSinceStart.toFixed(0)}ms - chunk with ${personBuffer.length} persons WAITING`);
          logger.warn(`QUEUE_STATE: ${personBuffer.length} persons ready but blocked by capacity`);
          console.log(`At max capacity (${activePipelines}/${maxConcurrency}) - continuing to read...`);
        }
      }
      
      // Move start position past this person
      startPos = personEnd + 9;
    }
    
    // Keep remaining incomplete data in buffer
    buffer = buffer.substring(startPos);
    
    const extractionEnd = performance.now();
    const extractionLatency = extractionEnd - extractionStart;
    
    if (personsExtractedThisIteration > 0) {
      logger.debug(`EXTRACTION_BATCH: Extracted ${personsExtractedThisIteration} persons in ${extractionLatency.toFixed(2)}ms`);
      logger.debug(`EXTRACTION_BATCH: Remaining buffer size: ${buffer.length} chars`);
    }
  }
  
  // Process final chunk if there are remaining persons
  if (personBuffer.length > 0) {
    const chunkXml = personBuffer.join('\n');
    
    const currentChunkIndex = chunkIndex;
    activePipelines++;
    chunkIndex++;
    
    logger.info(`FINAL CHUNK: ${personBuffer.length} persons`);
    console.log(`Final chunk ${currentChunkIndex + 1} started (Active: ${activePipelines}/${maxConcurrency})`);
    
    personPipelineProcessor(
      chunkXml,
      dbConnection,
      binSize,
      arrayMutex
    ).then(result => {
      result.localBins.forEach(bin => globalUniqueBins.add(bin));
      totalProcessed += result.processedCount;
      console.log(`Final chunk ${currentChunkIndex + 1} completed: ${result.processedCount} persons`);
    }).catch(error => {
      console.error(`Final chunk ${currentChunkIndex + 1} failed:`, error);
    }).finally(() => {
      activePipelines--;
      console.log(`Final slot freed. Active: ${activePipelines}/${maxConcurrency}`);
    });
  }

  console.log(`File reading complete: ${chunkIndex} chunks created`);
  
  // Wait for all processing to complete
  logger.info(`WAITING FOR COMPLETION: ${activePipelines} active pipelines remaining`);
  while (activePipelines > 0) {
    console.log(`Waiting for ${activePipelines} active pipelines to complete...`);
    await new Promise(resolve => setTimeout(resolve, 500));
  }

  console.log(`All processing complete:`);
  console.log(`- Total chunks processed: ${chunkIndex}`);
  console.log(`- Total persons processed: ${totalProcessed}`);
  console.log(`- Unique bins discovered: ${globalUniqueBins.size}`);

  return {
    totalProcessed,
    uniqueBins: globalUniqueBins,
    chunkCount: chunkIndex
  };
}