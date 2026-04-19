use std::{
    fs,
    io::{BufWriter, Write},
    path::{Path, PathBuf},
};

use rusqlite::Connection;
use tauri::State;

use crate::ez::{
    current_crs,
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, SessionManager, SourceKind, TransitMetadata},
};
use crate::projection::Projector;

mod helpers;
mod routes;
mod stops;

use helpers::{io_write_err, sqlite_err};

#[tauri::command]
pub fn export_transit(
    source_name: String,
    output_path: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let (work_dir, metadata) = resolve_session(&manager, &source_name)?;
    let connection = open_connection(&work_dir, &source_name)?;
    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let output_path = PathBuf::from(output_path);
    write_schedule_xml(
        &connection,
        &projector,
        &work_dir,
        &source_name,
        &metadata,
        &output_path,
    )
}

fn resolve_session(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<(PathBuf, TransitMetadata), EzError> {
    let guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
    let entry = session
        .ui_state
        .sources
        .iter()
        .find(|entry| entry.name == source_name)
        .ok_or_else(|| EzError::SourceNotFound {
            name: source_name.to_string(),
        })?;
    if entry.kind != SourceKind::Transit {
        return Err(EzError::UnsupportedSource {
            message: format!("Source '{source_name}' is not a transit source."),
        });
    }
    let metadata = entry
        .transit_metadata
        .clone()
        .ok_or_else(|| EzError::UnsupportedSource {
            message: format!("Transit source '{source_name}' is missing schedule metadata."),
        })?;
    Ok((session.work_dir.clone(), metadata))
}

fn open_connection(work_dir: &Path, source_name: &str) -> Result<Connection, EzError> {
    let path = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    Connection::open(path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })
}

fn write_schedule_xml(
    connection: &Connection,
    projector: &Projector,
    work_dir: &Path,
    source_name: &str,
    metadata: &TransitMetadata,
    output_path: &Path,
) -> Result<(), EzError> {
    let file = fs::File::create(output_path).map_err(|err| EzError::Io {
        message: format!("Failed to create export file: {err}"),
    })?;
    let mut writer = BufWriter::new(file);

    let preamble = fs::read(preamble_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
        message: format!("Failed to read transit preamble sidecar: {err}"),
    })?;
    writer.write_all(&preamble).map_err(io_write_err)?;
    if !preamble.ends_with(b"\n") {
        writer.write_all(b"\n").map_err(io_write_err)?;
    }

    if let Some(blob) = metadata.transit_schedule_attributes_blob.as_deref() {
        helpers::write_indented(&mut writer, blob.trim(), "  ")?;
    }

    helpers::write_open_tag(
        &mut writer,
        "transitStops",
        &metadata.transit_stops_tag_blob,
        "  ",
    )?;
    stops::write_stop_facilities(&mut writer, connection, projector)?;
    writer.write_all(b"  </transitStops>\n").map_err(io_write_err)?;

    let transfer_count: i64 = connection
        .query_row("SELECT COUNT(*) FROM minimal_transfer_times", [], |row| row.get(0))
        .map_err(sqlite_err)?;
    if transfer_count > 0 {
        helpers::write_open_tag(
            &mut writer,
            "minimalTransferTimes",
            &metadata.transit_minimal_transfers_tag_blob,
            "  ",
        )?;
        stops::write_transfer_times(&mut writer, connection)?;
        writer
            .write_all(b"  </minimalTransferTimes>\n")
            .map_err(io_write_err)?;
    }

    helpers::write_open_tag(
        &mut writer,
        "transitLines",
        &metadata.transit_lines_tag_blob,
        "  ",
    )?;
    routes::write_transit_lines(&mut writer, connection)?;
    writer.write_all(b"  </transitLines>\n").map_err(io_write_err)?;

    let postamble = fs::read(postamble_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
        message: format!("Failed to read transit postamble sidecar: {err}"),
    })?;
    writer.write_all(&postamble).map_err(io_write_err)?;
    writer.flush().map_err(io_write_err)?;
    Ok(())
}
