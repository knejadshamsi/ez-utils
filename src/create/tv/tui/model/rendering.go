package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete UI
func (m Model) View() string {
	if m.modal == ModalDeleteConfirmation {
		return m.renderDeleteConfirmationModal()
	}
	if m.modal != ModalNone && m.currentForm != nil {
		return m.renderModalView()
	}
	return m.renderMainView()
}

// renderMainView renders the main two-column interface
func (m Model) renderMainView() string {
	// Check if window size is initialized
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Calculate dimensions
	totalWidth := m.width
	if totalWidth < 80 {
		totalWidth = 80
	}
	// Calculate column widths - each column gets half the space minus the gap
	// Add safety margin to ensure we don't exceed terminal width
	gap := 2
	safetyMargin := 4 // Extra margin to prevent cutoff
	columnWidth := (totalWidth - gap - safetyMargin) / 2
	contentHeight := m.height - 8 // Space for title, counts, and help

	// Title section
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Width(totalWidth).
		Align(lipgloss.Center)

	title := titleStyle.Render("Transit Vehicle Manager")

	// Vehicle counts
	countsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Width(totalWidth).
		Align(lipgloss.Center)

	counts := fmt.Sprintf("Vehicle Types: %d | Vehicles: %d",
		len(m.manager.VehicleTypes), len(m.manager.Vehicles))
	countsLine := countsStyle.Render(counts)

	// Status/Error message section - placed right after counts
	var statusSection string
	if m.statusMsg != "" {
		statusStyle := m.styles.Success.
			Width(totalWidth).
			Align(lipgloss.Center).
			MarginTop(1).
			MarginBottom(1)
		statusSection = "\n" + statusStyle.Render(m.statusMsg)
	} else if m.errorMsg != "" {
		errorStyle := m.styles.Error.
			Width(totalWidth).
			Align(lipgloss.Center).
			MarginTop(1).
			MarginBottom(1)
		statusSection = "\n" + errorStyle.Render(m.errorMsg)
	}

	// Adjust content height if status message is shown
	if statusSection != "" {
		contentHeight = contentHeight - 2 // Account for status message space
	}

	// Create columns
	leftColumn := m.renderTypesList(columnWidth, contentHeight)
	rightColumn := m.renderVehiclesList(columnWidth, contentHeight)

	// Use lipgloss to join columns properly
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		"  ", // Gap
		rightColumn,
	)

	// Help section
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		Width(totalWidth).
		Align(lipgloss.Center)

	help := helpStyle.Render("↑↓ Navigate | ←→ Switch | Enter Edit | A Add | D Delete | S Save | Q Quit")

	// Combine all sections
	result := fmt.Sprintf("%s\n%s%s\n\n%s\n\n%s",
		title,
		countsLine,
		statusSection, // Status message right after counts
		columns,
		help)
	return result
}

