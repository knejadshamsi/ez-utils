package main

import (
	"fmt"
	"os"
	"strconv"

	"ez-utils/src/population"
)

// TODO: Move help functionality to a dedicated file and implement styled output
func printUsage() {
	fmt.Println("Usage: ez-utils <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  population <file> [--db]    Process a population XML file")
	fmt.Println("\nFlags:")
	fmt.Println("  --db                        Store data in PostgreSQL database")
	fmt.Println("\nExamples:")
	fmt.Println("  ez-utils population population.xml")
	fmt.Println("  ez-utils population population.xml --db")
}

// Handles CLI argument parsing and routes commands to appropriate handlers.
// TODO: Add more commands
// TODO: Add more error handling
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

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

		// POPULATION_CHUNK_SIZE env var controls the batch size for processing
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
