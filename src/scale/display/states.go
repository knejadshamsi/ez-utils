package display

import (
	"strings"

	"ez-utils/src/scale/display/i18n"
)

// Global text manager instance
var textManager *i18n.TextManager

// InitializeTextManager initializes the global text manager with the specified language
func InitializeTextManager(language string) error {
	tm, err := i18n.NewTextManager(language)
	if err != nil {
		return err
	}
	textManager = tm
	return nil
}

// GetTextManager returns the global text manager instance
func GetTextManager() *i18n.TextManager {
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
