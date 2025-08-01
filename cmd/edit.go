package cmd

import (
	"ez-utils/gui"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// RunEdit executes the edit command logic, which will launch the GUI.
func RunEdit() {
	// Set up logging to file
	logFile, err := os.OpenFile("ez-utils.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Warning: Could not open log file: %v\n", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	var startupFile string
	var editMode string
	args := os.Args[1:] // Ignore the program name

	log.Printf("RunEdit called with args: %v", args)
	log.Printf("Full os.Args: %v", os.Args)

	if len(args) >= 3 && args[0] == "edit" {
		switch args[1] {
		case "population":
			editMode = "population"
		case "network":
			editMode = "network"
		case "pt", "transit", "public transportation":
			editMode = "public transportation"
		default:
			errMsg := fmt.Sprintf("Error: Unknown edit mode: %s", args[1])
			log.Printf(errMsg)
			fmt.Println(errMsg)
			os.Exit(1)
		}
		startupFile = args[2]
		
		log.Printf("Original file path: %s", startupFile)
		
		// Convert relative path to absolute path
		if startupFile != "" {
			absPath, err := filepath.Abs(startupFile)
			if err == nil {
				startupFile = absPath
				log.Printf("Resolved to absolute path: %s", startupFile)
				fmt.Printf("Resolved file path: %s\n", startupFile)
				
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
	}

	log.Printf("Launching GUI with file: %s, mode: %s", startupFile, editMode)
	gui.Run(startupFile, editMode)
}