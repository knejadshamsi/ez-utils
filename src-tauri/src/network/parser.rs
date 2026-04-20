//! Network import parser.
//!
//! Parses a MATSim network XML file into a SQLite database, capturing node
//! elements as raw XML blobs and link elements as separated tag-level and
//! child-attributes blobs for lossless round-tripping.

use std::{
    fs::{self, File},
    io::BufReader,
    path::Path,
};

use quick_xml::{
    events::{BytesStart, Event},
    reader::Reader,
};
use rusqlite::{params, Connection};

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, NetworkMetadata},
};

use crate::projection::{CrsConfig, Projector};

use super::{
    capture::{
        append_event_bytes, capture_empty_element, capture_started_element,
        capture_tag_attributes,
    },
    types::{LinkRecord, NodeRecord},
};

pub(crate) fn run(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<NetworkMetadata, EzError> {
    let file = File::open(file_path).map_err(|err| EzError::Io {
        message: format!("Failed to open network XML: {err}"),
    })?;
    let mut reader = Reader::from_reader(BufReader::new(file));
    reader.config_mut().trim_text(false);
    reader.config_mut().trim_markup_names_in_closing_tags = false;

    let projector = Projector::new(crs_config)?;
    let mut event_buf = Vec::new();
    let mut capture_buf = Vec::new();

    let mut network_attributes_blob: Option<String> = None;
    let mut links_tag_blob = String::new();
    let mut preamble_buf: Vec<u8> = Vec::new();
    let mut postamble_buf: Vec<u8> = Vec::new();
    let mut in_root = false;
    let mut root_closed = false;

    let mut connection =
        Connection::open(db_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
            message: format!("Failed to open network SQLite database: {err}"),
        })?;
    let tx = connection.transaction().map_err(|err| EzError::Io {
        message: format!("Failed to begin network import transaction: {err}"),
    })?;

    {
        let mut insert_node = tx
            .prepare(
                "INSERT INTO nodes (id, lng, lat, raw_xml) VALUES (?1, ?2, ?3, ?4)",
            )
            .map_err(sqlite_err)?;
        let mut insert_link = tx
            .prepare(
                "INSERT INTO links (id, from_node, to_node, tag_blob, attributes_blob)
                 VALUES (?1, ?2, ?3, ?4, ?5)",
            )
            .map_err(sqlite_err)?;

        let mut in_nodes = false;
        let mut in_links = false;

        loop {
            event_buf.clear();
            let event = reader
                .read_event_into(&mut event_buf)
                .map(|event| event.into_owned())
                .map_err(|err| EzError::Io {
                    message: format!("Failed to parse network XML: {err}"),
                })?;

            // Phase 1: capture everything before the root <network> tag verbatim
            // into the preamble buffer. Includes the root Start tag itself so the
            // preamble file ends with `<network ...>` ready to concatenate.
            if !in_root {
                match &event {
                    Event::Start(start) if start.local_name().into_inner() == b"network" => {
                        append_event_bytes(&event, &mut preamble_buf);
                        in_root = true;
                        continue;
                    }
                    Event::Eof => {
                        return Err(EzError::Io {
                            message: "Network XML ended before root <network> element.".into(),
                        });
                    }
                    _ => {
                        append_event_bytes(&event, &mut preamble_buf);
                        continue;
                    }
                }
            }

            // Phase 3: capture everything after the root </network> tag verbatim
            // into the postamble buffer.
            if root_closed {
                match &event {
                    Event::Eof => break,
                    _ => {
                        append_event_bytes(&event, &mut postamble_buf);
                        continue;
                    }
                }
            }

            // Phase 2: body parsing. Emits nodes/links into SQLite and collects
            // the root `<attributes>` blob + `<links>` tag attrs as metadata.
            match event {
                Event::Start(ref start)
                    if !in_nodes
                        && !in_links
                        && start.local_name().into_inner() == b"attributes" =>
                {
                    let blob = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    network_attributes_blob = Some(blob);
                }
                Event::Start(ref start) if start.local_name().into_inner() == b"nodes" => {
                    in_nodes = true;
                }
                Event::End(ref end) if end.local_name().as_ref() == b"nodes" => {
                    in_nodes = false;
                }
                Event::Start(ref start)
                    if in_nodes && start.local_name().into_inner() == b"node" =>
                {
                    let raw_xml = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    let record = parse_node_from_tag(start, &raw_xml, &projector)?;
                    insert_node
                        .execute(params![record.id, record.lng, record.lat, record.raw_xml])
                        .map_err(sqlite_err)?;
                }
                Event::Empty(ref start)
                    if in_nodes && start.local_name().into_inner() == b"node" =>
                {
                    let raw_xml = capture_empty_element(start)?;
                    let record = parse_node_from_tag(start, &raw_xml, &projector)?;
                    insert_node
                        .execute(params![record.id, record.lng, record.lat, record.raw_xml])
                        .map_err(sqlite_err)?;
                }
                Event::Start(ref start) if start.local_name().into_inner() == b"links" => {
                    in_links = true;
                    links_tag_blob = capture_tag_attributes(start)?;
                }
                Event::End(ref end) if end.local_name().as_ref() == b"links" => {
                    in_links = false;
                }
                Event::Start(ref start)
                    if in_links && start.local_name().into_inner() == b"link" =>
                {
                    let record = parse_link_started(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    insert_link
                        .execute(params![
                            record.id,
                            record.from_node,
                            record.to_node,
                            record.tag_blob,
                            record.attributes_blob
                        ])
                        .map_err(sqlite_err)?;
                }
                Event::Empty(ref start)
                    if in_links && start.local_name().into_inner() == b"link" =>
                {
                    let record = parse_link_empty(start)?;
                    insert_link
                        .execute(params![
                            record.id,
                            record.from_node,
                            record.to_node,
                            record.tag_blob,
                            record.attributes_blob
                        ])
                        .map_err(sqlite_err)?;
                }
                Event::End(ref end) if end.local_name().as_ref() == b"network" => {
                    postamble_buf.push(b'<');
                    postamble_buf.push(b'/');
                    postamble_buf.extend_from_slice(end.as_ref());
                    postamble_buf.push(b'>');
                    root_closed = true;
                }
                Event::Eof => {
                    return Err(EzError::Io {
                        message: "Network XML ended before </network> closing tag.".into(),
                    });
                }
                _ => {}
            }
        }
    }

    tx.commit().map_err(|err| EzError::Io {
        message: format!("Failed to commit network import transaction: {err}"),
    })?;

    fs::write(preamble_file_path(work_dir, source_name), &preamble_buf).map_err(|err| {
        EzError::Io {
            message: format!("Failed to write network preamble file: {err}"),
        }
    })?;
    fs::write(postamble_file_path(work_dir, source_name), &postamble_buf).map_err(|err| {
        EzError::Io {
            message: format!("Failed to write network postamble file: {err}"),
        }
    })?;

    Ok(NetworkMetadata {
        network_attributes_blob,
        links_tag_blob,
    })
}

fn parse_node_from_tag(
    start: &BytesStart<'_>,
    raw_xml: &str,
    projector: &Projector,
) -> Result<NodeRecord, EzError> {
    let id = required_attr_string(start, b"id", "node")?;
    let x = required_attr_f64(start, b"x", "node")?;
    let y = required_attr_f64(start, b"y", "node")?;
    let projected = projector.project_xy(x, y)?;

    Ok(NodeRecord {
        id,
        lng: projected.lng,
        lat: projected.lat,
        raw_xml: raw_xml.to_string(),
    })
}

fn parse_link_started<R: std::io::BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
) -> Result<LinkRecord, EzError> {
    let id = required_attr_string(start, b"id", "link")?;
    let from_node = required_attr_string(start, b"from", "link")?;
    let to_node = required_attr_string(start, b"to", "link")?;
    let tag_blob = capture_tag_attributes(start)?;

    let mut attributes_blob: Option<String> = None;
    let mut depth = 0_i32;

    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse network XML: {err}"),
            })?;

        match event {
            Event::Start(ref inner) if inner.local_name().into_inner() == b"attributes" && depth == 0 => {
                let blob = capture_started_element(reader, inner, event_buf, capture_buf)?;
                attributes_blob = Some(blob);
            }
            Event::Empty(ref inner) if inner.local_name().into_inner() == b"attributes" && depth == 0 => {
                let blob = capture_empty_element(inner)?;
                attributes_blob = Some(blob);
            }
            Event::Start(ref inner) => {
                if inner.local_name().into_inner() == b"link" {
                    depth += 1;
                }
            }
            Event::End(ref inner) => {
                if inner.local_name().as_ref() == b"link" {
                    if depth == 0 {
                        break;
                    }
                    depth -= 1;
                }
            }
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF inside <link> element.".into(),
                });
            }
            _ => {}
        }
    }

    Ok(LinkRecord {
        id,
        from_node,
        to_node,
        tag_blob,
        attributes_blob,
    })
}

