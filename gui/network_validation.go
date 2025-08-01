package gui

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
)

// NetworkValidator validates network XML files
type NetworkValidator struct{}

// NewNetworkValidator creates a new network validator
func NewNetworkValidator() *NetworkValidator {
	return &NetworkValidator{}
}

// ValidateNetworkFile validates a MATSim network XML file
func (v *NetworkValidator) ValidateNetworkFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Check if it's valid XML
	decoder := xml.NewDecoder(file)
	var rootFound bool
	var hasNodes bool
	var hasLinks bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid XML structure: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "network":
				rootFound = true
			case "nodes":
				hasNodes = true
			case "links":
				hasLinks = true
			}
		}
	}

	// Validate required elements
	if !rootFound {
		return fmt.Errorf("missing required root element 'network'")
	}

	if !hasNodes {
		return fmt.Errorf("missing required element 'nodes'")
	}

	if !hasLinks {
		return fmt.Errorf("missing required element 'links'")
	}

	// Reset and perform detailed validation
	file.Seek(0, 0)
	return v.validateDetailedStructure(file)
}

// validateDetailedStructure performs detailed validation of the XML structure
func (v *NetworkValidator) validateDetailedStructure(reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	
	nodeIDs := make(map[string]bool)
	linkIDs := make(map[string]bool)
	var inNodes, inLinks bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "nodes":
				inNodes = true
				inLinks = false
			case "links":
				inLinks = true
				inNodes = false
			case "node":
				if inNodes {
					if err := v.validateNode(se.Attr, nodeIDs); err != nil {
						return err
					}
				}
			case "link":
				if inLinks {
					if err := v.validateLink(se.Attr, linkIDs, nodeIDs); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// validateNode validates node attributes
func (v *NetworkValidator) validateNode(attrs []xml.Attr, nodeIDs map[string]bool) error {
	var id string
	var hasX, hasY bool

	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id = attr.Value
		case "x":
			hasX = true
			// Could add coordinate validation here
		case "y":
			hasY = true
			// Could add coordinate validation here
		}
	}

	if id == "" {
		return fmt.Errorf("node missing required attribute 'id'")
	}

	if nodeIDs[id] {
		return fmt.Errorf("duplicate node ID: %s", id)
	}
	nodeIDs[id] = true

	if !hasX {
		return fmt.Errorf("node %s missing required attribute 'x'", id)
	}

	if !hasY {
		return fmt.Errorf("node %s missing required attribute 'y'", id)
	}

	return nil
}

// validateLink validates link attributes
func (v *NetworkValidator) validateLink(attrs []xml.Attr, linkIDs map[string]bool, nodeIDs map[string]bool) error {
	var id, from, to string

	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id = attr.Value
		case "from":
			from = attr.Value
		case "to":
			to = attr.Value
		}
	}

	if id == "" {
		return fmt.Errorf("link missing required attribute 'id'")
	}

	if linkIDs[id] {
		return fmt.Errorf("duplicate link ID: %s", id)
	}
	linkIDs[id] = true

	if from == "" {
		return fmt.Errorf("link %s missing required attribute 'from'", id)
	}

	if to == "" {
		return fmt.Errorf("link %s missing required attribute 'to'", id)
	}

	// Validate that from and to nodes exist
	if !nodeIDs[from] {
		return fmt.Errorf("link %s references non-existent from node: %s", id, from)
	}

	if !nodeIDs[to] {
		return fmt.Errorf("link %s references non-existent to node: %s", id, to)
	}

	return nil
}

// ValidateNetworkConsistency checks for network connectivity issues
func (v *NetworkValidator) ValidateNetworkConsistency(db *Database, processID int) error {
	// Get all nodes
	nodes, err := db.GetAllNodes(processID)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get all links
	links, err := db.GetAllLinks(processID)
	if err != nil {
		return fmt.Errorf("failed to get links: %w", err)
	}

	// Create node ID set
	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node.ID] = true
	}

	// Check all links reference valid nodes
	for _, link := range links {
		if !nodeSet[link.FromNode] {
			return fmt.Errorf("link %s references non-existent from node: %s", link.ID, link.FromNode)
		}
		if !nodeSet[link.ToNode] {
			return fmt.Errorf("link %s references non-existent to node: %s", link.ID, link.ToNode)
		}
	}

	// Check for isolated nodes (nodes with no connected links)
	connectedNodes := make(map[string]bool)
	for _, link := range links {
		connectedNodes[link.FromNode] = true
		connectedNodes[link.ToNode] = true
	}

	var isolatedCount int
	for _, node := range nodes {
		if !connectedNodes[node.ID] {
			isolatedCount++
		}
	}

	if isolatedCount > 0 {
		// This is a warning, not an error
		log.Printf("Warning: %d isolated nodes found in network", isolatedCount)
	}

	return nil
}