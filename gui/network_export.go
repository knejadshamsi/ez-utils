// Package gui provides network export functionality.
package gui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NetworkExporter handles exporting network data to XML
type NetworkExporter struct {
	processID int
	db        *Database
	app       *App // Need app for context to emit events
}

const NETWORK_EXPORT_BATCH_SIZE = 1000 // Process 1000 elements at a time

// NewNetworkExporter creates a new exporter.
func NewNetworkExporter(db *Database, processID int, app *App) *NetworkExporter {
	return &NetworkExporter{
		processID: processID,
		db:        db,
		app:       app,
	}
}

// ExportToFile exports network to XML file with batching and progress events
func (e *NetworkExporter) ExportToFile(filePath string, loadedRegions []BoundingBox) error {
	log.Printf("Starting network export from process: %d to file: %s", e.processID, filePath)
	
	// First, get the total count of nodes and links
	var totalNodes, totalLinks int
	nodesTable := fmt.Sprintf("network_nodes_%d", e.processID)
	linksTable := fmt.Sprintf("network_links_%d", e.processID)
	
	if err := e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", nodesTable), "").Scan(&totalNodes); err != nil {
		return fmt.Errorf("failed to count network nodes: %w", err)
	}
	if err := e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", linksTable), "").Scan(&totalLinks); err != nil {
		return fmt.Errorf("failed to count network links: %w", err)
	}
	
	totalElements := totalNodes + totalLinks
	
	// Emit initial progress
	runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
		"current": 0,
		"total":   totalElements,
	})
	
	// Create output file
	file, err := os.Create(filePath)
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
	_, err = file.WriteString(`<!DOCTYPE network SYSTEM "http://www.matsim.org/files/dtd/network_v2.dtd">` + "\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write DOCTYPE: %w", err)
	}
	
	// Write opening network tag
	_, err = file.WriteString("<network>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write opening network tag: %w", err)
	}
	
	// Export nodes
	processed := 0
	processed, err = e.exportNodesWithProgress(file, totalElements, processed)
	if err != nil {
		return err
	}
	
	// Export links
	processed, err = e.exportLinksWithProgress(file, totalElements, processed)
	if err != nil {
		return err
	}
	
	// Add final separator and closing tag
	_, err = file.WriteString("\n\n<!-- ====================================================================== -->\n</network>\n")
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
	
	log.Printf("Successfully exported %d network elements to %s", processed, filePath)
	return nil
}

// exportNodesWithProgress exports nodes in batches with progress updates
func (e *NetworkExporter) exportNodesWithProgress(file *os.File, totalElements, startProcessed int) (int, error) {
	processed := startProcessed
	nodesTable := fmt.Sprintf("network_nodes_%d", e.processID)
	
	// Get total nodes count for batching
	var totalNodes int
	if err := e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", nodesTable), "").Scan(&totalNodes); err != nil {
		return processed, fmt.Errorf("failed to count nodes: %w", err)
	}
	
	if totalNodes == 0 {
		// Write empty nodes section
		_, err := file.WriteString("    <nodes>\n    </nodes>\n")
		return processed, err
	}
	
	// Write opening nodes tag
	_, err := file.WriteString("    <nodes>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write opening nodes tag: %w", err)
	}
	
	// Write initial separator
	_, err = file.WriteString("<!-- ====================================================================== -->\n\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write initial nodes separator: %w", err)
	}
	
	// Process nodes in batches
	for offset := 0; offset < totalNodes; offset += NETWORK_EXPORT_BATCH_SIZE {
		// Query batch
		query := fmt.Sprintf(`
			SELECT id, x, y, raw_xml 
			FROM %s 
			ORDER BY id
			LIMIT %d OFFSET %d
		`, nodesTable, NETWORK_EXPORT_BATCH_SIZE, offset)
		
		rows, err := e.db.queryRows(query, "failed to query node batch")
		if err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return processed, fmt.Errorf("failed to query node batch at offset %d: %w", offset, err)
		}
		
		// Process batch
		batchCount := 0
		for rows.Next() {
			var id string
			var x, y float64
			var rawXML string
			err := rows.Scan(&id, &x, &y, &rawXML)
			if err != nil {
				log.Printf("Error scanning node row: %v", err)
				continue
			}
			
			// Format the node XML properly
			var formattedXML string
			if rawXML != "" {
				// Format the XML from database
				formatted, err := formatNetworkXML(rawXML)
				if err != nil {
					// Fall back to basic indentation if formatting fails
					log.Printf("Failed to format node %s XML: %v", id, err)
					formattedXML = indentNetworkXML(rawXML, "        ")
				} else {
					formattedXML = formatted
				}
			} else {
				// Generate basic node XML if no raw XML
				formattedXML = fmt.Sprintf("        <node id=\"%s\" x=\"%f\" y=\"%f\" />", escapeXMLNetwork(id), x, y)
			}
			
			// Write the formatted node XML
			_, err = file.WriteString(formattedXML)
			if err != nil {
				rows.Close()
				runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
					"error": err.Error(),
				})
				return processed, fmt.Errorf("failed to write node XML: %w", err)
			}
			
			batchCount++
			processed++
			
			// Add separator after each node except the last one
			if processed-startProcessed < totalNodes {
				_, err = file.WriteString("\n\n<!-- ====================================================================== -->\n\n")
				if err != nil {
					rows.Close()
					runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
						"error": err.Error(),
					})
					return processed, fmt.Errorf("failed to write node separator: %w", err)
				}
			}
		}
		rows.Close()
		
		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return processed, fmt.Errorf("error iterating over node rows: %w", err)
		}
		
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": processed,
			"total":   totalElements,
		})
		
		log.Printf("Exported node batch: %d-%d of %d", offset, offset+batchCount, totalNodes)
	}
	
	// Write closing nodes tag
	_, err = file.WriteString("\n    </nodes>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write closing nodes tag: %w", err)
	}
	
	return processed, nil
}

