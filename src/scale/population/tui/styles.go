package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions for TUI
var (
	colorWhite      = lipgloss.Color("#FFFFFF")
	colorCyan       = lipgloss.Color("#00FFFF") 
	colorOrange     = lipgloss.Color("#FFA500")
	colorPurple     = lipgloss.Color("#AA00FF")
	colorMagenta    = lipgloss.Color("#FF00FF")
	colorBlue       = lipgloss.Color("#0066FF")
	colorYellow     = lipgloss.Color("#FFFF00")
	colorGreenBright = lipgloss.Color("#00FF7F")
	colorTurquoise  = lipgloss.Color("#40E0D0")
	colorLightBlue  = lipgloss.Color("#87CEEB")
	colorGreen      = lipgloss.Color("#00FF00")
	colorRed        = lipgloss.Color("#FF0000")
)

// UnifiedStyles contains all styling definitions for the TUI
type UnifiedStyles struct {
	// Header styles
	titleStyle       lipgloss.Style
	timeStyle        lipgloss.Style
	phaseStyle       lipgloss.Style
	actionStyle      lipgloss.Style
	headerStyle      lipgloss.Style
	
	// System metric styles
	cpuStyle         lipgloss.Style
	ramStyle         lipgloss.Style
	diskStyle        lipgloss.Style
	
	// Worker table styles
	tableHeaderStyle lipgloss.Style
	workerActiveStyle lipgloss.Style
	statusActiveStyle lipgloss.Style
	statusCompletedStyle lipgloss.Style
	metricStyle      lipgloss.Style
	
	// Container styles
	mainContainer    lipgloss.Style
	sectionContainer lipgloss.Style
}

// NewStyles creates and returns style definitions
func NewStyles() UnifiedStyles {
	return UnifiedStyles{
		// Header styles
		titleStyle: lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true).
			Align(lipgloss.Center),
		
		timeStyle: lipgloss.NewStyle().
			Foreground(colorOrange).
			Align(lipgloss.Center),
		
		phaseStyle: lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true).
			Align(lipgloss.Center),
		
		actionStyle: lipgloss.NewStyle().
			Foreground(colorYellow).
			Align(lipgloss.Center),
		
		headerStyle: lipgloss.NewStyle().
			Foreground(colorWhite),
		
		// System metric styles
		cpuStyle: lipgloss.NewStyle().
			Foreground(colorTurquoise),
		
		ramStyle: lipgloss.NewStyle().
			Foreground(colorGreenBright),
		
		diskStyle: lipgloss.NewStyle().
			Foreground(colorLightBlue),
		
		// Worker table styles
		tableHeaderStyle: lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true).
			MarginTop(1).
			MarginBottom(1),
		
		workerActiveStyle: lipgloss.NewStyle().
			Foreground(colorWhite),
		
		statusActiveStyle: lipgloss.NewStyle().
			Foreground(colorGreen),
		
		statusCompletedStyle: lipgloss.NewStyle().
			Foreground(colorRed),
		
		metricStyle: lipgloss.NewStyle().
			Foreground(colorWhite),
		
		// Container styles
		mainContainer: lipgloss.NewStyle().
			Padding(0, 1),
		
		sectionContainer: lipgloss.NewStyle().
			Padding(1).
			Margin(1, 0),
	}
}

// GetStatusSymbol returns the appropriate status symbol with styling
func (s UnifiedStyles) GetStatusSymbol(status WorkerStatus) string {
	switch status {
	case StatusActive:
		return s.statusActiveStyle.Render("ACTIVE")
	case StatusCompleted:
		return s.statusCompletedStyle.Render("DONE") 
	default:
		return s.statusCompletedStyle.Render("IDLE")
	}
}