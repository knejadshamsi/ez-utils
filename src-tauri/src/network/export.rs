use std::{
    fs,
    io::{BufWriter, Write},
    path::{Path, PathBuf},
};

use rusqlite::Connection;
use tauri::State;

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, NetworkMetadata, SessionManager, SourceKind},
};

#[tauri::command]
pub fn export_network(
    source_name: String,
    output_path: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let (work_dir, metadata) = resolve_session(&manager, &source_name)?;
    let connection = open_connection(&work_dir, &source_name)?;
    let output_path = PathBuf::from(output_path);
    write_network_xml(&connection, &work_dir, &source_name, &metadata, &output_path)
}

/// Pulls the work dir and network metadata for `source_name` out of the active
/// session. The UI state is the single source of truth for metadata - it is
/// populated on import and persisted via ui.json inside the .ez archive.
fn resolve_session(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<(PathBuf, NetworkMetadata), EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;

    let entry = session
        .ui_state
        .sources
        .iter()
        .find(|entry| entry.name == source_name)
        .ok_or_else(|| EzError::SourceNotFound {
            name: source_name.to_string(),
        })?;

    if entry.kind != SourceKind::Network {
        return Err(EzError::UnsupportedSource {
            message: format!("Source '{source_name}' is not a network source."),
        });
    }

    let metadata = entry
        .network_metadata
        .clone()
        .unwrap_or_else(|| NetworkMetadata {
            network_attributes_blob: None,
            links_tag_blob: String::new(),
        });

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
        message: format!("Failed to open network SQLite database: {err}"),
    })
}

fn write_network_xml(
    connection: &Connection,
    work_dir: &Path,
    source_name: &str,
    metadata: &NetworkMetadata,
    output_path: &Path,
) -> Result<(), EzError> {
    let file = fs::File::create(output_path).map_err(|err| EzError::Io {
        message: format!("Failed to create export file: {err}"),
    })?;
    let mut writer = BufWriter::new(file);

    // 1. Preamble: XML decl, DOCTYPE, leading comments/whitespace, and the
    // root <network ...> start tag, captured verbatim at import time.
    let preamble = fs::read(preamble_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to read network preamble sidecar: {err}"),
        }
    })?;
    writer.write_all(&preamble).map_err(io_write_err)?;
    ensure_trailing_newline(&mut writer, &preamble)?;

    // 2. Optional root <attributes>...</attributes> block.
    if let Some(blob) = metadata.network_attributes_blob.as_deref() {
        write_indented(&mut writer, blob.trim(), "  ")?;
    }

    // 3. <nodes> section - stream raw_xml blobs in insertion order.
    writer.write_all(b"  <nodes>\n").map_err(io_write_err)?;
    {
        let mut stmt = connection
            .prepare("SELECT raw_xml FROM nodes")
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map([], |row| row.get::<_, String>(0))
            .map_err(sqlite_err)?;
        for row in rows {
            let raw_xml = row.map_err(sqlite_err)?;
            write_indented(&mut writer, raw_xml.trim(), "    ")?;
        }
    }
    writer.write_all(b"  </nodes>\n").map_err(io_write_err)?;

    // 4. <links> section - preserve the original <links ...> tag attributes.
    if metadata.links_tag_blob.is_empty() {
        writer.write_all(b"  <links>\n").map_err(io_write_err)?;
    } else {
        write!(writer, "  <links {}>\n", metadata.links_tag_blob).map_err(io_write_err)?;
    }
    {
        let mut stmt = connection
            .prepare("SELECT tag_blob, attributes_blob FROM links")
            .map_err(sqlite_err)?;
        let rows = stmt
            .query_map([], |row| {
                Ok((row.get::<_, String>(0)?, row.get::<_, Option<String>>(1)?))
            })
            .map_err(sqlite_err)?;
        for row in rows {
            let (tag_blob, attributes_blob) = row.map_err(sqlite_err)?;
            match attributes_blob {
                None => {
                    write!(writer, "    <link {tag_blob}/>\n").map_err(io_write_err)?;
                }
                Some(attrs) => {
                    write!(writer, "    <link {tag_blob}>\n").map_err(io_write_err)?;
                    write_indented(&mut writer, attrs.trim(), "      ")?;
                    writer.write_all(b"    </link>\n").map_err(io_write_err)?;
                }
            }
        }
    }
    writer.write_all(b"  </links>\n").map_err(io_write_err)?;

    // 5. Postamble: </network> plus anything after it, captured verbatim.
    let postamble = fs::read(postamble_file_path(work_dir, source_name)).map_err(|err| {
        EzError::Io {
            message: format!("Failed to read network postamble sidecar: {err}"),
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

/// Guarantees the body starts on a fresh line after the preamble, regardless
/// of whether the source file's root tag had a trailing newline.
fn ensure_trailing_newline<W: Write>(writer: &mut W, preamble: &[u8]) -> Result<(), EzError> {
    if !preamble.ends_with(b"\n") {
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
        message: format!("Network export query failed: {err}"),
    }
}
