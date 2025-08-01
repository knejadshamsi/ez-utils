package tui

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/model"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// RunBubbleTeaUI starts the Bubbletea-based TUI
func RunBubbleTeaUI(manager *core.VehicleManager) error {
	// Create the model
	m := model.NewModel(manager)

	// Create the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
