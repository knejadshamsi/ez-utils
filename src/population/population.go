package population

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ez-utils/src/display"
	"ez-utils/src/population/steps"
)

// ProcessPopulation orchestrates the XML population processing workflow
func ProcessPopulation(inputPath, outputDir string, flagDB bool, chunkSize int) error {
	display.InitDisplay("POPULATION", false, flagDB)

	// Step 1: Create Population Chunks
	display.SetStep(1)
	chunksDir := filepath.Join(outputDir, "population", "01_split_population_chunks", "chunks")
	os.MkdirAll(chunksDir, os.ModePerm)
	if err := steps.SplitPopulation(inputPath, chunkSize); err != nil {
		return fmt.Errorf("step 1 (Create Population Chunks) failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Step 2: Fix XML Structure
	display.SetStep(2)
	if err := steps.SplitPopulationPart2(inputPath, 0); err != nil {
		return fmt.Errorf("step 2 (Fix XML Structure) failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Step 3: Extract Agent Locations
	display.SetStep(3)
	indexesDir := filepath.Join(outputDir, "population", "02_extract_agent_locations", "indexes")
	if err := os.MkdirAll(indexesDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create indexes directory: %w", err)
	}
	if err := steps.ExtractAgentLocations(); err != nil {
		return fmt.Errorf("step 3 (Extract Agent Locations) failed: %w", err)
	}

	if flagDB {
		time.Sleep(500 * time.Millisecond)

		// Step 4: Store agent data in database
		display.SetStep(4)
		if err := steps.StoreAgentsInDatabase(indexesDir); err != nil {
			return fmt.Errorf("step 4 (Store Agent Data in Database) failed: %w", err)
		}
	}

	display.SetProcessComplete(true)
	return nil
}
