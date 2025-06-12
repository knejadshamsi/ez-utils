package main

import (
	"ez-utils/src/config"
	"ez-utils/src/help"
	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/processing"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

			// Validate input file exists
			if !population.FileExists(populationFile) {
				fmt.Printf("Error: Input file does not exist: %s\n", populationFile)
				os.Exit(1)
			}
			// fmt.Printf("DEBUG: Input file validated and exists\n")

			// Create Phase One configuration from loaded config
			phaseOneConfig := cfg.ToPhaseOneConfig()
			// fmt.Printf("DEBUG: Phase One config created - Extractors: %d, Mappers: %d\n",
			//	phaseOneConfig.ExtractorCount, phaseOneConfig.HashMapCount)

			// Queue configuration is now handled internally by PopulationProcessor

			// Create the simplified population processor
			processor := processing.NewPopulationProcessor(phaseOneConfig)

			// Disable TUI in non-interactive environments
			if os.Getenv("TERM") == "" || os.Getenv("CI") != "" {
				processor.DisableTUI()
			}

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down gracefully...")
				os.Exit(0)
			}()

			// Process the population file (all phases handled internally)
			if err := processor.Process(populationFile); err != nil {
				fmt.Printf("Population processing failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Phase One completed successfully!")

			// Start Phase Two processing
			fmt.Println("\nStarting Phase Two...")

			// Create Phase Two configuration
			phaseTwoConfig := cfg.ToPhaseTwoConfig()

			// Create Phase Two processor with proper parallel architecture
			phaseTwoProcessor := processing.NewPhaseTwoProcessor(
				phaseTwoConfig,
				populationFile,
			)

			// Disable TUI in non-interactive environments
			if os.Getenv("TERM") == "" || os.Getenv("CI") != "" {
				phaseTwoProcessor.DisableTUI()
			}

			// Process Phase Two
			if err := phaseTwoProcessor.Process(); err != nil {
				fmt.Printf("Phase Two failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Phase Two completed successfully!")

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
