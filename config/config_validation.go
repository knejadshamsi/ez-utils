package config

import "fmt"

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
		c.Workers.SafePointInterval = 10000
	}
	if c.Workers.GridExpansionBatch < 1 {
		c.Workers.GridExpansionBatch = 1000
	}

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