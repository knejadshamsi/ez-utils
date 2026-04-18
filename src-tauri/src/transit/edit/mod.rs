//! Edit-side commands for transit lines and routes (create, rename/update,
//! delete with cascade + derived_transit_modes recompute).
//! Stop facilities are never deleted here - top-level MATSim elements are shared.

use std::path::Path;

use rusqlite::{params, Connection};
use tauri::State;

use crate::ez::{
    fs::db_file_path,
    types::{EzError, SessionManager},
};

use super::queries::resolve_transit_work_dir;
use super::schema::ensure_transit_schema;

pub(crate) mod departures;
pub(crate) mod lines;
pub(crate) mod path_links;
pub(crate) mod profile_stops;
pub(crate) mod routes;
pub(crate) mod stop_facilities;
pub(crate) mod transfers;

pub use departures::{apply_route_departures_edits_cmd, list_transit_vehicles_cmd};
pub use lines::{create_line_cmd, delete_line_cmd, preview_delete_line_cmd, update_line_cmd};
pub use path_links::apply_route_path_links_cmd;
pub use profile_stops::apply_route_profile_edits_cmd;
pub use routes::{
    create_route_cmd, delete_route_cmd, preview_delete_route_cmd, update_route_cmd,
};
pub use stop_facilities::{
    create_stop_facility_cmd, delete_stop_facility_cmd, preview_delete_stop_facility_cmd,
    update_stop_facility_cmd,
};
pub use transfers::{delete_transit_transfer_cmd, upsert_transit_transfer_cmd};

pub(super) fn open_transit_connection(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<Connection, EzError> {
    let work_dir = resolve_transit_work_dir(manager, source_name)?;
    let conn = open_transit_db(&work_dir, source_name)?;
    ensure_transit_schema(&conn)?;
    Ok(conn)
}

fn open_transit_db(work_dir: &Path, source_name: &str) -> Result<Connection, EzError> {
    Connection::open(db_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })
}

pub(super) fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Transit edit failed: {err}"),
    }
}

pub(super) fn mark_dirty(manager: &State<'_, SessionManager>) -> Result<(), EzError> {
    let mut guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;
    session.dirty = true;
    Ok(())
}

pub(super) fn line_exists(connection: &Connection, line_id: &str) -> Result<bool, EzError> {
    let exists: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM lines WHERE id = ?1",
            params![line_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    Ok(exists > 0)
}

pub(super) fn route_exists(
    connection: &Connection,
    line_id: &str,
    route_id: &str,
) -> Result<bool, EzError> {
    let exists: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM routes WHERE line_id = ?1 AND id = ?2",
            params![line_id, route_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    Ok(exists > 0)
}
