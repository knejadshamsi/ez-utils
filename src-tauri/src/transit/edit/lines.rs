use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::super::queries::LineDeletePreview;
use super::{line_exists, mark_dirty, open_transit_connection, sqlite_err};

fn validate_line_id(id: &str) -> Result<(), EzError> {
    if id.is_empty() {
        return Err(EzError::TransitLineIdInvalid {
            message: "Line ID must not be empty.".into(),
        });
    }
    if id != id.trim() {
        return Err(EzError::TransitLineIdInvalid {
            message: "Line ID must not start or end with whitespace.".into(),
        });
    }
    Ok(())
}

fn validate_transport_mode(mode: &str) -> Result<(), EzError> {
    let trimmed = mode.trim();
    if trimmed.is_empty() {
        return Err(EzError::TransitTransportModeInvalid {
            message: "Transport mode must not be empty.".into(),
        });
    }
    Ok(())
}

#[tauri::command]
pub fn create_line_cmd(
    source_name: String,
    line_id: String,
    name: Option<String>,
    transport_mode: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    validate_line_id(&line_id)?;
    validate_transport_mode(&transport_mode)?;
    let connection = open_transit_connection(&manager, &source_name)?;
    if line_exists(&connection, &line_id)? {
        return Err(EzError::TransitDuplicateLineId { line_id });
    }
    let name_value = name.and_then(|s| {
        let t = s.trim().to_string();
        if t.is_empty() {
            None
        } else {
            Some(t)
        }
    });
    connection
        .execute(
            "INSERT INTO lines (id, name, attributes_blob, transport_mode)
             VALUES (?1, ?2, NULL, ?3)",
            params![line_id, name_value, transport_mode.trim()],
        )
        .map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn update_line_cmd(
    source_name: String,
    line_id: String,
    new_id: Option<String>,
    name: Option<String>,
    transport_mode: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !line_exists(&connection, &line_id)? {
        return Err(EzError::TransitLineNotFound { line_id });
    }

    let rename_to = match &new_id {
        Some(candidate) if candidate != &line_id => {
            validate_line_id(candidate)?;
            if line_exists(&connection, candidate)? {
                return Err(EzError::TransitDuplicateLineId {
                    line_id: candidate.clone(),
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

    let new_mode = match &transport_mode {
        Some(mode) => {
            validate_transport_mode(mode)?;
            Some(mode.trim().to_string())
        }
        None => None,
    };

    let tx = connection.transaction().map_err(sqlite_err)?;

    if let Some(new) = &rename_to {
        tx.execute(
            "UPDATE departures SET line_id = ?1 WHERE line_id = ?2",
            params![new, line_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE route_path_links SET line_id = ?1 WHERE line_id = ?2",
            params![new, line_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE route_profile_stops SET line_id = ?1 WHERE line_id = ?2",
            params![new, line_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE routes SET line_id = ?1 WHERE line_id = ?2",
            params![new, line_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE lines SET id = ?1 WHERE id = ?2",
            params![new, line_id],
        )
        .map_err(sqlite_err)?;
    }

    let final_id = rename_to.as_deref().unwrap_or(&line_id);

    if let Some(opt_name) = name_value {
        tx.execute(
            "UPDATE lines SET name = ?1 WHERE id = ?2",
            params![opt_name, final_id],
        )
        .map_err(sqlite_err)?;
    }

    if let Some(mode) = &new_mode {
        tx.execute(
            "UPDATE lines SET transport_mode = ?1 WHERE id = ?2",
            params![mode, final_id],
        )
        .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE routes SET transport_mode = ?1 WHERE line_id = ?2",
            params![mode, final_id],
        )
        .map_err(sqlite_err)?;
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn preview_delete_line_cmd(
    source_name: String,
    line_id: String,
    manager: State<'_, SessionManager>,
) -> Result<LineDeletePreview, EzError> {
    let connection = open_transit_connection(&manager, &source_name)?;
    if !line_exists(&connection, &line_id)? {
        return Err(EzError::TransitLineNotFound { line_id });
    }

    let routes: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM routes WHERE line_id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let profile_stops: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_profile_stops WHERE line_id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let path_links: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_path_links WHERE line_id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    let departures: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM departures WHERE line_id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;

    Ok(LineDeletePreview {
        routes,
        profile_stops,
        path_links,
        departures,
    })
}

#[tauri::command]
pub fn delete_line_cmd(
    source_name: String,
    line_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !line_exists(&connection, &line_id)? {
        return Err(EzError::TransitLineNotFound { line_id });
    }

    let tx = connection.transaction().map_err(sqlite_err)?;

    tx.execute("DELETE FROM departures WHERE line_id = ?1", params![line_id])
        .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM route_path_links WHERE line_id = ?1",
        params![line_id],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM route_profile_stops WHERE line_id = ?1",
        params![line_id],
    )
    .map_err(sqlite_err)?;
    tx.execute("DELETE FROM routes WHERE line_id = ?1", params![line_id])
        .map_err(sqlite_err)?;
    tx.execute("DELETE FROM lines WHERE id = ?1", params![line_id])
        .map_err(sqlite_err)?;

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
