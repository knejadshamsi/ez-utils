package tui

import (
	"fmt"
	
	"github.com/charmbracelet/lipgloss"
)

// Color palette - following existing patterns
var (
	colorGreen   = lipgloss.Color("#00ff00")
	colorMagenta = lipgloss.Color("#ff00ff")
	colorYellow  = lipgloss.Color("#ffff00")
	colorOrange  = lipgloss.Color("#ffa500")
	colorRed     = lipgloss.Color("#ff0000")
	colorBlue    = lipgloss.Color("#0000ff")
	colorCyan    = lipgloss.Color("#00ffff")
	colorWhite   = lipgloss.Color("#ffffff")
	colorGray    = lipgloss.Color("#808080")
)

// Style definitions for the TUI
type StyleDefinitions struct {
	// Header styles
	titleEZ           lipgloss.Style
	titleText         lipgloss.Style
	timeInfo          lipgloss.Style
	phaseInfo         lipgloss.Style
	
	// Container styles
	mainContainer     lipgloss.Style
	sectionContainer  lipgloss.Style
	agentContainer    lipgloss.Style
	
	// System metrics styles
	systemLabel       lipgloss.Style
	cpuInfo           lipgloss.Style
	ramInfo           lipgloss.Style
	diskInfo          lipgloss.Style
	
	// Agent status styles
	agentActive       lipgloss.Style
	agentBusy         lipgloss.Style
	agentPaused       lipgloss.Style
	agentAvailable    lipgloss.Style
	
	// Progress styles
	progressBar       lipgloss.Style
	progressFilled    lipgloss.Style
	progressEmpty     lipgloss.Style
	
	// Grid and data styles
	gridInfo          lipgloss.Style
	densityInfo       lipgloss.Style
	counterInfo       lipgloss.Style
	
	// Error and warning styles
	errorText         lipgloss.Style
	warningText       lipgloss.Style
	successText       lipgloss.Style
}

// NewStyleDefinitions creates and returns a new set of style definitions
func NewStyleDefinitions() StyleDefinitions {
	return StyleDefinitions{
		// Header styles
		titleEZ: lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1),
		
		titleText: lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1),
		
		timeInfo: lipgloss.NewStyle().
			Foreground(colorYellow).
			PaddingLeft(2),
		
		phaseInfo: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			PaddingLeft(2),
		
		// Container styles
		mainContainer: lipgloss.NewStyle().
			Padding(0, 1),
		
		sectionContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(1).
			Margin(1, 0),
		
		agentContainer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorGray).
			Padding(0, 1).
			Margin(0, 1),
		
		// System metrics styles
		systemLabel: lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true),
		
		cpuInfo: lipgloss.NewStyle().
			Foreground(colorGreen),
		
		ramInfo: lipgloss.NewStyle().
			Foreground(colorBlue),
		
		diskInfo: lipgloss.NewStyle().
			Foreground(colorMagenta),
		
		// Agent status styles
		agentActive: lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true),
		
		agentBusy: lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true),
		
		agentPaused: lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true),
		
		agentAvailable: lipgloss.NewStyle().
			Foreground(colorCyan),
		
		// Progress styles
		progressBar: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorGray),
		
		progressFilled: lipgloss.NewStyle().
			Background(colorGreen).
			Foreground(colorWhite),
		
		progressEmpty: lipgloss.NewStyle().
			Background(colorGray),
		
		// Grid and data styles
		gridInfo: lipgloss.NewStyle().
			Foreground(colorCyan),
		
		densityInfo: lipgloss.NewStyle().
			Foreground(colorMagenta),
		
		counterInfo: lipgloss.NewStyle().
			Foreground(colorYellow),
		
		// Error and warning styles
		errorText: lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true),
		
		warningText: lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true),
		
		successText: lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true),
	}
}

// Helper functions for common styling patterns
func (s StyleDefinitions) GetAgentStateStyle(state string) lipgloss.Style {
	switch state {
	case "busy":
		return s.agentBusy
	case "paused":
		return s.agentPaused
	case "available":
		return s.agentAvailable
	default:
		return s.agentActive
	}
}

func (s StyleDefinitions) FormatCounter(label string, value interface{}) string {
	return s.systemLabel.Render(label+":") + " " + s.counterInfo.Render(formatValue(value))
}

func (s StyleDefinitions) FormatSystemMetric(label string, value interface{}, style lipgloss.Style) string {
	return s.systemLabel.Render(label+":") + " " + style.Render(formatValue(value))
}

// Helper function to format values consistently
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case int:
		return formatNumber(int64(v))
	case int64:
		return formatNumber(v)
	case float64:
		if v < 10 {
			return fmt.Sprintf("%.2f", v)
		}
		return formatNumber(int64(v))
	case string:
		return v
	default:
		return ""
	}
}

// Format numbers with commas for readability
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
}