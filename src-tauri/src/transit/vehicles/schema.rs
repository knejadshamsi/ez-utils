use std::path::Path;

use rusqlite::Connection;

use crate::ez::{fs::vehicles_db_file_path, types::EzError};

const VEHICLES_SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS vehicle_types (
  id      TEXT PRIMARY KEY,
  raw_xml TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicles (
  id      TEXT PRIMARY KEY,
  type    TEXT,
  raw_xml TEXT NOT NULL
);
"#;

pub(crate) fn create_transit_vehicles_db(
    work_dir: &Path,
    source_name: &str,
) -> Result<(), EzError> {
    let connection = Connection::open(vehicles_db_file_path(work_dir, source_name)).map_err(
        |err| EzError::Io {
            message: format!("Failed to create SQLite database: {err}"),
        },
    )?;
    connection
        .execute_batch(VEHICLES_SCHEMA)
        .map_err(|err| EzError::Io {
            message: format!("Failed to initialize transit vehicles schema: {err}"),
        })?;
    Ok(())
}
