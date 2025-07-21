package gui

import (
	"fmt"
	"log"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SaveFile opens a save file dialog and returns the selected file path
func (a *App) SaveFile(defaultFileName string, fileType string) (string, error) {
	log.Printf("SaveFile called with defaultFileName: %s, fileType: %s", defaultFileName, fileType)
	
	var filters []runtime.FileFilter
	var defaultExt string
	
	switch fileType {
	case "population", "network", "pt":
		filters = []runtime.FileFilter{
			{
				DisplayName: "XML Files",
				Pattern:     "*.xml",
			},
			{
				DisplayName: "All Files", 
				Pattern:     "*.*",
			},
		}
		defaultExt = ".xml"
	default:
		filters = []runtime.FileFilter{
			{
				DisplayName: "All Files",
				Pattern:     "*.*",
			},
		}
	}
	
	// Ensure default filename has extension
	if !strings.Contains(defaultFileName, ".") && defaultExt != "" {
		defaultFileName += defaultExt
	}
	
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("Export %s File", strings.Title(fileType)),
		DefaultFilename: defaultFileName,
		Filters:         filters,
		ShowHiddenFiles: false,
		CanCreateDirectories: true,
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}
	
	// User cancelled
	if filePath == "" {
		return "", nil
	}
	
	return filePath, nil
}

// ExportPopulationFile exports population data to XML file
func (a *App) ExportPopulationFile(tableName string, outputPath string) error {
	log.Printf("ExportPopulationFile called with tableName: %s, outputPath: %s", tableName, outputPath)
	
	if tableName == "" {
		return fmt.Errorf("table name is required")
	}
	
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	
	// Create population exporter
	exporter := &PopulationExporter{
		db:  a.db.conn,
		app: a,
	}
	
	// Export the population data
	err := exporter.ExportPopulationFile(tableName, outputPath)
	if err != nil {
		log.Printf("Error exporting population: %v", err)
		return fmt.Errorf("failed to export population: %w", err)
	}
	
	log.Printf("Successfully exported population to: %s", outputPath)
	return nil
}

// GetExportInfo returns information about what can be exported
func (a *App) GetExportInfo(fileEditMode string) map[string]interface{} {
	info := map[string]interface{}{
		"canExport": false,
		"exportTypes": []string{},
	}
	
	switch fileEditMode {
	case "POPULATION":
		info["canExport"] = true
		info["exportTypes"] = []string{"population"}
		info["defaultFileName"] = "population_export"
	case "NETWORK":
		info["canExport"] = true
		info["exportTypes"] = []string{"network"}
		info["defaultFileName"] = "network_export"
	case "PT":
		info["canExport"] = true
		info["exportTypes"] = []string{"pt"}
		info["defaultFileName"] = "pt_export"
	}
	
	return info
}