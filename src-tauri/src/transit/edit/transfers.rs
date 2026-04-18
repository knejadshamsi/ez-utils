use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::{mark_dirty, open_transit_connection, sqlite_err};

/// Insert or update a `minimal_transfer_times` row. Source-wide edit.
#[tauri::command]
pub fn upsert_transit_transfer_cmd(
    source_name: String,
    from_stop: String,
    to_stop: String,
    transfer_time: f64,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    if from_stop.trim().is_empty() || to_stop.trim().is_empty() {
        return Err(EzError::Io {
            message: "from_stop and to_stop must be non-empty.".into(),
        });
    }
    if from_stop == to_stop {
        return Err(EzError::Io {
            message: "from_stop and to_stop must differ.".into(),
        });
    }
    if !transfer_time.is_finite() || transfer_time < 0.0 {
        return Err(EzError::Io {
            message: "transfer_time must be a non-negative number.".into(),
        });
    }
    let connection = open_transit_connection(&manager, &source_name)?;
    // Validate both stops exist.
    for id in [&from_stop, &to_stop] {
        let exists: i64 = connection
            .query_row(
                "SELECT COUNT(*) FROM stop_facilities WHERE id = ?1",
                params![id],
                |row| row.get(0),
            )
            .map_err(sqlite_err)?;
        if exists == 0 {
            return Err(EzError::TransitStopFacilityNotFound {
                stop_id: id.clone(),
            });
        }
    }
    connection
        .execute(
            "INSERT INTO minimal_transfer_times (from_stop, to_stop, transfer_time)
             VALUES (?1, ?2, ?3)
             ON CONFLICT(from_stop, to_stop) DO UPDATE SET transfer_time = excluded.transfer_time",
            params![from_stop, to_stop, transfer_time],
        )
        .map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn delete_transit_transfer_cmd(
    source_name: String,
    from_stop: String,
    to_stop: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_transit_connection(&manager, &source_name)?;
    connection
        .execute(
            "DELETE FROM minimal_transfer_times WHERE from_stop = ?1 AND to_stop = ?2",
            params![from_stop, to_stop],
        )
        .map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}
