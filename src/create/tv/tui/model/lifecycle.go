package model

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"
	"ez-utils/src/create/tv/tui/modals"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

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

	case tea.KeyMsg:
		if m.modal != ModalNone && m.currentForm != nil {
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
			if m.modal == ModalNone {
				switch m.focus {
				case FocusTypesList:
					// Get the currently selected vehicle type
					if selectedItem := m.typesList.SelectedItem(); selectedItem != nil {
						if typeItem, ok := selectedItem.(VehicleTypeItem); ok {
							// Remove the vehicle type
							m.manager.RemoveVehicleType(typeItem.vt.ID)
							m.manager.HasChanges = true
							// Update the vehicles list to reflect the removal
							m.updateVehiclesList()
							// Display status message
							m.SetStatusMsg("Deleted vehicle type: " + typeItem.vt.ID)
						}
					}
				case FocusVehiclesList:
					// Get the currently selected vehicle
					if selectedItem := m.vehiclesList.SelectedItem(); selectedItem != nil {
						if vehicleItem, ok := selectedItem.(VehicleItem); ok {
							// Remove the vehicle
							m.manager.RemoveVehicle(vehicleItem.v.ID)
							m.manager.HasChanges = true
							// Update the vehicles list to reflect the removal
							m.updateVehiclesList()
							// Display status message
							m.SetStatusMsg("Deleted vehicle: " + vehicleItem.v.ID)
						}
					}
				}
			}
		case "s":
			if m.modal == ModalNone {
				m.SetErrorMsg("Save not implemented yet")
			}
		case "esc":
			if m.modal != ModalNone {
				m.modal = ModalNone
				m.currentForm = nil
				m.errorMsg = ""
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
