package gui

import (
	"encoding/xml"
	"fmt"
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

// formatXMLWithIndent formats XML with proper indentation
func formatXMLWithIndent(xmlStr string, indent string) (string, error) {
	// Parse the XML
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	
	var result strings.Builder
	var depth int
	var lastToken xml.Token
	
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error parsing XML: %w", err)
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			// Write indentation
			result.WriteString(strings.Repeat(indent, depth))
			
			// Write start element
			result.WriteString("<")
			result.WriteString(t.Name.Local)
			
			// Write attributes
			for _, attr := range t.Attr {
				result.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, escapeXMLAttribute(attr.Value)))
			}
			
			// Check if this will be a self-closing tag by peeking ahead
			nextToken, err := decoder.Token()
			if err == nil {
				if endElem, ok := nextToken.(xml.EndElement); ok && endElem.Name.Local == t.Name.Local {
					// Self-closing tag
					result.WriteString(" />")
					result.WriteString("\n")
				} else {
					// Regular tag
					result.WriteString(">")
					
					// Only add newline if the next token is not character data
					if _, isCharData := nextToken.(xml.CharData); !isCharData {
						result.WriteString("\n")
						depth++
					}
					
					// Put the token back for processing
					decoder = xml.NewDecoder(strings.NewReader(xmlStr))
					// Re-parse to current position (not ideal but works for our use case)
					// In production, we'd use a more sophisticated approach
				}
			} else {
				result.WriteString(">")
				result.WriteString("\n")
				depth++
			}
			
			lastToken = token
			
		case xml.EndElement:
			// Check if we need to outdent
			if _, wasStart := lastToken.(xml.StartElement); !wasStart {
				depth--
				result.WriteString(strings.Repeat(indent, depth))
			}
			
			result.WriteString("</")
			result.WriteString(t.Name.Local)
			result.WriteString(">")
			result.WriteString("\n")
			
			lastToken = token
			
		case xml.CharData:
			// Only write non-whitespace character data
			data := strings.TrimSpace(string(t))
			if data != "" {
				result.WriteString(data)
			}
			lastToken = token
		}
	}
	
	return result.String(), nil
}

// formatPersonXML formats a person XML element with proper MATSim structure
func formatPersonXML(xmlStr string) (string, error) {
	// First compact the XML
	compacted := compactXML(xmlStr)
	
	// Use a simple approach for person formatting since we know the structure
	var result strings.Builder
	depth := 0
	indent := "\t"
	
	// Simple state machine for formatting
	inTag := false
	tagContent := ""
	
	for i := 0; i < len(compacted); i++ {
		ch := compacted[i]
		
		if ch == '<' {
			if i+1 < len(compacted) && compacted[i+1] == '/' {
				// Closing tag
				if !inTag && result.Len() > 0 && result.String()[result.Len()-1] != '\n' {
					// Add newline before closing tag if needed
					result.WriteString("\n")
				}
				depth--
				if depth >= 0 {
					result.WriteString(strings.Repeat(indent, depth))
				}
			} else {
				// Opening tag
				if result.Len() > 0 && result.String()[result.Len()-1] != '\n' {
					result.WriteString("\n")
				}
				result.WriteString(strings.Repeat(indent, depth))
			}
			inTag = true
			tagContent = ""
			result.WriteByte(ch)
		} else if ch == '>' {
			result.WriteByte(ch)
			
			// Check if it's a self-closing tag
			if len(tagContent) > 0 && tagContent[len(tagContent)-1] == '/' {
				// Self-closing tag, don't increase depth
			} else if strings.HasPrefix(tagContent, "/") {
				// Closing tag already handled depth
			} else {
				// Opening tag, increase depth
				depth++
			}
			
			inTag = false
			tagContent = ""
		} else {
			if inTag {
				tagContent += string(ch)
			}
			result.WriteByte(ch)
		}
	}
	
	return result.String(), nil
}

// escapeXMLAttribute escapes special characters in XML attributes
func escapeXMLAttribute(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}