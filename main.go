package main

import (
	"ez-utils/src/config"
	"ez-utils/src/help"
	"ez-utils/src/scale/population"
	"fmt"
	"os"
)

func main() {
	// Check for --new-config flag first (before any config loading)
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
		help.PrintDefaultHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle help requests
	if len(os.Args) >= 2 && (command == "--help" || command == "help") {
		help.PrintDefaultHelp()
		os.Exit(0)
	}

	// Handle help for specific commands
	if len(os.Args) >= 3 && os.Args[len(os.Args)-1] == "--help" {
		if command == "scale" && len(os.Args) >= 4 {
			subcommand := os.Args[2]
			switch subcommand {
			case "population":
				help.PrintPopulationHelp()
			default:
				fmt.Printf("Unknown scale subcommand: %s\n", subcommand)
				help.PrintDefaultHelp()
			}
		} else {
			switch command {
			case "network":
				help.PrintNetworkHelp()
			case "pt":
				help.PrintPTHelp()
			default:
				fmt.Printf("Unknown command: %s\n", command)
				help.PrintDefaultHelp()
			}
		}
		os.Exit(0)
	}

	switch command {
	case "scale":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a subcommand")
			fmt.Println("Usage: ez-utils scale <subcommand> [arguments]")
			fmt.Println("\nAvailable subcommands:")
			fmt.Println("  population    Process population data")
			os.Exit(1)
		}

		subcommand := os.Args[2]
		switch subcommand {
		case "population":
			if len(os.Args) < 4 {
				fmt.Println("Error: Please provide a path to the population file")
				fmt.Println("Usage: ez-utils scale population <population-file-path>")
				os.Exit(1)
			}

			populationFile := os.Args[3]
			outputDir := cfg.Population.OutputDir

			flagDB := false
			for i := 4; i < len(os.Args); i++ {
				if os.Args[i] == "--db" {
					flagDB = true
					break
				}
			}

			// Use chunk size from config
			chunkSize := cfg.Population.ChunkSize

			err := population.ProcessPopulation(populationFile, outputDir, flagDB, chunkSize)
			if err != nil {
				fmt.Printf("Error processing population: %v\n", err)
				os.Exit(1)
			}

		default:
			fmt.Printf("Unknown scale subcommand: %s\n", subcommand)
			fmt.Println("\nAvailable subcommands:")
			fmt.Println("  population    Process population data")
			os.Exit(1)
		}

	case "network":
		fmt.Println("Network command not yet implemented")
		os.Exit(1)

	case "pt":
		fmt.Println("PT command not yet implemented")
		os.Exit(1)

	case "help":
		help.PrintDefaultHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		help.PrintDefaultHelp()
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
		fmt.Println("Using local config: ./config.yaml")
	} else if source == "global" {
		localPath, _, globalDisplay, err := config.GetConfigPaths()
		_ = localPath // Avoid unused variable warning
		if err == nil {
			fmt.Printf("Using global config: %s\n", globalDisplay)
		} else {
			fmt.Println("Using global config")
		}
	}
}
