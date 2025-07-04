//go:build wails && !production

package main

import (
	"ez-utils/gui"
	"log"
	"os"
)

// isProductionBuild returns false when the 'production' build tag is not active.
// This is used during binding generation and development.
func isProductionBuild() bool {
	return false
}

func main() {
	// This entry point is used for:
	// 1. Binding generation (during wails build)
	// 2. Development mode (wails dev)
	//
	// In both cases, we need a CLEAN path to wails.Run() without any
	// complex logic that could fail during binding generation.
	
	// Optional: Set up logging to track binding generation
	logFile, err := os.OpenFile("ez-utils-binding-gen.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}
	
	log.Printf("Binding generation/dev mode - args: %v", os.Args)
	log.Printf("isProductionBuild: %v", isProductionBuild())
	
	// Simple, direct path to gui.Run() for binding generation
	// No argument parsing, no file checks, no early exits
	gui.Run("", "")
}