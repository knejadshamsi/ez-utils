package main

import (
	"fmt"
	"os"
	"strconv"

	"ez-utils/src/help"
	"ez-utils/src/scale/population"
)

func main() {
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
			outputDir := "temp"

			flagDB := false
			for i := 4; i < len(os.Args); i++ {
				if os.Args[i] == "--db" {
					flagDB = true
					break
				}
			}

			chunkSize := 2000
			if envSize := os.Getenv("POPULATION_CHUNK_SIZE"); envSize != "" {
				if size, err := strconv.Atoi(envSize); err == nil && size > 0 {
					chunkSize = size
				}
			}

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
