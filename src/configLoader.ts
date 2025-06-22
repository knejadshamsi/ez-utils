import { existsSync, readFileSync, writeFileSync } from 'fs';
import { resolve } from 'path';
import { load } from 'js-yaml';

export interface Config {
  maxConcurrency: number;
  personsPerChunk: number;
  binSize: number;
  targetDensity: number;
  database: {
    filename: string;
    timeout: number;
  };
  outputPath: string;
  defaultScales: number[];
}

const CONFIG_TEMPLATE = `# Population Scaling Configuration
# Maximum number of concurrent processing pipelines
maxConcurrency: 4

# Chunk processing configuration
personsPerChunk: 1000  # Number of persons per processing chunk

# Spatial binning configuration
binSize: 1000  # Square bin side length in meters
targetDensity: 100  # Desired average population density per bin

# Database configuration
database:
  filename: "population.db"  # SQLite database filename
  timeout: 30000  # Connection timeout in milliseconds

# Output configuration
outputPath: "./output"  # Directory for final scaled population files

# Default scales to generate (percentages 1-10, can be overridden via CLI)
defaultScales: [1, 5, 10]
`;

export function loadConfig(configFilePath?: string): Config {
  const configPath = configFilePath ? resolve(configFilePath) : resolve(process.cwd(), 'config.yaml');
  
  if (!existsSync(configPath)) {
    console.error('Config file not found at:', configPath);
    console.log('Creating default config.yaml file...');
    writeFileSync(configPath, CONFIG_TEMPLATE);
    console.log('Please edit config.yaml with your settings and run the command again.');
    process.exit(1);
  }

  try {
    const configContent = readFileSync(configPath, 'utf-8');
    const parsedConfig = load(configContent) as any;
    
    const config: Config = {
      maxConcurrency: parsedConfig.maxConcurrency || 4,
      personsPerChunk: parsedConfig.personsPerChunk || 1000,
      binSize: parsedConfig.binSize || 1000,
      targetDensity: parsedConfig.targetDensity || 100,
      database: {
        filename: parsedConfig.database?.filename || 'population.db',
        timeout: parsedConfig.database?.timeout || 30000
      },
      outputPath: parsedConfig.outputPath || './output',
      defaultScales: parsedConfig.defaultScales || [1, 5, 10]
    };
    
    return config;
  } catch (error) {
    console.error('Error parsing config.yaml:', error);
    process.exit(1);
  }
}