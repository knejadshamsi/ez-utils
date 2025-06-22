import { readdirSync, readFileSync, writeFileSync, unlinkSync, existsSync } from 'fs';
import { join, basename } from 'path';

interface MergeStats {
  chunkCount: number;
  personCount: number;
  fileSize: number;
}

interface FileMergerResult {
  finalFiles: string[];
  mergeStats: Record<number, MergeStats>;
  cleanupStatus: {
    deletedFiles: number;
    errors: string[];
  };
}

export async function scaleFileMerger(
  requestedScales: number[],
  outputPath: string,
  originalFileName: string
): Promise<FileMergerResult> {
  const finalFiles: string[] = [];
  const mergeStats: Record<number, MergeStats> = {};
  const cleanupStatus = {
    deletedFiles: 0,
    errors: [] as string[]
  };

  // Extract base name without extension from original file
  const baseFileName = basename(originalFileName, '.xml');

  for (const scaleValue of requestedScales) {
    console.log(`Merging chunk files for scale ${scaleValue}%...`);
    
    // Identify all chunk files for this scale
    const scaleDir = join(outputPath, `scale_${scaleValue}`);
    if (!existsSync(scaleDir)) {
      console.log(`No chunk files found for scale ${scaleValue}`);
      continue;
    }

    // Pattern for chunk files: scale-chunk-XX-ROWxCOLUMN.xml (row and column can be negative)
    const chunkPattern = new RegExp(`^scale-chunk-${scaleValue.toString().padStart(2, '0')}--?\\d+x-?\\d+\\.xml$`);
    const chunkFiles = readdirSync(scaleDir)
      .filter(file => chunkPattern.test(file))
      .map(file => join(scaleDir, file))
      .sort(); // Sort to ensure consistent order

    if (chunkFiles.length === 0) {
      console.log(`No matching chunk files found for scale ${scaleValue}`);
      continue;
    }

    // Initialize merged content with XML declaration and population start tag
    let mergedContent = '<?xml version="1.0" encoding="UTF-8"?>\n';
    mergedContent += '<population name="population">\n';
    
    let personCount = 0;

    // Read and merge all chunk files
    for (const chunkFile of chunkFiles) {
      try {
        const chunkContent = readFileSync(chunkFile, 'utf-8');
        
        // Count persons in chunk (simple count of <person occurrences)
        const personMatches = chunkContent.match(/<person/g);
        if (personMatches) {
          personCount += personMatches.length;
        }
        
        // Add chunk content (already contains just person elements without wrapper)
        mergedContent += chunkContent;
        if (!chunkContent.endsWith('\n')) {
          mergedContent += '\n';
        }
      } catch (error) {
        console.error(`Error reading chunk file ${chunkFile}:`, error);
        cleanupStatus.errors.push(`Failed to read chunk: ${chunkFile}`);
      }
    }

    // Close population tag
    mergedContent += '</population>\n';

    // Generate final file name: original-name_pct_XX.xml
    const finalFileName = `${baseFileName}_pct_${scaleValue.toString().padStart(2, '0')}.xml`;
    const finalFilePath = join(outputPath, finalFileName);

    // Write final merged file
    try {
      writeFileSync(finalFilePath, mergedContent, 'utf-8');
      finalFiles.push(finalFilePath);
      
      // Store merge statistics
      mergeStats[scaleValue] = {
        chunkCount: chunkFiles.length,
        personCount,
        fileSize: Buffer.byteLength(mergedContent)
      };
      
      console.log(`Created ${finalFileName}: ${personCount} persons from ${chunkFiles.length} chunks`);
    } catch (error) {
      console.error(`Error writing final file ${finalFileName}:`, error);
      cleanupStatus.errors.push(`Failed to write final file: ${finalFileName}`);
      continue;
    }

    // Clean up chunk files after successful merge
    for (const chunkFile of chunkFiles) {
      try {
        unlinkSync(chunkFile);
        cleanupStatus.deletedFiles++;
      } catch (error) {
        console.error(`Error deleting chunk file ${chunkFile}:`, error);
        cleanupStatus.errors.push(`Failed to delete chunk: ${chunkFile}`);
      }
    }

    // Try to remove the scale directory if empty
    try {
      const remainingFiles = readdirSync(scaleDir);
      if (remainingFiles.length === 0) {
        const { rmdirSync } = await import('fs');
        rmdirSync(scaleDir);
      }
    } catch (error) {
      // Directory removal is optional, don't fail the process
      console.log(`Could not remove scale directory ${scaleDir}:`, error);
    }
  }

  console.log(`File merger complete: ${finalFiles.length} final files created`);
  console.log(`Cleanup: ${cleanupStatus.deletedFiles} chunk files deleted`);
  if (cleanupStatus.errors.length > 0) {
    console.log(`Cleanup errors: ${cleanupStatus.errors.length}`);
  }

  return {
    finalFiles,
    mergeStats,
    cleanupStatus
  };
}