package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LocaleCache represents a cached locale with metadata
type LocaleCache struct {
	data      *LocaleData
	loadTime  time.Time
	filePath  string
	fileModTime time.Time
}

// Global cache for loaded locales
var (
	localeCache = make(map[string]*LocaleCache)
	cacheMutex  sync.RWMutex
	cacheMaxAge = 5 * time.Minute // Cache for 5 minutes
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
// Uses caching to improve performance on repeated loads
func LoadLocale(language string) (*LocaleData, error) {
	// Check cache first
	if cachedData := getCachedLocale(language); cachedData != nil {
		return cachedData, nil
	}

	return loadLocaleFromFile(language)
}

// getCachedLocale retrieves locale from cache if valid
func getCachedLocale(language string) *LocaleData {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	cached, exists := localeCache[language]
	if !exists {
		return nil
	}

	// Check if cache is still valid
	if time.Since(cached.loadTime) > cacheMaxAge {
		return nil
	}

	// Check if file has been modified
	if fileInfo, err := os.Stat(cached.filePath); err == nil {
		if fileInfo.ModTime().After(cached.fileModTime) {
			// File has been modified, cache is invalid
			return nil
		}
	}

	return cached.data
}

// loadLocaleFromFile loads locale from file and caches it
func loadLocaleFromFile(language string) (*LocaleData, error) {
	if language == "" {
		return nil, fmt.Errorf("language cannot be empty")
	}

	// Try multiple possible paths for locales
	var localesPath string
	var err error

	// Method 1: Relative to executable
	execPath, err := os.Executable()
	if err == nil {
		projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(execPath))))
		localesPath = filepath.Join(projectRoot, "src", "scale", "display", "locales")
		if _, err := os.Stat(localesPath); err == nil {
			// Found it
		} else {
			localesPath = ""
		}
	}

	// Method 2: Relative to current working directory
	if localesPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		localesPath = filepath.Join(cwd, "src", "scale", "display", "locales")
		if _, err := os.Stat(localesPath); os.IsNotExist(err) {
			localesPath = ""
		}
	}

	// Method 3: Try common development paths
	if localesPath == "" {
		commonPaths := []string{
			"./src/scale/display/locales",
			"../src/scale/display/locales",
			"../../src/scale/display/locales",
			"./display/locales",
			"./locales",
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				localesPath = path
				break
			}
		}
	}

	if localesPath == "" {
		return nil, fmt.Errorf("could not find locales directory for language '%s'")
	}
	
	filePath := filepath.Join(localesPath, language+".json")
	
	// Check if the file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("locale file not found: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access locale file %s: %w", filePath, err)
	}

	// Check if it's actually a file and not empty
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("locale path is a directory, not a file: %s", filePath)
	}
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("locale file is empty: %s", filePath)
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

	// Validate the parsed data
	if err := validateLocaleData(&localeData); err != nil {
		return nil, fmt.Errorf("invalid locale data in %s: %w", filePath, err)
	}

	// Cache the loaded data
	cacheLocale(language, &localeData, filePath, fileInfo.ModTime())
	
	return &localeData, nil
}

// cacheLocale stores locale data in cache
func cacheLocale(language string, data *LocaleData, filePath string, modTime time.Time) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	localeCache[language] = &LocaleCache{
		data:        data,
		loadTime:    time.Now(),
		filePath:    filePath,
		fileModTime: modTime,
	}
}

// ClearLocaleCache clears all cached locale data
func ClearLocaleCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	localeCache = make(map[string]*LocaleCache)
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

// validateLocaleData validates that the loaded locale data has required fields
func validateLocaleData(data *LocaleData) error {
	if data == nil {
		return fmt.Errorf("locale data is nil")
	}

	// Check that essential UI fields exist
	if data.UI.TitlePrefix == "" {
		return fmt.Errorf("missing required UI field: title_prefix")
	}
	if data.UI.TitleDescription == "" {
		return fmt.Errorf("missing required UI field: title_description")
	}
	if data.UI.TimeFormat == "" {
		return fmt.Errorf("missing required UI field: time_format")
	}

	// Check that modules exist
	if data.Modules == nil {
		return fmt.Errorf("modules section is missing")
	}

	// Check that at least one module has steps
	hasSteps := false
	for _, moduleData := range data.Modules {
		if moduleData.Steps != nil && len(moduleData.Steps) > 0 {
			hasSteps = true
			break
		}
	}
	if !hasSteps {
		return fmt.Errorf("no modules contain step definitions")
	}

	return nil
}