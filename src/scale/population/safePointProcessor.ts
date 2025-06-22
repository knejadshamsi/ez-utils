import { createReadStream } from 'fs';
import { createInterface } from 'readline';

export interface SafePointResult {
  boundaryCount: number;
  fileSize: number;
  chunkSizes: number[];
}

export async function safePointProcessor(
  inputFilePath: string,
  safePoints: Array<{start: number, end: number}>,
  arrayMutex: any,
  personsPerChunk: number
): Promise<SafePointResult> {
  const fileStream = createReadStream(inputFilePath);
  const rl = createInterface({
    input: fileStream,
    crlfDelay: Infinity
  });

  let lineNumber = 0;
  let personCount = 0;
  let previousBoundary = 1;
  let totalLines = 0;

  for await (const line of rl) {
    lineNumber++;
    totalLines = lineNumber;

    if (line.includes("<person ")) {
      personCount++;

      // At 1001st, 2001st person etc, save boundary before this person
      if (personCount > 1 && (personCount - 1) % personsPerChunk === 0) {
        await arrayMutex.runExclusive(() => {
          safePoints.push({
            start: previousBoundary,
            end: lineNumber - 1
          });
        });
        previousBoundary = lineNumber;
      }
    }
  }

  // Add final chunk if there are remaining persons
  if (personCount > 0 && previousBoundary < totalLines) {
    await arrayMutex.runExclusive(() => {
      safePoints.push({
        start: previousBoundary,
        end: totalLines
      });
    });
  }

  rl.close();
  fileStream.close();

  // Calculate chunk sizes  
  const chunkSizes = safePoints.map(sp => sp.end - sp.start + 1);

  return {
    boundaryCount: safePoints.length,
    fileSize: totalLines,
    chunkSizes
  };
}