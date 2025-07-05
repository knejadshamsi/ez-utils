//go:build wails && production

package main

import (
	"ez-utils/cmd"
	"ez-utils/gui"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// isProductionBuild returns true when the 'production' build tag is active.
// This is defined here for the production build.
func isProductionBuild() bool {
	return true
}

func main() {
	// During binding generation, the 'production' tag is NOT set, so we need
	// a version of this file without the production tag that provides a clean
	// path to wails.Run(). That's handled by wails_dev.go.
	
	// This file handles the actual production runtime.
	
	// Set up logging to file
	logFile, err := os.OpenFile("ez-utils.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}
	
	args := os.Args[1:] // Ignore the program name
	log.Printf("Production build main called with args: %v", args)
	
	// Route based on command
	if len(args) > 0 {
		switch args[0] {
		case "scale":
			// Run CLI scale command
			log.Printf("Running scale command")
			cmd.RunScale()
			return
			
		case "create":
			// Run CLI create command
			log.Printf("Running create command")
			cmd.RunCreate()
			return
			
		case "edit":
			// Edit command requires proper subcommand
			if len(args) < 2 {
				fmt.Println("Error: edit command requires a subcommand (e.g., 'edit population <file>')")
				log.Printf("Invalid edit command - missing subcommand")
				os.Exit(1)
			}
			
			// Check for valid edit subcommands
			switch args[1] {
			case "population", "network", "pt", "transit", "public transportation":
				if len(args) < 3 {
					fmt.Printf("Error: edit %s requires a file path\n", args[1])
					log.Printf("Invalid edit %s command - missing file path", args[1])
					os.Exit(1)
				}
				
				var editMode string
				switch args[1] {
				case "population":
					editMode = "population"
				case "network":
					editMode = "network"
				case "pt", "transit", "public transportation":
					editMode = "public transportation"
				}
				
				startupFile := args[2]
				
				log.Printf("Original file path: %s", startupFile)
				
				// Convert relative path to absolute path
				if startupFile != "" {
					absPath, err := filepath.Abs(startupFile)
					if err == nil {
						startupFile = absPath
						log.Printf("Resolved to absolute path: %s", startupFile)
						
						// Check if file exists
						if _, err := os.Stat(startupFile); os.IsNotExist(err) {
							errMsg := fmt.Sprintf("Error: File does not exist: %s", startupFile)
							log.Printf(errMsg)
							fmt.Println(errMsg)
							os.Exit(1)
						} else if err != nil {
							errMsg := fmt.Sprintf("Error: Cannot access file %s: %v", startupFile, err)
							log.Printf(errMsg)
							fmt.Println(errMsg)
							os.Exit(1)
						} else {
							log.Printf("File exists and is accessible: %s", startupFile)
						}
					} else {
						errMsg := fmt.Sprintf("Error resolving path %s: %v", startupFile, err)
						log.Printf(errMsg)
						fmt.Println(errMsg)
						os.Exit(1)
					}
				}
				
				log.Printf("Launching GUI with file: %s, mode: %s", startupFile, editMode)
				gui.Run(startupFile, editMode)
				return
				
			default:
				fmt.Printf("Error: Unknown edit subcommand '%s'. Available: population, network, pt/transit/public transportation\n", args[1])
				log.Printf("Invalid edit subcommand: %s", args[1])
				os.Exit(1)
			}
		}
	}
	
	// No arguments or unrecognized command - launch GUI
	log.Printf("Launching GUI with no arguments")
	gui.Run("", "")
}