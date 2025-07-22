package cmd

import (
	"ez-utils/src/create/pt"
	"ez-utils/src/create/tv"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunCreate executes the create command logic.
func RunCreate() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Please provide a subcommand")
		fmt.Println("Usage: ez-utils create <subcommand> [arguments]")
		fmt.Println("\nAvailable subcommands:")
		fmt.Println("  pt    Create public transit schedule from GTFS data")
		fmt.Println("  tv    Create transit vehicles from schedule")
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

	case "tv":
		if err := tv.Run(os.Args[3:]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown create subcommand: %s\n", subcommand)
		fmt.Println("\nAvailable subcommands:")
		fmt.Println("  pt    Create public transit schedule from GTFS data")
		fmt.Println("  tv    Create transit vehicles from schedule")
		os.Exit(1)
	}
}