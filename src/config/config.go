package config

import (
	"fmt"
	"time"
)

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

	// Note: Database configuration is validated when --db flag is used
	// This allows users to run ez-utils without database config if not using --db

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

	// Validate workers config using centralized validator
	if err := c.ValidatePopulationConfig(); err != nil {
		return fmt.Errorf("population config validation failed: %w", err)
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
