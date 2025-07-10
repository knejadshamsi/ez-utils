// Package gui provides network database operations.
package gui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Network table queries
const (
	createNodesTableQuery = `CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		x REAL,
		y REAL,
		raw_xml TEXT
	)`
	createLinksTableQuery = `CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		from_node TEXT,
		to_node TEXT,
		raw_xml TEXT
	)`
	createNodesIndexQuery       = `CREATE INDEX IF NOT EXISTS idx_nodes_coords_%d ON %s(x, y)`
	selectAllNodesQuery         = `SELECT id, x, y, raw_xml FROM %s`
	selectNodesByBboxQuery      = `SELECT id, x, y, raw_xml FROM %s WHERE x >= ? AND x <= ? AND y >= ? AND y <= ?`
	selectAllLinksQuery         = `SELECT id, from_node, to_node, raw_xml FROM %s`
	selectLinksByNodesQuery     = `SELECT id, from_node, to_node, raw_xml FROM %s WHERE from_node IN (%s) OR to_node IN (%s)`
	updateNodeQuery             = `UPDATE %s SET x = ?, y = ?, raw_xml = ? WHERE id = ?`
	updateLinkQuery             = `UPDATE %s SET raw_xml = ? WHERE id = ?`
	insertLinkQuery             = `INSERT INTO %s (id, from_node, to_node, raw_xml) VALUES (?, ?, ?, ?)`
	insertNodeBatchBaseQuery    = "INSERT OR REPLACE INTO %s (id, x, y, raw_xml) VALUES "
	insertLinkBatchBaseQuery    = "INSERT INTO %s (id, from_node, to_node, raw_xml) VALUES "
	networkInsertBatchPlaceholder = "(?, ?, ?, ?)"
	updateTelemetryQueryNetwork = `UPDATE process_telemetry SET bytes_read = ?, nodes_read = ?, links_read = ?, error_count = ?, last_updated = CURRENT_TIMESTAMP WHERE process_id = ?`
)

// Network table error messages
const (
	createNodesTableError       = "failed to create nodes table %s"
	createLinksTableError       = "failed to create links table %s"
	createNodesIndexError       = "failed to create nodes index for %s"
	selectAllNodesError         = "failed to query nodes from %s"
	selectNodesByBboxError      = "failed to query nodes by bbox from %s"
	selectAllLinksError         = "failed to query links from %s"
	selectLinksByNodesError     = "failed to query links by nodes from %s"
	updateNodeError             = "failed to update node in %s"
	updateLinkError             = "failed to update link in %s"
	insertLinkError             = "failed to insert link into %s"
	insertNodeBatchError        = "failed to insert node batch"
	insertLinkBatchError        = "failed to insert link batch"
	updateTelemetryErrorNetwork = "failed to update network telemetry for process %d"
)

// CreateNetworkTables creates tables and index for network data.
func (db *Database) CreateNetworkTables(processID int) error {
	// Ensure telemetry table has network columns
	if err := db.AlterTelemetryTableForNetwork(); err != nil {
		return fmt.Errorf("failed to update telemetry table: %w", err)
	}
	
	nodesTable := fmt.Sprintf("network_nodes_%d", processID)
	linksTable := fmt.Sprintf("network_links_%d", processID)

	if err := db.execTableQuery(fmt.Sprintf(createNodesTableQuery, nodesTable)); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(createNodesTableError, nodesTable), err)
	}

	if err := db.execTableQuery(fmt.Sprintf(createLinksTableQuery, linksTable)); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(createLinksTableError, linksTable), err)
	}

	if err := db.execTableQuery(fmt.Sprintf(createNodesIndexQuery, processID, nodesTable)); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(createNodesIndexError, nodesTable), err)
	}

	return nil
}

