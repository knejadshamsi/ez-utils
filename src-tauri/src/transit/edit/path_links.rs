use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::{mark_dirty, open_transit_connection, route_exists, sqlite_err};

/// Replace the full ordered link-id list for a route's network path. Bulk
/// replacement (wipe + reinsert with sequence = index). The UI composes the
/// new list per pair-of-stops and sends the whole thing.
#[tauri::command]
pub fn apply_route_path_links_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    links: Vec<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut connection = open_transit_connection(&manager, &source_name)?;
    if !route_exists(&connection, &line_id, &route_id)? {
        return Err(EzError::TransitRouteNotFound { line_id, route_id });
    }

    for (idx, link) in links.iter().enumerate() {
        if link.trim().is_empty() {
            return Err(EzError::Io {
                message: format!("Path link at index {idx} is empty."),
            });
        }
    }

    let tx = connection.transaction().map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM route_path_links WHERE line_id = ?1 AND route_id = ?2",
        params![line_id, route_id],
    )
    .map_err(sqlite_err)?;

    {
        let mut insert = tx
            .prepare(
                "INSERT INTO route_path_links (line_id, route_id, sequence, link_id)
                 VALUES (?1, ?2, ?3, ?4)",
            )
            .map_err(sqlite_err)?;
        for (idx, link) in links.iter().enumerate() {
            insert
                .execute(params![line_id, route_id, idx as i64, link.trim()])
                .map_err(sqlite_err)?;
        }
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
