package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete UI
func (m Model) View() string {
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
	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s", title, countsLine, columns, help)
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
