#!/usr/bin/env bun
import { existsSync } from 'fs';
import { resolve } from 'path';
import { loadConfig } from './src/configLoader';
import { printUsage } from './src/cli/helpText';
import { populationScalingManager } from './src/scale/population/populationScalingManager';

async function main() {
  const args = process.argv.slice(2);
  
  if (args.length === 0 || args.includes('--help') || args.includes('-h')) {
    printUsage();
    process.exit(0);
  }

  if (args[0] !== 'scale' || args[1] !== 'population') {
    console.error('Invalid command. Use: ez-utils scale population <input-file>');
    printUsage();
    process.exit(1);
  }

  const inputFile = args[2];
  if (!inputFile) {
    console.error('Input file is required');
    printUsage();
    process.exit(1);
  }

  if (!existsSync(inputFile)) {
    console.error('Input file does not exist:', inputFile);
    process.exit(1);
  }

  // Parse config path if provided
  let configPath: string | undefined;
  const configIndex = args.indexOf('--config');
  if (configIndex !== -1 && args[configIndex + 1]) {
    configPath = args[configIndex + 1];
  }
  
  const config = loadConfig(configPath);
  const cleanDatabase = args.includes('--clean');
  
  // Parse custom scales if provided
  let requestedScales = config.defaultScales;
  const scalesIndex = args.indexOf('--scales') !== -1 ? args.indexOf('--scales') : args.indexOf('-s');
  if (scalesIndex !== -1 && args[scalesIndex + 1]) {
    try {
      requestedScales = args[scalesIndex + 1].split(',').map(s => parseInt(s.trim()));
      if (requestedScales.some(s => isNaN(s) || s < 1 || s > 10)) {
        throw new Error('Invalid scale values');
      }
    } catch (error) {
      console.error('Invalid scales format. Use comma-separated percentage values between 1 and 10');
      process.exit(1);
    }
  }

  console.log('Population Scaling Configuration:');
  console.log('- Input file:', inputFile);
  console.log('- Requested scales:', requestedScales);
  console.log('- Max concurrency:', config.maxConcurrency);
  console.log('- Bin size:', config.binSize);
  console.log('- Target density:', config.targetDensity);
  console.log('- Output path:', config.outputPath);
  console.log('- Clean database:', cleanDatabase);
  console.log('');

  try {
    await populationScalingManager({
      inputFilePath: resolve(inputFile),
      requestedScales,
      maxConcurrency: config.maxConcurrency,
      dbConfig: config.database,
      binSize: config.binSize,
      targetDensity: config.targetDensity,
      outputPath: config.outputPath,
      cleanDatabase
    });
    
    console.log('Population scaling completed successfully!');
  } catch (error) {
    console.error('Population scaling failed:', error);
    process.exit(1);
  }
}

if (import.meta.main) {
  main();
}