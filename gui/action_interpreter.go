package gui

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// ExecuteAction executes a single action from the frontend
func (a *App) ExecuteAction(actionJSON json.RawMessage) error {
	// First, get the type info to determine which action to execute
	var typeInfo struct {
		Type        string `json:"type"`
		ElementType string `json:"elementType"`
		Action      string `json:"action"`
	}
	
	if err := json.Unmarshal(actionJSON, &typeInfo); err != nil {
		return fmt.Errorf("failed to unmarshal action type info: %w", err)
	}
	
	// Create the switch key
	key := fmt.Sprintf("%s.%s.%s", typeInfo.Type, typeInfo.ElementType, typeInfo.Action)
	
	switch key {
	// ===== POPULATION ACTIONS =====
	case "population.person.add":
		var act struct {
			TableName string `json:"tableName"`
			Data struct {
				ID     string `json:"id"`
				Coords string `json:"coords"`
				RawXML string `json:"rawXML"`
			} `json:"data"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal add person action: %w", err)
		}
		// Call the internal function that has business logic
		personData := map[string]any{
			"id":      act.Data.ID,
			"coords":  act.Data.Coords,
			"raw_xml": act.Data.RawXML,
		}
		_, err := a.addPerson(act.TableName, personData)
		return err
		
	case "population.person.update":
		var act struct {
			TableName string `json:"tableName"`
			PersonID  string `json:"personId"`
			PlanXML   string `json:"planXML,omitempty"`
			Coords    string `json:"coords,omitempty"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal update person action: %w", err)
		}
		if act.PlanXML != "" {
			// Call internal function with coordinate extraction logic
			_, err := a.updatePersonPlan(act.TableName, act.PersonID, act.PlanXML)
			return err
		}
		if act.Coords != "" {
			return a.db.UpdatePersonCoords(act.TableName, act.PersonID, act.Coords)
		}
		return nil
		
	case "population.person.delete":
		var act struct {
			TableName string `json:"tableName"`
			PersonID  string `json:"personId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete person action: %w", err)
		}
		_, err := a.deletePerson(act.TableName, act.PersonID)
		return err
		
	case "population.person.batchUpdate":
		var act struct {
			TableName string `json:"tableName"`
			Updates []struct {
				PersonID string `json:"personId"`
				Coords   string `json:"coords,omitempty"`
				PlanXML  string `json:"planXML,omitempty"`
			} `json:"updates"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal batch update persons action: %w", err)
		}
		var updates []PersonUpdate
		for _, u := range act.Updates {
			updates = append(updates, PersonUpdate{
				ID:     u.PersonID,
				Coords: u.Coords,
				RawXML: u.PlanXML,
			})
		}
		return a.db.BatchUpdatePersons(act.TableName, updates)
		
	// ===== NETWORK ACTIONS =====
	case "network.node.update":
		var act struct {
			ProcessID int     `json:"processId"`
			NodeID    string  `json:"nodeId"`
			X         float64 `json:"x"`
			Y         float64 `json:"y"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal update node action: %w", err)
		}
		return a.db.UpdateNode(act.ProcessID, act.NodeID, act.X, act.Y)
		
	case "network.node.delete":
		var act struct {
			ProcessID            int    `json:"processId"`
			NodeID               string `json:"nodeId"`
			DeleteConnectedLinks bool   `json:"deleteConnectedLinks"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete node action: %w", err)
		}
		return a.db.DeleteNode(act.ProcessID, act.NodeID, act.DeleteConnectedLinks)
		
	case "network.node.batchUpdate":
		var act struct {
			ProcessID int                          `json:"processId"`
			Updates   map[string]struct{
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"updates"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal batch update nodes action: %w", err)
		}
		updates := make(map[string]NodeData)
		for id, data := range act.Updates {
			updates[id] = NodeData{
				ID:     id,
				Coords: fmt.Sprintf("%f,%f", data.X, data.Y),
			}
		}
		return a.db.BatchUpdateNodes(act.ProcessID, updates)
		
	case "network.node.batchDelete":
		var act struct {
			ProcessID            int      `json:"processId"`
			NodeIDs              []string `json:"nodeIds"`
			DeleteConnectedLinks bool     `json:"deleteConnectedLinks"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal batch delete nodes action: %w", err)
		}
		return a.db.BatchDeleteNodes(act.ProcessID, act.NodeIDs, act.DeleteConnectedLinks)
		
	case "network.link.create":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
			FromNode  string `json:"fromNode"`
			ToNode    string `json:"toNode"`
			RawXML    string `json:"rawXML"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal create link action: %w", err)
		}
		return a.db.CreateNetworkLink(act.ProcessID, act.LinkID, act.FromNode, act.ToNode, act.RawXML)
		
	case "network.link.update":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
			RawXML    string `json:"rawXML"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal update link action: %w", err)
		}
		return a.db.UpdateNetworkLink(act.ProcessID, act.LinkID, act.RawXML)
		
	case "network.link.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete link action: %w", err)
		}
		return a.db.DeleteLink(act.ProcessID, act.LinkID)
		
	// ===== PT ACTIONS =====
	case "pt.stop.add":
		var act struct {
			ProcessID int    `json:"processId"`
			Stop      PTStop `json:"stop"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal add stop action: %w", err)
		}
		return a.db.AddPTStop(act.ProcessID, act.Stop)
		
	case "pt.stop.update":
		var act struct {
			ProcessID int          `json:"processId"`
			StopID    string       `json:"stopId"`
			Update    PTStopUpdate `json:"update"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal update stop action: %w", err)
		}
		return a.db.UpdatePTStop(act.ProcessID, act.StopID, act.Update)
		
	case "pt.stop.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			StopID    string `json:"stopId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete stop action: %w", err)
		}
		return a.db.DeletePTStop(act.ProcessID, act.StopID)
		
	case "pt.stop.batchUpdate":
		var act struct {
			ProcessID int                       `json:"processId"`
			Updates   map[string]PTStopUpdate `json:"updates"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal batch update stops action: %w", err)
		}
		return a.db.BatchUpdatePTStops(act.ProcessID, act.Updates)
		
	case "pt.line.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			LineID    string `json:"lineId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete line action: %w", err)
		}
		return a.db.DeletePTLine(act.ProcessID, act.LineID)
		
	case "pt.route.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			RouteID   string `json:"routeId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete route action: %w", err)
		}
		return a.db.DeletePTRoute(act.ProcessID, act.RouteID)
		
	case "pt.routeStop.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			RouteID   string `json:"routeId"`
			StopOrder int    `json:"stopOrder"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete route stop action: %w", err)
		}
		return a.db.DeletePTRouteStop(act.ProcessID, act.RouteID, act.StopOrder)
		
	case "pt.departure.delete":
		var act struct {
			ProcessID   int    `json:"processId"`
			DepartureID string `json:"departureId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete departure action: %w", err)
		}
		return a.db.DeletePTDeparture(act.ProcessID, act.DepartureID)
		
	// ===== PROCESS ACTIONS =====
	case "process.process.delete":
		var act struct {
			ProcessID int `json:"processId"`
		}
		if err := json.Unmarshal(actionJSON, &act); err != nil {
			return fmt.Errorf("failed to unmarshal delete process action: %w", err)
		}
		return a.deleteProcess(act.ProcessID)
		
	default:
		return fmt.Errorf("unknown action: %s", key)
	}
}

// SyncChanges applies a batch of changes from the frontend
func (a *App) SyncChanges(actions []json.RawMessage) error {
	return a.db.WithTransaction(func(tx *sql.Tx) error {
		for i, action := range actions {
			if err := a.ExecuteAction(action); err != nil {
				return fmt.Errorf("action %d failed: %w", i, err)
			}
		}
		return nil
	})
}