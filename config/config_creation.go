package config

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"
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

	if err := createConfigAt(targetPath, isLocal); err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	fmt.Printf("\nConfiguration file created at: %s\n", displayPath)
	fmt.Println("Please review and update the configuration settings, then run the command again.")
	fmt.Println("\nYou can edit the file with your preferred text editor:")
	fmt.Printf("  nano %s\n", targetPath)
	fmt.Printf("  vim %s\n", targetPath)
	fmt.Printf("  code %s\n", targetPath)
	fmt.Println("\nThe configuration file contains detailed comments explaining each setting.")

	os.Exit(0)
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