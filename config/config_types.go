package config

import "time"

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
	return "host=" + d.Host + " port=" + string(d.Port) + " user=" + d.User + " password=" + d.Password + " dbname=" + d.Name + " sslmode=disable"
}

// GetConnectionTimeout returns the timeout as a Duration
func (d DatabaseConfig) GetConnectionTimeout() (time.Duration, error) {
	if d.ConnectionTimeout == "" {
		return 30 * time.Second, nil
	}
	return time.ParseDuration(d.ConnectionTimeout)
}