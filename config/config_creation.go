package config

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed config.template.yaml
var templateContent []byte

// CheckAndCreateConfig checks if config exists anywhere, creates one if not
func CheckAndCreateConfig() error {
	localPath, globalPath, _, err := GetConfigPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %v", err)
	}

	if _, err := os.Stat(localPath); err == nil {
		return nil
	}

	if _, err := os.Stat(globalPath); err == nil {
		return nil
	}

	return createConfigInteractively()
}

// CreateNewConfig forces creation of a new config file (for --new-config flag)
func CreateNewConfig() error {
	return createConfigInteractively()
}

// createConfigInteractively creates local config automatically
func createConfigInteractively() error {
	localPath, _, _, err := GetConfigPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %v", err)
	}

	// Always create local config automatically
	targetPath := localPath
	displayPath := localPath
	isLocal := true

	if err := createConfigAt(targetPath, isLocal); err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	fmt.Printf("Created configuration file: %s\n", displayPath)
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