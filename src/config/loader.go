package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// global config instance
var currentConfig *Config

// CheckAndCreateConfig checks if config exists anywhere, creates one if not
func CheckAndCreateConfig() error {
	localPath, globalPath, _, err := GetConfigPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %v", err)
	}

	// Check local config first
	if _, err := os.Stat(localPath); err == nil {
		return nil // Local config exists, we're good
	}

	// Check global config
	if _, err := os.Stat(globalPath); err == nil {
		return nil // Global config exists, we're good
	}

	// No config found anywhere - create one interactively
	return createConfigInteractively()
}

// CreateNewConfig forces creation of a new config file (for --new-config flag)
func CreateNewConfig() error {
	return createConfigInteractively()
}

// createConfigInteractively prompts user for config location choice
func createConfigInteractively() error {
	localPath, globalPath, globalDisplay, err := GetConfigPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %v", err)
	}

	fmt.Println("\nNo configuration found. Where would you like to create it?")
	fmt.Printf("[1] Local config (%s) - Project-specific settings\n", localPath)
	fmt.Printf("[2] Global config (%s) - Use across all projects\n", globalDisplay)
	fmt.Print("\nEnter choice [1-2]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	choice := strings.TrimSpace(input)
	var targetPath, displayPath string
	var isLocal bool

	switch choice {
	case "1":
		targetPath = localPath
		displayPath = localPath
		isLocal = true
	case "2":
		targetPath = globalPath
		displayPath = globalDisplay
		isLocal = false
	default:
		return fmt.Errorf("invalid choice. Please enter 1 or 2")
	}

	// Create the config file
	if err := createConfigAt(targetPath, isLocal); err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	// Print helpful message and exit
	fmt.Printf("\nConfiguration file created at: %s\n", displayPath)
	fmt.Println("Please review and update the configuration settings, then run the command again.")
	fmt.Println("\nYou can edit the file with your preferred text editor:")
	fmt.Printf("  nano %s\n", targetPath)
	fmt.Printf("  vim %s\n", targetPath)
	fmt.Printf("  code %s\n", targetPath)
	fmt.Println("\nThe configuration file contains detailed comments explaining each setting.")

	os.Exit(0) // Exit with success since this is expected behavior
	return nil // This line will never be reached, but Go requires it
}

// createConfigAt creates config file at the specified path
func createConfigAt(configPath string, isLocal bool) error {
	// Ensure directory exists (only for global configs)
	if !isLocal {
		if err := EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	// Get embedded template configuration
	templateData := []byte(GetEmbeddedTemplate())

	// Write to target location
	if err := os.WriteFile(configPath, templateData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// LoadConfig loads and validates the configuration using priority system
func LoadConfig() (*Config, string, error) {
	if currentConfig != nil {
		return currentConfig, "", nil // Return empty path since already loaded
	}

	localPath, globalPath, _, err := GetConfigPaths()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get config paths: %v", err)
	}

	// Priority: local first, then global
	var configPath string
	var configSource string

	if _, err := os.Stat(localPath); err == nil {
		configPath = localPath
		configSource = "local"
	} else if _, err := os.Stat(globalPath); err == nil {
		configPath = globalPath
		configSource = "global"
	} else {
		return nil, "", fmt.Errorf("no configuration file found. Run the command again to create one")
	}

	config, err := loadConfigFromPath(configPath)
	if err != nil {
		return nil, "", err
	}

	currentConfig = config
	return currentConfig, configSource, nil
}

// loadConfigFromPath loads config from a specific file path
func loadConfigFromPath(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Apply defaults for any missing values
	applyDefaults(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &config, nil
}

// applyDefaults sets default values for any missing configuration
func applyDefaults(config *Config) {
	if config.Language == "" {
		config.Language = "en"
	}

	// Database configuration must be explicitly set by user - no defaults
	// Only set defaults for optional pool settings
	if config.Database.MaxConnections == 0 {
		config.Database.MaxConnections = 25
	}
	if config.Database.ConnectionTimeout == "" {
		config.Database.ConnectionTimeout = "30s"
	}
	// Set default port only if database is configured
	if config.Database.Host != "" && config.Database.Port == 0 {
		config.Database.Port = 5432
	}

	if config.Population.ChunkSize == 0 {
		config.Population.ChunkSize = 2000
	}
	if config.Population.OutputDir == "" {
		config.Population.OutputDir = "output"
	}
}

// GetConfig returns the current loaded configuration
func GetConfig() *Config {
	return currentConfig
}

// LoadConfigWithPath loads config from a specific path (for --config flag)
func LoadConfigWithPath(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Apply defaults for any missing values
	applyDefaults(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	currentConfig = &config
	return currentConfig, nil
}
