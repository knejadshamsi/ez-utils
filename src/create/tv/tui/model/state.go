package model

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Focus states for navigation
type FocusState int

const (
	FocusTypesList FocusState = iota
	FocusVehiclesList
)

// Modal states
type ModalState int

const (
	ModalNone ModalState = iota
	ModalAddType
	ModalEditType
	ModalAddVehicles
	ModalImportSchedule
	ModalImportFile
	ModalExport
	ModalDeleteConfirmation
)

// Styles for UI components
type Styles struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Header       lipgloss.Style
	List         lipgloss.Style
	StatusBar    lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
	Modal        lipgloss.Style
	ModalTitle   lipgloss.Style
	ModalContent lipgloss.Style
	Button       lipgloss.Style
	ActiveButton lipgloss.Style
}

// NewStyles creates the default style set
func NewStyles() *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginBottom(1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#374151")).
			Padding(0, 1),
		List: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(1),
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1).
			Background(lipgloss.Color("#1F2937")),
		ModalTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1),
		ModalContent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")),
		Button: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(0, 2).
			Foreground(lipgloss.Color("#FFFFFF")),
		ActiveButton: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 2).
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true),
	}
}

// Model represents the application state
type Model struct {
	// Core data
	manager *core.VehicleManager

	// UI state
	width  int
	height int
	focus  FocusState
	modal  ModalState

	// Lists
	typesList    list.Model
	vehiclesList list.Model

	// Forms and modals
	currentForm *form.Form

	// Messages
	statusMsg string
	errorMsg  string

	// Styling
	styles *Styles

	// Selected vehicle type (for filtering vehicles)
	selectedType *core.VehicleType

	// Delete confirmation context
	itemToDeleteID   string
	itemToDeleteType FocusState
}

// NewModel creates a new model instance
func NewModel(manager *core.VehicleManager) Model {
	styles := NewStyles()

	// Create lists
	typesList := list.New([]list.Item{}, NewVehicleTypeDelegate(), 0, 0)
	typesList.Title = "Vehicle Types"
	typesList.SetShowStatusBar(false)
	typesList.SetFilteringEnabled(false)

	vehiclesList := list.New([]list.Item{}, NewVehicleDelegate(), 0, 0)
	vehiclesList.Title = "Vehicles"
	vehiclesList.SetShowStatusBar(false)
	vehiclesList.SetFilteringEnabled(false)

	model := Model{
		manager:      manager,
		focus:        FocusTypesList,
		modal:        ModalNone,
		typesList:    typesList,
		vehiclesList: vehiclesList,
		styles:       styles,
		selectedType: nil,
	}

	model.updateLists()
	return model
}
