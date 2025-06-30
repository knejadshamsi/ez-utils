package cmd

import (
	"ez-utils/config"
	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/processing"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// RunScale executes the scale command logic.
func RunScale() {
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

		// Load configuration
		cfg, _, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
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
}