package gui

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PopulationExporter handles exporting population data to XML
type PopulationExporter struct {
	db  *sql.DB
	app *App // Need app for context to emit events
}

const EXPORT_BATCH_SIZE = 1000 // Process 1000 persons at a time

// PopulationXML represents the root element of a population XML file
type PopulationXML struct {
	XMLName xml.Name `xml:"population"`
	Persons []string `xml:",innerxml"`
}

// ExportPopulationFile exports population data from database to XML file with batching
func (e *PopulationExporter) ExportPopulationFile(tableName string, outputPath string) error {
	log.Printf("Starting population export from table: %s to file: %s", tableName, outputPath)
	
	// First, get the total count
	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err := e.db.QueryRow(countQuery).Scan(&totalCount); err != nil {
		return fmt.Errorf("failed to count population: %w", err)
	}
	
	// Emit initial progress
	runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
		"current": 0,
		"total":   totalCount,
	})
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Write XML header
	_, err = file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write XML header: %w", err)
	}
	
	// Write DOCTYPE
	_, err = file.WriteString(`<!DOCTYPE population SYSTEM "http://www.matsim.org/files/dtd/population_v6.dtd">` + "\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write DOCTYPE: %w", err)
	}
	
	// Write opening population tag
	_, err = file.WriteString("<population>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write opening tag: %w", err)
	}
	
	// Write initial separator
	_, err = file.WriteString("<!-- ====================================================================== -->\n\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write separator: %w", err)
	}
	
	// Process in batches
	processed := 0
	for offset := 0; offset < totalCount; offset += EXPORT_BATCH_SIZE {
		// Query batch
		query := fmt.Sprintf(`
			SELECT id, coords, raw_xml 
			FROM %s 
			ORDER BY id
			LIMIT %d OFFSET %d
		`, tableName, EXPORT_BATCH_SIZE, offset)
		
		rows, err := e.db.Query(query)
		if err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("failed to query batch at offset %d: %w", offset, err)
		}
		
		// Process batch
		batchCount := 0
		for rows.Next() {
			var id, coords, rawXML string
			err := rows.Scan(&id, &coords, &rawXML)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			
			// Format the person XML properly
			var formattedXML string
			if rawXML != "" {
				// Format the compact XML from database
				formatted, err := formatPersonXML(rawXML)
				if err != nil {
					// Fall back to basic indentation if formatting fails
					log.Printf("Failed to format person %s XML: %v", id, err)
					formattedXML = indentXML(rawXML, "\t")
				} else {
					formattedXML = formatted
				}
			} else {
				// Generate basic person XML if no raw XML
				formattedXML = fmt.Sprintf("\t<person id=\"%s\">\n\t</person>", escapeXMLPopulation(id))
			}
			
			// Write the formatted person XML
			_, err = file.WriteString(formattedXML)
			if err != nil {
				rows.Close()
				runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to write person XML: %w", err)
			}
			
			batchCount++
			processed++
			
			// Add separator after each person except the last one
			if processed < totalCount {
				_, err = file.WriteString("\n\n<!-- ====================================================================== -->\n\n")
				if err != nil {
					rows.Close()
					runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
						"error": err.Error(),
					})
					return fmt.Errorf("failed to write separator: %w", err)
				}
			}
		}
		rows.Close()
		
		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("error iterating over rows: %w", err)
		}
		
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": processed,
			"total":   totalCount,
		})
		
		log.Printf("Exported batch: %d-%d of %d", offset, offset+batchCount, totalCount)
	}
	
	// Add final separator and closing tag
	_, err = file.WriteString("\n\n<!-- ====================================================================== -->\n</population>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write closing tag: %w", err)
	}
	
	// Emit completion
	runtime.EventsEmit(e.app.ctx, "export:complete", map[string]interface{}{
		"total": processed,
	})
	
	log.Printf("Successfully exported %d persons to %s", processed, outputPath)
	return nil
}

// indentXML adds proper indentation to XML string
func indentXML(xmlStr string, indent string) string {
	lines := strings.Split(xmlStr, "\n")
	var result []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, indent+trimmed)
		}
	}
	
	return strings.Join(result, "\n")
}

// escapeXMLPopulation escapes special XML characters for population export
func escapeXMLPopulation(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}