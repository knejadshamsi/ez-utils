use std::{
    fs::{self, File},
    path::{Path, PathBuf},
};

use zip::{write::FileOptions, CompressionMethod, ZipArchive, ZipWriter};

use super::types::{EzError, UiState, DEFAULT_AUTOSAVE_INTERVAL_MINUTES};

const UI_FILE_NAME: &str = "ui.json";

pub(crate) fn ui_json_path(work_dir: &Path) -> PathBuf {
    work_dir.join(UI_FILE_NAME)
}

pub(crate) fn db_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.db"))
}

pub(crate) fn preamble_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.preamble.xml"))
}

pub(crate) fn postamble_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.postamble.xml"))
}

pub(crate) fn vehicles_db_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.vehicles.db"))
}

pub(crate) fn vehicles_preamble_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.vehicles.preamble.xml"))
}

pub(crate) fn vehicles_postamble_file_path(work_dir: &Path, source_name: &str) -> PathBuf {
    work_dir.join(format!("{source_name}.vehicles.postamble.xml"))
}

pub(crate) fn work_dir_path(ez_path: &Path) -> PathBuf {
    let file_name = ez_path
        .file_name()
        .and_then(|value| value.to_str())
        .unwrap_or("file.ez");
    ez_path.with_file_name(format!("{file_name}work"))
}

fn temp_archive_path(ez_path: &Path) -> PathBuf {
    let file_name = ez_path
        .file_name()
        .and_then(|value| value.to_str())
        .unwrap_or("file.ez");
    ez_path.with_file_name(format!("{file_name}.tmp"))
}

pub(crate) fn write_ui_state(
    work_dir: &Path,
    ui_state: &UiState,
) -> Result<(), EzError> {
    let content = serde_json::to_vec_pretty(ui_state).map_err(|err| EzError::Io {
        message: format!("Failed to serialize ui.json: {err}"),
    })?;
    fs::write(ui_json_path(work_dir), content).map_err(|err| EzError::Io {
        message: format!("Failed to write ui.json: {err}"),
    })
}

pub(crate) fn read_ui_state(work_dir: &Path) -> Result<UiState, EzError> {
    let path = ui_json_path(work_dir);
    if !path.exists() {
        let state = UiState::default();
        write_ui_state(work_dir, &state)?;
        return Ok(state);
    }

    let content = fs::read_to_string(&path).map_err(|_| EzError::ArchiveCorrupt)?;
    let mut state: UiState =
        serde_json::from_str(&content).map_err(|_| EzError::ArchiveCorrupt)?;
    if state.version > 1 {
        return Err(EzError::UnsupportedUiVersion {
            version: state.version,
        });
    }
    if state.settings.autosave_settings.interval_minutes == 0 {
        state.settings.autosave_settings.interval_minutes = DEFAULT_AUTOSAVE_INTERVAL_MINUTES;
    }
    Ok(state)
}

pub(crate) fn ensure_writable_parent(ez_path: &Path) -> Result<(), EzError> {
    let parent = ez_path
        .parent()
        .ok_or_else(|| EzError::UnwritableLocation {
            path: ez_path.to_string_lossy().to_string(),
        })?;

    fs::create_dir_all(parent).map_err(|_| EzError::UnwritableLocation {
        path: parent.to_string_lossy().to_string(),
    })?;

    let probe = parent.join(".ez-write-test.tmp");
    File::create(&probe)
        .and_then(|_| fs::remove_file(&probe))
        .map_err(|_| EzError::UnwritableLocation {
            path: parent.to_string_lossy().to_string(),
        })?;

    Ok(())
}

fn atomic_replace(temp_path: &Path, target_path: &Path) -> Result<(), EzError> {
    #[cfg(target_os = "windows")]
    {
        use std::os::windows::ffi::OsStrExt;
        #[link(name = "Kernel32")]
        extern "system" {
            fn ReplaceFileW(
                lpReplacedFileName: *const u16,
                lpReplacementFileName: *const u16,
                lpBackupFileName: *const u16,
                dwReplaceFlags: u32,
                lpExclude: *mut core::ffi::c_void,
                lpReserved: *mut core::ffi::c_void,
            ) -> i32;
        }
        const REPLACEFILE_IGNORE_MERGE_ERRORS: u32 = 0x2;

        let to_wide = |value: &Path| {
            value
                .as_os_str()
                .encode_wide()
                .chain(std::iter::once(0))
                .collect::<Vec<u16>>()
        };

        if target_path.exists() {
            let replaced = unsafe {
                ReplaceFileW(
                    to_wide(target_path).as_ptr(),
                    to_wide(temp_path).as_ptr(),
                    std::ptr::null(),
                    REPLACEFILE_IGNORE_MERGE_ERRORS,
                    std::ptr::null_mut(),
                    std::ptr::null_mut(),
                )
            };
            if replaced == 0 {
                return Err(EzError::Io {
                    message: "Failed to replace .ez archive.".into(),
                });
            }
        } else {
            fs::rename(temp_path, target_path).map_err(|err| EzError::Io {
                message: format!("Failed to place .ez archive: {err}"),
            })?;
        }
        Ok(())
    }
    #[cfg(not(target_os = "windows"))]
    {
        fs::rename(temp_path, target_path).map_err(|err| EzError::Io {
            message: format!("Failed to replace .ez archive: {err}"),
        })
    }
}

