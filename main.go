package main

import (
	"ez-utils/config"
	"ez-utils/src/create/pt"
	"ez-utils/src/help"
	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/processing"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a subcommand")
			fmt.Println("Usage: ez-utils create <subcommand> [arguments]")
			fmt.Println("\nAvailable subcommands:")
			fmt.Println("  pt    Create public transit schedule from GTFS data")
			os.Exit(1)
		}

		subcommand := os.Args[2]
		switch subcommand {
		case "pt":
			if len(os.Args) < 4 {
				fmt.Println("Error: Please provide GTFS directory")
				fmt.Println("Usage: ez-utils create pt <gtfs-directory> [time-period]")
				fmt.Println("\nTime periods: week, work, or DD-MM-YY (default: work)")
				os.Exit(1)
			}

			gtfsDir := os.Args[3]

			// Default time period to "work" if not provided
			timePeriod := "work"
			if len(os.Args) >= 5 {
				timePeriod = os.Args[4]
			}

			// Validate input directory exists
			if _, err := os.Stat(gtfsDir); os.IsNotExist(err) {
				fmt.Printf("Error: GTFS directory does not exist: %s\n", gtfsDir)
				os.Exit(1)
			}

			// Parse service day from time period
			var serviceDay pt.ServiceDay
			switch timePeriod {
			case "week":
				serviceDay = pt.ServiceDayWeekend
			case "work":
				serviceDay = pt.ServiceDayWeekday
			default:
				// For specific dates (DD-MM-YY), default to weekday for now
				// TODO: Parse specific date and determine day of week
				serviceDay = pt.ServiceDayWeekday
			}

			// Check for clean flag
			cleanFlag := false
			for _, arg := range os.Args {
				if arg == "--clean" || arg == "-c" {
					cleanFlag = true
					break
				}
			}

			// Create the PT orchestrator
			orchestrator := pt.NewPTOrchestrator(gtfsDir, serviceDay, cleanFlag)

			// Disable TUI in non-interactive environments
			if os.Getenv("TERM") == "" || os.Getenv("CI") != "" {
				orchestrator.DisableTUI()
			}

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				os.Exit(0)
			}()

			// Process all 9 steps with quit handling
			done := make(chan error, 1)
			go func() {
				done <- orchestrator.Process()
			}()

			// Wait for either completion or quit signal
			for {
				select {
				case err := <-done:
					if err != nil {
						os.Exit(1)
					}
					// Wait a moment for cleanup to complete
					time.Sleep(200 * time.Millisecond)
					return // Process completed successfully
				default:
					if orchestrator.IsQuitRequested() {
						// Wait a moment for cleanup to complete
						time.Sleep(200 * time.Millisecond)
						return // User requested quit, exit cleanly
					}
					time.Sleep(100 * time.Millisecond) // Small delay to prevent busy waiting
				}
			}

		default:
			fmt.Printf("Unknown create subcommand: %s\n", subcommand)
			fmt.Println("\nAvailable subcommands:")
			fmt.Println("  pt    Create public transit schedule from GTFS data")
			os.Exit(1)
		}

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

			// Create Phase One configuration from loaded config
			phaseOneConfig := cfg.ToPhaseOneConfig()


			// Create Phase Two configuration
			phaseTwoConfig := cfg.ToPhaseTwoConfig()

			// Create the 5-phase population scaling orchestrator
			orchestrator := processing.NewPopulationScalingOrchestrator(
				phaseOneConfig,
				phaseTwoConfig,
				populationFile,
			)

			// Disable TUI in non-interactive environments
			if os.Getenv("TERM") == "" || os.Getenv("CI") != "" {
				orchestrator.DisableTUI()
			}

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				os.Exit(0)
			}()

			// Process all 5 phases
			if err := orchestrator.Process(); err != nil {
				fmt.Printf("Population scaling failed: %v\n", err)
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
