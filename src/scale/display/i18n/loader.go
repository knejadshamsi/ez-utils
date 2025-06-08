package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LocaleData represents the structure of a locale JSON file
type LocaleData struct {
	Modules map[string]ModuleData `json:"modules"`
	UI      UIData                `json:"ui"`
	System  SystemData            `json:"system"`
	Status  StatusData            `json:"status"`
}

// ModuleData represents module-specific localization data
type ModuleData struct {
	Steps    map[string]string `json:"steps"`
	Counters map[string]string `json:"counters"`
}

// UIData represents UI-specific localization data
type UIData struct {
	TitlePrefix     string `json:"title_prefix"`
	TitleDescription string `json:"title_description"`
	TimeFormat      string `json:"time_format"`
	AbortHint       string `json:"abort_hint"`
	ProcessComplete string `json:"process_complete"`
	CurrentlyPrefix string `json:"currently_prefix"`
}

// SystemData represents system-specific localization data
type SystemData struct {
	Label     string `json:"label"`
	CPUFormat string `json:"cpu_format"`
	RAMFormat string `json:"ram_format"`
}

// StatusData represents status-specific localization data
type StatusData struct {
	FirstChunk string `json:"first_chunk"`
	LastChunk  string `json:"last_chunk"`
	Fixed      string `json:"fixed"`
}

// LoadLocale loads a locale JSON file and returns the parsed data
func LoadLocale(language string) (*LocaleData, error) {
	// Get the directory where the executable is located
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Navigate to the locales directory relative to the project structure
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(execPath))))
	localesPath := filepath.Join(projectRoot, "src", "scale", "display", "locales")
	
	// If that doesn't exist, try relative to current working directory
	if _, err := os.Stat(localesPath); os.IsNotExist(err) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		localesPath = filepath.Join(cwd, "src", "scale", "display", "locales")
	}
	
	filePath := filepath.Join(localesPath, language+".json")
	
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("locale file not found: %s", filePath)
	}
	
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read locale file %s: %w", filePath, err)
	}
	
	// Parse JSON
	var localeData LocaleData
	if err := json.Unmarshal(data, &localeData); err != nil {
		return nil, fmt.Errorf("failed to parse locale file %s: %w", filePath, err)
	}
	
	return &localeData, nil
}

// LoadLocaleWithFallback loads a locale with English fallback
func LoadLocaleWithFallback(language string) (*LocaleData, error) {
	// Try to load the requested language
	locale, err := LoadLocale(language)
	if err == nil {
		return locale, nil
	}
	
	// If the requested language is not English, try English as fallback
	if language != "en" {
		fallbackLocale, fallbackErr := LoadLocale("en")
		if fallbackErr == nil {
			return fallbackLocale, nil
		}
		// Return the original error if even English fails
		return nil, fmt.Errorf("failed to load locale %s and fallback failed: %w", language, err)
	}
	
	// If English itself failed, return the error
	return nil, err
}