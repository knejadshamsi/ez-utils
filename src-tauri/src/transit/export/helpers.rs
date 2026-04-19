use std::io::Write;

use crate::ez::types::EzError;

pub(super) fn write_open_tag<W: Write>(
    writer: &mut W,
    tag: &str,
    tag_blob: &str,
    indent: &str,
) -> Result<(), EzError> {
    if tag_blob.trim().is_empty() {
        write!(writer, "{indent}<{tag}>\n").map_err(io_write_err)?;
    } else {
        write!(writer, "{indent}<{tag} {tag_blob}>\n").map_err(io_write_err)?;
    }
    Ok(())
}

pub(super) fn write_indented<W: Write>(
    writer: &mut W,
    content: &str,
    indent: &str,
) -> Result<(), EzError> {
    for line in content.lines() {
        writer.write_all(indent.as_bytes()).map_err(io_write_err)?;
        writer.write_all(line.as_bytes()).map_err(io_write_err)?;
        writer.write_all(b"\n").map_err(io_write_err)?;
    }
    Ok(())
}

pub(super) fn format_decimal(v: f64) -> String {
    if (v.round() - v).abs() < 1e-9 {
        format!("{:.1}", v)
    } else {
        format!("{:.6}", v)
    }
}

pub(super) fn io_write_err(err: std::io::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write export file: {err}"),
    }
}

pub(super) fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Transit export query failed: {err}"),
    }
}
