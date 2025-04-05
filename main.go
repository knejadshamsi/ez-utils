package main

import (
	"fmt"
	"os"
	"strconv"

	"ez-utils/src/help"
	"ez-utils/src/population"
)

func printUsage() {
	fmt.Println("Usage: ez-utils <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  population    Process population data")
	fmt.Println("  network       Analyze network connections")
	fmt.Println("  pt           Process public transit data")
	fmt.Println("\nFor detailed help on each command, use:")
	fmt.Println("  <command> --help")
	fmt.Println("\nExample:")
	fmt.Println("  ez-utils population --help")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle help requests
	if len(os.Args) == 3 && os.Args[2] == "--help" {
		switch command {
		case "population":
			help.PrintPopulationHelp()
		case "network":
			help.PrintNetworkHelp()
		case "pt":
			help.PrintPTHelp()
		default:
			fmt.Printf("Unknown command: %s\n", command)
			printUsage()
		}
		os.Exit(0)
	}

	switch command {
	case "population":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a path to the population file")
			fmt.Println("Usage: ez-utils population <population-file-path>")
			os.Exit(1)
		}

		populationFile := os.Args[2]
		outputDir := "temp"

		flagDB := false
		for i := 3; i < len(os.Args); i++ {
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

	case "help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
