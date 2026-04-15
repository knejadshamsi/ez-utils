use std::path::Path;

use rusqlite::Connection;

use crate::ez::{fs::db_file_path, types::EzError};

const NETWORK_SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS nodes (
  id      TEXT PRIMARY KEY,
  lng     REAL NOT NULL,
  lat     REAL NOT NULL,
  raw_xml TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS links (
  id              TEXT PRIMARY KEY,
  from_node       TEXT NOT NULL,
  to_node         TEXT NOT NULL,
  tag_blob        TEXT NOT NULL,
  attributes_blob TEXT
);
"#;

pub(crate) fn create_network_db(
    work_dir: &Path,
    source_name: &str,
) -> Result<(), EzError> {
    let connection = Connection::open(db_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to create SQLite database: {err}"),
        }
    })?;
    connection
        .execute_batch(NETWORK_SCHEMA)
        .map_err(|err| EzError::Io {
            message: format!("Failed to initialize network schema: {err}"),
        })?;
    Ok(())
}
