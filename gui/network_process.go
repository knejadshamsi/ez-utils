// Package gui provides network processing logic.
package gui

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// NetworkProcessor for network processing.
type NetworkProcessor struct {
	db                      *Database
	processID               int
	telemetry               *ProcessTelemetry
	telemetryMutex          sync.Mutex
	nodesRead               int64
	linksRead               int64
	errorCounter            int64
	telemetryFailureCounter int64
}

// NewNetworkProcessor creates a new network processor.
func NewNetworkProcessor(db *Database, processID int) (*NetworkProcessor, error) {
	return &NetworkProcessor{
		db:        db,
		processID: processID,
		telemetry: &ProcessTelemetry{
			ProcessID:   processID,
			ErrorCount:  0,
			LastUpdated: time.Now(),
		},
	}, nil
}

// ProcessNetworkFile starts processing a network file.
func (a *App) ProcessNetworkFile(filePath string) (map[string]any, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	result, err := a.db.execQuery(createProcessQuery, fmt.Sprintf("failed to create process for file %s", filePath), filePath, "pending")
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	processID := int(id)

	go a.processNetworkFile(filePath, processID)

	return map[string]any{
		"message":   "Processing started",
		"processId": processID,
	}, nil
}

// processNetworkFile processes the network XML.
func (a *App) processNetworkFile(filePath string, processID int) {
	updateStatus := func(status string) {
		if _, err := a.db.execQuery(updateProcessStatusQuery, fmt.Sprintf("failed to update process status for ID %d", processID), status, processID); err != nil {
			log.Printf("Failed to update status: %v", err)
		}
	}

	updateStatus("Initializing processor...")
	processor, err := NewNetworkProcessor(a.db, processID)
	if err != nil {
		updateStatus(fmt.Sprintf("failed to create processor: %v", err))
		return
	}

	updateStatus("Processing file...")
	if err := processor.ProcessNetworkFile(filePath); err != nil {
		updateStatus(fmt.Sprintf("failed to process file: %v", err))
		return
	}

	updateStatus("Completed")
}

// ProcessNetworkFile method for NetworkProcessor.
func (p *NetworkProcessor) ProcessNetworkFile(filePath string) error {
	// Validate XML structure first
	validator := NewNetworkValidator()
	if err := validator.ValidateNetworkFile(filePath); err != nil {
		return fmt.Errorf("network XML validation failed: %w", err)
	}
	log.Printf("Network XML validation passed")

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	p.telemetry.TotalFileSize = fileInfo.Size()

	_, err = p.db.execQuery(initTelemetryQuery, fmt.Sprintf("failed to initialize telemetry for %d", p.processID), p.processID, p.telemetry.TotalFileSize)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	countingReader := &CountingReader{reader: file}

	if err := p.db.CreateNetworkTables(p.processID); err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan error, 1)
	telemetryDone := make(chan struct{})

	go func() {
		defer close(telemetryDone)
		for {
			select {
			case <-ticker.C:
				p.updateTelemetry(countingReader)
			case <-done:
				p.updateTelemetry(countingReader)
				return
			}
		}
	}()

	err = p.processXML(countingReader)
	done <- err
	<-telemetryDone

	if err != nil {
		return err
	}

	p.telemetryMutex.Lock()
	failCount := p.telemetryFailureCounter
	p.telemetryMutex.Unlock()

	if failCount > 0 {
		log.Printf("Process %d completed with %d telemetry failures", p.processID, failCount)
	}

	return nil
}

