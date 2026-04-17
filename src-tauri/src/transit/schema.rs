use std::path::Path;

use rusqlite::Connection;

use crate::ez::{fs::db_file_path, types::EzError};

const TRANSIT_SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS stop_facilities (
  id                    TEXT PRIMARY KEY,
  lng                   REAL NOT NULL,
  lat                   REAL NOT NULL,
  name                  TEXT,
  link_ref_id           TEXT,
  stop_area_id          TEXT,
  is_blocking           INTEGER NOT NULL DEFAULT 0,
  attributes_blob       TEXT
);

CREATE TABLE IF NOT EXISTS lines (
  id              TEXT PRIMARY KEY,
  name            TEXT,
  attributes_blob TEXT,
  transport_mode  TEXT
);

CREATE TABLE IF NOT EXISTS routes (
  id              TEXT NOT NULL,
  line_id         TEXT NOT NULL,
  transport_mode  TEXT NOT NULL,
  description     TEXT,
  attributes_blob TEXT,
  PRIMARY KEY (line_id, id)
);

CREATE TABLE IF NOT EXISTS route_profile_stops (
  line_id           TEXT NOT NULL,
  route_id          TEXT NOT NULL,
  sequence          INTEGER NOT NULL,
  stop_ref_id       TEXT NOT NULL,
  arrival_offset    TEXT,
  departure_offset  TEXT,
  allow_boarding    INTEGER NOT NULL DEFAULT 1,
  allow_alighting   INTEGER NOT NULL DEFAULT 1,
  await_departure   INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (line_id, route_id, sequence)
);

CREATE TABLE IF NOT EXISTS route_path_links (
  line_id   TEXT NOT NULL,
  route_id  TEXT NOT NULL,
  sequence  INTEGER NOT NULL,
  link_id   TEXT NOT NULL,
  PRIMARY KEY (line_id, route_id, sequence)
);

CREATE TABLE IF NOT EXISTS departures (
  id              TEXT NOT NULL,
  line_id         TEXT NOT NULL,
  route_id        TEXT NOT NULL,
  departure_time  TEXT NOT NULL,
  vehicle_ref_id  TEXT,
  attributes_blob TEXT,
  PRIMARY KEY (line_id, route_id, id)
);

CREATE TABLE IF NOT EXISTS minimal_transfer_times (
  from_stop     TEXT NOT NULL,
  to_stop       TEXT NOT NULL,
  transfer_time REAL NOT NULL,
  PRIMARY KEY (from_stop, to_stop)
);

CREATE INDEX IF NOT EXISTS idx_rps_stop_ref ON route_profile_stops(stop_ref_id);
CREATE INDEX IF NOT EXISTS idx_routes_mode  ON routes(transport_mode);
"#;

pub(crate) fn create_transit_db(work_dir: &Path, source_name: &str) -> Result<(), EzError> {
    let connection = Connection::open(db_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to create SQLite database: {err}"),
        }
    })?;
    connection
        .execute_batch(TRANSIT_SCHEMA)
        .map_err(|err| EzError::Io {
            message: format!("Failed to initialize transit schema: {err}"),
        })?;
    Ok(())
}

/// Populate `lines.transport_mode` from the first available route's mode.
/// Called at the end of import once all routes are inserted; the column is
/// internal (not exported) but lets lines own a canonical mode.
pub(crate) fn backfill_lines_mode(tx: &rusqlite::Transaction) -> Result<(), EzError> {
    tx.execute(
        "UPDATE lines SET transport_mode = (
             SELECT transport_mode FROM routes
             WHERE routes.line_id = lines.id LIMIT 1
         ) WHERE transport_mode IS NULL",
        [],
    )
    .map_err(|err| EzError::Io {
        message: format!("Failed to back-fill lines.transport_mode: {err}"),
    })?;
    Ok(())
}

/// Idempotent migration for post-initial schema changes. Runs on open.
pub(crate) fn ensure_transit_schema(connection: &Connection) -> Result<(), EzError> {
    let has_lines_mode: bool = connection
        .query_row(
            "SELECT COUNT(*) > 0 FROM pragma_table_info('lines') WHERE name = 'transport_mode'",
            [],
            |row| row.get::<_, bool>(0),
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to inspect transit schema: {err}"),
        })?;
    if !has_lines_mode {
        connection
            .execute("ALTER TABLE lines ADD COLUMN transport_mode TEXT", [])
            .map_err(|err| EzError::Io {
                message: format!("Failed to migrate lines schema: {err}"),
            })?;
        connection
            .execute(
                "UPDATE lines SET transport_mode = (
                     SELECT transport_mode FROM routes
                     WHERE routes.line_id = lines.id LIMIT 1
                 ) WHERE transport_mode IS NULL",
                [],
            )
            .map_err(|err| EzError::Io {
                message: format!("Failed to back-fill lines.transport_mode: {err}"),
            })?;
    }
    Ok(())
}
