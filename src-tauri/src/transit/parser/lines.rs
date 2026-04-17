use std::io::BufRead;

use quick_xml::{
    events::{BytesStart, Event},
    reader::Reader,
};
use rusqlite::{params, Statement};

use crate::ez::types::EzError;
use crate::network::capture::{capture_empty_element, capture_started_element};
use crate::transit::types::{
    DepartureRecord, RoutePathLinkRecord, RouteProfileStopRecord,
};
use crate::transit::xml_attrs::{
    optional_attr_bool, optional_attr_string, required_attr_string,
};

use super::util::{consume_to_end, sqlite_err};

pub(super) struct LineContext {
    pub id: String,
    pub name: Option<String>,
    pub attributes_blob: Option<String>,
}

pub(super) struct RouteContext {
    pub id: String,
    pub line_id: String,
    pub transport_mode: String,
    pub description: Option<String>,
    pub attributes_blob: Option<String>,
}

pub(super) fn parse_line_tag(
    start: &BytesStart<'_>,
) -> Result<(String, Option<String>), EzError> {
    let id = required_attr_string(start, b"id", "transitLine")?;
    let name = optional_attr_string(start, b"name")?;
    Ok((id, name))
}

pub(super) fn handle_profile_stop_empty(
    start: &BytesStart<'_>,
    route: &RouteContext,
    sequence: i32,
    insert_profile_stop: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record = parse_profile_stop(start, &route.line_id, &route.id, sequence)?;
    insert_profile_stop
        .execute(params![
            record.line_id,
            record.route_id,
            record.sequence,
            record.stop_ref_id,
            record.arrival_offset,
            record.departure_offset,
            record.allow_boarding as i32,
            record.allow_alighting as i32,
            record.await_departure as i32
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_profile_stop_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    route: &RouteContext,
    sequence: i32,
    event_buf: &mut Vec<u8>,
    insert_profile_stop: &mut Statement<'_>,
) -> Result<(), EzError> {
    handle_profile_stop_empty(start, route, sequence, insert_profile_stop)?;
    consume_to_end(reader, event_buf, b"stop")
}

pub(super) fn handle_path_link_empty(
    start: &BytesStart<'_>,
    route: &RouteContext,
    sequence: i32,
    insert_path_link: &mut Statement<'_>,
) -> Result<(), EzError> {
    let link_id = required_attr_string(start, b"refId", "link")?;
    let record = RoutePathLinkRecord {
        line_id: route.line_id.clone(),
        route_id: route.id.clone(),
        sequence,
        link_id,
    };
    insert_path_link
        .execute(params![
            record.line_id,
            record.route_id,
            record.sequence,
            record.link_id
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_path_link_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    route: &RouteContext,
    sequence: i32,
    event_buf: &mut Vec<u8>,
    insert_path_link: &mut Statement<'_>,
) -> Result<(), EzError> {
    handle_path_link_empty(start, route, sequence, insert_path_link)?;
    consume_to_end(reader, event_buf, b"link")
}

pub(super) fn handle_departure_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    route: &RouteContext,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
    insert_departure: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record =
        parse_departure_started(reader, start, &route.line_id, &route.id, event_buf, capture_buf)?;
    insert_departure
        .execute(params![
            record.id,
            record.line_id,
            record.route_id,
            record.departure_time,
            record.vehicle_ref_id,
            record.attributes_blob
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

pub(super) fn handle_departure_empty(
    start: &BytesStart<'_>,
    route: &RouteContext,
    insert_departure: &mut Statement<'_>,
) -> Result<(), EzError> {
    let record = parse_departure_empty(start, &route.line_id, &route.id)?;
    insert_departure
        .execute(params![
            record.id,
            record.line_id,
            record.route_id,
            record.departure_time,
            record.vehicle_ref_id,
            record.attributes_blob
        ])
        .map_err(sqlite_err)?;
    Ok(())
}

fn parse_profile_stop(
    start: &BytesStart<'_>,
    line_id: &str,
    route_id: &str,
    sequence: i32,
) -> Result<RouteProfileStopRecord, EzError> {
    let stop_ref_id = required_attr_string(start, b"refId", "stop")?;
    let arrival_offset = optional_attr_string(start, b"arrivalOffset")?;
    let departure_offset = optional_attr_string(start, b"departureOffset")?;
    let allow_boarding = optional_attr_bool(start, b"allowBoarding")?.unwrap_or(true);
    let allow_alighting = optional_attr_bool(start, b"allowAlighting")?.unwrap_or(true);
    let await_departure = optional_attr_bool(start, b"awaitDeparture")?.unwrap_or(false);

    Ok(RouteProfileStopRecord {
        line_id: line_id.to_string(),
        route_id: route_id.to_string(),
        sequence,
        stop_ref_id,
        arrival_offset,
        departure_offset,
        allow_boarding,
        allow_alighting,
        await_departure,
    })
}

fn parse_departure_started<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    line_id: &str,
    route_id: &str,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
) -> Result<DepartureRecord, EzError> {
    let id = required_attr_string(start, b"id", "departure")?;
    let departure_time = required_attr_string(start, b"departureTime", "departure")?;
    let vehicle_ref_id = optional_attr_string(start, b"vehicleRefId")?;

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
            Event::End(ref inner) if inner.local_name().as_ref() == b"departure" => {
                break;
            }
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF inside <departure> element.".into(),
                });
            }
            _ => {}
        }
    }

    Ok(DepartureRecord {
        id,
        line_id: line_id.to_string(),
        route_id: route_id.to_string(),
        departure_time,
        vehicle_ref_id,
        attributes_blob,
    })
}

fn parse_departure_empty(
    start: &BytesStart<'_>,
    line_id: &str,
    route_id: &str,
) -> Result<DepartureRecord, EzError> {
    let id = required_attr_string(start, b"id", "departure")?;
    let departure_time = required_attr_string(start, b"departureTime", "departure")?;
    let vehicle_ref_id = optional_attr_string(start, b"vehicleRefId")?;

    Ok(DepartureRecord {
        id,
        line_id: line_id.to_string(),
        route_id: route_id.to_string(),
        departure_time,
        vehicle_ref_id,
        attributes_blob: None,
    })
}
