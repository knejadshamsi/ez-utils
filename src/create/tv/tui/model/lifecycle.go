package model

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/io"
	"ez-utils/src/create/tv/tui/form"
	"ez-utils/src/create/tv/tui/modals"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// clearStatusMsg is sent when the status message timer expires
type clearStatusMsg struct{}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateListSizes()

	case clearStatusMsg:
		// Clear the status message when timer expires
		m.statusMsg = ""
		return m, nil

	case tea.KeyMsg:
		// Handle delete confirmation modal input first
		if m.modal == ModalDeleteConfirmation {
			switch msg.String() {
			case "y", "Y":
				// Perform the actual deletion
				var successMsg string

				switch m.itemToDeleteType {
				case FocusTypesList:
					// Delete the vehicle type
					m.manager.RemoveVehicleType(m.itemToDeleteID)
					m.manager.HasChanges = true
					successMsg = fmt.Sprintf("Vehicle type '%s' deleted.", m.itemToDeleteID)
					// Update both lists since deleting a type affects both
					m.updateLists()

				case FocusVehiclesList:
					// Delete the vehicle
					m.manager.RemoveVehicle(m.itemToDeleteID)
					m.manager.HasChanges = true
					successMsg = fmt.Sprintf("Vehicle '%s' deleted.", m.itemToDeleteID)
					// Update only vehicles list
					m.updateVehiclesList()
				}

				// Reset modal state and show success message
				m.modal = ModalNone
				m.itemToDeleteID = ""
				m.itemToDeleteType = 0
				if successMsg != "" {
					cmd = m.SetTimedStatusMessage(successMsg, 3*time.Second)
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)

			case "n", "N", "esc":
				// Cancel deletion
				m.modal = ModalNone
				m.itemToDeleteID = ""
				m.itemToDeleteType = 0
				return m, nil
			}
			return m, nil
		}

		if m.modal != ModalNone && m.currentForm != nil {
			// Check for ESC first before handling form input
			if msg.String() == "esc" {
				m.modal = ModalNone
				m.currentForm = nil
				m.errorMsg = ""
				return m, nil
			}
			
			// Handle form input
			*m.currentForm, cmd = m.currentForm.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.modal == ModalNone {
				return m, tea.Quit
			}
		case "a":
			if m.modal == ModalNone && m.focus == FocusTypesList {
				modals.ShowAddTypeModal(&m)
			}
		case "d":
			// Check if we're in the delete confirmation modal
			if m.modal == ModalDeleteConfirmation {
				return m, nil // Don't process 'd' key when modal is open
			}

			if m.modal == ModalNone {
				switch m.focus {
				case FocusTypesList:
					// Get the currently selected vehicle type
					if selectedItem := m.typesList.SelectedItem(); selectedItem != nil {
						if typeItem, ok := selectedItem.(VehicleTypeItem); ok {
							// Store the context for deletion
							m.itemToDeleteID = typeItem.vt.ID
							m.itemToDeleteType = FocusTypesList
							m.modal = ModalDeleteConfirmation
						}
					}
				case FocusVehiclesList:
					// Get the currently selected vehicle
					if selectedItem := m.vehiclesList.SelectedItem(); selectedItem != nil {
						if vehicleItem, ok := selectedItem.(VehicleItem); ok {
							// Store the context for deletion
							m.itemToDeleteID = vehicleItem.v.ID
							m.itemToDeleteType = FocusVehiclesList
							m.modal = ModalDeleteConfirmation
						}
					}
				}
			}
		case "s":
			if m.modal == ModalNone {
				// Call the SaveToFile function
				if err := io.SaveToFile(m.manager, m.manager.VehiclesFile); err != nil {
					// Display error message with details
					m.SetErrorMsg(fmt.Sprintf("Failed to save: %v", err))
				} else {
					// Update manager state and show success message
					m.manager.HasChanges = false
					cmd = m.SetTimedStatusMessage(fmt.Sprintf("✅ Saved changes to %s", m.manager.VehiclesFile), 5*time.Second)
					cmds = append(cmds, cmd)
				}
			}
		case "esc":
			if m.modal != ModalNone {
				m.modal = ModalNone
				m.currentForm = nil
				m.errorMsg = ""
				// Clear delete confirmation context
				m.itemToDeleteID = ""
				m.itemToDeleteType = 0
			}
		case "right":
			if m.modal == ModalNone && m.focus == FocusTypesList {
				m.focus = FocusVehiclesList
			}
		case "left":
			if m.modal == ModalNone && m.focus == FocusVehiclesList {
				m.focus = FocusTypesList
			}
		case "enter":
			if m.modal == ModalNone {
				switch m.focus {
				case FocusTypesList:
					// Edit selected vehicle type
					if selectedItem := m.typesList.SelectedItem(); selectedItem != nil {
						if typeItem, ok := selectedItem.(VehicleTypeItem); ok {
							modals.ShowEditTypeModal(&m, typeItem.vt)
						}
					}
				case FocusVehiclesList:
					// Edit selected vehicle
					if selectedItem := m.vehiclesList.SelectedItem(); selectedItem != nil {
						if vehicleItem, ok := selectedItem.(VehicleItem); ok {
							modals.ShowEditVehicleModal(&m, vehicleItem.v)
						}
					}
				}
			}
		}

	case form.FormSubmittedMsg:
		// Form was successfully submitted
		m.modal = ModalNone
		m.currentForm = nil
		m.updateLists()

	case form.ModalCloseMsg:
		// Modal should be closed
		m.modal = ModalNone
		m.currentForm = nil
	}

	// Update lists if not in modal
	if m.modal == ModalNone {
		switch m.focus {
		case FocusTypesList:
			prevIndex := m.typesList.Index()
			m.typesList, cmd = m.typesList.Update(msg)
			cmds = append(cmds, cmd)
			// If selection changed, update vehicles list
			if prevIndex != m.typesList.Index() {
				m.updateVehiclesList()
			}
		case FocusVehiclesList:
			m.vehiclesList, cmd = m.vehiclesList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateListSizes updates the size of lists based on window size
func (m *Model) updateListSizes() {
	listHeight := (m.height - 6) / 2 // Split between two lists, account for headers/status
	listWidth := (m.width - 4) / 3   // Three columns: types, vehicles, menu

	m.typesList.SetSize(listWidth, listHeight)
	m.vehiclesList.SetSize(listWidth, listHeight)
}

// updateLists refreshes the content of both lists
func (m *Model) updateLists() {
	// Update vehicle types list
	typeItems := make([]list.Item, len(m.manager.VehicleTypes))
	for i, vt := range m.manager.VehicleTypes {
		typeItems[i] = VehicleTypeItem{vt: vt}
	}
	m.typesList.SetItems(typeItems)

	// Update vehicles list based on selected type
	m.updateVehiclesList()
}

// updateVehiclesList updates the vehicles list based on the selected vehicle type
func (m *Model) updateVehiclesList() {
	var vehicleItems []list.Item

	// Get selected vehicle type
	if selectedItem := m.typesList.SelectedItem(); selectedItem != nil {
		if typeItem, ok := selectedItem.(VehicleTypeItem); ok {
			// Filter vehicles by selected type
			for _, v := range m.manager.Vehicles {
				if v.TypeID == typeItem.vt.ID {
					vehicleItems = append(vehicleItems, VehicleItem{v: v})
				}
			}
		}
	} else {
		// Show all vehicles if no type selected
		for _, v := range m.manager.Vehicles {
			vehicleItems = append(vehicleItems, VehicleItem{v: v})
		}
	}

	m.vehiclesList.SetItems(vehicleItems)
}

// Implement modals.ModelInterface
func (m *Model) GetManager() *core.VehicleManager {
	return m.manager
}

func (m *Model) SetModal(modal int) {
	m.modal = ModalState(modal)
}

func (m *Model) SetStatusMsg(msg string) {
	m.statusMsg = msg
	m.errorMsg = "" // Clear error when setting status
}

// SetTimedStatusMessage sets a status message that will automatically clear after the specified duration
func (m *Model) SetTimedStatusMessage(msg string, duration time.Duration) tea.Cmd {
	m.statusMsg = msg
	m.errorMsg = "" // Clear error when setting status
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) SetErrorMsg(msg string) {
	m.errorMsg = msg
	m.statusMsg = "" // Clear status when setting error
}

func (m *Model) SetCurrentForm(f *form.Form) {
	m.currentForm = f
}

func (m *Model) GetTypesList() any {
	return &m.typesList
}

func (m *Model) GetStyles() *form.Styles {
	// Convert model styles to form styles
	return &form.Styles{}
}
