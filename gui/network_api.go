package gui

import (
	"fmt"
	"log"
)

// Network API methods use NodeResult and LinkResult types defined in network_export.go

// GetNodesInBBox retrieves network nodes within a bounding box
func (a *App) GetNodesInBBox(processID int, bbox BoundingBox) ([]NodeResult, error) {
	log.Printf("GetNodesInBBox called with processID: %d, bbox: %+v", processID, bbox)
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	nodes, err := a.db.GetNodesInBBox(processID, bbox)
	if err != nil {
		log.Printf("Error getting nodes: %v", err)
		return nil, err
	}
	
	log.Printf("Retrieved %d nodes", len(nodes))
	return nodes, nil
}

// GetLinksInBBox retrieves network links connected to the given node IDs
func (a *App) GetLinksInBBox(processID int, nodeIDs []string) ([]LinkResult, error) {
	log.Printf("GetLinksInBBox called with processID: %d, nodeIDs count: %d", processID, len(nodeIDs))
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	links, err := a.db.GetLinksInBBox(processID, nodeIDs)
	if err != nil {
		log.Printf("Error getting links: %v", err)
		return nil, err
	}
	
	log.Printf("Retrieved %d links", len(links))
	return links, nil
}

// UpdateNetworkNode updates a network node's properties
func (a *App) UpdateNetworkNode(processID int, nodeID string, x, y float64, rawXML string) error {
	log.Printf("UpdateNetworkNode called with processID: %d, nodeID: %s", processID, nodeID)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	err := a.db.UpdateNetworkNode(processID, nodeID, x, y, rawXML)
	if err != nil {
		log.Printf("Error updating node: %v", err)
		return err
	}
	
	log.Printf("Node updated successfully")
	return nil
}

// UpdateNetworkLink updates a network link's properties
func (a *App) UpdateNetworkLink(processID int, linkID string, rawXML string) error {
	log.Printf("UpdateNetworkLink called with processID: %d, linkID: %s", processID, linkID)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	err := a.db.UpdateNetworkLink(processID, linkID, rawXML)
	if err != nil {
		log.Printf("Error updating link: %v", err)
		return err
	}
	
	log.Printf("Link updated successfully")
	return nil
}

// DeleteNetworkNode deletes a network node and its connected links
func (a *App) DeleteNetworkNode(processID int, nodeID string) error {
	log.Printf("DeleteNetworkNode called with processID: %d, nodeID: %s", processID, nodeID)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// First delete all connected links
	err := a.db.DeleteLinksByNode(processID, nodeID)
	if err != nil {
		log.Printf("Error deleting connected links: %v", err)
		return err
	}
	
	// Then delete the node (true = deleteConnectedLinks)
	err = a.db.DeleteNode(processID, nodeID, true)
	if err != nil {
		log.Printf("Error deleting node: %v", err)
		return err
	}
	
	log.Printf("Node and connected links deleted successfully")
	return nil
}

// DeleteNetworkLink deletes a network link
func (a *App) DeleteNetworkLink(processID int, linkID string) error {
	log.Printf("DeleteNetworkLink called with processID: %d, linkID: %s", processID, linkID)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	err := a.db.DeleteLink(processID, linkID)
	if err != nil {
		log.Printf("Error deleting link: %v", err)
		return err
	}
	
	log.Printf("Link deleted successfully")
	return nil
}

// CreateNetworkNode creates a new network node
func (a *App) CreateNetworkNode(processID int, nodeID string, x, y float64, rawXML string) error {
	log.Printf("CreateNetworkNode called with processID: %d, nodeID: %s", processID, nodeID)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Insert the node
	err := a.db.InsertNetworkNode(processID, nodeID, x, y, rawXML)
	if err != nil {
		log.Printf("Error creating node: %v", err)
		return err
	}
	
	log.Printf("Node created successfully")
	return nil
}

// CreateNetworkLink creates a new network link
func (a *App) CreateNetworkLink(processID int, linkID string, fromNode, toNode string, rawXML string) error {
	log.Printf("CreateNetworkLink called with processID: %d, linkID: %s, from: %s, to: %s", processID, linkID, fromNode, toNode)
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	err := a.db.InsertNetworkLink(processID, linkID, fromNode, toNode, rawXML)
	if err != nil {
		log.Printf("Error creating link: %v", err)
		return err
	}
	
	log.Printf("Link created successfully")
	return nil
}