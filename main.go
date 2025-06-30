//go:build !wails

package main

import (
	"ez-utils/cmd"
	"ez-utils/config"
	"ez-utils/src/help"
	"fmt"
	"os"
)

func main() {
	if hasNewConfigFlag() {
		if err := config.CreateNewConfig(); err != nil {
			fmt.Printf("Error creating new configuration: %v\n", err)
			os.Exit(1)
		}
		return // CreateNewConfig will exit, but this is defensive
	}

	// Check and create config file as the very first action
	if err := config.CheckAndCreateConfig(); err != nil {
		fmt.Printf("Error checking configuration: %v\n", err)
		os.Exit(1)
	}

	// Load configuration and get which type is being used
	cfg, configSource, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Show which config is being used (except for help commands)
	if len(os.Args) >= 2 && !isHelpCommand() {
		showConfigSource(configSource)
	}

	if len(os.Args) < 2 {
		help.PrintHelp("", "", cfg)
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle general help requests
	if command == "help" || command == "--help" {
		help.PrintHelp("", "", cfg)
		os.Exit(0)
	}

	// Handle help for specific commands and subcommands
	if os.Args[len(os.Args)-1] == "--help" {
		if len(os.Args) > 2 {
			subcommand := os.Args[2]
			help.PrintHelp(command, subcommand, cfg)
		} else {
			help.PrintHelp(command, "", cfg)
		}
		os.Exit(0)
	}

	switch command {
	case "create":
		cmd.RunCreate()
	case "edit":
		cmd.RunEdit()
	case "scale":
		cmd.RunScale()
	case "network":
		fmt.Println("Network command not yet implemented")
		os.Exit(1)
	case "help":
		help.PrintHelp("", "", cfg)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		help.PrintHelp("", "", cfg)
		os.Exit(1)
	}
}

// hasNewConfigFlag checks if --new-config flag is present anywhere in args
func hasNewConfigFlag() bool {
	for _, arg := range os.Args {
		if arg == "--new-config" {
			return true
		}
	}
	return false
}

// isHelpCommand checks if this is a help-related command
func isHelpCommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	command := os.Args[1]
	return command == "help" || command == "--help" ||
		(len(os.Args) >= 2 && os.Args[len(os.Args)-1] == "--help")
}

// showConfigSource displays which config file is being used
func showConfigSource(source string) {
	if source == "local" {
		// Using local config silently
	} else if source == "global" {
		_, _, globalDisplay, err := config.GetConfigPaths()
		if err == nil {
			fmt.Printf("Using global config: %s\n", globalDisplay)
		} else {
			fmt.Println("Using global config")
		}
	}
}
