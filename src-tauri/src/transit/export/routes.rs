use std::io::Write;

use rusqlite::Connection;

use crate::ez::types::EzError;
use crate::utils::xml_escape;

use super::helpers::{io_write_err, sqlite_err, write_indented};

pub(super) fn write_transit_lines<W: Write>(
    writer: &mut W,
    connection: &Connection,
) -> Result<(), EzError> {
    let mut lines_stmt = connection
        .prepare("SELECT id, name, attributes_blob FROM lines ORDER BY id")
        .map_err(sqlite_err)?;
    let lines = lines_stmt
        .query_map([], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, Option<String>>(1)?,
                row.get::<_, Option<String>>(2)?,
            ))
        })
        .map_err(sqlite_err)?;
    for line_row in lines {
        let (line_id, line_name, line_attrs) = line_row.map_err(sqlite_err)?;
        let mut line_attrs_tag = format!("id=\"{}\"", xml_escape(&line_id));
        if let Some(n) = line_name.as_deref() {
            line_attrs_tag.push_str(&format!(" name=\"{}\"", xml_escape(n)));
        }
        write!(writer, "    <transitLine {line_attrs_tag}>\n").map_err(io_write_err)?;
        if let Some(blob) = line_attrs.as_deref() {
            write_indented(writer, blob.trim(), "      ")?;
        }
        write_routes_for_line(writer, connection, &line_id)?;
        writer.write_all(b"    </transitLine>\n").map_err(io_write_err)?;
    }
    Ok(())
}

fn write_routes_for_line<W: Write>(
    writer: &mut W,
    connection: &Connection,
    line_id: &str,
) -> Result<(), EzError> {
    let mut stmt = connection
        .prepare(
            "SELECT id, transport_mode, description, attributes_blob
             FROM routes WHERE line_id = ?1 ORDER BY id",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([line_id], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, String>(1)?,
                row.get::<_, Option<String>>(2)?,
                row.get::<_, Option<String>>(3)?,
            ))
        })
        .map_err(sqlite_err)?;
    for row in rows {
        let (route_id, transport_mode, description, route_attrs) = row.map_err(sqlite_err)?;
        write!(
            writer,
            "      <transitRoute id=\"{}\">\n",
            xml_escape(&route_id),
        )
        .map_err(io_write_err)?;
        if let Some(blob) = route_attrs.as_deref() {
            write_indented(writer, blob.trim(), "        ")?;
        }
        write!(
            writer,
            "        <transportMode>{}</transportMode>\n",
            xml_escape(transport_mode.trim()),
        )
        .map_err(io_write_err)?;
        if let Some(desc) = description.as_deref() {
            let trimmed = desc.trim();
            if !trimmed.is_empty() {
                write!(
                    writer,
                    "        <description>{}</description>\n",
                    xml_escape(trimmed),
                )
                .map_err(io_write_err)?;
            }
        }
        write_route_profile(writer, connection, line_id, &route_id)?;
        write_route_path(writer, connection, line_id, &route_id)?;
        write_route_departures(writer, connection, line_id, &route_id)?;
        writer.write_all(b"      </transitRoute>\n").map_err(io_write_err)?;
    }
    Ok(())
}

fn write_route_profile<W: Write>(
    writer: &mut W,
    connection: &Connection,
    line_id: &str,
    route_id: &str,
) -> Result<(), EzError> {
    writer.write_all(b"        <routeProfile>\n").map_err(io_write_err)?;
    let mut stmt = connection
        .prepare(
            "SELECT stop_ref_id, arrival_offset, departure_offset,
                    allow_boarding, allow_alighting, await_departure
             FROM route_profile_stops
             WHERE line_id = ?1 AND route_id = ?2
             ORDER BY sequence",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([line_id, route_id], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, Option<String>>(1)?,
                row.get::<_, Option<String>>(2)?,
                row.get::<_, i32>(3)? != 0,
                row.get::<_, i32>(4)? != 0,
                row.get::<_, i32>(5)? != 0,
            ))
        })
        .map_err(sqlite_err)?;
    for row in rows {
        let (ref_id, arr, dep, allow_b, allow_a, await_dep) = row.map_err(sqlite_err)?;
        let mut attrs = format!("refId=\"{}\"", xml_escape(&ref_id));
        if let Some(a) = arr.as_deref() {
            attrs.push_str(&format!(" arrivalOffset=\"{}\"", xml_escape(a)));
        }
        if let Some(d) = dep.as_deref() {
            attrs.push_str(&format!(" departureOffset=\"{}\"", xml_escape(d)));
        }
        if !allow_b {
            attrs.push_str(" allowBoarding=\"false\"");
        }
        if !allow_a {
            attrs.push_str(" allowAlighting=\"false\"");
        }
        if await_dep {
            attrs.push_str(" awaitDeparture=\"true\"");
        }
        write!(writer, "          <stop {attrs}/>\n").map_err(io_write_err)?;
    }
    writer.write_all(b"        </routeProfile>\n").map_err(io_write_err)?;
    Ok(())
}

fn write_route_path<W: Write>(
    writer: &mut W,
    connection: &Connection,
    line_id: &str,
    route_id: &str,
) -> Result<(), EzError> {
    let count: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM route_path_links WHERE line_id = ?1 AND route_id = ?2",
            [line_id, route_id],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;
    if count == 0 {
        return Ok(());
    }
    writer.write_all(b"        <route>\n").map_err(io_write_err)?;
    let mut stmt = connection
        .prepare(
            "SELECT link_id FROM route_path_links
             WHERE line_id = ?1 AND route_id = ?2 ORDER BY sequence",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([line_id, route_id], |row| row.get::<_, String>(0))
        .map_err(sqlite_err)?;
    for row in rows {
        let link_id = row.map_err(sqlite_err)?;
        write!(writer, "          <link refId=\"{}\"/>\n", xml_escape(&link_id))
            .map_err(io_write_err)?;
    }
    writer.write_all(b"        </route>\n").map_err(io_write_err)?;
    Ok(())
}

fn write_route_departures<W: Write>(
    writer: &mut W,
    connection: &Connection,
    line_id: &str,
    route_id: &str,
) -> Result<(), EzError> {
    writer.write_all(b"        <departures>\n").map_err(io_write_err)?;
    let mut stmt = connection
        .prepare(
            "SELECT id, departure_time, vehicle_ref_id, attributes_blob
             FROM departures WHERE line_id = ?1 AND route_id = ?2
             ORDER BY departure_time, id",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([line_id, route_id], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, String>(1)?,
                row.get::<_, Option<String>>(2)?,
                row.get::<_, Option<String>>(3)?,
            ))
        })
        .map_err(sqlite_err)?;
    for row in rows {
        let (id, time, veh, attrs_blob) = row.map_err(sqlite_err)?;
        let mut attrs = format!(
            "id=\"{}\" departureTime=\"{}\"",
            xml_escape(&id),
            xml_escape(&time),
        );
        if let Some(v) = veh.as_deref() {
            attrs.push_str(&format!(" vehicleRefId=\"{}\"", xml_escape(v)));
        }
        match attrs_blob.as_deref() {
            None => write!(writer, "          <departure {attrs}/>\n").map_err(io_write_err)?,
            Some(blob) => {
                write!(writer, "          <departure {attrs}>\n").map_err(io_write_err)?;
                write_indented(writer, blob.trim(), "            ")?;
                writer.write_all(b"          </departure>\n").map_err(io_write_err)?;
            }
        }
    }
    writer.write_all(b"        </departures>\n").map_err(io_write_err)?;
    Ok(())
}
