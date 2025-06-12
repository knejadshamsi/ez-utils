package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the platform-specific configuration directory
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%/ez-utils
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "ez-utils")

	case "darwin":
		// macOS: ~/Library/Application Support/ez-utils
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "ez-utils")

	default:
		// Linux and other Unix-like systems: ~/.config/ez-utils
		// First try XDG_CONFIG_HOME
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, "ez-utils")
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(homeDir, ".config", "ez-utils")
		}
	}

	return configDir, nil
}

// GetLocalConfigPath returns the path to the local config file
func GetLocalConfigPath() string {
	return "config.yaml"
}

// GetGlobalConfigPath returns the full path to the global config file
func GetGlobalConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// GetConfigPath returns the full path to the config file (deprecated - use priority system)
func GetConfigPath() (string, error) {
	return GetGlobalConfigPath()
}

// GetConfigPaths returns both local and global config paths with their display locations
func GetConfigPaths() (localPath string, globalPath string, globalDisplay string, err error) {
	localPath = GetLocalConfigPath()

	globalPath, err = GetGlobalConfigPath()
	if err != nil {
		return "", "", "", err
	}

	// Get user-friendly display path for global config
	switch runtime.GOOS {
	case "windows":
		globalDisplay = "%APPDATA%/ez-utils/config.yaml"
	case "darwin":
		globalDisplay = "~/Library/Application Support/ez-utils/config.yaml"
	default:
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			globalDisplay = "$XDG_CONFIG_HOME/ez-utils/config.yaml"
		} else {
			globalDisplay = "~/.config/ez-utils/config.yaml"
		}
	}

	return localPath, globalPath, globalDisplay, nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0755)
}

// Embedded template configuration
func GetEmbeddedTemplate() string {
	return `# ================================================================
# EZ-UTILS CONFIGURATION FILE
# ================================================================
# This file was automatically generated on first run.
# Please review and update the settings below before using ez-utils.
# ================================================================

# ----------------------------------------------------------------
# GLOBAL SETTINGS
# ----------------------------------------------------------------
# Language for user interface and messages
# Supported values: en (English), fr (French)
language: en

# ----------------------------------------------------------------
# DATABASE CONFIGURATION
# ----------------------------------------------------------------
# PostgreSQL connection settings for storing population data
# These settings are only used when the --db flag is provided
database:
  # Database server hostname or IP address
  host: localhost
  
  # PostgreSQL port (default: 5432)
  port: 5432
  
  # Database username
  user: postgres
  
  # Database password
  # NOTE: This is stored in plain text. Ensure appropriate
  # file permissions (600) to protect sensitive information
  password: postgres
  
  # Name of the database to connect to
  # This database must already exist
  name: simulation
  
  # Maximum number of open connections to the database
  # Increase for better performance with large datasets
  max_connections: 25
  
  # Connection timeout duration
  # Format: 30s, 1m, 5m, etc.
  connection_timeout: 30s

# ----------------------------------------------------------------
# POPULATION MODULE SETTINGS
# ----------------------------------------------------------------
# Configuration for the population scaling operations
population:
  # Number of persons to process per chunk for Phase Two parallel processing
  # Lower values create more chunks but allow better parallelization
  # Recommended: 1000-5000 depending on available RAM and number of reducers
  chunk_size: 1000
  
  # Directory for output files
  # Can be absolute or relative to current directory
  output_dir: "output"
  
  # Scales to generate (1-10% only)
  # If not specified, defaults to all scales: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
  # scales: [2, 5, 8, 10]
  
# ----------------------------------------------------------------
# WORKER CONFIGURATION
# ----------------------------------------------------------------
# Configuration for parallel processing workers
workers:
  # Phase One Workers
  readers: 4           # Parallel file reading workers
  extractors: 4        # Parallel XML parsing workers
  hashmap: 3           # Coordinate processing workers
  
  # Phase Two Workers  
  reducers: 8          # Parallel file processing workers
  file_writers: 4      # Chunk writing workers
  database: 4          # Database insertion workers (if --db flag used)
  
  # Safe point discovery settings
  safe_point_interval: 10000   # Lines between safe points for parallel processing
  
  # Grid expansion settings
  grid_expansion_batch: 1000   # Coordinates per HashMap worker before expansion
  
  # Queue sizes for buffering between workers
  queue_sizes:
    extractor: 100     # Buffer size for each Extractor worker
    hashmap: 500       # Buffer size for each HashMap worker
    reducer: 200       # Buffer size for each Reducer worker
    file_writer: 150   # Buffer size for each File Writer worker
    database: 300      # Buffer size for each Database worker

# ----------------------------------------------------------------
# PATH OVERRIDES (Optional)
# ----------------------------------------------------------------
# Custom paths for various operations
# Leave empty to use defaults
paths:
  # Temporary directory for intermediate files and chunk files
  # Default: system temp directory
  temp_dir: ""
  
  # Directory for chunk files during processing
  # Default: temp_dir/chunks
  chunks_dir: ""

# ================================================================
# For more information, visit:
# https://github.com/knejadshamsi/ez-utils
# ================================================================`
}
