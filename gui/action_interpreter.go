package gui

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// ExecuteAction executes a single action from the frontend
func (a *App) ExecuteAction(actionJSON json.RawMessage) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	var typeInfo struct {
		Type        string `json:"type"`
		ElementType string `json:"elementType"`
		Action      string `json:"action"`
	}

	json.Unmarshal(actionJSON, &typeInfo)
	key := fmt.Sprintf("%s.%s.%s", typeInfo.Type, typeInfo.ElementType, typeInfo.Action)

	switch key {
	// ===== POPULATION ACTIONS =====
	case "population.person.add":
		var act struct {
			TableName string `json:"tableName"`
			Data      struct {
				ID     string  `json:"id"`
				Lng    float64 `json:"lng"`
				Lat    float64 `json:"lat"`
				RawXML string  `json:"rawXML"`
			} `json:"data"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.AddPerson(act.TableName, act.Data.ID, act.Data.Lng, act.Data.Lat, act.Data.RawXML)

	case "population.person.update":
		var act struct {
			TableName string  `json:"tableName"`
			PersonID  string  `json:"personId"`
			PlanXML   string  `json:"planXML,omitempty"`
			Lng       float64 `json:"lng,omitempty"`
			Lat       float64 `json:"lat,omitempty"`
		}
		json.Unmarshal(actionJSON, &act)
		if act.PlanXML != "" {
			_, err := a.updatePersonPlan(act.TableName, act.PersonID, act.PlanXML)
			return err
		}
		return a.db.UpdatePersonCoords(act.TableName, act.PersonID, act.Lng, act.Lat)

	case "population.person.delete":
		var act struct {
			TableName string `json:"tableName"`
			PersonID  string `json:"personId"`
		}
		json.Unmarshal(actionJSON, &act)
		_, err := a.deletePerson(act.TableName, act.PersonID)
		return err

	case "population.person.batchUpdate":
		var act struct {
			TableName string `json:"tableName"`
			Updates   []struct {
				PersonID string  `json:"personId"`
				Lng      float64 `json:"lng,omitempty"`
				Lat      float64 `json:"lat,omitempty"`
				PlanXML  string  `json:"planXML,omitempty"`
			} `json:"updates"`
		}
		json.Unmarshal(actionJSON, &act)
		var updates []PersonUpdate
		for _, u := range act.Updates {
			updates = append(updates, PersonUpdate{
				ID:     u.PersonID,
				Lng:    u.Lng,
				Lat:    u.Lat,
				RawXML: u.PlanXML,
			})
		}
		return a.db.BatchUpdatePersons(act.TableName, updates)

	// ===== NETWORK ACTIONS =====
	case "network.node.update":
		var act struct {
			ProcessID int     `json:"processId"`
			NodeID    string  `json:"nodeId"`
			Lng       float64 `json:"lng"`
			Lat       float64 `json:"lat"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.UpdateNode(act.ProcessID, act.NodeID, act.Lng, act.Lat)

	case "network.node.delete":
		var act struct {
			ProcessID            int    `json:"processId"`
			NodeID               string `json:"nodeId"`
			DeleteConnectedLinks bool   `json:"deleteConnectedLinks"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.DeleteNode(act.ProcessID, act.NodeID, act.DeleteConnectedLinks)

	case "network.node.batchUpdate":
		var act struct {
			ProcessID int `json:"processId"`
			Updates   map[string]struct {
				Lng float64 `json:"lng"`
				Lat float64 `json:"lat"`
			} `json:"updates"`
		}
		json.Unmarshal(actionJSON, &act)
		updates := make(map[string]NodeData)
		for id, data := range act.Updates {
			updates[id] = NodeData{
				ID:  id,
				Lng: data.Lng,
				Lat: data.Lat,
			}
		}
		return a.db.BatchUpdateNodes(act.ProcessID, updates)

	case "network.node.batchDelete":
		var act struct {
			ProcessID            int      `json:"processId"`
			NodeIDs              []string `json:"nodeIds"`
			DeleteConnectedLinks bool     `json:"deleteConnectedLinks"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.BatchDeleteNodes(act.ProcessID, act.NodeIDs, act.DeleteConnectedLinks)

	case "network.node.create":
		var act struct {
			ProcessID int     `json:"processId"`
			NodeID    string  `json:"nodeId"`
			Lng       float64 `json:"lng"`
			Lat       float64 `json:"lat"`
			RawXML    string  `json:"rawXML"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.InsertNetworkNode(act.ProcessID, act.NodeID, act.Lng, act.Lat, act.RawXML)

	case "network.link.create":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
			FromNode  string `json:"fromNode"`
			ToNode    string `json:"toNode"`
			RawXML    string `json:"rawXML"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.CreateNetworkLink(act.ProcessID, act.LinkID, act.FromNode, act.ToNode, act.RawXML)

	case "network.link.update":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
			RawXML    string `json:"rawXML"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.UpdateNetworkLink(act.ProcessID, act.LinkID, act.RawXML)

	case "network.link.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			LinkID    string `json:"linkId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.db.DeleteLink(act.ProcessID, act.LinkID)

	// ===== PT ACTIONS =====
	case "pt.line.save":
		var act struct {
			ProcessID int  `json:"processId"`
			Line      Line `json:"line"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.SavePTLines(act.ProcessID, []Line{act.Line})

	case "pt.route.save":
		var act struct {
			ProcessID int   `json:"processId"`
			Route     Route `json:"route"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.SavePTRoutes(act.ProcessID, []Route{act.Route})

	case "pt.departure.save":
		var act struct {
			ProcessID int       `json:"processId"`
			Departure Departure `json:"departure"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.SavePTDepartures(act.ProcessID, []Departure{act.Departure})

	case "pt.stop.save":
		var act struct {
			ProcessID int  `json:"processId"`
			Stop      Stop `json:"stop"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.SavePTStops(act.ProcessID, []Stop{act.Stop})

	case "pt.routeStop.save":
		var act struct {
			ProcessID int       `json:"processId"`
			RouteStop RouteStop `json:"routeStop"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.SavePTRouteStops(act.ProcessID, []RouteStop{act.RouteStop})

	case "pt.routeStop.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			StopID    string `json:"stopId"`
			RouteID   string `json:"routeId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.RemovePTStopsFromRoute(act.ProcessID, act.RouteID, []string{act.StopID})


	case "pt.line.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			LineID    string `json:"lineId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.DeletePTLines(act.ProcessID, []string{act.LineID})

	case "pt.route.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			RouteID   string `json:"routeId"`
		}
		json.Unmarshal(actionJSON, &act)
		// With CASCADE DELETE, just delete the route
		// Stops and departures will be automatically deleted
		return a.DeletePTRoutes(act.ProcessID, []string{act.RouteID})

	case "pt.departure.delete":
		var act struct {
			ProcessID   int    `json:"processId"`
			DepartureID string `json:"departureId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.DeletePTDepartures(act.ProcessID, []string{act.DepartureID})

	case "pt.stop.delete":
		var act struct {
			ProcessID int    `json:"processId"`
			StopID    string `json:"stopId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.DeletePTStops(act.ProcessID, []string{act.StopID})

	// ===== PROCESS ACTIONS =====
	case "process.process.delete":
		var act struct {
			ProcessID int `json:"processId"`
		}
		json.Unmarshal(actionJSON, &act)
		return a.deleteProcess(act.ProcessID)

	default:
		return fmt.Errorf("unknown action: %s", key)
	}
}

// SyncChanges applies a batch of changes from the frontend
func (a *App) SyncChanges(actions []json.RawMessage) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return a.db.WithTransaction(func(tx *sql.Tx) error {
		for _, action := range actions {
			if err := a.ExecuteAction(action); err != nil {
				return err
			}
		}
		return nil
	})
}