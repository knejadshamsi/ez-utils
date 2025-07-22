package model

import (
	"fmt"
	"io"
	"ez-utils/src/create/tv/core"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VehicleTypeItem represents a vehicle type in the list
type VehicleTypeItem struct {
	vt *core.VehicleType
}

func (i VehicleTypeItem) Title() string {
	return i.vt.ID
}

func (i VehicleTypeItem) Description() string {
	return fmt.Sprintf("%s - Seats: %d, Standing: %d", i.vt.Description, i.vt.Capacity.Seats, i.vt.Capacity.Standing)
}

func (i VehicleTypeItem) FilterValue() string {
	return i.vt.ID
}

// VehicleItem represents a vehicle in the list
type VehicleItem struct {
	v *core.Vehicle
}

func (i VehicleItem) Title() string {
	return i.v.ID
}

func (i VehicleItem) Description() string {
	return fmt.Sprintf("Type: %s", i.v.TypeID)
}

func (i VehicleItem) FilterValue() string {
	return i.v.ID
}

// VehicleTypeDelegate for rendering vehicle type list items
type VehicleTypeDelegate struct{}

func NewVehicleTypeDelegate() VehicleTypeDelegate {
	return VehicleTypeDelegate{}
}

func (d VehicleTypeDelegate) Height() int {
	return 2
}

func (d VehicleTypeDelegate) Spacing() int {
	return 1
}

func (d VehicleTypeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d VehicleTypeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(VehicleTypeItem)
	if !ok {
		return
	}

	title := item.Title()
	desc := item.Description()

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true).Render(title)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(desc)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(title)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// VehicleDelegate for rendering vehicle list items
type VehicleDelegate struct{}

func NewVehicleDelegate() VehicleDelegate {
	return VehicleDelegate{}
}

func (d VehicleDelegate) Height() int {
	return 2
}

func (d VehicleDelegate) Spacing() int {
	return 1
}

func (d VehicleDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d VehicleDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(VehicleItem)
	if !ok {
		return
	}

	title := item.Title()
	desc := item.Description()

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true).Render(title)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(desc)
	} else {
		title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(title)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}