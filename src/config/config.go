package config

import (
	"fmt"
	"time"
)

// Config represents the main configuration structure
type Config struct {
	Language   string             `yaml:"language"`
	Database   DatabaseConfig     `yaml:"database"`
	Population PopulationConfig   `yaml:"population"`
	Paths      PathsConfig        `yaml:"paths"`
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	Name               string        `yaml:"name"`
	MaxConnections     int           `yaml:"max_connections"`
	ConnectionTimeout  string        `yaml:"connection_timeout"`
}

// PopulationConfig holds population module settings
type PopulationConfig struct {
	ChunkSize  int    `yaml:"chunk_size"`
	OutputDir  string `yaml:"output_dir"`
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

	return nil
}