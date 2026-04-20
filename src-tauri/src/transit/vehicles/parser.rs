//! Transit vehicles import parser.
//!
//! Parses a MATSim vehicleDefinitions XML file (schema
//! `vehicleDefinitions_v2.0.xsd`) into a SQLite database, capturing
//! `<vehicleType>` and `<vehicle>` elements as raw XML blobs for lossless
//! round-tripping. The root element's attributes (namespaces,
//! schemaLocation) are captured separately into `TransitVehiclesMetadata`
//! and stored on `SourceEntry.transit_metadata.vehicles`. Leading content
//! before the root tag and trailing content after its close tag are written
//! verbatim to preamble/postamble sidecar files beside the database.

use std::{
    fs::{self, File},
    io::BufReader,
    path::Path,
};

use quick_xml::{events::Event, reader::Reader};
use rusqlite::{params, Connection};

use crate::ez::{
    fs::{
        vehicles_db_file_path, vehicles_postamble_file_path, vehicles_preamble_file_path,
    },
    types::EzError,
};
use crate::network::capture::{
    append_event_bytes, capture_empty_element, capture_started_element,
    capture_tag_attributes,
};
use crate::transit::xml_attrs::{optional_attr_string, required_attr_string};

pub(crate) fn run(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
) -> Result<String, EzError> {
    let file = File::open(file_path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit vehicles XML: {err}"),
    })?;
    let mut reader = Reader::from_reader(BufReader::new(file));
    reader.config_mut().trim_text(false);
    reader.config_mut().trim_markup_names_in_closing_tags = false;

    let mut root_tag_attributes_blob = String::new();
    let mut preamble_buf: Vec<u8> = Vec::new();
    let mut postamble_buf: Vec<u8> = Vec::new();
    let mut in_root = false;
    let mut root_closed = false;

    let mut event_buf = Vec::new();
    let mut capture_buf = Vec::new();

    let mut connection = Connection::open(vehicles_db_file_path(work_dir, source_name))
        .map_err(|err| EzError::Io {
            message: format!("Failed to open transit vehicles SQLite database: {err}"),
        })?;
    let tx = connection.transaction().map_err(|err| EzError::Io {
        message: format!("Failed to begin transit vehicles import transaction: {err}"),
    })?;

    {
        let mut insert_vehicle_type = tx
            .prepare("INSERT INTO vehicle_types (id, raw_xml) VALUES (?1, ?2)")
            .map_err(sqlite_err)?;
        let mut insert_vehicle = tx
            .prepare("INSERT INTO vehicles (id, type, raw_xml) VALUES (?1, ?2, ?3)")
            .map_err(sqlite_err)?;

        loop {
            event_buf.clear();
            let event = reader
                .read_event_into(&mut event_buf)
                .map(|event| event.into_owned())
                .map_err(|err| EzError::Io {
                    message: format!("Failed to parse transit vehicles XML: {err}"),
                })?;

            // Phase 1: capture preamble before <vehicleDefinitions>
            if !in_root {
                match &event {
                    Event::Start(start)
                        if start.local_name().into_inner() == b"vehicleDefinitions" =>
                    {
                        append_event_bytes(&event, &mut preamble_buf);
                        root_tag_attributes_blob = capture_tag_attributes(start)?;
                        in_root = true;
                        continue;
                    }
                    Event::Eof => {
                        return Err(EzError::Io {
                            message:
                                "Vehicles XML ended before root <vehicleDefinitions> element."
                                    .into(),
                        });
                    }
                    _ => {
                        append_event_bytes(&event, &mut preamble_buf);
                        continue;
                    }
                }
            }

            // Phase 3: capture postamble after </vehicleDefinitions>
            if root_closed {
                match &event {
                    Event::Eof => break,
                    _ => {
                        append_event_bytes(&event, &mut postamble_buf);
                        continue;
                    }
                }
            }

            // Phase 2: body parsing
            match event {
                Event::Start(ref start)
                    if start.local_name().into_inner() == b"vehicleType" =>
                {
                    let id = required_attr_string(start, b"id", "vehicleType")?;
                    let raw = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    insert_vehicle_type
                        .execute(params![id, raw])
                        .map_err(sqlite_err)?;
                }

                Event::Start(ref start)
                    if start.local_name().into_inner() == b"vehicle" =>
                {
                    let id = required_attr_string(start, b"id", "vehicle")?;
                    let vtype = optional_attr_string(start, b"type")?;
                    let raw = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    insert_vehicle
                        .execute(params![id, vtype, raw])
                        .map_err(sqlite_err)?;
                }

                Event::Empty(ref start)
                    if start.local_name().into_inner() == b"vehicle" =>
                {
                    let id = required_attr_string(start, b"id", "vehicle")?;
                    let vtype = optional_attr_string(start, b"type")?;
                    let raw = capture_empty_element(start)?;
                    insert_vehicle
                        .execute(params![id, vtype, raw])
                        .map_err(sqlite_err)?;
                }

                Event::End(ref end)
                    if end.local_name().as_ref() == b"vehicleDefinitions" =>
                {
                    postamble_buf.push(b'<');
                    postamble_buf.push(b'/');
                    postamble_buf.extend_from_slice(end.as_ref());
                    postamble_buf.push(b'>');
                    root_closed = true;
                }

                Event::Eof => {
                    return Err(EzError::Io {
                        message:
                            "Vehicles XML ended before </vehicleDefinitions> closing tag."
                                .into(),
                    });
                }

                _ => {}
            }
        }
    }

    tx.commit().map_err(|err| EzError::Io {
        message: format!("Failed to commit transit vehicles import transaction: {err}"),
    })?;

    fs::write(vehicles_preamble_file_path(work_dir, source_name), &preamble_buf).map_err(
        |err| EzError::Io {
            message: format!("Failed to write transit vehicles preamble file: {err}"),
        },
    )?;
    fs::write(
        vehicles_postamble_file_path(work_dir, source_name),
        &postamble_buf,
    )
    .map_err(|err| EzError::Io {
        message: format!("Failed to write transit vehicles postamble file: {err}"),
    })?;

    Ok(root_tag_attributes_blob)
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write transit vehicles SQLite rows: {err}"),
    }
}
