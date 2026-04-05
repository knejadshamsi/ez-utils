pub(crate) mod fs;
mod source;
pub(crate) mod types;

use std::{ffi::OsStr, path::PathBuf};

use tauri::State;

use crate::network::{import_network, CrsConfig as NetworkCrsConfig};
use crate::population::{import_population, CrsConfig};
use crate::transit::import_transit;

pub use self::types::SessionManager;

use self::{
    fs::{
        cleanup_work_dir, db_file_path, ensure_writable_parent, pack_work_dir,
        postamble_file_path, preamble_file_path, read_ui_state, unpack_archive,
        vehicles_db_file_path, work_dir_path, write_ui_state,
    },
    source::{
        detect_source_kind, existing_source_names_from_disk, normalize_source_name,
        validate_source_name, MAX_SOURCE_DATABASES,
    },
    types::{
        session_payload, CloseMode, EzError, ImportedSourcePayload, NewSessionPayload,
        Session, SourceKind, StatePayload, UiState,
    },
};

fn prune_stale_transit_links(work_dir: &std::path::Path, ui_state: &mut UiState) -> bool {
    let network_source_names: std::collections::HashSet<String> = ui_state
        .sources
        .iter()
        .filter(|s| s.kind == SourceKind::Network)
        .map(|s| s.name.clone())
        .collect();

    let mut mutated = false;
    for source in ui_state.sources.iter_mut() {
        if source.kind != SourceKind::Transit {
            continue;
        }
        let Some(meta) = source.transit_metadata.as_mut() else {
            continue;
        };
        if meta.vehicles.is_some()
            && !vehicles_db_file_path(work_dir, &source.name).exists()
        {
            meta.vehicles = None;
            mutated = true;
        }
        if let Some(linked) = meta.linked_network_source_name.as_ref() {
            if !network_source_names.contains(linked) {
                meta.linked_network_source_name = None;
                mutated = true;
            }
        }
    }
    mutated
}

#[tauri::command]
pub fn validate_xml(file_path: String) -> Result<SourceKind, EzError> {
    let file_path = PathBuf::from(file_path);
    detect_source_kind(&file_path)
}

#[tauri::command]
pub async fn new_from_xml(
    xml_path: String,
    ez_path: String,
    manager: State<'_, SessionManager>,
) -> Result<NewSessionPayload, EzError> {
    let xml_path = PathBuf::from(xml_path);
    let ez_path = PathBuf::from(&ez_path);
    let ez_path = if ez_path.extension().map_or(true, |e| e != "ez") {
        ez_path.with_extension("ez")
    } else {
        ez_path
    };

    let kind = detect_source_kind(&xml_path)?;
    ensure_writable_parent(&ez_path)?;

    let work_dir = work_dir_path(&ez_path);
    if work_dir.exists() {
        std::fs::remove_dir_all(&work_dir).map_err(|err| EzError::Io {
            message: format!("Failed to reset working directory: {err}"),
        })?;
    }
    std::fs::create_dir_all(&work_dir).map_err(|err| EzError::Io {
        message: format!("Failed to create working directory: {err}"),
    })?;

    let ui_state = UiState::default();
    write_ui_state(&work_dir, &ui_state)?;

    let file_stem = xml_path
        .file_stem()
        .and_then(OsStr::to_str)
        .ok_or_else(|| EzError::UnsupportedSource {
            message: "Could not determine source file name.".into(),
        })?;
    let existing_names = existing_source_names_from_disk(&work_dir);
    let name = validate_source_name(file_stem, existing_names.into_iter())?;

    let mut network_metadata = None;
    let mut transit_metadata = None;
    let import_result = match kind {
        SourceKind::Network => {
            import_network(&work_dir, &xml_path, &name, NetworkCrsConfig::MontrealMtm8)
                .map(|meta| { network_metadata = Some(meta); })
        }
        SourceKind::Population => {
            import_population(&work_dir, &xml_path, &name, CrsConfig::MontrealMtm8)
        }
        SourceKind::Transit => {
            import_transit(&work_dir, &xml_path, &name, NetworkCrsConfig::MontrealMtm8)
                .map(|meta| { transit_metadata = Some(meta); })
        }
    };

    if let Err(err) = import_result {
        cleanup_work_dir(&work_dir);
        return Err(err);
    }

    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;

    let session = Session {
        ez_path,
        work_dir,
        dirty: true,
        ui_state,
    };
    let payload = NewSessionPayload {
        state: session_payload(&session),
        imported: ImportedSourcePayload {
            name,
            kind,
            network_metadata,
            transit_metadata,
        },
    };
    *guard = Some(session);
    Ok(payload)
}

#[tauri::command]
pub async fn unpack_ez(ez_path: String) -> Result<(), EzError> {
    let ez_path = PathBuf::from(ez_path);
    if !ez_path.exists() {
        return Err(EzError::ArchiveMissing);
    }
    ensure_writable_parent(&ez_path)?;

    let work_dir = work_dir_path(&ez_path);
    unpack_archive(&ez_path, &work_dir)?;
    Ok(())
}

