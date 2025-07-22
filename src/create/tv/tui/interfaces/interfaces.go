package interfaces

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"
)

// ModelInterface defines what modals need from the main model
type ModelInterface interface {
	GetManager() *core.VehicleManager
	SetModal(modal int)
	SetStatusMsg(msg string)
	SetErrorMsg(msg string)
	SetCurrentForm(f *form.Form)
	GetTypesList() any // Will be properly typed later
	GetStyles() *form.Styles
}
