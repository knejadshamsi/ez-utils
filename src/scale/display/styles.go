package display

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorGreen      = lipgloss.Color("#00FF00")
	colorLightGreen = lipgloss.Color("#98FB98")
	colorDarkGreen  = lipgloss.Color("#004000")
	colorMagenta    = lipgloss.Color("#FF00FF")
	colorYellow     = lipgloss.Color("#FFFF00")
	colorOrange     = lipgloss.Color("#FFA500")
	colorLightBlue  = lipgloss.Color("#00BBFF")
	colorPurple     = lipgloss.Color("#BB00FF")
	colorGray       = lipgloss.Color("#999999")
	colorDarkGray   = lipgloss.Color("#303030")
	colorWhite      = lipgloss.Color("#FFFFFF")
)

// StyleDefinitions
type StyleDefinitions struct {
	// Title section
	titleEZ   lipgloss.Style
	titleText lipgloss.Style
	timeInfo  lipgloss.Style

	// Steps styling
	completedTask        lipgloss.Style
	completedDuration    lipgloss.Style
	currentlyPrefix      lipgloss.Style
	currentStepContainer lipgloss.Style
	upcomingTask         lipgloss.Style
	counterValue         lipgloss.Style

	// Status indicators
	doneIndicator   lipgloss.Style
	brightIndicator lipgloss.Style
	dimIndicator    lipgloss.Style

	// System info section
	systemLabel     lipgloss.Style
	footerContainer lipgloss.Style
	cpuInfo         lipgloss.Style
	ramInfo         lipgloss.Style

	// Components
	spinner         lipgloss.Style
	processComplete lipgloss.Style
}

// Defines complete UI styling, padding and borders
var lipglossStyle = StyleDefinitions{
	// Title section - simplified
	titleEZ: lipgloss.NewStyle().
		Foreground(colorOrange).
		Bold(true),

	titleText: lipgloss.NewStyle().
		Foreground(colorWhite).
		Bold(true),

	timeInfo: lipgloss.NewStyle().
		Foreground(colorOrange),

	// Steps styling - more consistent
	completedTask: lipgloss.NewStyle().
		Background(colorDarkGreen).
		Foreground(colorLightGreen).
		Padding(0, 1).
		MarginLeft(1),

	completedDuration: lipgloss.NewStyle().
		Background(colorDarkGray).
		Foreground(colorGray).
		Padding(0, 1),

	currentlyPrefix: lipgloss.NewStyle().
		Background(colorMagenta).
		Foreground(colorWhite).
		Bold(true).
		Padding(0, 1),

	currentStepContainer: lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorMagenta).
		Width(80),

	upcomingTask: lipgloss.NewStyle().
		Foreground(colorGray).
		MarginLeft(1),

	counterValue: lipgloss.NewStyle().
		Foreground(colorYellow),

	// Status indicators - simplified
	doneIndicator: lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true),

	brightIndicator: lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true),

	dimIndicator: lipgloss.NewStyle().
		Foreground(colorDarkGreen),

	// System info section - consistent styling
	systemLabel: lipgloss.NewStyle().
		Foreground(colorWhite).
		Bold(true).
		PaddingLeft(1),

	footerContainer: lipgloss.NewStyle().
		PaddingBottom(1),

	cpuInfo: lipgloss.NewStyle().
		Foreground(colorLightBlue),

	ramInfo: lipgloss.NewStyle().
		Foreground(colorPurple),

	// Components
	spinner: lipgloss.NewStyle().
		Foreground(colorMagenta),

	processComplete: lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGreen).
		Padding(1, 4).
		Align(lipgloss.Center).
		Width(80).
		MarginTop(1),
}
