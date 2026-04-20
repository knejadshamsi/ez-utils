use std::{
    ffi::OsStr,
    fs::{self, File},
    io::Read,
    path::Path,
};

use super::types::{EzError, SourceKind};

pub(crate) const MAX_SOURCE_DATABASES: usize = 5;
const MAX_SNIFF_BYTES: usize = 4096;
const RESERVED_WINDOWS_NAMES: &[&str] = &[
    "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8",
    "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
];

pub(crate) fn normalize_source_name(name: &str) -> String {
    name.trim().trim_end_matches(['.', ' ']).to_string()
}

pub(crate) fn validate_source_name(
    name: &str,
    existing_names: impl Iterator<Item = String>,
) -> Result<String, EzError> {
    let normalized = normalize_source_name(name);
    if name != normalized {
        return Err(EzError::SourceNameInvalid {
            message:
                "Source name cannot have leading or trailing whitespace, trailing dots, or trailing spaces."
                    .into(),
        });
    }
    if normalized.is_empty() {
        return Err(EzError::SourceNameInvalid {
            message: "Source name cannot be empty.".into(),
        });
    }

    if normalized.chars().any(|c| {
        matches!(c, '<' | '>' | ':' | '"' | '/' | '\\' | '|' | '?' | '*') || c.is_control()
    }) {
        return Err(EzError::SourceNameInvalid {
            message: "Source name contains invalid filename characters.".into(),
        });
    }

    let uppercase = normalized.to_ascii_uppercase();
    if RESERVED_WINDOWS_NAMES
        .iter()
        .any(|reserved| *reserved == uppercase)
    {
        return Err(EzError::SourceNameInvalid {
            message: "Source name is a reserved Windows filename.".into(),
        });
    }

    let duplicate = existing_names
        .map(|existing| normalize_source_name(&existing).to_ascii_lowercase())
        .any(|existing| existing == normalized.to_ascii_lowercase());
    if duplicate {
        return Err(EzError::SourceNameInvalid {
            message: "A source with this name already exists.".into(),
        });
    }

    Ok(normalized)
}

pub(crate) fn detect_source_kind(file_path: &Path) -> Result<SourceKind, EzError> {
    let mut file = File::open(file_path).map_err(|err| EzError::Io {
        message: format!("Failed to open source file: {err}"),
    })?;
    let mut buffer = vec![0_u8; MAX_SNIFF_BYTES];
    let bytes_read = file.read(&mut buffer).map_err(|err| EzError::Io {
        message: format!("Failed to read source file: {err}"),
    })?;
    buffer.truncate(bytes_read);
    let content = String::from_utf8_lossy(&buffer).to_ascii_lowercase();

    if content.contains("<network") {
        return Ok(SourceKind::Network);
    }
    if content.contains("<population") || content.contains("<plans") {
        return Ok(SourceKind::Population);
    }
    if content.contains("<transitschedule") {
        return Ok(SourceKind::Transit);
    }

    match file_path
        .extension()
        .and_then(OsStr::to_str)
        .map(|value| value.to_ascii_lowercase())
    {
        Some(ext) if ext == "zip" => Err(EzError::ImportNotImplemented {
            source_kind: "transit".into(),
        }),
        _ => Err(EzError::UnsupportedSource {
            message: "Source file does not look like a supported MATSim network XML.".into(),
        }),
    }
}

pub(crate) fn existing_source_names_from_disk(work_dir: &Path) -> Vec<String> {
    fs::read_dir(work_dir)
        .ok()
        .into_iter()
        .flat_map(|iter| iter.filter_map(Result::ok))
        .filter_map(|entry| {
            let path = entry.path();
            if path.extension().and_then(OsStr::to_str) == Some("db") {
                path.file_stem()
                    .and_then(OsStr::to_str)
                    .filter(|value| !value.ends_with(".vehicles"))
                    .map(|value| value.to_string())
            } else {
                None
            }
        })
        .collect()
}

