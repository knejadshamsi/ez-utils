use std::io::BufRead;

use quick_xml::{
    events::{BytesStart, Event},
    reader::Reader,
};
use rusqlite::{params, Statement};

use crate::ez::types::EzError;
use crate::network::capture::{capture_empty_element, capture_started_element};
use crate::projection::Projector;
use crate::transit::types::{MinimalTransferTimeRecord, StopFacilityRecord};
use crate::transit::xml_attrs::{
    optional_attr_bool, optional_attr_string, required_attr_f64, required_attr_string,
};

use super::util::{consume_to_end, sqlite_err};

pub(super) fn handle_stop_facility_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    projector: &Projector,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
    insert_stop: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record = parse_stop_facility_started(reader, start, projector, event_buf, capture_buf)?;
    insert_stop
        .execute(params![
            record.id,
            record.lng,
            record.lat,
            record.name,
            record.link_ref_id,
            record.stop_area_id,
            record.is_blocking as i32,
            record.attributes_blob
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_stop_facility_empty(
    start: &BytesStart<'_>,
    projector: &Projector,
    insert_stop: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record = parse_stop_facility_empty(start, projector)?;
    insert_stop
        .execute(params![
            record.id,
            record.lng,
            record.lat,
            record.name,
            record.link_ref_id,
            record.stop_area_id,
            record.is_blocking as i32,
            record.attributes_blob
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_transfer_relation_empty(
    start: &BytesStart<'_>,
    insert_transfer: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record = parse_transfer_relation(start)?;
    insert_transfer
        .execute(params![record.from_stop, record.to_stop, record.transfer_time])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_transfer_relation_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    event_buf: &mut Vec<u8>,
    insert_transfer: &mut Statement<'_>,
) -> Result<(), EzError> {
    handle_transfer_relation_empty(start, insert_transfer)?;
    consume_to_end(reader, event_buf, b"relation")
}

fn parse_stop_facility_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    projector: &Projector,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
) -> Result<StopFacilityRecord, EzError> {
    let id = required_attr_string(start, b"id", "stopFacility")?;
    let x = required_attr_f64(start, b"x", "stopFacility")?;
    let y = required_attr_f64(start, b"y", "stopFacility")?;
    let name = optional_attr_string(start, b"name")?;
    let link_ref_id = optional_attr_string(start, b"linkRefId")?;
    let stop_area_id = optional_attr_string(start, b"stopAreaId")?;
    let is_blocking = optional_attr_bool(start, b"isBlocking")?.unwrap_or(false);
    let projected = projector.project_xy(x, y)?;

    let mut attributes_blob: Option<String> = None;

    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse transit XML: {err}"),
            })?;

        match event {
            Event::Start(ref inner) if inner.local_name().into_inner() == b"attributes" => {
                let blob = capture_started_element(reader, inner, event_buf, capture_buf)?;
                attributes_blob = Some(blob);
            }
            Event::Empty(ref inner) if inner.local_name().into_inner() == b"attributes" => {
                let blob = capture_empty_element(inner)?;
                attributes_blob = Some(blob);
            }
            Event::End(ref inner) if inner.local_name().as_ref() == b"stopFacility" => {
                break;
            }
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF inside <stopFacility> element.".into(),
                });
            }
            _ => {}
        }
    }

    Ok(StopFacilityRecord {
        id,
        lng: projected.lng,
        lat: projected.lat,
        name,
        link_ref_id,
        stop_area_id,
        is_blocking,
        attributes_blob,
    })
}

fn parse_stop_facility_empty(
    start: &BytesStart<'_>,
    projector: &Projector,
) -> Result<StopFacilityRecord, EzError> {
    let id = required_attr_string(start, b"id", "stopFacility")?;
    let x = required_attr_f64(start, b"x", "stopFacility")?;
    let y = required_attr_f64(start, b"y", "stopFacility")?;
    let name = optional_attr_string(start, b"name")?;
    let link_ref_id = optional_attr_string(start, b"linkRefId")?;
    let stop_area_id = optional_attr_string(start, b"stopAreaId")?;
    let is_blocking = optional_attr_bool(start, b"isBlocking")?.unwrap_or(false);
    let projected = projector.project_xy(x, y)?;

    Ok(StopFacilityRecord {
        id,
        lng: projected.lng,
        lat: projected.lat,
        name,
        link_ref_id,
        stop_area_id,
        is_blocking,
        attributes_blob: None,
    })
}

fn parse_transfer_relation(start: &BytesStart<'_>) -> Result<MinimalTransferTimeRecord, EzError> {
    let from_stop = required_attr_string(start, b"fromStop", "relation")?;
    let to_stop = required_attr_string(start, b"toStop", "relation")?;
    let transfer_time = required_attr_f64(start, b"transferTime", "relation")?;
    Ok(MinimalTransferTimeRecord {
        from_stop,
        to_stop,
        transfer_time,
    })
}