#[tauri::command]
pub async fn load_ez(
    ez_path: String,
    manager: State<'_, SessionManager>,
) -> Result<StatePayload, EzError> {
    let ez_path = PathBuf::from(ez_path);
    let work_dir = work_dir_path(&ez_path);
    let mut ui_state = read_ui_state(&work_dir)?;

    if prune_stale_transit_links(&work_dir, &mut ui_state) {
        write_ui_state(&work_dir, &ui_state)?;
    }

    // Idempotent schema migrations for pre-existing transit DBs.
    for source in ui_state.sources.iter() {
        if source.kind == SourceKind::Transit {
            let db_path = crate::ez::fs::db_file_path(&work_dir, &source.name);
            if db_path.exists() {
                let conn = rusqlite::Connection::open(&db_path).map_err(|err| EzError::Io {
                    message: format!("Failed to open transit DB for migration: {err}"),
                })?;
                crate::transit::schema::ensure_transit_schema(&conn)?;
            }
        }
    }

    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;

    let session = Session {
        ez_path,
        work_dir,
        dirty: false,
        ui_state,
    };
    let payload = session_payload(&session);
    *guard = Some(session);
    Ok(payload)
}

#[tauri::command]
pub async fn save_ez(
    manager: State<'_, SessionManager>,
) -> Result<StatePayload, EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    write_ui_state(&session.work_dir, &session.ui_state)?;
    pack_work_dir(&session.work_dir, &session.ez_path)?;
    session.dirty = false;
    Ok(session_payload(session))
}

#[tauri::command]
pub async fn close_ez(
    mode: CloseMode,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.take().ok_or(EzError::NoActiveSession)?;

    match mode {
        CloseMode::Save => {
            write_ui_state(&session.work_dir, &session.ui_state)?;
            pack_work_dir(&session.work_dir, &session.ez_path)?;
            cleanup_work_dir(&session.work_dir);
        }
        CloseMode::Discard => {
            cleanup_work_dir(&session.work_dir);
        }
    }
    Ok(())
}

#[tauri::command]
pub fn get_state(
    manager: State<'_, SessionManager>,
) -> Result<StatePayload, EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
    Ok(session_payload(session))
}

#[tauri::command]
pub async fn import_source(
    file_path: String,
    manager: State<'_, SessionManager>,
) -> Result<ImportedSourcePayload, EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    let file_path = PathBuf::from(file_path);
    let kind = detect_source_kind(&file_path)?;
    let existing_names = existing_source_names_from_disk(&session.work_dir);
    if existing_names.len() >= MAX_SOURCE_DATABASES {
        return Err(EzError::SourceLimitReached {
            limit: MAX_SOURCE_DATABASES,
        });
    }
    let file_stem = file_path
        .file_stem()
        .and_then(OsStr::to_str)
        .ok_or_else(|| EzError::UnsupportedSource {
            message: "Could not determine source file name.".into(),
        })?;

    let name = validate_source_name(file_stem, existing_names.into_iter())?;
    let mut network_metadata = None;
    let mut transit_metadata = None;
    match kind {
        SourceKind::Network => {
            let meta = import_network(
                &session.work_dir,
                &file_path,
                &name,
                NetworkCrsConfig::MontrealMtm8,
            )?;
            network_metadata = Some(meta);
        }
        SourceKind::Population => {
            import_population(
                &session.work_dir,
                &file_path,
                &name,
                CrsConfig::MontrealMtm8,
            )?;
        }
        SourceKind::Transit => {
            let meta = import_transit(
                &session.work_dir,
                &file_path,
                &name,
                NetworkCrsConfig::MontrealMtm8,
            )?;
            transit_metadata = Some(meta);
        }
    }
    session.dirty = true;
    Ok(ImportedSourcePayload { name, kind, network_metadata, transit_metadata })
}

#[tauri::command]
pub fn rename_source(
    source_name: String,
    new_name: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    let source_name = normalize_source_name(&source_name);
    let existing_names = existing_source_names_from_disk(&session.work_dir)
        .into_iter()
        .filter(|existing| existing != &source_name);
    let new_name = validate_source_name(&new_name, existing_names)?;

    let old_path = db_file_path(&session.work_dir, &source_name);
    let new_path = db_file_path(&session.work_dir, &new_name);
    if !old_path.exists() {
        return Err(EzError::SourceNotFound { name: source_name });
    }

    std::fs::rename(&old_path, &new_path).map_err(|err| EzError::Io {
        message: format!("Failed to rename source database: {err}"),
    })?;

    // Rename sidecar preamble/postamble files when present. Older sources
    // imported before sidecar capture was introduced may not have them;
    // ignore missing-file errors.
    let old_preamble = preamble_file_path(&session.work_dir, &source_name);
    if old_preamble.exists() {
        let _ = std::fs::rename(
            &old_preamble,
            preamble_file_path(&session.work_dir, &new_name),
        );
    }
    let old_postamble = postamble_file_path(&session.work_dir, &source_name);
    if old_postamble.exists() {
        let _ = std::fs::rename(
            &old_postamble,
            postamble_file_path(&session.work_dir, &new_name),
        );
    }

    session.dirty = true;
    Ok(())
}

#[tauri::command]
pub fn remove_source(
    source_name: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    let source_name = normalize_source_name(&source_name);
    let path = db_file_path(&session.work_dir, &source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound { name: source_name });
    }

    std::fs::remove_file(&path).map_err(|err| EzError::Io {
        message: format!("Failed to remove source database: {err}"),
    })?;

    // Best-effort cleanup of sidecar files. Older sources imported before
    // sidecar capture was introduced may not have them; ignore errors.
    let _ = std::fs::remove_file(preamble_file_path(&session.work_dir, &source_name));
    let _ = std::fs::remove_file(postamble_file_path(&session.work_dir, &source_name));

    session.dirty = true;
    Ok(())
}

#[tauri::command]
pub fn save_ui_state(
    ui_state: UiState,
    manager: State<'_, SessionManager>,
) -> Result<StatePayload, EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;

    write_ui_state(&session.work_dir, &ui_state)?;
    session.ui_state = ui_state;
    session.dirty = true;
    Ok(session_payload(session))
}
