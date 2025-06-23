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