// renderTypesList renders the vehicle types list
func (m Model) renderTypesList(width, height int) string {
	// Create border box
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Width(width).
		Height(height)

	if m.focus == FocusTypesList {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Width(width - 2). // Account for border
		Align(lipgloss.Center)

	if m.focus == FocusTypesList {
		headerStyle = headerStyle.Background(lipgloss.Color("#A78BFA"))
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("Vehicle Types") + "\n")

	// List items
	items := m.typesList.Items()
	selectedIndex := m.typesList.Index()

	availableHeight := height - 3 // Header + borders (no divider anymore)
	visibleStart := 0
	visibleEnd := len(items)

	// Calculate visible range for scrolling
	if len(items) > availableHeight {
		if selectedIndex >= availableHeight/2 {
			visibleStart = selectedIndex - availableHeight/2
		}
		visibleEnd = visibleStart + availableHeight
		if visibleEnd > len(items) {
			visibleEnd = len(items)
			visibleStart = visibleEnd - availableHeight
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		if typeItem, ok := items[i].(VehicleTypeItem); ok {
			id := typeItem.vt.ID
			if len(id) > 15 {
				id = id[:12] + "..."
			}

			line := fmt.Sprintf(" %-15s Seats:%3d Stand:%3d",
				id,
				typeItem.vt.Capacity.Seats,
				typeItem.vt.Capacity.Standing)

			if i == selectedIndex && m.focus == FocusTypesList {
				lineStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("#7C3AED")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Width(width - 2)
				content.WriteString(lineStyle.Render("▶"+line[1:]) + "\n")
			} else {
				content.WriteString(line + "\n")
			}
		}
	}

	if len(items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Width(width - 2).
			Align(lipgloss.Center)
		content.WriteString(emptyStyle.Render("No vehicle types") + "\n")
	}

	// Fill remaining space
	linesWritten := visibleEnd - visibleStart + 3 // +3 for header, divider, and potential empty message
	for i := linesWritten; i < height-2; i++ {
		content.WriteString("\n")
	}

	return borderStyle.Render(strings.TrimRight(content.String(), "\n"))
}

// renderVehiclesList renders the vehicles list
func (m Model) renderVehiclesList(width, height int) string {
	// Create border box
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Width(width).
		Height(height)

	if m.focus == FocusVehiclesList {
		borderStyle = borderStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Width(width - 2). // Account for border
		Align(lipgloss.Center)

	if m.focus == FocusVehiclesList {
		headerStyle = headerStyle.Background(lipgloss.Color("#A78BFA"))
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("Vehicles") + "\n")

	// Filter vehicles
	var filteredVehicles []VehicleItem
	selectedTypeItem := m.typesList.SelectedItem()

	if selectedTypeItem != nil {
		if typeItem, ok := selectedTypeItem.(VehicleTypeItem); ok {
			for _, v := range m.manager.Vehicles {
				if v.TypeID == typeItem.vt.ID {
					filteredVehicles = append(filteredVehicles, VehicleItem{v: v})
				}
			}
		}
	} else {
		for _, v := range m.manager.Vehicles {
			filteredVehicles = append(filteredVehicles, VehicleItem{v: v})
		}
	}

	// List items
	selectedIndex := m.vehiclesList.Index()
	availableHeight := height - 3 // Header + borders (no divider anymore)
	visibleStart := 0
	visibleEnd := len(filteredVehicles)

	// Calculate visible range for scrolling
	if len(filteredVehicles) > availableHeight {
		if selectedIndex >= availableHeight/2 {
			visibleStart = selectedIndex - availableHeight/2
		}
		visibleEnd = visibleStart + availableHeight
		if visibleEnd > len(filteredVehicles) {
			visibleEnd = len(filteredVehicles)
			visibleStart = visibleEnd - availableHeight
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		vItem := filteredVehicles[i]
		vID := vItem.v.ID
		if len(vID) > width-6 {
			vID = vID[:width-9] + "..."
		}

		line := fmt.Sprintf(" %s", vID)

		if i == selectedIndex && m.focus == FocusVehiclesList {
			lineStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("#7C3AED")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Width(width - 2)
			content.WriteString(lineStyle.Render("▶"+line[1:]) + "\n")
		} else {
			content.WriteString(line + "\n")
		}
	}

	if len(filteredVehicles) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Width(width - 2).
			Align(lipgloss.Center)
		content.WriteString(emptyStyle.Render("No vehicles") + "\n")
	}

	// Fill remaining space
	linesWritten := visibleEnd - visibleStart + 3
	for i := linesWritten; i < height-2; i++ {
		content.WriteString("\n")
	}

	return borderStyle.Render(strings.TrimRight(content.String(), "\n"))
}

// renderModalView renders modal overlays
func (m Model) renderModalView() string {
	// Dim background
	background := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4B5563")).
		Render(m.renderMainView())

	// Modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A78BFA")).
		Background(lipgloss.Color("#1F2937")).
		Padding(1, 2).
		MarginTop(5).
		MarginLeft(10)

	modalContent := m.currentForm.View()
	styledModal := modalStyle.Render(modalContent)

	// Overlay modal on dimmed background
	lines := strings.Split(background, "\n")
	modalLines := strings.Split(styledModal, "\n")

	// Simple overlay (replace middle lines with modal)
	startLine := 5
	for i, mLine := range modalLines {
		if startLine+i < len(lines) {
			lines[startLine+i] = mLine
		}
	}

	return strings.Join(lines, "\n")
}

// renderDeleteConfirmationModal renders the delete confirmation modal
func (m Model) renderDeleteConfirmationModal() string {
	// First render the normal main view
	mainView := m.renderMainView()

	// Create the delete confirmation modal content
	modalWidth := 60

	// Build the modal content
	var modalContent strings.Builder

	// Title
	title := "DELETE CONFIRMATION"
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	modalContent.WriteString(titleStyle.Render(title) + "\n\n")

	// Add specific warning for vehicle types
	if m.itemToDeleteType == FocusTypesList {
		// Count vehicles of this type
		vehicleCount := 0
		for _, vehicle := range m.manager.Vehicles {
			if vehicle.TypeID == m.itemToDeleteID {
				vehicleCount++
			}
		}

		if vehicleCount > 0 {
			warningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true).
				Width(modalWidth - 4).
				Align(lipgloss.Center)
			warningMsg := fmt.Sprintf("⚠️  This will also delete all %d vehicles of this type.", vehicleCount)
			modalContent.WriteString(warningStyle.Render(warningMsg) + "\n\n")
		}
	}

	// Add the confirmation question
	questionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	confirmMsg := fmt.Sprintf("Are you sure you want to delete '%s'?", m.itemToDeleteID)
	modalContent.WriteString(questionStyle.Render(confirmMsg) + "\n\n")

	// Add instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	modalContent.WriteString(instructionStyle.Render("Press 'y' to confirm or 'n' to cancel"))

	// Style the modal with red border for delete
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#EF4444")). // Red border for delete
		Padding(1, 2).
		Width(modalWidth).
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#FFFFFF"))

	styledModal := modalStyle.Render(modalContent.String())

	// Place modal in center
	centeredModal := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		styledModal,
	)

	// Split views into lines
	mainLines := strings.Split(mainView, "\n")
	modalLines := strings.Split(centeredModal, "\n")

	// Overlay modal on top of main view
	result := make([]string, len(mainLines))
	copy(result, mainLines)

	for i, modalLine := range modalLines {
		if i < len(result) && strings.TrimSpace(modalLine) != "" {
			result[i] = modalLine
		}
	}

	return strings.Join(result, "\n")
}
