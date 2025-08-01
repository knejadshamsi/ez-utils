package gui

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// compactXML removes unnecessary whitespace from XML while preserving structure
func compactXML(xmlStr string) string {
	// Remove leading/trailing whitespace from each line and join
	lines := strings.Split(xmlStr, "\n")
	var compacted []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			compacted = append(compacted, trimmed)
		}
	}

	// Join without newlines to create compact XML
	return strings.Join(compacted, "")
}

// formatPersonXML formats a person XML element with proper MATSim structure
func formatPersonXML(xmlStr string) (string, error) {
	var result bytes.Buffer
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	encoder := xml.NewEncoder(&result)
	encoder.Indent("", "\t")

	// Copy tokens from decoder to encoder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		encoder.EncodeToken(tok)
	}
	encoder.Flush()

	return result.String(), nil
}

// formatNetworkXML formats a network XML element (node/link) with proper MATSim structure
func formatNetworkXML(xmlStr string) (string, error) {
	var result bytes.Buffer
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	encoder := xml.NewEncoder(&result)
	encoder.Indent("        ", "\t") // Start with 8 spaces, then use tabs for nested elements

	// Copy tokens from decoder to encoder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		encoder.EncodeToken(tok)
	}
	encoder.Flush()

	return result.String(), nil
}
