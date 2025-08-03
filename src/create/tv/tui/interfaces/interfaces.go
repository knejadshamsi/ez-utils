package interfaces

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"
	"github.com/charmbracelet/bubbles/list"
)

// ModelInterface defines what modals need from the main model
type ModelInterface interface {
	GetManager() *core.VehicleManager
	SetModal(modal int)
	SetStatusMsg(msg string)
	SetErrorMsg(msg string)
	SetCurrentForm(f *form.Form)
	GetTypesList() *list.Model
	GetSelectedVehicleType() *core.VehicleType
	GetStyles() *form.Styles
}