// processXML processes the XML content.
func (p *NetworkProcessor) processXML(reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	nodeBatch := make([]NodeData, 0, 100)
	linkBatch := make([]LinkData, 0, 100)
	var nodeBuffer, linkBuffer bytes.Buffer
	nodeEncoder := xml.NewEncoder(&nodeBuffer)
	linkEncoder := xml.NewEncoder(&linkBuffer)
	nodeDepth, linkDepth := 0, 0

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if nodeDepth > 0 {
			nodeEncoder.EncodeToken(token)
		} else if linkDepth > 0 {
			linkEncoder.EncodeToken(token)
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "node" {
				if nodeDepth == 0 {
					nodeBuffer.Reset()
					nodeEncoder.EncodeToken(se)
				}
				nodeDepth++
			} else if se.Name.Local == "link" {
				if linkDepth == 0 {
					linkBuffer.Reset()
					linkEncoder.EncodeToken(se)
				}
				linkDepth++
			}
		case xml.EndElement:
			if se.Name.Local == "node" {
				nodeDepth--
				if nodeDepth == 0 {
					nodeEncoder.Flush()
					nodeData, err := p.extractNodeData(nodeBuffer.String())
					if err == nil {
						nodeBatch = append(nodeBatch, nodeData)
						p.telemetryMutex.Lock()
						p.nodesRead++
						shouldUpdate := p.nodesRead%100 == 0
						p.telemetryMutex.Unlock()

						if len(nodeBatch) >= 100 {
							if err := p.db.InsertNodeBatch(p.processID, nodeBatch); err != nil {
								return err
							}
							nodeBatch = make([]NodeData, 0, 100)
						}

						if shouldUpdate {
							p.updateTelemetry(reader.(*CountingReader))
						}
					} else {
						p.telemetryMutex.Lock()
						p.errorCounter++
						p.telemetryMutex.Unlock()
					}
				}
			} else if se.Name.Local == "link" {
				linkDepth--
				if linkDepth == 0 {
					linkEncoder.Flush()
					linkData, err := p.extractLinkData(linkBuffer.String())
					if err == nil {
						linkBatch = append(linkBatch, linkData)
						p.telemetryMutex.Lock()
						p.linksRead++
						shouldUpdate := p.linksRead%100 == 0
						p.telemetryMutex.Unlock()

						if len(linkBatch) >= 100 {
							if err := p.db.InsertLinkBatch(p.processID, linkBatch); err != nil {
								return err
							}
							linkBatch = make([]LinkData, 0, 100)
						}

						if shouldUpdate {
							p.updateTelemetry(reader.(*CountingReader))
						}
					} else {
						p.telemetryMutex.Lock()
						p.errorCounter++
						p.telemetryMutex.Unlock()
					}
				}
			}
		}
	}

	if len(nodeBatch) > 0 {
		if err := p.db.InsertNodeBatch(p.processID, nodeBatch); err != nil {
			return err
		}
	}
	if len(linkBatch) > 0 {
		if err := p.db.InsertLinkBatch(p.processID, linkBatch); err != nil {
			return err
		}
	}

	return nil
}

// extractNodeData extracts data from node XML.
func (p *NetworkProcessor) extractNodeData(xmlString string) (NodeData, error) {
	// Parse XML to extract attributes
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlString)))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return NodeData{}, fmt.Errorf("failed to parse node XML: %w", err)
		}
		
		if se, ok := token.(xml.StartElement); ok && se.Name.Local == "node" {
			var id, x, y string
			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "id":
					id = attr.Value
				case "x":
					x = attr.Value
				case "y":
					y = attr.Value
				}
			}
			
			if id == "" || x == "" || y == "" {
				return NodeData{}, fmt.Errorf("node missing required attributes")
			}
			
			coords := fmt.Sprintf("%s,%s", x, y)
			return NodeData{ID: id, Coords: coords, RawXML: xmlString}, nil
		}
	}
	
	return NodeData{}, fmt.Errorf("no node element found in XML")
}

// extractLinkData extracts data from link XML.
func (p *NetworkProcessor) extractLinkData(xmlString string) (LinkData, error) {
	// Parse XML to extract attributes
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlString)))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return LinkData{}, fmt.Errorf("failed to parse link XML: %w", err)
		}
		
		if se, ok := token.(xml.StartElement); ok && se.Name.Local == "link" {
			var from, to string
			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "from":
					from = attr.Value
				case "to":
					to = attr.Value
				}
			}
			
			if from == "" || to == "" {
				return LinkData{}, fmt.Errorf("link missing required attributes")
			}
			
			return LinkData{FromNode: from, ToNode: to, RawXML: xmlString}, nil
		}
	}
	
	return LinkData{}, fmt.Errorf("no link element found in XML")
}

// updateTelemetry updates network telemetry.
func (p *NetworkProcessor) updateTelemetry(reader *CountingReader) {
	p.telemetryMutex.Lock()
	p.telemetry.BytesRead = reader.BytesRead()
	p.telemetry.NodesRead = p.nodesRead // Assume extension in types
	p.telemetry.LinksRead = p.linksRead
	p.telemetry.ErrorCount = p.errorCounter
	p.telemetry.LastUpdated = time.Now()
	telemetry := *p.telemetry
	p.telemetryMutex.Unlock()

	err := p.db.UpdateNetworkTelemetry(p.processID, telemetry.BytesRead, telemetry.NodesRead, telemetry.LinksRead, telemetry.ErrorCount)
	if err != nil {
		log.Printf("Failed to update telemetry: %v", err)
		time.Sleep(100 * time.Millisecond)
		err = p.db.UpdateNetworkTelemetry(p.processID, telemetry.BytesRead, telemetry.NodesRead, telemetry.LinksRead, telemetry.ErrorCount)
		if err != nil {
			p.telemetryMutex.Lock()
			p.telemetryFailureCounter++
			p.telemetryMutex.Unlock()
		}
	}
}

// updateNetworkNode updates a network node's coordinates (internal function for interpreter)
func (a *App) updateNetworkNode(processID int, nodeID string, x, y float64) error {
	return a.db.UpdateNode(processID, nodeID, x, y)
}