// exportLinksWithProgress exports links in batches with progress updates
func (e *NetworkExporter) exportLinksWithProgress(file *os.File, totalElements, startProcessed int) (int, error) {
	processed := startProcessed
	linksTable := fmt.Sprintf("network_links_%d", e.processID)
	
	// Get total links count for batching
	var totalLinks int
	if err := e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", linksTable), "").Scan(&totalLinks); err != nil {
		return processed, fmt.Errorf("failed to count links: %w", err)
	}
	
	if totalLinks == 0 {
		// Write empty links section
		_, err := file.WriteString("    <links>\n    </links>\n")
		return processed, err
	}
	
	// Write opening links tag
	_, err := file.WriteString("    <links>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write opening links tag: %w", err)
	}
	
	// Write initial separator
	_, err = file.WriteString("<!-- ====================================================================== -->\n\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write initial links separator: %w", err)
	}
	
	// Process links in batches
	for offset := 0; offset < totalLinks; offset += NETWORK_EXPORT_BATCH_SIZE {
		// Query batch
		query := fmt.Sprintf(`
			SELECT id, from_node, to_node, raw_xml 
			FROM %s 
			ORDER BY id
			LIMIT %d OFFSET %d
		`, linksTable, NETWORK_EXPORT_BATCH_SIZE, offset)
		
		rows, err := e.db.queryRows(query, "failed to query link batch")
		if err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return processed, fmt.Errorf("failed to query link batch at offset %d: %w", offset, err)
		}
		
		// Process batch
		batchCount := 0
		for rows.Next() {
			var id, fromNode, toNode, rawXML string
			err := rows.Scan(&id, &fromNode, &toNode, &rawXML)
			if err != nil {
				log.Printf("Error scanning link row: %v", err)
				continue
			}
			
			// Format the link XML properly
			var formattedXML string
			if rawXML != "" {
				// Format the XML from database
				formatted, err := formatNetworkXML(rawXML)
				if err != nil {
					// Fall back to basic indentation if formatting fails
					log.Printf("Failed to format link %s XML: %v", id, err)
					formattedXML = indentNetworkXML(rawXML, "        ")
				} else {
					formattedXML = formatted
				}
			} else {
				// Generate basic link XML if no raw XML
				formattedXML = fmt.Sprintf("        <link id=\"%s\" from=\"%s\" to=\"%s\" />", 
					escapeXMLNetwork(id), escapeXMLNetwork(fromNode), escapeXMLNetwork(toNode))
			}
			
			// Write the formatted link XML
			_, err = file.WriteString(formattedXML)
			if err != nil {
				rows.Close()
				runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
					"error": err.Error(),
				})
				return processed, fmt.Errorf("failed to write link XML: %w", err)
			}
			
			batchCount++
			processed++
			
			// Add separator after each link except the last one
			linksProcessed := processed - startProcessed
			if linksProcessed < totalLinks {
				_, err = file.WriteString("\n\n<!-- ====================================================================== -->\n\n")
				if err != nil {
					rows.Close()
					runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
						"error": err.Error(),
					})
					return processed, fmt.Errorf("failed to write link separator: %w", err)
				}
			}
		}
		rows.Close()
		
		// Check for errors from iterating over rows
		if err = rows.Err(); err != nil {
			runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
				"error": err.Error(),
			})
			return processed, fmt.Errorf("error iterating over link rows: %w", err)
		}
		
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": processed,
			"total":   totalElements,
		})
		
		log.Printf("Exported link batch: %d-%d of %d", offset, offset+batchCount, totalLinks)
	}
	
	// Write closing links tag
	_, err = file.WriteString("\n    </links>\n")
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return processed, fmt.Errorf("failed to write closing links tag: %w", err)
	}
	
	return processed, nil
}

// GetNetworkStats returns network statistics.
func (e *NetworkExporter) GetNetworkStats() (totalNodes, totalLinks int, err error) {
	// Use db to query counts
	nodesTable := fmt.Sprintf("network_nodes_%d", e.processID)
	linksTable := fmt.Sprintf("network_links_%d", e.processID)
	err = e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", nodesTable), "").Scan(&totalNodes)
	if err != nil {
		return
	}
	err = e.db.queryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", linksTable), "").Scan(&totalLinks)
	return
}

// ExportNetworkFile is the App method for exporting network files
func (a *App) ExportNetworkFile(processID int, outputPath string, loadedRegions []BoundingBox) error {
	exporter := NewNetworkExporter(a.db, processID, a)
	return exporter.ExportToFile(outputPath, loadedRegions)
}

// indentNetworkXML adds proper indentation to XML string for networks
func indentNetworkXML(xmlStr string, indent string) string {
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

// escapeXMLNetwork escapes special XML characters for network export
func escapeXMLNetwork(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}