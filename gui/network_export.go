// Package gui provides network export functionality.
package gui

import (
	"fmt"
	"io"
	"os"
)

// NetworkExporter exports network data.
type NetworkExporter struct {
	processID int
	db        *Database
}

// NewNetworkExporter creates a new exporter.
func NewNetworkExporter(db *Database, processID int) *NetworkExporter {
	return &NetworkExporter{
		processID: processID,
		db:        db,
	}
}

// ExportToFile exports network to XML file.
func (e *NetworkExporter) ExportToFile(filePath string, loadedRegions []BoundingBox) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE network SYSTEM "http://www.matsim.org/files/dtd/network_v2.dtd">
<network>
`)
	if err != nil {
		return err
	}

	if err = e.exportNodes(file, loadedRegions); err != nil {
		return err
	}

	if err = e.exportLinks(file, loadedRegions); err != nil {
		return err
	}

	_, err = file.WriteString("</network>\n")
	return err
}

// exportNodes exports nodes to writer.
func (e *NetworkExporter) exportNodes(w io.Writer, regions []BoundingBox) error {
	_, err := w.Write([]byte("    <nodes>\n"))
	if err != nil {
		return err
	}

	exportedNodes := make(map[string]bool)
	for _, bbox := range regions {
		nodes, err := e.db.GetNodesInBBox(e.processID, bbox) // Assume added to Database
		if err != nil {
			return err
		}
		for _, node := range nodes {
			if exportedNodes[node.ID] {
				continue
			}
			if node.RawXML != "" {
				_, err = w.Write([]byte("        " + node.RawXML + "\n"))
			} else {
				_, err = fmt.Fprintf(w, `        <node id="%s" x="%f" y="%f" />\n`, node.ID, node.X, node.Y)
			}
			if err != nil {
				return err
			}
			exportedNodes[node.ID] = true
		}
	}
	_, err = w.Write([]byte("    </nodes>\n"))
	return err
}

// exportLinks exports links to writer.
func (e *NetworkExporter) exportLinks(w io.Writer, regions []BoundingBox) error {
	_, err := w.Write([]byte("    <links>\n"))
	if err != nil {
		return err
	}

	allNodeIDs := make(map[string]bool)
	for _, bbox := range regions {
		nodes, err := e.db.GetNodesInBBox(e.processID, bbox)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			allNodeIDs[node.ID] = true
		}
	}

	nodeIDSlice := make([]string, 0, len(allNodeIDs))
	for id := range allNodeIDs {
		nodeIDSlice = append(nodeIDSlice, id)
	}

	links, err := e.db.GetLinksInBBox(e.processID, nodeIDSlice)
	if err != nil {
		return err
	}

	exportedLinks := make(map[string]bool)
	for _, link := range links {
		if exportedLinks[link.ID] || !allNodeIDs[link.FromNode] || !allNodeIDs[link.ToNode] {
			continue
		}
		if link.RawXML != "" {
			_, err = w.Write([]byte("        " + link.RawXML + "\n"))
		} else {
			_, err = fmt.Fprintf(w, `        <link id="%s" from="%s" to="%s" />\n`, link.ID, link.FromNode, link.ToNode)
		}
		if err != nil {
			return err
		}
		exportedLinks[link.ID] = true
	}

	_, err = w.Write([]byte("    </links>\n"))
	return err
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
	exporter := NewNetworkExporter(a.db, processID)
	return exporter.ExportToFile(outputPath, loadedRegions)
}
