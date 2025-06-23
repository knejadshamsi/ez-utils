package display

import (
	"fmt"
	"strings"

	"ez-utils/src/display/config"
	"ez-utils/src/display/lang"
)

// Global text manager instance
var textManager *lang.TextManager

// InitializeTextManager initializes the global text manager with the specified language
func InitializeTextManager(language string) error {
	if language == "" {
		language = "en" // Default to English
	}

	tm, err := lang.NewTextManager(language)
	if err != nil {
		// Try fallback to English if the requested language fails
		if language != "en" {
			tm, fallbackErr := lang.NewTextManager("en")
			if fallbackErr != nil {
				return fmt.Errorf("failed to initialize text manager for language '%s' and English fallback failed: %v (original error: %v)", language, fallbackErr, err)
			}
			textManager = tm
			return nil
		}
		return fmt.Errorf("failed to initialize text manager for English: %v", err)
	}
	textManager = tm
	return nil
}

// GetTextManager returns the global text manager instance
func GetTextManager() *lang.TextManager {
	return textManager
}

func GetStepLabel(module string, step int) string {
	if textManager == nil {
		// Fallback if text manager not initialized
		return "Processing Step"
	}
	
	// Convert module name to lowercase for consistency with JSON keys
	moduleKey := strings.ToLower(module)
	return textManager.GetStepLabel(moduleKey, step)
}

// GetStepLabelFromConfig returns step label from configured steps, with locale fallback
func GetStepLabelFromConfig(steps []config.StepConfig, module string, step int) string { // Changed to config.StepConfig
	// First try to get from configured steps (step is 0-based index)
	if step >= 0 && step < len(steps) {
		if steps[step].Title != "" {
			return steps[step].Title
		}
	}
	
	// Fall back to locale files
	return GetStepLabel(module, step)
}
