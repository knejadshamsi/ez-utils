use std::path::Path;

use rusqlite::Connection;

use crate::ez::{fs::db_file_path, types::EzError};

const POPULATION_SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS person_attributes (
  person_id       TEXT PRIMARY KEY,
  attributes_blob TEXT
);

CREATE TABLE IF NOT EXISTS person_plans (
  person_id   TEXT NOT NULL,
  plan_id     TEXT NOT NULL,
  plan_index  INTEGER NOT NULL,
  selected    INTEGER NOT NULL,
  plan_blob   TEXT NOT NULL,
  PRIMARY KEY (person_id, plan_id)
);

CREATE TABLE IF NOT EXISTS plan_activity_coords (
  person_id      TEXT NOT NULL,
  plan_id        TEXT NOT NULL,
  plan_index     INTEGER NOT NULL,
  activity_index INTEGER NOT NULL,
  lng            REAL NOT NULL,
  lat            REAL NOT NULL,
  PRIMARY KEY (person_id, plan_id, activity_index)
);
"#;

pub(crate) fn create_population_db(
    work_dir: &Path,
    source_name: &str,
) -> Result<(), EzError> {
    let connection = Connection::open(db_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to create SQLite database: {err}"),
        }
    })?;
    connection
        .execute_batch(POPULATION_SCHEMA)
        .map_err(|err| EzError::Io {
            message: format!("Failed to initialize population schema: {err}"),
        })?;
    Ok(())
}
