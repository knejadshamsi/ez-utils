use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::super::queries::ProfileStopEditInput;
use super::{mark_dirty, open_transit_connection, route_exists, sqlite_err};

/// Replace the full ordered list of profile stops for a route in one transaction.
/// Frontend sends the authoritative sequence; backend wipes existing rows and
/// re-inserts with `sequence = index`. Missing `stop_ref_id` values reject.
#[tauri::command]
pub fn apply_route_profile_edits_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    stops: Vec<ProfileStopEditInput>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    // Validate every referenced stop_facility exists before touching anything.
    for (idx, stop) in stops.iter().enumerate() {
        if stop.stop_ref_id.trim().is_empty() {
            return Err(EzError::TransitProfileStopInvalid {
                message: format!("Profile stop at index {idx} has empty stopRefId."),
            });
        }
        let exists: i64 = connection
            .query_row(
                "SELECT COUNT(*) FROM stop_facilities WHERE id = ?1",
                params![stop.stop_ref_id],
                |row| row.get(0),
            )
            .map_err(sqlite_err)?;
        if exists == 0 {
            return Err(EzError::TransitStopFacilityNotFound {
                stop_id: stop.stop_ref_id.clone(),
            });
        }
    }

    let tx = connection.transaction().map_err(sqlite_err)?;

    // Snapshot the existing ordered stop set so we can tell whether the
    // sequence changed. If it did, the current `route_path_links` (network-linked
    // path) is no longer trustworthy and must be wiped - the user has moved or
    // added stops that the stored link sequence doesn't account for. Flag/offset
    // tweaks alone preserve the link path.
    let old_stop_refs: Vec<String> = {
        let mut stmt = tx
            .prepare(
                "SELECT stop_ref_id FROM route_profile_stops
                 WHERE line_id = ?1 AND route_id = ?2 ORDER BY sequence",
            )
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map(params![line_id, route_id], |row| row.get::<_, String>(0))
            .map_err(sqlite_err)?;
        rows.collect::<Result<Vec<_>, _>>().map_err(sqlite_err)?
    };
    let new_stop_refs: Vec<String> = stops
        .iter()
        .map(|s| s.stop_ref_id.trim().to_string())
        .collect();
    let stop_set_changed = old_stop_refs != new_stop_refs;

    tx.execute(
        "DELETE FROM route_profile_stops WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;

    {
        let mut insert = tx
            .prepare(
                "INSERT INTO route_profile_stops
                 (line_id, route_id, sequence, stop_ref_id, arrival_offset, departure_offset,
                  allow_boarding, allow_alighting, await_departure)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)",
            )
            .map_err(sqlite_err)?;
        for (idx, stop) in stops.iter().enumerate() {
            let arrival = stop
                .arrival_offset
                .as_ref()
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty());
            let departure = stop
                .departure_offset
                .as_ref()
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty());
            insert
                .execute(params![
                    line_id,
                    route_id,
                    idx as i64,
                    stop.stop_ref_id.trim(),
                    arrival,
                    departure,
                    stop.allow_boarding as i32,
                    stop.allow_alighting as i32,
                    stop.await_departure as i32,
                ])
                .map_err(sqlite_err)?;
        }
    }

    if stop_set_changed {
        tx.execute(
            "DELETE FROM route_path_links WHERE line_id = ?1 AND route_id = ?2",
            params![line_id, route_id],
        )
        .map_err(sqlite_err)?;
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
