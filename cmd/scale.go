package cmd

import (
	"ez-utils/config"
	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/processing"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
			fmt.Println("Usage: ez-utils scale population <population-file-path> [-s|--scales scale1,scale2,...]")
			os.Exit(1)
		}

		populationFile := os.Args[3]

		// Parse command line flags for scales
		var customScales []int
		for i := 4; i < len(os.Args); i++ {
			if os.Args[i] == "-s" || os.Args[i] == "--scales" {
				if i+1 < len(os.Args) {
					scalesStr := os.Args[i+1]
					scales, err := parseScales(scalesStr)
					if err != nil {
						fmt.Printf("Error parsing scales: %v\n", err)
						os.Exit(1)
					}
					customScales = scales
					break
				} else {
					fmt.Println("Error: -s/--scales flag requires a value")
					os.Exit(1)
				}
			}
		}

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

		// Override scales if provided via CLI
		if len(customScales) > 0 {
			cfg.Population.Scales = customScales
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

// parseScales parses a comma-separated string of scale values
func parseScales(scalesStr string) ([]int, error) {
	parts := strings.Split(scalesStr, ",")
	scales := make([]int, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		scale, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid scale value '%s': must be integer", part)
		}
		
		if scale < 1 || scale > 100 {
			return nil, fmt.Errorf("scale value %d out of range: must be between 1-100", scale)
		}
		
		scales = append(scales, scale)
	}
	
	if len(scales) == 0 {
		return nil, fmt.Errorf("no valid scale values provided")
	}
	
	return scales, nil
}