fn zip_options() -> FileOptions<'static, ()> {
    FileOptions::default().compression_method(CompressionMethod::Deflated)
}

fn archive_relative_path(base_dir: &Path, path: &Path) -> Result<String, EzError> {
    path.strip_prefix(base_dir)
        .map_err(|err| EzError::Io {
            message: format!("Failed to compute archive path: {err}"),
        })
        .map(|relative| relative.to_string_lossy().replace('\\', "/"))
}

fn add_directory_to_archive(
    writer: &mut ZipWriter<File>,
    base_dir: &Path,
    current_dir: &Path,
) -> Result<(), EzError> {
    let mut entries = fs::read_dir(current_dir)
        .map_err(|err| EzError::Io {
            message: format!("Failed to read working directory: {err}"),
        })?
        .collect::<Result<Vec<_>, _>>()
        .map_err(|err| EzError::Io {
            message: format!("Failed to enumerate working directory: {err}"),
        })?;
    entries.sort_by_key(|entry| entry.path());

    for entry in entries {
        let path = entry.path();
        let metadata = entry.metadata().map_err(|err| EzError::Io {
            message: format!("Failed to inspect working directory entry: {err}"),
        })?;

        if metadata.is_dir() {
            let relative = archive_relative_path(base_dir, &path)?;
            writer
                .add_directory(format!("{relative}/"), zip_options())
                .map_err(|err| EzError::Io {
                    message: format!("Failed to add directory to archive: {err}"),
                })?;
            add_directory_to_archive(writer, base_dir, &path)?;
        } else if metadata.is_file() {
            let relative = archive_relative_path(base_dir, &path)?;
            writer
                .start_file(relative, zip_options())
                .map_err(|err| EzError::Io {
                    message: format!("Failed to add file to archive: {err}"),
                })?;
            let mut file = File::open(&path).map_err(|err| EzError::Io {
                message: format!("Failed to open file for archiving: {err}"),
            })?;
            std::io::copy(&mut file, writer).map_err(|err| EzError::Io {
                message: format!("Failed to write file into archive: {err}"),
            })?;
        }
    }

    Ok(())
}

pub(crate) fn pack_work_dir(work_dir: &Path, ez_path: &Path) -> Result<(), EzError> {
    if let Some(parent) = ez_path.parent() {
        fs::create_dir_all(parent).map_err(|err| EzError::Io {
            message: format!("Failed to create archive parent directory: {err}"),
        })?;
    }

    let temp_path = temp_archive_path(ez_path);
    if temp_path.exists() {
        let _ = fs::remove_file(&temp_path);
    }

    let file = File::create(&temp_path).map_err(|err| EzError::Io {
        message: format!("Failed to create temp archive: {err}"),
    })?;
    let mut writer = ZipWriter::new(file);
    add_directory_to_archive(&mut writer, work_dir, work_dir)?;
    writer.finish().map_err(|err| EzError::Io {
        message: format!("Failed to finalize archive: {err}"),
    })?;

    atomic_replace(&temp_path, ez_path)
}

pub(crate) fn unpack_archive(ez_path: &Path, work_dir: &Path) -> Result<(), EzError> {
    if work_dir.exists() {
        fs::remove_dir_all(work_dir).map_err(|err| EzError::Io {
            message: format!("Failed to reset working directory: {err}"),
        })?;
    }
    fs::create_dir_all(work_dir).map_err(|err| EzError::Io {
        message: format!("Failed to create working directory: {err}"),
    })?;

    let file = File::open(ez_path).map_err(|_| EzError::ArchiveCorrupt)?;
    let mut archive = ZipArchive::new(file).map_err(|_| EzError::ArchiveCorrupt)?;
    for index in 0..archive.len() {
        let mut entry = archive
            .by_index(index)
            .map_err(|_| EzError::ArchiveCorrupt)?;
        let Some(enclosed_name) = entry.enclosed_name().map(|name| name.to_owned()) else {
            return Err(EzError::ArchiveCorrupt);
        };
        let output_path = work_dir.join(enclosed_name);
        if entry.name().ends_with('/') {
            fs::create_dir_all(&output_path).map_err(|err| EzError::Io {
                message: format!("Failed to create archive directory: {err}"),
            })?;
            continue;
        }

        if let Some(parent) = output_path.parent() {
            fs::create_dir_all(parent).map_err(|err| EzError::Io {
                message: format!("Failed to create archive parent directory: {err}"),
            })?;
        }

        let mut output = File::create(&output_path).map_err(|err| EzError::Io {
            message: format!("Failed to create unpacked file: {err}"),
        })?;
        std::io::copy(&mut entry, &mut output).map_err(|err| EzError::Io {
            message: format!("Failed to unpack file: {err}"),
        })?;
    }

    Ok(())
}

pub(crate) fn cleanup_work_dir(path: &Path) {
    let _ = fs::remove_dir_all(path);
}