// GetNodesInBBox retrieves nodes within a bounding box.
func (db *Database) GetNodesInBBox(processID int, bbox BoundingBox) ([]NodeResult, error) {
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf(selectNodesByBboxQuery, tableName)
	rows, err := db.queryRows(query, fmt.Sprintf(selectNodesByBboxError, tableName), bbox.West, bbox.East, bbox.South, bbox.North)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []NodeResult
	for rows.Next() {
		var node NodeResult
		if err := rows.Scan(&node.ID, &node.X, &node.Y, &node.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

// GetLinksInBBox retrieves links connected to nodes in the given node IDs.
func (db *Database) GetLinksInBBox(processID int, nodeIDs []string) ([]LinkResult, error) {
	if len(nodeIDs) == 0 {
		return []LinkResult{}, nil
	}
	tableName := fmt.Sprintf("network_links_%d", processID)
	placeholders := make([]string, len(nodeIDs))
	args := make([]interface{}, len(nodeIDs)*2)
	for i, id := range nodeIDs {
		placeholders[i] = "?"
		args[i] = id
		args[i+len(nodeIDs)] = id
	}
	query := fmt.Sprintf(selectLinksByNodesQuery, tableName, strings.Join(placeholders, ","), strings.Join(placeholders, ","))
	rows, err := db.queryRows(query, fmt.Sprintf(selectLinksByNodesError, tableName), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []LinkResult
	for rows.Next() {
		var link LinkResult
		if err := rows.Scan(&link.ID, &link.FromNode, &link.ToNode, &link.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// UpdateNetworkNode updates a node's data.
func (db *Database) UpdateNetworkNode(processID int, nodeID string, x, y float64, rawXML string) error {
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf(updateNodeQuery, tableName)
	result, err := db.execQuery(query, fmt.Sprintf(updateNodeError, tableName), x, y, rawXML, nodeID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("node %s not found in %s", nodeID, tableName)
	}
	return nil
}

// UpdateNetworkLink updates a link's raw XML.
func (db *Database) UpdateNetworkLink(processID int, linkID string, rawXML string) error {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf(updateLinkQuery, tableName)
	result, err := db.execQuery(query, fmt.Sprintf(updateLinkError, tableName), rawXML, linkID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("link %s not found in %s", linkID, tableName)
	}
	return nil
}

// CreateNetworkLink inserts a new link.
func (db *Database) CreateNetworkLink(processID int, linkID, fromNode, toNode, rawXML string) error {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf(insertLinkQuery, tableName)
	_, err := db.execQuery(query, fmt.Sprintf(insertLinkError, tableName), linkID, fromNode, toNode, rawXML)
	return err
}

// InsertNodeBatch inserts a batch of nodes using transaction.
func (db *Database) InsertNodeBatch(processID int, nodes []NodeData) error {
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	return db.WithTransaction(func(tx *sql.Tx) error {
		if len(nodes) == 0 {
			return nil
		}
		query := fmt.Sprintf(insertNodeBatchBaseQuery, tableName)
		var values []string
		var args []interface{}
		for _, node := range nodes {
			coords := strings.Split(node.Coords, ",")
			if len(coords) != 2 {
				return fmt.Errorf("invalid coordinates for node %s", node.ID)
			}
			x, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
			if err != nil {
				return err
			}
			values = append(values, networkInsertBatchPlaceholder)
			args = append(args, node.ID, x, y, node.RawXML)
		}
		query += strings.Join(values, ", ")
		_, err := tx.Exec(query, args...)
		return err
	})
}

// InsertLinkBatch inserts a batch of links using transaction.
func (db *Database) InsertLinkBatch(processID int, links []LinkData) error {
	tableName := fmt.Sprintf("network_links_%d", processID)
	return db.WithTransaction(func(tx *sql.Tx) error {
		if len(links) == 0 {
			return nil
		}
		query := fmt.Sprintf(insertLinkBatchBaseQuery, tableName)
		var values []string
		var args []interface{}
		for _, link := range links {
			linkID := fmt.Sprintf("%s_%s", link.FromNode, link.ToNode)
			values = append(values, networkInsertBatchPlaceholder)
			args = append(args, linkID, link.FromNode, link.ToNode, link.RawXML)
			// Note: Handling duplicates as in original, but simplified
		}
		query += strings.Join(values, ", ")
		_, err := tx.Exec(query, args...)
		return err
	})
}

// AlterTelemetryTableForNetwork adds network-specific columns to the telemetry table if they don't exist
func (db *Database) AlterTelemetryTableForNetwork() error {
	// Try to add network-specific columns - ignore errors if columns already exist
	alterQueries := []string{
		"ALTER TABLE process_telemetry ADD COLUMN nodes_read INTEGER DEFAULT 0",
		"ALTER TABLE process_telemetry ADD COLUMN links_read INTEGER DEFAULT 0",
	}
	
	for _, query := range alterQueries {
		// Execute but ignore "duplicate column" errors
		_, _ = db.conn.Exec(query)
	}
	
	return nil
}

// UpdateNetworkTelemetry updates telemetry for network.
func (db *Database) UpdateNetworkTelemetry(processID int, bytesRead, nodesRead, linksRead, errorCount int64) error {
	// First ensure network columns exist
	_ = db.AlterTelemetryTableForNetwork()
	
	_, err := db.execQuery(updateTelemetryQueryNetwork, fmt.Sprintf(updateTelemetryErrorNetwork, processID), bytesRead, nodesRead, linksRead, errorCount, processID)
	return err
}

// UpdateNode updates a network node's coordinates in the database
func (db *Database) UpdateNode(processID int, nodeID string, x, y float64) error {
	// Generate the new XML representation
	rawXML := fmt.Sprintf(`<node id="%s" x="%.6f" y="%.6f"/>`, nodeID, x, y)
	
	// Build the table name
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf(updateNodeQuery, tableName)
	
	// Execute the update using the consistent execQuery pattern
	result, err := db.execQuery(query, fmt.Sprintf(updateNodeError, tableName), x, y, rawXML, nodeID)
	if err != nil {
		return err
	}
	
	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("node not found: %s", nodeID)
	}
	
	return nil
}

// GetNode retrieves a single node by ID
func (db *Database) GetNode(processID int, nodeID string) (*NodeResult, error) {
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf("SELECT id, x, y, raw_xml FROM %s WHERE id = ?", tableName)
	
	var node NodeResult
	err := db.queryRow(query, "", nodeID).Scan(&node.ID, &node.X, &node.Y, &node.RawXML)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	
	return &node, nil
}

// GetLink retrieves a single link by ID
func (db *Database) GetLink(processID int, linkID string) (*LinkResult, error) {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf("SELECT id, from_node, to_node, raw_xml FROM %s WHERE id = ?", tableName)
	
	var link LinkResult
	err := db.queryRow(query, "", linkID).Scan(&link.ID, &link.FromNode, &link.ToNode, &link.RawXML)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("link not found: %s", linkID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	
	return &link, nil
}

// GetAllNodes retrieves all nodes for a process
func (db *Database) GetAllNodes(processID int) ([]NodeResult, error) {
	tableName := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf(selectAllNodesQuery, tableName)
	
	rows, err := db.queryRows(query, fmt.Sprintf(selectAllNodesError, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var nodes []NodeResult
	for rows.Next() {
		var node NodeResult
		if err := rows.Scan(&node.ID, &node.X, &node.Y, &node.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}
	
	return nodes, nil
}

// GetAllLinks retrieves all links for a process
func (db *Database) GetAllLinks(processID int) ([]LinkResult, error) {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf(selectAllLinksQuery, tableName)
	
	rows, err := db.queryRows(query, fmt.Sprintf(selectAllLinksError, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var links []LinkResult
	for rows.Next() {
		var link LinkResult
		if err := rows.Scan(&link.ID, &link.FromNode, &link.ToNode, &link.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}
	
	return links, nil
}

// GetLinksForNode retrieves all links connected to a specific node
func (db *Database) GetLinksForNode(processID int, nodeID string) ([]LinkResult, error) {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf("SELECT id, from_node, to_node, raw_xml FROM %s WHERE from_node = ? OR to_node = ?", tableName)
	
	rows, err := db.queryRows(query, "failed to query links for node", nodeID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var links []LinkResult
	for rows.Next() {
		var link LinkResult
		if err := rows.Scan(&link.ID, &link.FromNode, &link.ToNode, &link.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}
	
	return links, nil
}

// DeleteNode deletes a node and optionally its connected links
func (db *Database) DeleteNode(processID int, nodeID string, deleteConnectedLinks bool) error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		// Delete connected links first if requested
		if deleteConnectedLinks {
			linksTable := fmt.Sprintf("network_links_%d", processID)
			query := fmt.Sprintf("DELETE FROM %s WHERE from_node = ? OR to_node = ?", linksTable)
			if _, err := tx.Exec(query, nodeID, nodeID); err != nil {
				return fmt.Errorf("failed to delete connected links: %w", err)
			}
		}
		
		// Delete the node
		nodesTable := fmt.Sprintf("network_nodes_%d", processID)
		query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", nodesTable)
		result, err := tx.Exec(query, nodeID)
		if err != nil {
			return fmt.Errorf("failed to delete node: %w", err)
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		
		if rowsAffected == 0 {
			return fmt.Errorf("node not found: %s", nodeID)
		}
		
		return nil
	})
}

// DeleteLink deletes a link
func (db *Database) DeleteLink(processID int, linkID string) error {
	tableName := fmt.Sprintf("network_links_%d", processID)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	
	result, err := db.execQuery(query, "failed to delete link", linkID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("link not found: %s", linkID)
	}
	
	return nil
}

// BatchUpdateNodes updates multiple nodes in a transaction
func (db *Database) BatchUpdateNodes(processID int, updates map[string]NodeData) error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		tableName := fmt.Sprintf("network_nodes_%d", processID)
		stmt, err := tx.Prepare(fmt.Sprintf(updateNodeQuery, tableName))
		if err != nil {
			return fmt.Errorf("failed to prepare update statement: %w", err)
		}
		defer stmt.Close()
		
		for nodeID, data := range updates {
			// Parse coordinates
			coords := strings.Split(data.Coords, ",")
			if len(coords) != 2 {
				return fmt.Errorf("invalid coordinates for node %s: %s", nodeID, data.Coords)
			}
			
			x, err := strconv.ParseFloat(coords[0], 64)
			if err != nil {
				return fmt.Errorf("invalid x coordinate for node %s: %s", nodeID, coords[0])
			}
			
			y, err := strconv.ParseFloat(coords[1], 64)
			if err != nil {
				return fmt.Errorf("invalid y coordinate for node %s: %s", nodeID, coords[1])
			}
			
			if _, err := stmt.Exec(x, y, data.RawXML, nodeID); err != nil {
				return fmt.Errorf("failed to update node %s: %w", nodeID, err)
			}
		}
		
		return nil
	})
}

// BatchDeleteNodes deletes multiple nodes in a transaction
func (db *Database) BatchDeleteNodes(processID int, nodeIDs []string, deleteConnectedLinks bool) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		// Delete connected links first if requested
		if deleteConnectedLinks {
			linksTable := fmt.Sprintf("network_links_%d", processID)
			placeholders := make([]string, len(nodeIDs))
			args := make([]interface{}, len(nodeIDs)*2)
			for i, id := range nodeIDs {
				placeholders[i] = "?"
				args[i] = id
				args[len(nodeIDs)+i] = id
			}
			query := fmt.Sprintf("DELETE FROM %s WHERE from_node IN (%s) OR to_node IN (%s)", 
				linksTable, strings.Join(placeholders, ","), strings.Join(placeholders, ","))
			if _, err := tx.Exec(query, args...); err != nil {
				return fmt.Errorf("failed to delete connected links: %w", err)
			}
		}
		
		// Delete nodes
		nodesTable := fmt.Sprintf("network_nodes_%d", processID)
		stmt, err := tx.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id = ?", nodesTable))
		if err != nil {
			return fmt.Errorf("failed to prepare delete statement: %w", err)
		}
		defer stmt.Close()
		
		for _, nodeID := range nodeIDs {
			if _, err := stmt.Exec(nodeID); err != nil {
				return fmt.Errorf("failed to delete node %s: %w", nodeID, err)
			}
		}
		
		return nil
	})
}

// GetNetworkStatistics retrieves statistics about the network
func (db *Database) GetNetworkStatistics(processID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Count nodes
	nodesTable := fmt.Sprintf("network_nodes_%d", processID)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", nodesTable)
	var nodeCount int
	if err := db.queryRow(query, "").Scan(&nodeCount); err != nil {
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}
	stats["node_count"] = nodeCount
	
	// Count links
	linksTable := fmt.Sprintf("network_links_%d", processID)
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s", linksTable)
	var linkCount int
	if err := db.queryRow(query, "").Scan(&linkCount); err != nil {
		return nil, fmt.Errorf("failed to count links: %w", err)
	}
	stats["link_count"] = linkCount
	
	// Get bounding box
	query = fmt.Sprintf("SELECT MIN(x), MAX(x), MIN(y), MAX(y) FROM %s", nodesTable)
	var minX, maxX, minY, maxY sql.NullFloat64
	if err := db.queryRow(query, "").Scan(&minX, &maxX, &minY, &maxY); err != nil {
		return nil, fmt.Errorf("failed to get bounding box: %w", err)
	}
	
	if minX.Valid && maxX.Valid && minY.Valid && maxY.Valid {
		stats["bounding_box"] = BoundingBox{
			West:  minX.Float64,
			East:  maxX.Float64,
			South: minY.Float64,
			North: maxY.Float64,
		}
	}
	
	return stats, nil
}
