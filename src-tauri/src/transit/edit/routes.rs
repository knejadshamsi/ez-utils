use rusqlite::{params, OptionalExtension};
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::super::queries::RouteDeletePreview;
use super::{line_exists, mark_dirty, open_transit_connection, route_exists, sqlite_err};

fn validate_route_id(id: &str) -> Result<(), EzError> {
    if id.is_empty() {
        return Err(EzError::TransitRouteIdInvalid {
            message: "Route ID must not be empty.".into(),
        });
    }
    if id != id.trim() {
        return Err(EzError::TransitRouteIdInvalid {
            message: "Route ID must not start or end with whitespace.".into(),
        });
    }
    Ok(())
}

#[tauri::command]
pub fn create_route_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    description: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    validate_route_id(&route_id)?;
    let connection = open_transit_connection(&manager, &source_name)?;
    if !line_exists(&connection, &line_id)? {
        return Err(EzError::TransitLineNotFound {
            line_id: line_id.clone(),
        });
    }
    if route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitDuplicateRouteId { line_id, route_id });
    }
    // Mode is owned by the line; inherit it for the new route.
    let line_mode: Option<String> = connection
        .query_row(
            "SELECT transport_mode FROM lines WHERE id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .optional()
        .map_err(sqlite_err)?;
    let mode = line_mode.ok_or(EzError::TransitTransportModeInvalid {
        message: "Line has no transport mode; set it on the line first.".into(),
    })?;
    let description_value = description.and_then(|s| {
        let t = s.trim().to_string();
        if t.is_empty() {
            None
        } else {
            Some(t)
        }
    });
    connection
        .execute(
            "INSERT INTO routes (id, line_id, transport_mode, description, attributes_blob)
             VALUES (?1, ?2, ?3, ?4, NULL)",
            params![route_id, line_id, mode, description_value],
        )
        .map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn update_route_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    new_route_id: Option<String>,
    description: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    let rename_to = match &new_route_id {
        Some(candidate) if candidate != &route_id => {
            validate_route_id(candidate)?;
            if route_exists(&connection, &line_id, candidate)? {
                return Err(EzError::TransitDuplicateRouteId {
                    line_id: line_id.clone(),
                    route_id: candidate.clone(),
                });
            }
            Some(candidate.clone())
        }
        _ => None,
    };

    let description_value = description.map(|s| {
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
            "UPDATE departures SET route_id = ?1 WHERE line_id = ?2 AND route_id = ?3",
            params![new, line_id, route_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE route_path_links SET route_id = ?1 WHERE line_id = ?2 AND route_id = ?3",
            params![new, line_id, route_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE route_profile_stops SET route_id = ?1 WHERE line_id = ?2 AND route_id = ?3",
            params![new, line_id, route_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE routes SET id = ?1 WHERE line_id = ?2 AND id = ?3",
            params![new, line_id, route_id],
        )
        .map_err(sqlite_err)?;
    }

    let final_route_id: &str = rename_to.as_deref().unwrap_or(&route_id);

    if let Some(opt_desc) = description_value {
        tx.execute(
            "UPDATE routes SET description = ?1 WHERE line_id = ?2 AND id = ?3",
            params![opt_desc, line_id, final_route_id],
        )
        .map_err(sqlite_err)?;
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn preview_delete_route_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<RouteDeletePreview, EzError> {
    let connection = open_transit_connection(&manager, &source_name)?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    let profile_stops: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_profile_stops WHERE line_id = ?1 AND route_id = ?2",
            params![line_id, route_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let path_links: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_path_links WHERE line_id = ?1 AND route_id = ?2",
            params![line_id, route_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let departures: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM departures WHERE line_id = ?1 AND route_id = ?2",
            params![line_id, route_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;

    Ok(RouteDeletePreview {
        profile_stops,
        path_links,
        departures,
    })
}

#[tauri::command]
pub fn delete_route_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    let tx = connection.transaction().map_err(sqlite_err)?;

    tx.execute(
        "DELETE FROM departures WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM route_path_links WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM route_profile_stops WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM routes WHERE line_id = ?1 AND id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
