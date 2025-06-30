//go:build wails

package main

import (
	"ez-utils/gui"
)

func main() {
	// This is the dedicated entry point for the Wails GUI application.
	// It directly calls the Run function from the gui package,
	// which is exactly what the `wails build` process needs to see.
	// The empty string argument indicates that no specific file
	// is being opened at startup.
	gui.Run("")
}