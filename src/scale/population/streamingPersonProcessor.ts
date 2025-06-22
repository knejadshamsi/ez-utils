import { Database } from 'bun:sqlite';
import { Mutex } from 'async-mutex';
import winston from 'winston';

interface StreamingPersonProcessorParams {
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
    new winston.transports.File({ filename: 'concurrency-debug.log' })
  ]
});

export async function streamingPersonProcessor(params: StreamingPersonProcessorParams): Promise<{
  totalProcessed: number;
  uniqueBins: Set<string>;
  chunkCount: number;
}> {
  const { inputFilePath, personsPerChunk, maxConcurrency, dbConnection, binSize, arrayMutex } = params;
  
  logger.info(`=== STARTING STREAMING PROCESSOR ===`);
  logger.info(`Chunk size: ${personsPerChunk} persons per chunk`);
  logger.info(`Max concurrency: ${maxConcurrency}`);
  
  console.log(`Chunk size: ${personsPerChunk} persons per chunk`);
  console.log(`Max concurrency: ${maxConcurrency}`);
  
  // Use Bun's streaming API for large files
  const file = Bun.file(inputFilePath);
  const fileSize = file.size;
  console.log(`Processing file: ${(fileSize / 1024 / 1024).toFixed(2)} MB`);
  
  const stream = file.stream();
  const reader = stream.getReader();
  const decoder = new TextDecoder();

  let buffer = '';
  let lineBuffer = '';
  let personCount = 0;
  let chunkPersonCount = 0;
  let chunkIndex = 0;
  let totalProcessed = 0;
  let isCapturing = false;
  let activePipelines = 0;
  let completedChunks = 0;
  
  const globalUniqueBins = new Set<string>();
  const pipelinePromises: Promise<ProcessingResult>[] = [];

  // Import the processor
  const { personPipelineProcessor } = await import('./personPipelineProcessor');

  // Process stream chunks
  let bytesRead = 0;
  let lastLogBytes = 0;
  let fileReadStartTime = Date.now();
  let readOperationCount = 0;
  let totalReadTime = 0;
  let totalDecodeTime = 0;
  let totalLineSplitTime = 0;
  let totalLineProcessingTime = 0;
  
  logger.info(`Starting file reading loop`);
  logger.info(`INITIAL STATE: activePipelines=${activePipelines}, maxConcurrency=${maxConcurrency}, chunkIndex=${chunkIndex}, personCount=${personCount}, chunkPersonCount=${chunkPersonCount}`);
  logger.info(`INITIAL BUFFERS: buffer.length=${buffer.length}, lineBuffer.length=${lineBuffer.length}, isCapturing=${isCapturing}`);
  
  while (true) {
    const readStartTime = Date.now();
    logger.debug(`READ OPERATION ${++readOperationCount} STARTING at ${readStartTime}`);
    
    const { done, value } = await reader.read();
    const readEndTime = Date.now();
    const readDuration = readEndTime - readStartTime;
    totalReadTime += readDuration;
    
    logger.debug(`READ OPERATION ${readOperationCount} COMPLETE: duration=${readDuration}ms, done=${done}, valueSize=${value?.byteLength || 0}, cumulativeReadTime=${totalReadTime}ms`);
    logger.debug(`READ STATS: avgReadTime=${(totalReadTime / readOperationCount).toFixed(2)}ms per operation`);
    
    if (done) {
      logger.info(`FILE READING COMPLETE: Total time: ${Date.now() - fileReadStartTime}ms, total operations: ${readOperationCount}, avg per operation: ${(totalReadTime / readOperationCount).toFixed(2)}ms`);
      logger.info(`PERFORMANCE BREAKDOWN: totalReadTime=${totalReadTime}ms, totalDecodeTime=${totalDecodeTime}ms, totalLineSplitTime=${totalLineSplitTime}ms, totalLineProcessingTime=${totalLineProcessingTime}ms`);
      break;
    }
    
    // Track progress
    if (value) {
      bytesRead += value.byteLength;
      logger.debug(`BYTES TRACKING: bytesRead=${bytesRead}, thisChunk=${value.byteLength}, progress=${((bytesRead / fileSize) * 100).toFixed(3)}%`);
      
      // Log progress every 100MB initially to see speed
      if (bytesRead - lastLogBytes >= 100 * 1024 * 1024) {
        const progressPct = ((bytesRead / fileSize) * 100).toFixed(1);
        console.log(`Reading: ${progressPct}% (${(bytesRead / 1024 / 1024).toFixed(0)} MB / ${(fileSize / 1024 / 1024).toFixed(0)} MB)`);
        logger.info(`PROGRESS MILESTONE: ${progressPct}% (${(bytesRead / 1024 / 1024).toFixed(0)} MB / ${(fileSize / 1024 / 1024).toFixed(0)} MB), activePipelines=${activePipelines}/${maxConcurrency}`);
        lastLogBytes = bytesRead;
      }
    }
    
    // Decode chunk and add to line buffer
    const decodeStartTime = Date.now();
    const decodedText = decoder.decode(value, { stream: true });
    const decodeEndTime = Date.now();
    const decodeDuration = decodeEndTime - decodeStartTime;
    totalDecodeTime += decodeDuration;
    
    logger.debug(`DECODE OPERATION: duration=${decodeDuration}ms, decodedLength=${decodedText.length}, cumulativeDecodeTime=${totalDecodeTime}ms`);
    
    const preBufferLength = lineBuffer.length;
    lineBuffer += decodedText;
    const postBufferLength = lineBuffer.length;
    
    logger.debug(`BUFFER UPDATE: preLength=${preBufferLength}, addedLength=${decodedText.length}, postLength=${postBufferLength}, growth=${postBufferLength - preBufferLength}`);
    
    // Process complete lines - use lastIndexOf for efficiency with large chunks
    const lineSplitStartTime = Date.now();
    const lastNewline = lineBuffer.lastIndexOf('\n');
    logger.debug(`LINE SPLIT CHECK: lastNewline=${lastNewline}, lineBuffer.length=${lineBuffer.length}`);
    
    if (lastNewline === -1) {
      logger.debug(`NO COMPLETE LINES: continuing to next read operation, lineBuffer.length=${lineBuffer.length}`);
      continue; // No complete lines yet
    }
    
    const completeLines = lineBuffer.substring(0, lastNewline);
    lineBuffer = lineBuffer.substring(lastNewline + 1); // Keep incomplete line
    const lineSplitEndTime = Date.now();
    const lineSplitDuration = lineSplitEndTime - lineSplitStartTime;
    totalLineSplitTime += lineSplitDuration;
    
    logger.debug(`LINE SPLIT COMPLETE: duration=${lineSplitDuration}ms, completeLines.length=${completeLines.length}, remainingBuffer.length=${lineBuffer.length}`);
    
    // Process all lines at once
    const lines = completeLines.split('\n');
    logger.debug(`LINES ARRAY: ${lines.length} lines to process, first10chars="${lines[0]?.substring(0, 10) || 'EMPTY'}", last10chars="${lines[lines.length - 1]?.substring(0, 10) || 'EMPTY'}"`);
    
    const lineProcessingStartTime = Date.now();
    
    for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
    const line = lines[lineIndex];
    const lineStartTime = Date.now();
    logger.debug(`LINE ${lineIndex + 1}/${lines.length}: length=${line.length} chars, isCapturing=${isCapturing}, chunkPersonCount=${chunkPersonCount}, personCount=${personCount}`);
    logger.debug(`LINE CONTENT PREVIEW: "${line.substring(0, 50)}${line.length > 50 ? '...' : ''}"`);
    
    // Check for person start tag
    const hasPersonStart = line.includes('<person ');
    logger.debug(`PERSON START CHECK: hasPersonStart=${hasPersonStart}, line="${line}"`);
    
    if (hasPersonStart) {
      const prevIsCapturing = isCapturing;
      const prevPersonCount = personCount;
      const prevChunkPersonCount = chunkPersonCount;
      
      isCapturing = true;
      personCount++;
      chunkPersonCount++;
      
      logger.info(`PERSON START FOUND: line=${lineIndex + 1}, prevIsCapturing=${prevIsCapturing}, newIsCapturing=${isCapturing}, prevPersonCount=${prevPersonCount}, newPersonCount=${personCount}, prevChunkPersonCount=${prevChunkPersonCount}, newChunkPersonCount=${chunkPersonCount}`);
    }

    // Add to buffer if we're capturing a person or have persons in current chunk
    const shouldAddToBuffer = isCapturing || chunkPersonCount > 0;
    logger.debug(`BUFFER ADD CHECK: shouldAddToBuffer=${shouldAddToBuffer}, isCapturing=${isCapturing}, chunkPersonCount=${chunkPersonCount}`);
    
    if (shouldAddToBuffer) {
      const prevBufferLength = buffer.length;
      buffer += line + '\n';
      const newBufferLength = buffer.length;
      logger.debug(`BUFFER ADDED: prevLength=${prevBufferLength}, addedLength=${line.length + 1}, newLength=${newBufferLength}`);
    }

    // Check for person end tag
    const hasPersonEnd = line.includes('</person>');
    logger.debug(`PERSON END CHECK: hasPersonEnd=${hasPersonEnd}, line="${line}"`);
    
    if (isCapturing && hasPersonEnd) {
      const prevIsCapturing = isCapturing;
      isCapturing = false;
      logger.info(`PERSON END FOUND: line=${lineIndex + 1}, prevIsCapturing=${prevIsCapturing}, newIsCapturing=${isCapturing}, currentPersonCount=${personCount}, currentChunkPersonCount=${chunkPersonCount}`);
    }

    // Process chunk when we have enough persons and reached end of a person
    const isChunkReady = chunkPersonCount >= personsPerChunk && !isCapturing;
    logger.debug(`CHUNK READY CHECK: chunkPersonCount=${chunkPersonCount} >= personsPerChunk=${personsPerChunk} = ${chunkPersonCount >= personsPerChunk}, isCapturing=${isCapturing}, isChunkReady=${isChunkReady}`);
    
    if (isChunkReady) {
      logger.info(`CHUNK READY CONFIRMED: chunkPersonCount=${chunkPersonCount} >= personsPerChunk=${personsPerChunk}, isCapturing=${isCapturing}`);
      logger.info(`CONCURRENCY STATE: activePipelines=${activePipelines}, maxConcurrency=${maxConcurrency}, canFire=${activePipelines < maxConcurrency}`);
      logger.info(`BUFFER STATE: buffer.length=${buffer.length}, chunkIndex=${chunkIndex}, personCount=${personCount}`);
      
      // Check if we can start a new chunk
      if (activePipelines < maxConcurrency) {
        // Process this chunk directly
        const chunkBuffer = buffer;
        const currentChunkIndex = chunkIndex;
        const preActivePipelines = activePipelines;
        const preChunkIndex = chunkIndex;
        
        activePipelines++;
        chunkIndex++;
        
        logger.info(`CHUNK FIRING INITIATED: currentChunkIndex=${currentChunkIndex + 1}, preActivePipelines=${preActivePipelines}, newActivePipelines=${activePipelines}, preChunkIndex=${preChunkIndex}, newChunkIndex=${chunkIndex}`);
        logger.info(`CHUNK FIRING DETAILS: bufferSize=${chunkBuffer.length}, personsInChunk=${chunkPersonCount}, totalPersonsSoFar=${personCount}`);
        
        console.log(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} FIRE (Active: ${activePipelines}/${maxConcurrency})`);
        
        const chunkFireTime = Date.now();
        logger.info(`CHUNK FIRE TIMESTAMP: ${chunkFireTime}, chunk=${currentChunkIndex + 1}`);
        
        // Call processor directly without await - fire and forget
        personPipelineProcessor(
          chunkBuffer,
          dbConnection,
          binSize,
          arrayMutex
        ).then(result => {
          const chunkCompleteTime = Date.now();
          const chunkDuration = chunkCompleteTime - chunkFireTime;
          logger.info(`CHUNK SUCCESS: chunk=${currentChunkIndex + 1}, duration=${chunkDuration}ms, processedCount=${result.processedCount}, localBins=${result.localBins.size}, fireTime=${chunkFireTime}, completeTime=${chunkCompleteTime}`);
          
          // Merge local bins into global set
          const prevGlobalBinsSize = globalUniqueBins.size;
          result.localBins.forEach(bin => globalUniqueBins.add(bin));
          const newGlobalBinsSize = globalUniqueBins.size;
          
          const prevTotalProcessed = totalProcessed;
          totalProcessed += result.processedCount;
          
          logger.info(`GLOBAL STATE UPDATE: prevGlobalBins=${prevGlobalBinsSize}, newGlobalBins=${newGlobalBinsSize}, addedBins=${newGlobalBinsSize - prevGlobalBinsSize}, prevTotalProcessed=${prevTotalProcessed}, newTotalProcessed=${totalProcessed}`);
          
          console.log(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} DONE: ${result.processedCount} persons`);
        }).catch(error => {
          const chunkErrorTime = Date.now();
          const chunkDuration = chunkErrorTime - chunkFireTime;
          logger.error(`CHUNK ERROR: chunk=${currentChunkIndex + 1}, duration=${chunkDuration}ms, error=${error.message}, fireTime=${chunkFireTime}, errorTime=${chunkErrorTime}, stack=${error.stack}`);
          console.error(`[${new Date().toISOString()}] Chunk ${currentChunkIndex + 1} ERROR:`, error);
        }).finally(() => {
          const chunkFinalTime = Date.now();
          const chunkTotalDuration = chunkFinalTime - chunkFireTime;
          const prevActivePipelines = activePipelines;
          activePipelines--;
          const newActivePipelines = activePipelines;
          
          logger.info(`SLOT FREED: chunk=${currentChunkIndex + 1}, totalDuration=${chunkTotalDuration}ms, prevActivePipelines=${prevActivePipelines}, newActivePipelines=${newActivePipelines}, fireTime=${chunkFireTime}, finalTime=${chunkFinalTime}`);
          console.log(`[${new Date().toISOString()}] SLOT FREED. Active: ${activePipelines}/${maxConcurrency}`);
        });

        // Reset for next chunk
        const prevBuffer = buffer;
        const prevChunkPersonCount = chunkPersonCount;
        buffer = '';
        chunkPersonCount = 0;
        logger.info(`CHUNK RESET: prevBufferLength=${prevBuffer.length}, newBufferLength=${buffer.length}, prevChunkPersonCount=${prevChunkPersonCount}, newChunkPersonCount=${chunkPersonCount}`);
        logger.info(`POST-RESET STATE: activePipelines=${activePipelines}/${maxConcurrency}, chunkIndex=${chunkIndex}, personCount=${personCount}, totalProcessed=${totalProcessed}`);
      } else {
        // At max capacity - skip this iteration, keep building buffer
        logger.warn(`AT MAX CAPACITY: activePipelines=${activePipelines}/${maxConcurrency} - continuing to read, buffer building, currentBufferSize=${buffer.length}, chunkPersonCount=${chunkPersonCount}`);
        logger.warn(`CAPACITY DETAILS: wouldBeChunk=${chunkIndex + 1}, personsReady=${chunkPersonCount}, bufferLines=${buffer.split('\n').length}`);
        console.log(`At max capacity (${activePipelines}/${maxConcurrency}) - continuing to read...`);
      }
    }
    
    const lineEndTime = Date.now();
    const lineDuration = lineEndTime - lineStartTime;
    logger.debug(`LINE ${lineIndex + 1} COMPLETE: duration=${lineDuration}ms, finalState: isCapturing=${isCapturing}, chunkPersonCount=${chunkPersonCount}, activePipelines=${activePipelines}`);
  }
  
  const lineProcessingEndTime = Date.now();
  const lineProcessingDuration = lineProcessingEndTime - lineProcessingStartTime;
  totalLineProcessingTime += lineProcessingDuration;
  
  logger.info(`LINES BATCH COMPLETE: processedLines=${lines.length}, duration=${lineProcessingDuration}ms, cumulativeLineProcessingTime=${totalLineProcessingTime}ms, avgPerLine=${(lineProcessingDuration / lines.length).toFixed(2)}ms`);
  logger.info(`BATCH END STATE: activePipelines=${activePipelines}/${maxConcurrency}, chunkPersonCount=${chunkPersonCount}, personCount=${personCount}, chunkIndex=${chunkIndex}, bufferLength=${buffer.length}`);
  }
  
  // Process any remaining line in the buffer
  if (lineBuffer) {
    const line = lineBuffer;
    // Check for person start tag
    if (line.includes('<person ')) {
      isCapturing = true;
      personCount++;
      chunkPersonCount++;
    }

    // Add to buffer if we're capturing a person or have persons in current chunk
    if (isCapturing || chunkPersonCount > 0) {
      buffer += line + '\n';
    }

    // Check for person end tag
    if (isCapturing && line.includes('</person>')) {
      isCapturing = false;
    }
  }

  // Process final chunk if there's remaining data
  if (chunkPersonCount > 0 && buffer.length > 0) {
    const currentChunkIndex = chunkIndex;
    activePipelines++;
    chunkIndex++;
    
    console.log(`Final chunk ${currentChunkIndex + 1} started (Active: ${activePipelines}/${maxConcurrency})`);
    
    // Process final chunk directly
    personPipelineProcessor(
      buffer,
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

  console.log(`File reading complete: ${personCount} persons found, ${chunkIndex} chunks created`);
  
  // Wait for all processing to complete
  logger.info(`WAITING FOR COMPLETION: ${activePipelines} active pipelines remaining`);
  while (activePipelines > 0) {
    console.log(`Waiting for ${activePipelines} active pipelines to complete...`);
    logger.info(`WAITING LOOP: ${activePipelines} active pipelines still running`);
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  logger.info(`ALL PIPELINES COMPLETE: activePipelines=${activePipelines}`);

  // No need to close anything with Bun.file

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