use std::path::{Path, PathBuf};

use rusqlite::Connection;

use crate::ez::{fs::db_file_path, types::EzError};

use super::types::{ListedLine, ListedRoute};

pub(super) fn list_lines(
    work_dir: &Path,
    source_name: &str,
) -> Result<Vec<ListedLine>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT l.id,
                    l.name,
                    l.transport_mode,
                    COUNT(r.id) AS route_count
             FROM lines AS l
             LEFT JOIN routes AS r ON r.line_id = l.id
             GROUP BY l.id, l.name, l.transport_mode
             ORDER BY l.id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare lines query: {err}"),
        })?;

    let rows = stmt
        .query_map([], |row| {
            let mode: Option<String> = row.get(2)?;
            Ok(ListedLine {
                id: row.get(0)?,
                name: row.get(1)?,
                route_count: row.get(3)?,
                modes: mode.map(|m| vec![m]).unwrap_or_default(),
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run lines query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read line row: {err}"),
        })?);
    }
    Ok(out)
}

pub(super) fn list_routes(
    work_dir: &Path,
    source_name: &str,
    line_id: &str,
) -> Result<Vec<ListedRoute>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT r.id,
                    r.line_id,
                    r.transport_mode,
                    r.description,
                    COUNT(rps.sequence) AS stop_count
             FROM routes AS r
             LEFT JOIN route_profile_stops AS rps
                    ON rps.line_id = r.line_id AND rps.route_id = r.id
             WHERE r.line_id = ?1
             GROUP BY r.id, r.line_id, r.transport_mode, r.description
             ORDER BY r.id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare routes query: {err}"),
        })?;

    let rows = stmt
        .query_map([line_id], |row| {
            Ok(ListedRoute {
                id: row.get(0)?,
                line_id: row.get(1)?,
                transport_mode: row.get(2)?,
                description: row.get(3)?,
                stop_count: row.get(4)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run routes query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read route row: {err}"),
        })?);
    }
    Ok(out)
}
