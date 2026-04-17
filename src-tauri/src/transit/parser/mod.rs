//! Transit schedule import parser.
//!
//! Parses a MATSim transitSchedule XML file into a SQLite database, capturing
//! stop facilities and line/route/departure data with their attributes for
//! lossless round-tripping. The XML surrounding the root element is written
//! verbatim to preamble/postamble sidecar files beside the database.

use std::{
    fs::{self, File},
    io::BufReader,
    path::Path,
};

use quick_xml::reader::Reader;
use rusqlite::Connection;

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, TransitMetadata},
};
use crate::projection::{CrsConfig, Projector};

mod body;
mod lines;
mod stops;
mod util;

pub(crate) fn run(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<TransitMetadata, EzError> {
    let file = File::open(file_path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit XML: {err}"),
    })?;
    let mut reader = Reader::from_reader(BufReader::new(file));
    reader.config_mut().trim_text(false);
    reader.config_mut().trim_markup_names_in_closing_tags = false;

    let projector = Projector::new(crs_config)?;
    let mut event_buf = Vec::new();
    let mut capture_buf = Vec::new();

    let mut connection =
        Connection::open(db_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
            message: format!("Failed to open transit SQLite database: {err}"),
        })?;
    let tx = connection.transaction().map_err(|err| EzError::Io {
        message: format!("Failed to begin transit import transaction: {err}"),
    })?;

    let state = body::parse(
        &mut reader,
        &tx,
        &projector,
        &mut event_buf,
        &mut capture_buf,
    )?;

    // Internal: give each line its own mode (from first route) so the line
    // can own this concept at the UI level. Not exported.
    super::schema::backfill_lines_mode(&tx)?;

    tx.commit().map_err(|err| EzError::Io {
        message: format!("Failed to commit transit import transaction: {err}"),
    })?;

    fs::write(preamble_file_path(work_dir, source_name), &state.preamble_buf).map_err(|err| {
        EzError::Io {
            message: format!("Failed to write transit preamble file: {err}"),
        }
    })?;
    fs::write(
        postamble_file_path(work_dir, source_name),
        &state.postamble_buf,
    )
    .map_err(|err| EzError::Io {
        message: format!("Failed to write transit postamble file: {err}"),
    })?;

    Ok(TransitMetadata {
        transit_schedule_attributes_blob: state.transit_schedule_attributes_blob,
        transit_stops_tag_blob: state.transit_stops_tag_blob,
        transit_minimal_transfers_tag_blob: state.transit_minimal_transfers_tag_blob,
        transit_lines_tag_blob: state.transit_lines_tag_blob,
        vehicles: None,
        linked_network_source_name: None,
    })
}
