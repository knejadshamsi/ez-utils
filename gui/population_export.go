package gui

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
)

// PopulationExporter handles exporting population data to XML
type PopulationExporter struct {
	db *sql.DB
}

// PopulationXML represents the root element of a population XML file
type PopulationXML struct {
	XMLName xml.Name `xml:"population"`
	Persons []string `xml:",innerxml"`
}

// ExportPopulationFile exports population data from database to XML file
func (e *PopulationExporter) ExportPopulationFile(tableName string, outputPath string) error {
	log.Printf("Starting population export from table: %s to file: %s", tableName, outputPath)
	
	// Query all persons from the table
	query := fmt.Sprintf(`
		SELECT id, coords, raw_xml 
		FROM %s 
		ORDER BY id
	`, tableName)
	
	rows, err := e.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query population data: %w", err)
	}
	defer rows.Close()
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Write XML header
	_, err = file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	if err != nil {
		return fmt.Errorf("failed to write XML header: %w", err)
	}
	
	// Write DOCTYPE
	_, err = file.WriteString(`<!DOCTYPE population SYSTEM "http://www.matsim.org/files/dtd/population_v6.dtd">` + "\n")
	if err != nil {
		return fmt.Errorf("failed to write DOCTYPE: %w", err)
	}
	
	// Write opening population tag
	_, err = file.WriteString("<population>\n")
	if err != nil {
		return fmt.Errorf("failed to write opening tag: %w", err)
	}
	
	// Process each person
	count := 0
	for rows.Next() {
		var id, coords, rawXML string
		err := rows.Scan(&id, &coords, &rawXML)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		
		// If we have raw XML, use it
		if rawXML != "" {
			// Ensure proper indentation
			indentedXML := indentXML(rawXML, "\t")
			_, err = file.WriteString(indentedXML + "\n")
			if err != nil {
				return fmt.Errorf("failed to write person XML: %w", err)
			}
		} else {
			// Generate basic person XML if no raw XML
			personXML := fmt.Sprintf("\t<person id=\"%s\">\n\t</person>\n", escapeXMLPopulation(id))
			_, err = file.WriteString(personXML)
			if err != nil {
				return fmt.Errorf("failed to write person XML: %w", err)
			}
		}
		
		count++
	}
	
	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %w", err)
	}
	
	// Write closing population tag
	_, err = file.WriteString("</population>\n")
	if err != nil {
		return fmt.Errorf("failed to write closing tag: %w", err)
	}
	
	log.Printf("Successfully exported %d persons to %s", count, outputPath)
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