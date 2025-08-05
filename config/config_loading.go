package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

var currentConfig *Config

// LoadConfig loads and validates the configuration using priority system
func LoadConfig() (*Config, string, error) {
	if currentConfig != nil {
		return currentConfig, "", nil
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
		// No config found - create local config and continue
		if err := createConfigInteractively(); err != nil {
			return nil, "", fmt.Errorf("failed to create configuration: %v", err)
		}
		configPath = localPath
		configSource = "local"
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

	applyDefaults(&config)

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &config, nil
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

	applyDefaults(&config)

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	currentConfig = &config
	return currentConfig, nil
}