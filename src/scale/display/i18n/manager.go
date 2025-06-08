package i18n

import (
	"fmt"
	"strconv"
)

// TextManager manages localized text and provides methods to retrieve text
type TextManager struct {
	locale   *LocaleData
	language string
}

// NewTextManager creates a new TextManager with the specified language
func NewTextManager(language string) (*TextManager, error) {
	locale, err := LoadLocaleWithFallback(language)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize text manager: %w", err)
	}
	
	return &TextManager{
		locale:   locale,
		language: language,
	}, nil
}

// GetStepLabel returns the localized step label for a module and step number
func (tm *TextManager) GetStepLabel(module string, step int) string {
	if moduleData, exists := tm.locale.Modules[module]; exists {
		if stepLabel, exists := moduleData.Steps[strconv.Itoa(step)]; exists {
			return stepLabel
		}
	}
	
	// Fallback to default if specific step not found
	if moduleData, exists := tm.locale.Modules[module]; exists {
		if defaultLabel, exists := moduleData.Steps["default"]; exists {
			return defaultLabel
		}
	}
	
	// Ultimate fallback
	return "Processing Step"
}

// GetCounterText returns localized counter text with variable substitution
func (tm *TextManager) GetCounterText(module, key string, vars TemplateVars) string {
	if moduleData, exists := tm.locale.Modules[module]; exists {
		if template, exists := moduleData.Counters[key]; exists {
			return ProcessTemplate(template, vars)
		}
	}
	
	// Fallback if template not found
	return key
}

// GetUIText returns localized UI text
func (tm *TextManager) GetUIText(key string) string {
	switch key {
	case "title_prefix":
		return tm.locale.UI.TitlePrefix
	case "title_description":
		return tm.locale.UI.TitleDescription
	case "time_format":
		return tm.locale.UI.TimeFormat
	case "abort_hint":
		return tm.locale.UI.AbortHint
	case "process_complete":
		return tm.locale.UI.ProcessComplete
	case "currently_prefix":
		return tm.locale.UI.CurrentlyPrefix
	default:
		return key
	}
}

// GetSystemText returns localized system text with variable substitution
func (tm *TextManager) GetSystemText(key string, vars TemplateVars) string {
	var template string
	
	switch key {
	case "label":
		template = tm.locale.System.Label
	case "cpu_format":
		template = tm.locale.System.CPUFormat
	case "ram_format":
		template = tm.locale.System.RAMFormat
	default:
		return key
	}
	
	return ProcessTemplate(template, vars)
}

// GetStatusText returns localized status text
func (tm *TextManager) GetStatusText(key string) string {
	switch key {
	case "first_chunk":
		return tm.locale.Status.FirstChunk
	case "last_chunk":
		return tm.locale.Status.LastChunk
	case "fixed":
		return tm.locale.Status.Fixed
	default:
		return key
	}
}

// GetLanguage returns the current language
func (tm *TextManager) GetLanguage() string {
	return tm.language
}

// FormatTimeText formats time text with minutes and seconds
func (tm *TextManager) FormatTimeText(minutes, seconds int) string {
	vars := CreateTemplateVars().
		SetInt("minutes", minutes).
		SetInt("seconds", seconds)
	
	return ProcessTemplate(tm.locale.UI.TimeFormat, vars)
}