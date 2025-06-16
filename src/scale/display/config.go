package display

import (
	"fmt"
	"time"
)

// Config represents the complete display configuration for an instance
type Config struct {
	Process  *ProcessConfig `json:"process"`
	Theme    *ThemeConfig   `json:"theme"`
	Language string         `json:"language"`
	Async    *AsyncConfig   `json:"async"`
}

// ProcessConfig defines the process-specific display settings
type ProcessConfig struct {
	Title       string       `json:"title"`        // Custom process title (e.g., "GTFS Processing")
	Description string       `json:"description"`  // Process description
	Steps       []StepConfig `json:"steps"`        // Custom step configurations
	MaxSteps    int          `json:"max_steps"`    // Total number of steps
	Flags       map[string]bool `json:"flags"`     // Process flags (clean, db, etc.)
}

// StepConfig defines configuration for individual steps
type StepConfig struct {
	Title       string   `json:"title"`        // Custom step title
	Description string   `json:"description"`  // Step description
	LiveUpdates []string `json:"live_updates"` // List of live update keys for this step
}

// ThemeConfig defines the visual theme for the display
type ThemeConfig struct {
	Name      string            `json:"name"`       // Predefined theme name
	Colors    map[string]string `json:"colors"`     // Custom color overrides
	Profile   string            `json:"profile"`    // "dark", "light", "high-contrast"
}

// AsyncConfig defines settings for async message processing
type AsyncConfig struct {
	QueueSize      int           `json:"queue_size"`       // Message queue buffer size
	UpdateInterval time.Duration `json:"update_interval"`  // UI update frequency
	MaxBatchSize   int           `json:"max_batch_size"`   // Max messages per batch
}

// LiveUpdateData represents a live update for a step
type LiveUpdateData struct {
	Key   string      `json:"key"`   // Update key (e.g., "file_count", "progress")
	Value interface{} `json:"value"` // Update value
	Type  string      `json:"type"`  // Value type for formatting
}

// StepMessage represents a custom message for a step
type StepMessage struct {
	Template string                 `json:"template"` // Message template with placeholders
	Data     map[string]interface{} `json:"data"`     // Data for template substitution
}

// DefaultConfig creates a default configuration
func DefaultConfig() *Config {
	return &Config{
		Process: &ProcessConfig{
			Title:       "Process Execution",
			Description: "Executing process steps",
			Steps:       make([]StepConfig, 0),
			MaxSteps:    10,
			Flags:       make(map[string]bool),
		},
		Theme: &ThemeConfig{
			Name:    "default",
			Profile: "dark",
			Colors:  make(map[string]string),
		},
		Language: "en",
		Async: &AsyncConfig{
			QueueSize:      1000,
			UpdateInterval: 100 * time.Millisecond,
			MaxBatchSize:   50,
		},
	}
}

// NewConfig creates a new configuration with validation
func NewConfig(processTitle, description, language string, maxSteps int) (*Config, error) {
	if processTitle == "" {
		return nil, fmt.Errorf("process title cannot be empty")
	}
	if maxSteps <= 0 {
		return nil, fmt.Errorf("max steps must be positive")
	}
	if language != "en" && language != "fr" {
		return nil, fmt.Errorf("language must be 'en' or 'fr', got '%s'", language)
	}

	config := DefaultConfig()
	config.Process.Title = processTitle
	config.Process.Description = description
	config.Process.MaxSteps = maxSteps
	config.Language = language

	return config, nil
}

// AddStep adds a step configuration
func (c *Config) AddStep(title, description string, liveUpdates ...string) *Config {
	step := StepConfig{
		Title:       title,
		Description: description,
		LiveUpdates: liveUpdates,
	}
	c.Process.Steps = append(c.Process.Steps, step)
	return c
}

// SetTheme sets a predefined theme
func (c *Config) SetTheme(name string) *Config {
	c.Theme.Name = name
	return c
}

// SetCustomColor sets a custom color for a theme element
func (c *Config) SetCustomColor(element, color string) *Config {
	if c.Theme.Colors == nil {
		c.Theme.Colors = make(map[string]string)
	}
	c.Theme.Colors[element] = color
	return c
}

// SetFlag sets a process flag
func (c *Config) SetFlag(flag string, value bool) *Config {
	if c.Process.Flags == nil {
		c.Process.Flags = make(map[string]bool)
	}
	c.Process.Flags[flag] = value
	return c
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Process == nil {
		return fmt.Errorf("process configuration is required")
	}
	if c.Process.Title == "" {
		return fmt.Errorf("process title is required")
	}
	if c.Process.MaxSteps <= 0 {
		return fmt.Errorf("max steps must be positive")
	}
	if c.Language != "en" && c.Language != "fr" {
		return fmt.Errorf("language must be 'en' or 'fr'")
	}
	if c.Theme == nil {
		return fmt.Errorf("theme configuration is required")
	}
	if c.Async == nil {
		return fmt.Errorf("async configuration is required")
	}
	if c.Async.QueueSize <= 0 {
		return fmt.Errorf("async queue size must be positive")
	}

	return nil
}

// GetStepConfig returns the configuration for a specific step
func (c *Config) GetStepConfig(stepIndex int) *StepConfig {
	if stepIndex < 0 || stepIndex >= len(c.Process.Steps) {
		return nil
	}
	return &c.Process.Steps[stepIndex]
}

// HasLiveUpdate checks if a step supports a specific live update
func (c *Config) HasLiveUpdate(stepIndex int, updateKey string) bool {
	stepConfig := c.GetStepConfig(stepIndex)
	if stepConfig == nil {
		return false
	}

	for _, key := range stepConfig.LiveUpdates {
		if key == updateKey {
			return true
		}
	}
	return false
}

// Predefined theme configurations
var (
	ThemeDefault = map[string]string{
		"primary":   "#00FF00",
		"secondary": "#FF00FF",
		"success":   "#00AA00", 
		"warning":   "#FFAA00",
		"error":     "#FF0000",
		"info":      "#00BBFF",
		"muted":     "#999999",
	}

	ThemePTBlue = map[string]string{
		"primary":   "#0066CC",
		"secondary": "#0099FF",
		"success":   "#00AA00",
		"warning":   "#FF9900",
		"error":     "#CC0000",
		"info":      "#00BBFF",
		"muted":     "#666666",
	}

	ThemePopulationGreen = map[string]string{
		"primary":   "#00AA00",
		"secondary": "#66CC00",
		"success":   "#00FF00",
		"warning":   "#FFAA00",
		"error":     "#FF0000",
		"info":      "#00BBFF",
		"muted":     "#999999",
	}

	ThemeHighContrast = map[string]string{
		"primary":   "#FFFFFF",
		"secondary": "#FFFF00",
		"success":   "#00FF00",
		"warning":   "#FFFF00",
		"error":     "#FF0000",
		"info":      "#00FFFF",
		"muted":     "#CCCCCC",
	}
)

// GetThemeColors returns the colors for a theme
func GetThemeColors(themeName string) map[string]string {
	switch themeName {
	case "pt-blue":
		return ThemePTBlue
	case "population-green":
		return ThemePopulationGreen
	case "high-contrast":
		return ThemeHighContrast
	default:
		return ThemeDefault
	}
}