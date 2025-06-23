package main

import (
	"bufio"
	_ "embed" // Required for go:embed directive
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed config/config.template.yaml
var templateContent []byte

// Config represents the main configuration structure
type Config struct {
	Language   string           `yaml:"language"`
	Database   DatabaseConfig   `yaml:"database"`
	Population PopulationConfig `yaml:"population"`
	Workers    WorkersConfig    `yaml:"workers"`
	Paths      PathsConfig      `yaml:"paths"`
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	User              string `yaml:"user"`
	Password          string `yaml:"password"`
	Name              string `yaml:"name"`
	MaxConnections    int    `yaml:"max_connections"`
	ConnectionTimeout string `yaml:"connection_timeout"`
}

// PopulationConfig holds population module settings
type PopulationConfig struct {
	ChunkSize int    `yaml:"chunk_size"`
	OutputDir string `yaml:"output_dir"`
}

// WorkersConfig holds worker configuration settings
type WorkersConfig struct {
	// Phase One Workers
	Readers    int `yaml:"readers"`
	Extractors int `yaml:"extractors"`
	Hashmap    int `yaml:"hashmap"`

	// Phase Two Workers
	Reducers    int `yaml:"reducers"`
	FileWriters int `yaml:"file_writers"`
	Database    int `yaml:"database"`

	// Safe point discovery settings
	SafePointInterval int64 `yaml:"safe_point_interval"`

	// Grid expansion settings
	GridExpansionBatch int `yaml:"grid_expansion_batch"`

	// Queue sizes
	QueueSizes QueueSizesConfig `yaml:"queue_sizes"`
}

// QueueSizesConfig holds queue buffer sizes
type QueueSizesConfig struct {
	Extractor  int `yaml:"extractor"`
	Hashmap    int `yaml:"hashmap"`
	Reducer    int `yaml:"reducer"`
	FileWriter int `yaml:"file_writer"`
	Database   int `yaml:"database"`
}

// PathsConfig holds optional path overrides
type PathsConfig struct {
	TempDir string `yaml:"temp_dir"`
}

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
	return nil
}

// createConfigAt creates config file at the specified path
func createConfigAt(configPath string, isLocal bool) error {
	if !isLocal {
		if err := EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	if err := os.WriteFile(configPath, templateContent, 0644); err != nil {
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

// GetConnectionString returns a PostgreSQL connection string
func (d DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name,
	)
}

// GetConnectionTimeout returns the timeout as a Duration
func (d DatabaseConfig) GetConnectionTimeout() (time.Duration, error) {
	if d.ConnectionTimeout == "" {
		return 30 * time.Second, nil
	}
	return time.ParseDuration(d.ConnectionTimeout)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate language
	if c.Language != "en" && c.Language != "fr" {
		return fmt.Errorf("unsupported language: %s (must be 'en' or 'fr')", c.Language)
	}

	// Validate database port if specified
	if c.Database.Port != 0 && (c.Database.Port <= 0 || c.Database.Port > 65535) {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	if c.Database.MaxConnections < 1 {
		return fmt.Errorf("max_connections must be at least 1")
	}

	// Validate connection timeout
	if _, err := c.Database.GetConnectionTimeout(); err != nil {
		return fmt.Errorf("invalid connection_timeout format: %v", err)
	}

	// Validate population config
	if c.Population.ChunkSize < 1 {
		return fmt.Errorf("chunk_size must be at least 1")
	}

	if c.Population.OutputDir == "" {
		return fmt.Errorf("output_dir cannot be empty")
	}

	if c.Workers.SafePointInterval <= 0 {
		c.Workers.SafePointInterval = 10000 // Default to 10,000 lines
	}
	if c.Workers.GridExpansionBatch < 1 {
		c.Workers.GridExpansionBatch = 1000 // Default batch size
	}

	// Validate queue sizes
	if c.Workers.QueueSizes.Extractor < 1 {
		return fmt.Errorf("extractor queue size must be at least 1")
	}
	if c.Workers.QueueSizes.Hashmap < 1 {
		return fmt.Errorf("hashmap queue size must be at least 1")
	}
	if c.Workers.QueueSizes.Reducer < 1 {
		return fmt.Errorf("reducer queue size must be at least 1")
	}
	if c.Workers.QueueSizes.FileWriter < 1 {
		return fmt.Errorf("file_writer queue size must be at least 1")
	}
	if c.Workers.QueueSizes.Database < 1 {
		return fmt.Errorf("database queue size must be at least 1")
	}

	return nil
}