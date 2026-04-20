use std::{
    fs,
    io::{BufWriter, Write},
    path::{Path, PathBuf},
};

use rusqlite::{params, Connection};
use tauri::State;

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, SessionManager},
};
use crate::utils::xml_escape;

#[tauri::command]
pub fn export_population(
    source_name: String,
    output_path: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let work_dir = resolve_work_dir(&manager)?;
    let connection = open_population_connection(&work_dir, &source_name)?;
    let output_path = PathBuf::from(output_path);
    write_population_xml(&connection, &work_dir, &source_name, &output_path)
}

fn resolve_work_dir(manager: &State<'_, SessionManager>) -> Result<PathBuf, EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
    Ok(session.work_dir.clone())
}

fn open_population_connection(
    work_dir: &Path,
    source_name: &str,
) -> Result<Connection, EzError> {
    let path = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    Connection::open(path).map_err(|err| EzError::Io {
        message: format!("Failed to open population SQLite database: {err}"),
    })
}

fn write_population_xml(
    connection: &Connection,
    work_dir: &Path,
    source_name: &str,
    output_path: &PathBuf,
) -> Result<(), EzError> {
    let file = fs::File::create(output_path).map_err(|err| EzError::Io {
        message: format!("Failed to create export file: {err}"),
    })?;
    let mut writer = BufWriter::new(file);

    // Preamble: XML decl, DOCTYPE, leading content and the root <population>
    // start tag captured verbatim at import time.
    let preamble = fs::read(preamble_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to read population preamble sidecar: {err}"),
        }
    })?;
    writer.write_all(&preamble).map_err(io_write_err)?;
    if !preamble.ends_with(b"\n") {
        writer.write_all(b"\n").map_err(io_write_err)?;
    }

    let mut person_stmt = connection
        .prepare("SELECT person_id, attributes_blob FROM person_attributes ORDER BY person_id")
        .map_err(sqlite_err)?;

    let mut plan_stmt = connection
        .prepare("SELECT plan_blob FROM person_plans WHERE person_id = ?1 ORDER BY plan_index")
        .map_err(sqlite_err)?;

    let persons = person_stmt
        .query_map([], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, Option<String>>(1)?,
            ))
        })
        .map_err(sqlite_err)?;

    for person_result in persons {
        let (person_id, attributes_blob) = person_result.map_err(sqlite_err)?;

        writeln!(writer, "  <person id=\"{}\">", xml_escape(&person_id))
            .map_err(io_write_err)?;

        if let Some(ref blob) = attributes_blob {
            write_indented(&mut writer, blob.trim(), "    ")?;
        }

        let plans = plan_stmt
            .query_map(params![person_id], |row| row.get::<_, String>(0))
            .map_err(sqlite_err)?;

        for plan_result in plans {
            let plan_blob = plan_result.map_err(sqlite_err)?;
            write_indented(&mut writer, plan_blob.trim(), "    ")?;
        }

        writeln!(writer, "  </person>").map_err(io_write_err)?;
    }

    // Postamble: </population> plus any trailing content captured verbatim.
    let postamble = fs::read(postamble_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to read population postamble sidecar: {err}"),
        }
    })?;
    writer.write_all(&postamble).map_err(io_write_err)?;

    writer.flush().map_err(io_write_err)?;

    Ok(())
}

fn write_indented<W: Write>(writer: &mut W, content: &str, indent: &str) -> Result<(), EzError> {
    for line in content.lines() {
        writer.write_all(indent.as_bytes()).map_err(io_write_err)?;
        writer.write_all(line.as_bytes()).map_err(io_write_err)?;
        writer.write_all(b"\n").map_err(io_write_err)?;
    }
    Ok(())
}

fn io_write_err(err: std::io::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write export file: {err}"),
    }
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Population export query failed: {err}"),
    }
}
