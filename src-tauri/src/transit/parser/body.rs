use std::io::BufRead;

use quick_xml::{events::Event, reader::Reader};
use rusqlite::{params, Transaction};

use crate::ez::types::EzError;
use crate::network::capture::{
    append_event_bytes, capture_started_element, capture_tag_attributes,
};
use crate::projection::Projector;
use crate::transit::xml_attrs::required_attr_string;

use super::lines::{
    handle_departure_empty, handle_departure_started, handle_path_link_empty,
    handle_path_link_started, handle_profile_stop_empty, handle_profile_stop_started,
    parse_line_tag, LineContext, RouteContext,
};
use super::stops::{
    handle_stop_facility_empty, handle_stop_facility_started, handle_transfer_relation_empty,
    handle_transfer_relation_started,
};
use super::util::{read_text_content, sqlite_err};

pub(super) struct BodyState {
    pub transit_schedule_attributes_blob: Option<String>,
    pub transit_stops_tag_blob: String,
    pub transit_minimal_transfers_tag_blob: String,
    pub transit_lines_tag_blob: String,
    pub preamble_buf: Vec<u8>,
    pub postamble_buf: Vec<u8>,
}

pub(super) fn parse<R: BufRead>(
    reader: &mut Reader<R>,
    tx: &Transaction<'_>,
    projector: &Projector,
    event_buf: &mut Vec<u8>,
    capture_buf: &mut Vec<u8>,
) -> Result<BodyState, EzError> {
    let mut transit_schedule_attributes_blob: Option<String> = None;
    let mut transit_stops_tag_blob = String::new();
    let mut transit_minimal_transfers_tag_blob = String::new();
    let mut transit_lines_tag_blob = String::new();
    let mut preamble_buf: Vec<u8> = Vec::new();
    let mut postamble_buf: Vec<u8> = Vec::new();
    let mut in_root = false;
    let mut root_closed = false;

    let mut insert_stop = tx
        .prepare(
            "INSERT INTO stop_facilities
             (id, lng, lat, name, link_ref_id, stop_area_id, is_blocking, attributes_blob)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
        )
        .map_err(sqlite_err)?;
    let mut insert_line = tx
        .prepare("INSERT INTO lines (id, name, attributes_blob) VALUES (?1, ?2, ?3)")
        .map_err(sqlite_err)?;
    let mut insert_route = tx
        .prepare(
            "INSERT INTO routes (id, line_id, transport_mode, description, attributes_blob)
             VALUES (?1, ?2, ?3, ?4, ?5)",
        )
        .map_err(sqlite_err)?;
    let mut insert_profile_stop = tx
        .prepare(
            "INSERT INTO route_profile_stops
             (line_id, route_id, sequence, stop_ref_id, arrival_offset, departure_offset,
              allow_boarding, allow_alighting, await_departure)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)",
        )
        .map_err(sqlite_err)?;
    let mut insert_path_link = tx
        .prepare(
            "INSERT INTO route_path_links (line_id, route_id, sequence, link_id)
             VALUES (?1, ?2, ?3, ?4)",
        )
        .map_err(sqlite_err)?;
    let mut insert_departure = tx
        .prepare(
            "INSERT INTO departures
             (id, line_id, route_id, departure_time, vehicle_ref_id, attributes_blob)
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        )
        .map_err(sqlite_err)?;
    let mut insert_transfer = tx
        .prepare(
            "INSERT INTO minimal_transfer_times (from_stop, to_stop, transfer_time)
             VALUES (?1, ?2, ?3)",
        )
        .map_err(sqlite_err)?;

    let mut in_transit_stops = false;
    let mut in_minimal_transfers = false;
    let mut in_transit_lines = false;
    let mut current_line: Option<LineContext> = None;
    let mut current_route: Option<RouteContext> = None;
    let mut in_route_profile = false;
    let mut in_route_path = false;
    let mut in_departures = false;
    let mut profile_stop_sequence = 0;
    let mut path_link_sequence = 0;

    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse transit XML: {err}"),
            })?;

        // Phase 1: preamble
        if !in_root {
            match &event {
                Event::Start(start) if start.local_name().into_inner() == b"transitSchedule" => {
                    append_event_bytes(&event, &mut preamble_buf);
                    in_root = true;
                    continue;
                }
                Event::Eof => {
                    return Err(EzError::Io {
                        message: "Transit XML ended before root <transitSchedule> element.".into(),
                    });
                }
                _ => {
                    append_event_bytes(&event, &mut preamble_buf);
                    continue;
                }
            }
        }

        // Phase 3: postamble
        if root_closed {
            match &event {
                Event::Eof => break,
                _ => {
                    append_event_bytes(&event, &mut postamble_buf);
                    continue;
                }
            }
        }

        // Phase 2: body
        match event {
            Event::Start(ref start)
                if !in_transit_stops
                    && !in_minimal_transfers
                    && !in_transit_lines
                    && current_line.is_none()
                    && start.local_name().into_inner() == b"attributes" =>
            {
                let blob =
                    capture_started_element(reader, start, event_buf, capture_buf)?;
                transit_schedule_attributes_blob = Some(blob);
            }

            Event::Start(ref start) if start.local_name().into_inner() == b"transitStops" => {
                in_transit_stops = true;
                transit_stops_tag_blob = capture_tag_attributes(start)?;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"transitStops" => {
                in_transit_stops = false;
            }

            Event::Start(ref start)
                if in_transit_stops && start.local_name().into_inner() == b"stopFacility" =>
            {
                handle_stop_facility_started(
                    reader,
                    start,
                    projector,
                    event_buf,
                    capture_buf,
                    &mut insert_stop,
                )?;
            }
            Event::Empty(ref start)
                if in_transit_stops && start.local_name().into_inner() == b"stopFacility" =>
            {
                handle_stop_facility_empty(start, projector, &mut insert_stop)?;
            }

            Event::Start(ref start)
                if start.local_name().into_inner() == b"minimalTransferTimes" =>
            {
                in_minimal_transfers = true;
                transit_minimal_transfers_tag_blob = capture_tag_attributes(start)?;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"minimalTransferTimes" => {
                in_minimal_transfers = false;
            }

            Event::Empty(ref start)
                if in_minimal_transfers && start.local_name().into_inner() == b"relation" =>
            {
                handle_transfer_relation_empty(start, &mut insert_transfer)?;
            }
            Event::Start(ref start)
                if in_minimal_transfers && start.local_name().into_inner() == b"relation" =>
            {
                handle_transfer_relation_started(reader, start, event_buf, &mut insert_transfer)?;
            }

            Event::Start(ref start) if start.local_name().into_inner() == b"transitLines" => {
                in_transit_lines = true;
                transit_lines_tag_blob = capture_tag_attributes(start)?;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"transitLines" => {
                in_transit_lines = false;
            }

            Event::Start(ref start) if start.local_name().into_inner() == b"transitLine" => {
                let (id, name) = parse_line_tag(start)?;
                current_line = Some(LineContext { id, name, attributes_blob: None });
            }
            Event::End(ref end) if end.local_name().as_ref() == b"transitLine" => {
                if let Some(line) = current_line.take() {
                    insert_line
                        .execute(params![line.id, line.name, line.attributes_blob])
                        .map_err(sqlite_err)?;
                }
            }

            Event::Start(ref start)
                if current_line.is_some()
                    && current_route.is_none()
                    && start.local_name().into_inner() == b"attributes" =>
            {
                let blob =
                    capture_started_element(reader, start, event_buf, capture_buf)?;
                if let Some(ref mut line) = current_line {
                    line.attributes_blob = Some(blob);
                }
            }

            Event::Start(ref start)
                if current_line.is_some()
                    && start.local_name().into_inner() == b"transitRoute" =>
            {
                let id = required_attr_string(start, b"id", "transitRoute")?;
                current_route = Some(RouteContext {
                    id,
                    line_id: current_line.as_ref().unwrap().id.clone(),
                    transport_mode: String::new(),
                    description: None,
                    attributes_blob: None,
                });
                profile_stop_sequence = 0;
                path_link_sequence = 0;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"transitRoute" => {
                if let Some(route) = current_route.take() {
                    insert_route
                        .execute(params![
                            route.id,
                            route.line_id,
                            route.transport_mode,
                            route.description,
                            route.attributes_blob
                        ])
                        .map_err(sqlite_err)?;
                }
                in_route_profile = false;
                in_route_path = false;
                in_departures = false;
            }

            Event::Start(ref start)
                if current_route.is_some()
                    && !in_route_profile
                    && !in_departures
                    && start.local_name().into_inner() == b"attributes" =>
            {
                let blob =
                    capture_started_element(reader, start, event_buf, capture_buf)?;
                if let Some(ref mut route) = current_route {
                    route.attributes_blob = Some(blob);
                }
            }

            Event::Start(ref start)
                if current_route.is_some() && start.local_name().into_inner() == b"description" =>
            {
                let text = read_text_content(reader, event_buf)?;
                if let Some(ref mut route) = current_route {
                    route.description = Some(text);
                }
            }

            Event::Start(ref start)
                if current_route.is_some() && start.local_name().into_inner() == b"transportMode" =>
            {
                let text = read_text_content(reader, event_buf)?;
                if let Some(ref mut route) = current_route {
                    route.transport_mode = text;
                }
            }

            Event::Start(ref start)
                if current_route.is_some() && start.local_name().into_inner() == b"routeProfile" =>
            {
                in_route_profile = true;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"routeProfile" => {
                in_route_profile = false;
            }

            Event::Empty(ref start)
                if in_route_profile && start.local_name().into_inner() == b"stop" =>
            {
                if let Some(ref route) = current_route {
                    handle_profile_stop_empty(
                        start,
                        route,
                        profile_stop_sequence,
                        &mut insert_profile_stop,
                    )?;
                    profile_stop_sequence += 1;
                }
            }
            Event::Start(ref start)
                if in_route_profile && start.local_name().into_inner() == b"stop" =>
            {
                if let Some(ref route) = current_route {
                    handle_profile_stop_started(
                        reader,
                        start,
                        route,
                        profile_stop_sequence,
                        event_buf,
                        &mut insert_profile_stop,
                    )?;
                    profile_stop_sequence += 1;
                } else {
                    super::util::consume_to_end(reader, event_buf, b"stop")?;
                }
            }

            Event::Start(ref start)
                if current_route.is_some() && start.local_name().into_inner() == b"route" =>
            {
                in_route_path = true;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"route" => {
                in_route_path = false;
            }

            Event::Empty(ref start)
                if in_route_path && start.local_name().into_inner() == b"link" =>
            {
                if let Some(ref route) = current_route {
                    handle_path_link_empty(
                        start,
                        route,
                        path_link_sequence,
                        &mut insert_path_link,
                    )?;
                    path_link_sequence += 1;
                }
            }
            Event::Start(ref start)
                if in_route_path && start.local_name().into_inner() == b"link" =>
            {
                if let Some(ref route) = current_route {
                    handle_path_link_started(
                        reader,
                        start,
                        route,
                        path_link_sequence,
                        event_buf,
                        &mut insert_path_link,
                    )?;
                    path_link_sequence += 1;
                } else {
                    super::util::consume_to_end(reader, event_buf, b"link")?;
                }
            }

            Event::Start(ref start)
                if current_route.is_some() && start.local_name().into_inner() == b"departures" =>
            {
                in_departures = true;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"departures" => {
                in_departures = false;
            }

            Event::Start(ref start)
                if in_departures && start.local_name().into_inner() == b"departure" =>
            {
                if let Some(ref route) = current_route {
                    handle_departure_started(
                        reader,
                        start,
                        route,
                        event_buf,
                        capture_buf,
                        &mut insert_departure,
                    )?;
                }
            }
            Event::Empty(ref start)
                if in_departures && start.local_name().into_inner() == b"departure" =>
            {
                if let Some(ref route) = current_route {
                    handle_departure_empty(start, route, &mut insert_departure)?;
                }
            }

            Event::End(ref end) if end.local_name().as_ref() == b"transitSchedule" => {
                postamble_buf.push(b'<');
                postamble_buf.push(b'/');
                postamble_buf.extend_from_slice(end.as_ref());
                postamble_buf.push(b'>');
                root_closed = true;
            }

            Event::Eof => {
                return Err(EzError::Io {
                    message: "Transit XML ended before </transitSchedule> closing tag.".into(),
                });
            }

            _ => {}
        }
    }

    Ok(BodyState {
        transit_schedule_attributes_blob,
        transit_stops_tag_blob,
        transit_minimal_transfers_tag_blob,
        transit_lines_tag_blob,
        preamble_buf,
        postamble_buf,
    })
}
