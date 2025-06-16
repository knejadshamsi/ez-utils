package display

import (
	"fmt"
	
	"ez-utils/src/config"
)

// CreateConfigFromGlobal creates a display config using global application config
func CreateConfigFromGlobal(globalConfig *config.Config, processTitle, description string, maxSteps int) (*Config, error) {
	if globalConfig == nil {
		return nil, fmt.Errorf("global config cannot be nil")
	}
	
	// Validate language from global config
	language := globalConfig.Language
	if language == "" {
		language = "en" // Default to English
	}
	
	// Create display config with global language setting
	displayConfig, err := NewConfig(processTitle, description, language, maxSteps)
	if err != nil {
		return nil, fmt.Errorf("failed to create display config: %w", err)
	}
	
	return displayConfig, nil
}

// NewPTDisplayConfig creates a pre-configured display config for PT processing
func NewPTDisplayConfig(globalConfig *config.Config, cleanFlag bool) (*Config, error) {
	config, err := CreateConfigFromGlobal(
		globalConfig,
		"GTFS Transit Processing",
		"Converting GTFS data to MATSim transit schedule format",
		9, // PT has 9 steps
	)
	if err != nil {
		return nil, err
	}
	
	// Configure PT-specific steps
	config.AddStep("Validate GTFS Files", "Checking required GTFS files", "file", "status", "size").
		AddStep("Identify Service Patterns", "Analyzing calendar.txt for service patterns", "services", "progress").
		AddStep("Map Service Routes", "Linking services to routes", "routes", "mappings").
		AddStep("Select Services", "Choosing services for processing", "selected", "total").
		AddStep("Collect Trips", "Gathering trip data", "trips", "progress").
		AddStep("Extract Stop Sequences", "Processing stop sequences", "sequences", "stops").
		AddStep("Gather Stop Details", "Collecting stop information", "stops", "details").
		AddStep("Generate XML", "Creating transit schedule XML", "progress", "size").
		AddStep("Cleanup Files", "Removing temporary files", "files", "dirs", "bytes")
		
	// Set PT theme and flags
	config.SetTheme("pt-blue").
		SetFlag("clean", cleanFlag)
	
	return config, nil
}

// NewPopulationDisplayConfig creates a pre-configured display config for population processing
// Note: This is for future use if population ever switches to this display system
func NewPopulationDisplayConfig(globalConfig *config.Config, cleanFlag, dbFlag bool) (*Config, error) {
	config, err := CreateConfigFromGlobal(
		globalConfig,
		"Population Processing",
		"Scaling down MATSim population data",
		19, // Population has up to 19 steps
	)
	if err != nil {
		return nil, err
	}
	
	// Configure population-specific steps (based on current locale files)
	config.AddStep("Create Population Chunks", "Breaking population into manageable chunks", "count", "persons", "mb").
		AddStep("Fix XML Structure", "Correcting XML formatting issues", "status").
		AddStep("Extract Agent Locations", "Getting location data from agents", "persons", "current", "total").
		AddStep("Store Agent Data in Database", "Saving agent data for processing", "coordinates").
		SetTheme("population-green").
		SetFlag("clean", cleanFlag).
		SetFlag("db", dbFlag)
	
	// Add remaining steps...
	// (This would continue with all 19 steps from the existing locale files)
	
	return config, nil
}

// Helper function to validate global config language
func ValidateGlobalConfigLanguage(globalConfig *config.Config) error {
	if globalConfig == nil {
		return fmt.Errorf("global config is nil")
	}
	
	language := globalConfig.Language
	if language != "" && language != "en" && language != "fr" {
		return fmt.Errorf("unsupported language '%s' in global config, must be 'en' or 'fr'", language)
	}
	
	return nil
}