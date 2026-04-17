use std::{
    ffi::OsStr,
    fs,
    path::{Path, PathBuf},
};

use tauri::State;

use crate::ez::{
    fs::{
        vehicles_db_file_path, vehicles_postamble_file_path, vehicles_preamble_file_path,
    },
    types::{EzError, SessionManager, SourceKind, TransitVehiclesMetadata},
};

pub(crate) mod parser;
pub(crate) mod schema;

pub(crate) fn import_transit_vehicles(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
) -> Result<TransitVehiclesMetadata, EzError> {
    schema::create_transit_vehicles_db(work_dir, source_name)?;

    match parser::run(work_dir, file_path, source_name) {
        Ok(root_tag_attributes_blob) => {
            let source_file_name = file_path
                .file_name()
                .and_then(OsStr::to_str)
                .unwrap_or("")
                .to_string();
            Ok(TransitVehiclesMetadata {
                root_tag_attributes_blob,
                source_file_name,
            })
        }
        Err(err) => {
            let _ = fs::remove_file(vehicles_db_file_path(work_dir, source_name));
            let _ = fs::remove_file(vehicles_preamble_file_path(work_dir, source_name));
            let _ = fs::remove_file(vehicles_postamble_file_path(work_dir, source_name));
            Err(err)
        }
    }
}

#[tauri::command]
pub fn detach_transit_vehicles_cmd(
    source_name: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    let entry = session
        .ui_state
        .sources
        .iter_mut()
        .find(|entry| entry.name == source_name)
        .ok_or_else(|| EzError::SourceNotFound {
            name: source_name.clone(),
        })?;

    if entry.kind != SourceKind::Transit {
        return Err(EzError::UnsupportedSource {
            message: format!("Source '{source_name}' is not a transit source."),
        });
    }

    let transit_metadata = entry
        .transit_metadata
        .as_mut()
        .ok_or_else(|| EzError::UnsupportedSource {
            message: format!(
                "Transit source '{source_name}' is missing schedule metadata."
            ),
        })?;

    if transit_metadata.vehicles.is_none() {
        return Ok(());
    }

    let _ = fs::remove_file(vehicles_db_file_path(&session.work_dir, &source_name));
    let _ = fs::remove_file(vehicles_preamble_file_path(&session.work_dir, &source_name));
    let _ = fs::remove_file(vehicles_postamble_file_path(&session.work_dir, &source_name));

    transit_metadata.vehicles = None;
    session.dirty = true;

    Ok(())
}

#[tauri::command]
pub fn import_transit_vehicles_cmd(
    source_name: String,
    file_path: String,
    manager: State<'_, SessionManager>,
) -> Result<TransitVehiclesMetadata, EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    let entry = session
        .ui_state
        .sources
        .iter_mut()
        .find(|entry| entry.name == source_name)
        .ok_or_else(|| EzError::SourceNotFound {
            name: source_name.clone(),
        })?;

    if entry.kind != SourceKind::Transit {
        return Err(EzError::UnsupportedSource {
            message: format!("Source '{source_name}' is not a transit source."),
        });
    }

    let transit_metadata = entry
        .transit_metadata
        .as_mut()
        .ok_or_else(|| EzError::UnsupportedSource {
            message: format!(
                "Transit source '{source_name}' is missing schedule metadata."
            ),
        })?;

    if transit_metadata.vehicles.is_some() {
        return Err(EzError::Io {
            message: format!(
                "Source '{source_name}' already has a transit vehicles file attached."
            ),
        });
    }

    let file_path = PathBuf::from(file_path);
    let metadata = import_transit_vehicles(&session.work_dir, &file_path, &source_name)?;

    transit_metadata.vehicles = Some(metadata.clone());
    session.dirty = true;

    Ok(metadata)
}