fn parse_link_empty(start: &BytesStart<'_>) -> Result<LinkRecord, EzError> {
    let id = required_attr_string(start, b"id", "link")?;
    let from_node = required_attr_string(start, b"from", "link")?;
    let to_node = required_attr_string(start, b"to", "link")?;
    let tag_blob = capture_tag_attributes(start)?;

    Ok(LinkRecord {
        id,
        from_node,
        to_node,
        tag_blob,
        attributes_blob: None,
    })
}

fn required_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
    element_name: &str,
) -> Result<String, EzError> {
    optional_attr_string(start, key)?.ok_or_else(|| EzError::Io {
        message: format!("Missing required attribute on <{element_name}>."),
    })
}

fn required_attr_f64(
    start: &BytesStart<'_>,
    key: &[u8],
    element_name: &str,
) -> Result<f64, EzError> {
    let value = required_attr_string(start, key, element_name)?;
    value.parse::<f64>().map_err(|err| EzError::Io {
        message: format!("Failed to parse numeric attribute on <{element_name}>: {err}"),
    })
}

fn optional_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
) -> Result<Option<String>, EzError> {
    let attr = start
        .try_get_attribute(key)
        .map_err(|err| EzError::Io {
            message: format!("Failed to read XML attribute: {err}"),
        })?;

    attr.map(|attr| {
        std::str::from_utf8(attr.value.as_ref())
            .map(|value| value.to_string())
            .map_err(|err| EzError::Io {
                message: format!("XML attribute was not valid UTF-8: {err}"),
            })
    })
    .transpose()
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write network SQLite rows: {err}"),
    }
}
