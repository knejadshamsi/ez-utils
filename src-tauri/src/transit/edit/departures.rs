use std::collections::HashMap;

use rusqlite::{params, Connection};
use tauri::State;

use crate::ez::{
    fs::{db_file_path, vehicles_db_file_path},
    types::{EzError, SessionManager, SourceKind},
};

use super::super::queries::{DepartureEditInput, ListedVehicle};
use super::super::queries::resolve_transit_work_dir;
use super::{mark_dirty, route_exists, sqlite_err};

/// Replace the full departure list for a route in one transaction. Preserves
/// existing `attributes_blob` per-id: if the user kept a departure with the same
/// id, its custom XML attrs survive. New/renamed ids get NULL attrs_blob.
#[tauri::command]
pub fn apply_route_departures_edits_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    departures: Vec<DepartureEditInput>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    let mut connection = Connection::open(db_file_path(&work_dir, &source_name))
        .map_err(|err| EzError::Io {
            message: format!("Failed to open transit SQLite database: {err}"),
        })?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    // Validate every incoming id non-empty.
    for (idx, dep) in departures.iter().enumerate() {
        if dep.id.trim().is_empty() {
            return Err(EzError::Io {
                message: format!("Departure at index {idx} has empty id."),
            });
        }
        if dep.departure_time.trim().is_empty() {
            return Err(EzError::Io {
                message: format!(
                    "Departure '{}' at index {idx} has empty departureTime.",
                    dep.id
                ),
            });
        }
    }

    let tx = connection.transaction().map_err(sqlite_err)?;

    // Snapshot old attributes_blob by id so we can preserve on match.
    let old_attrs: HashMap<String, Option<String>> = {
        let mut stmt = tx
            .prepare(
                "SELECT id, attributes_blob FROM departures
                 WHERE line_id = ?1 AND route_id = ?2",
            )
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map(params![line_id, route_id], |row| {
                Ok((row.get::<_, String>(0)?, row.get::<_, Option<String>>(1)?))
            })
            .map_err(sqlite_err)?;
        rows.collect::<Result<HashMap<_, _>, _>>().map_err(sqlite_err)?
    };

    tx.execute(
        "DELETE FROM departures WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;

    {
        let mut insert = tx
            .prepare(
                "INSERT INTO departures
                 (id, line_id, route_id, departure_time, vehicle_ref_id, attributes_blob)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            )
            .map_err(sqlite_err)?;
        for dep in departures.iter() {
            let id = dep.id.trim();
            let time = dep.departure_time.trim();
            let vehicle = dep
                .vehicle_ref_id
                .as_ref()
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty());
            let attrs = old_attrs.get(id).cloned().unwrap_or(None);
            insert
                .execute(params![id, line_id, route_id, time, vehicle, attrs])
                .map_err(sqlite_err)?;
        }
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

/// List vehicles from the optional `.vehicles.db` sidecar. Returns an empty vec
/// when no vehicles file is attached to this source.
#[tauri::command]
pub fn list_transit_vehicles_cmd(
    source_name: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedVehicle>, EzError> {
    // Validate source exists and is transit; also gate on vehicles metadata being present.
    let work_dir = {
        let guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
        let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
        let entry = session
            .ui_state
            .sources
            .iter()
            .find(|e| e.name == source_name)
            .ok_or_else(|| EzError::SourceNotFound {
                name: source_name.clone(),
            })?;
        if entry.kind != SourceKind::Transit {
            return Err(EzError::UnsupportedSource {
                message: format!("Source '{source_name}' is not a transit source."),
            });
        }
        let has_vehicles = entry
            .transit_metadata
            .as_ref()
            .and_then(|m| m.vehicles.as_ref())
            .is_some();
        if !has_vehicles {
            return Ok(Vec::new());
        }
        session.work_dir.clone()
    };

    let vehicles_path = vehicles_db_file_path(&work_dir, &source_name);
    if !vehicles_path.exists() {
        return Ok(Vec::new());
    }

    // Collect the set of vehicle ids referenced by any departure in the transit
    // schedule DB, so the UI can mark "in use" vehicles separately from ones
    // that the attached vehicles file defines but nothing references yet.
    let schedule_conn = Connection::open(db_file_path(&work_dir, &source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to open transit SQLite database: {err}"),
        }
    })?;
    let referenced_ids: std::collections::HashSet<String> = {
        let mut stmt = schedule_conn
            .prepare(
                "SELECT DISTINCT vehicle_ref_id FROM departures
                 WHERE vehicle_ref_id IS NOT NULL",
            )
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map([], |row| row.get::<_, String>(0))
            .map_err(sqlite_err)?;
        let mut set = std::collections::HashSet::new();
        for r in rows {
            set.insert(r.map_err(sqlite_err)?);
        }
        set
    };

    let connection = Connection::open(&vehicles_path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit vehicles SQLite database: {err}"),
    })?;
    let mut stmt = connection
        .prepare("SELECT id, type FROM vehicles ORDER BY id")
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([], |row| {
            let id: String = row.get(0)?;
            let vehicle_type: Option<String> = row.get(1)?;
            Ok((id, vehicle_type))
        })
        .map_err(sqlite_err)?;
    let mut out = Vec::new();
    for row in rows {
        let (id, vehicle_type) = row.map_err(sqlite_err)?;
        let referenced = referenced_ids.contains(&id);
        out.push(ListedVehicle {
            id,
            vehicle_type,
            referenced,
        });
    }
    Ok(out)
}
