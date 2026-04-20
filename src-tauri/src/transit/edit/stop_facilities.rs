use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::super::queries::StopFacilityDeletePreview;
use super::{mark_dirty, open_transit_connection, sqlite_err};

fn validate_stop_id(id: &str) -> Result<(), EzError> {
    if id.is_empty() {
        return Err(EzError::TransitStopFacilityIdInvalid {
            message: "Stop facility ID must not be empty.".into(),
        });
    }
    if id != id.trim() {
        return Err(EzError::TransitStopFacilityIdInvalid {
            message: "Stop facility ID must not start or end with whitespace.".into(),
        });
    }
    Ok(())
}

fn stop_facility_exists(
    connection: &rusqlite::Connection,
    stop_id: &str,
) -> Result<bool, EzError> {
    let count: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM stop_facilities WHERE id = ?1",
            params![stop_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    Ok(count > 0)
}

#[tauri::command]
#[allow(clippy::too_many_arguments)]
pub fn update_stop_facility_cmd(
    source_name: String,
    stop_id: String,
    new_id: Option<String>,
    name: Option<String>,
    lng: Option<f64>,
    lat: Option<f64>,
    link_ref_id: Option<String>,
    stop_area_id: Option<String>,
    is_blocking: Option<bool>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !stop_facility_exists(&connection, &stop_id)? {
        return Err(EzError::TransitStopFacilityNotFound { stop_id });
    }

    let rename_to = match &new_id {
        Some(candidate) if candidate != &stop_id => {
            validate_stop_id(candidate)?;
            if stop_facility_exists(&connection, candidate)? {
                return Err(EzError::TransitDuplicateStopFacilityId {
                    stop_id: candidate.clone(),
                });
            }
            Some(candidate.clone())
        }
        _ => None,
    };

    let name_value = name.map(|s| {
        let t = s.trim().to_string();
        if t.is_empty() {
            None
        } else {
            Some(t)
        }
    });

    let tx = connection.transaction().map_err(sqlite_err)?;

    if let Some(new) = &rename_to {
        tx.execute(
            "UPDATE route_profile_stops SET stop_ref_id = ?1 WHERE stop_ref_id = ?2",
            params![new, stop_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE minimal_transfer_times SET from_stop = ?1 WHERE from_stop = ?2",
            params![new, stop_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE minimal_transfer_times SET to_stop = ?1 WHERE to_stop = ?2",
            params![new, stop_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE stop_facilities SET id = ?1 WHERE id = ?2",
            params![new, stop_id],
        )
        .map_err(sqlite_err)?;
    }

    let final_id = rename_to.as_deref().unwrap_or(&stop_id);

    if let Some(opt_name) = name_value {
        tx.execute(
            "UPDATE stop_facilities SET name = ?1 WHERE id = ?2",
            params![opt_name, final_id],
        )
        .map_err(sqlite_err)?;
    }

    if let (Some(new_lng), Some(new_lat)) = (lng, lat) {
        tx.execute(
            "UPDATE stop_facilities SET lng = ?1, lat = ?2 WHERE id = ?3",
            params![new_lng, new_lat, final_id],
        )
        .map_err(sqlite_err)?;
    } else if let Some(new_lng) = lng {
        tx.execute(
            "UPDATE stop_facilities SET lng = ?1 WHERE id = ?2",
            params![new_lng, final_id],
        )
        .map_err(sqlite_err)?;
    } else if let Some(new_lat) = lat {
        tx.execute(
            "UPDATE stop_facilities SET lat = ?1 WHERE id = ?2",
            params![new_lat, final_id],
        )
        .map_err(sqlite_err)?;
    }

    // link_ref_id: accept Some(non-empty) to set, Some(empty) to clear, None to leave unchanged.
    if let Some(raw) = link_ref_id {
        let trimmed = raw.trim().to_string();
        let stored: Option<String> = if trimmed.is_empty() { None } else { Some(trimmed) };
        tx.execute(
            "UPDATE stop_facilities SET link_ref_id = ?1 WHERE id = ?2",
            params![stored, final_id],
        )
        .map_err(sqlite_err)?;
    }

    if let Some(raw) = stop_area_id {
        let trimmed = raw.trim().to_string();
        let stored: Option<String> = if trimmed.is_empty() { None } else { Some(trimmed) };
        tx.execute(
            "UPDATE stop_facilities SET stop_area_id = ?1 WHERE id = ?2",
            params![stored, final_id],
        )
        .map_err(sqlite_err)?;
    }

    if let Some(flag) = is_blocking {
        tx.execute(
            "UPDATE stop_facilities SET is_blocking = ?1 WHERE id = ?2",
            params![flag as i32, final_id],
        )
        .map_err(sqlite_err)?;
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
#[allow(clippy::too_many_arguments)]
pub fn create_stop_facility_cmd(
    source_name: String,
    stop_id: String,
    lng: f64,
    lat: f64,
    name: Option<String>,
    link_ref_id: Option<String>,
    stop_area_id: Option<String>,
    is_blocking: Option<bool>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    validate_stop_id(&stop_id)?;
    let connection = open_transit_connection(&manager, &source_name)?;
    if stop_facility_exists(&connection, &stop_id)? {
        return Err(EzError::TransitDuplicateStopFacilityId { stop_id });
    }
    let name_value = trim_to_option(name);
    let link_value = trim_to_option(link_ref_id);
    let area_value = trim_to_option(stop_area_id);
    let blocking = is_blocking.unwrap_or(false);
    connection
        .execute(
            "INSERT INTO stop_facilities
             (id, lng, lat, name, link_ref_id, stop_area_id, is_blocking, attributes_blob)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, NULL)",
            params![stop_id, lng, lat, name_value, link_value, area_value, blocking as i32],
        )
        .map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

fn trim_to_option(raw: Option<String>) -> Option<String> {
    raw.and_then(|s| {
        let t = s.trim().to_string();
        if t.is_empty() { None } else { Some(t) }
    })
}

#[tauri::command]
pub fn preview_delete_stop_facility_cmd(
    source_name: String,
    stop_id: String,
    manager: State<'_, SessionManager>,
) -> Result<StopFacilityDeletePreview, EzError> {
    let connection = open_transit_connection(&manager, &source_name)?;
    if !stop_facility_exists(&connection, &stop_id)? {
        return Err(EzError::TransitStopFacilityNotFound { stop_id });
    }
    let profile_stops: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_profile_stops WHERE stop_ref_id = ?1",
            params![stop_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let routes: i64 = connection
        .query_row(
            "SELECT COUNT(DISTINCT line_id || ':::' || route_id)
             FROM route_profile_stops WHERE stop_ref_id = ?1",
            params![stop_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let transfers: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM minimal_transfer_times
             WHERE from_stop = ?1 OR to_stop = ?1",
            params![stop_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    Ok(StopFacilityDeletePreview {
        stop_id,
        profile_stops,
        routes,
        transfers,
    })
}

#[tauri::command]
pub fn delete_stop_facility_cmd(
    source_name: String,
    stop_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !stop_facility_exists(&connection, &stop_id)? {
        return Err(EzError::TransitStopFacilityNotFound { stop_id });
    }
    let tx = connection.transaction().map_err(sqlite_err)?;
    // After deleting this stop's profile-stop rows, some routes may have gaps
    // in their sequence numbering. Compact per affected (line, route).
    let affected_routes: Vec<(String, String)> = {
        let mut stmt = tx
            .prepare(
                "SELECT DISTINCT line_id, route_id
                 FROM route_profile_stops WHERE stop_ref_id = ?1",
            )
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map(params![stop_id], |row| {
                Ok((row.get::<_, String>(0)?, row.get::<_, String>(1)?))
            })
            .map_err(sqlite_err)?;
        rows.collect::<Result<Vec<_>, _>>().map_err(sqlite_err)?
    };

    tx.execute(
        "DELETE FROM route_profile_stops WHERE stop_ref_id = ?1",
        params![stop_id],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM minimal_transfer_times WHERE from_stop = ?1 OR to_stop = ?1",
        params![stop_id],
    )
    .map_err(sqlite_err)?;
    tx.execute("DELETE FROM stop_facilities WHERE id = ?1", params![stop_id])
        .map_err(sqlite_err)?;

    // Compact sequences per affected route so they stay 0-based contiguous.
    for (line_id, route_id) in affected_routes {
        let ids: Vec<(i64, i64)> = {
            let mut stmt = tx
                .prepare(
                    "SELECT sequence FROM route_profile_stops
                     WHERE line_id = ?1 AND route_id = ?2
                     ORDER BY sequence",
                )
                .map_err(sqlite_err)?;
            let rows = stmt
                .query_map(params![line_id, route_id], |row| row.get::<_, i64>(0))
                .map_err(sqlite_err)?;
            rows.enumerate()
                .map(|(new_idx, old)| old.map(|o| (o, new_idx as i64)))
                .collect::<Result<Vec<_>, _>>()
                .map_err(sqlite_err)?
        };
        for (old_seq, new_seq) in ids {
            if old_seq == new_seq {
                continue;
            }
            tx.execute(
                "UPDATE route_profile_stops SET sequence = ?1
                 WHERE line_id = ?2 AND route_id = ?3 AND sequence = ?4",
                params![new_seq, line_id, route_id, old_seq],
            )
            .map_err(sqlite_err)?;
        }
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
