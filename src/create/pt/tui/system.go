package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderSystem creates the system info section with CPU/RAM and quit hint
func RenderSystem(state *State) string {
	// Status indicator (blinking dot)
	statusIndicator := "●"
	if state.StatusBlink {
		statusIndicator = Styles.BrightIndicator.Render("●")
	} else {
		statusIndicator = Styles.DimIndicator.Render("●")
	}

	// System info parts
	var systemInfoParts []string
	systemLabel := Styles.SystemLabel.Render("SYSTEM: ")
	systemInfoParts = append(systemInfoParts, systemLabel)

	// CPU info
	if state.CPUUsage > 0 {
		cpuValue := fmt.Sprintf("%.1f%%", state.CPUUsage)
		systemInfoParts = append(systemInfoParts, Styles.CPUInfo.Render(fmt.Sprintf("[CPU: %s]", cpuValue)))
	}

	// RAM info
	if state.RAMUsage > 0 {
		ramValueGB := state.RAMUsage / 1024.0
		ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
		systemInfoParts = append(systemInfoParts, Styles.RAMInfo.Render(fmt.Sprintf("[RAM: %s]", ramValue)))
	}

	// Quit hint
	abortText := lipgloss.NewStyle().Foreground(ColorWhite).Render("[Press q to abort]")
	systemInfo := strings.Join(systemInfoParts, "")
	leftContent := fmt.Sprintf("%s%s", statusIndicator, systemInfo)

	// Use lipgloss to create a justified layout
	content := lipgloss.JoinHorizontal(lipgloss.Left, leftContent, " ", abortText)

	return Styles.FooterContainer.Render(content)
}
