package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorGreen       = lipgloss.Color("#00FF00")
	ColorLightGreen  = lipgloss.Color("#98FB98")
	ColorDarkGreen   = lipgloss.Color("#004000")
	ColorMagenta     = lipgloss.Color("#FF00FF")
	ColorYellow      = lipgloss.Color("#FFFF00")
	ColorOrange      = lipgloss.Color("#FFA500")
	ColorLightBlue   = lipgloss.Color("#00BBFF")
	ColorPurple      = lipgloss.Color("#BB00FF")
	ColorLightPurple = lipgloss.Color("#D8BFD8")
	ColorGray        = lipgloss.Color("#999999")
	ColorDarkGray    = lipgloss.Color("#303030")
	ColorWhite       = lipgloss.Color("#FFFFFF")
)

// StyleDefinitions
type StyleDefinitions struct {
	// Title section
	TitleEZ   lipgloss.Style
	TitleText lipgloss.Style
	TimeInfo  lipgloss.Style

	// Steps styling
	CompletedTask        lipgloss.Style
	CompletedDuration    lipgloss.Style
	CurrentlyPrefix      lipgloss.Style
	CurrentStepContainer lipgloss.Style
	UpcomingTask         lipgloss.Style
	CounterValue         lipgloss.Style

	// Status indicators
	DoneIndicator   lipgloss.Style
	BrightIndicator lipgloss.Style
	DimIndicator    lipgloss.Style

	// System info section
	SystemLabel     lipgloss.Style
	FooterContainer lipgloss.Style
	CPUInfo         lipgloss.Style
	RAMInfo         lipgloss.Style
	DiskInfo        lipgloss.Style

	// Components
	Spinner         lipgloss.Style
	ProcessComplete lipgloss.Style
}

// Styles - Complete UI styling to match existing appearance exactly
var Styles = StyleDefinitions{
	// Title section
	TitleEZ: lipgloss.NewStyle().
		Foreground(ColorOrange).
		Bold(true),

	TitleText: lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true),

	TimeInfo: lipgloss.NewStyle().
		Foreground(ColorOrange),

	// Steps styling
	CompletedTask: lipgloss.NewStyle().
		Background(ColorDarkGreen).
		Foreground(ColorLightGreen).
		Padding(0, 1).
		MarginLeft(1),

	CompletedDuration: lipgloss.NewStyle().
		Background(ColorDarkGray).
		Foreground(ColorGray).
		Padding(0, 1),

	CurrentlyPrefix: lipgloss.NewStyle().
		Background(ColorMagenta).
		Foreground(ColorWhite).
		Bold(true).
		Padding(0, 1),

	CurrentStepContainer: lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorMagenta).
		Width(80),

	UpcomingTask: lipgloss.NewStyle().
		Foreground(ColorGray).
		MarginLeft(1),

	CounterValue: lipgloss.NewStyle().
		Foreground(ColorYellow),

	// Status indicators
	DoneIndicator: lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true),

	BrightIndicator: lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true),

	DimIndicator: lipgloss.NewStyle().
		Foreground(ColorDarkGreen),

	// System info section
	SystemLabel: lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true).
		PaddingLeft(1),

	FooterContainer: lipgloss.NewStyle().
		PaddingBottom(1),

	CPUInfo: lipgloss.NewStyle().
		Foreground(ColorLightBlue),

	RAMInfo: lipgloss.NewStyle().
		Foreground(ColorPurple),

	DiskInfo: lipgloss.NewStyle().
		Foreground(ColorLightPurple),

	// Components
	Spinner: lipgloss.NewStyle().
		Foreground(ColorMagenta),

	ProcessComplete: lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Padding(1, 4).
		Align(lipgloss.Center).
		Width(80).
		MarginTop(1),
}
