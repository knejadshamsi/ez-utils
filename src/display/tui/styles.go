package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorGreen      = lipgloss.Color("#00FF00") // Made public
	ColorLightGreen = lipgloss.Color("#98FB98") // Made public
	ColorDarkGreen  = lipgloss.Color("#004000") // Made public
	ColorMagenta    = lipgloss.Color("#FF00FF") // Made public
	ColorYellow     = lipgloss.Color("#FFFF00") // Made public
	ColorOrange     = lipgloss.Color("#FFA500") // Made public
	ColorLightBlue  = lipgloss.Color("#00BBFF") // Made public
	ColorPurple     = lipgloss.Color("#BB00FF") // Made public
	ColorLightPurple = lipgloss.Color("#D8BFD8") // Made public
	ColorGray       = lipgloss.Color("#999999") // Made public
	ColorDarkGray   = lipgloss.Color("#303030") // Made public
	ColorWhite      = lipgloss.Color("#FFFFFF") // Made public
)

// StyleDefinitions
type StyleDefinitions struct {
	// Title section
	TitleEZ   lipgloss.Style // Made public
	TitleText lipgloss.Style // Made public
	TimeInfo  lipgloss.Style // Made public

	// Steps styling
	CompletedTask        lipgloss.Style // Made public
	CompletedDuration    lipgloss.Style // Made public
	CurrentlyPrefix      lipgloss.Style // Made public
	CurrentStepContainer lipgloss.Style // Made public
	UpcomingTask         lipgloss.Style // Made public
	CounterValue         lipgloss.Style // Made public

	// Status indicators
	DoneIndicator   lipgloss.Style // Made public
	BrightIndicator lipgloss.Style // Made public
	DimIndicator    lipgloss.Style // Made public

	// System info section
	SystemLabel     lipgloss.Style // Made public
	FooterContainer lipgloss.Style // Made public
	CPUInfo         lipgloss.Style // Made public
	RAMInfo         lipgloss.Style // Made public
	DiskInfo        lipgloss.Style // Made public

	// Components
	Spinner         lipgloss.Style // Made public
	ProcessComplete lipgloss.Style // Made public
}

// Defines complete UI styling, padding and borders
var LipglossStyle = StyleDefinitions{ // Made public
	// Title section - simplified
	TitleEZ: lipgloss.NewStyle().
		Foreground(ColorOrange).
		Bold(true),

	TitleText: lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true),

	TimeInfo: lipgloss.NewStyle().
		Foreground(ColorOrange),

	// Steps styling - more consistent
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

	// Status indicators - simplified
	DoneIndicator: lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true),

	BrightIndicator: lipgloss.NewStyle().
		Foreground(ColorGreen).
		Bold(true),

	DimIndicator: lipgloss.NewStyle().
		Foreground(ColorDarkGreen),

	// System info section - consistent styling
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
