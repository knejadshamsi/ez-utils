package cmd

import (
	"ez-utils/gui"
	"os"
)

// RunEdit executes the edit command logic, which will launch the GUI.
func RunEdit() {
	var startupFile string
	args := os.Args[1:] // Ignore the program name

	if len(args) >= 3 && args[0] == "edit" && args[1] == "population" {
		startupFile = args[2]
	}

	gui.Run(startupFile)
}