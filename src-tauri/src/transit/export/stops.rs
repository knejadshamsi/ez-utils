use std::io::Write;

use rusqlite::Connection;

use crate::ez::types::EzError;
use crate::projection::Projector;
use crate::utils::xml_escape;

use super::helpers::{format_decimal, io_write_err, sqlite_err, write_indented};

struct StopRow {
    id: String,
    lng: f64,
    lat: f64,
    name: Option<String>,
    link_ref_id: Option<String>,
    stop_area_id: Option<String>,
    is_blocking: bool,
    attributes_blob: Option<String>,
}

pub(super) fn write_stop_facilities<W: Write>(
    writer: &mut W,
    connection: &Connection,
    projector: &Projector,
) -> Result<(), EzError> {
    let mut stmt = connection
        .prepare(
            "SELECT id, lng, lat, name, link_ref_id, stop_area_id, is_blocking, attributes_blob
             FROM stop_facilities ORDER BY id",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([], |row| {
            Ok(StopRow {
                id: row.get(0)?,
                lng: row.get(1)?,
                lat: row.get(2)?,
                name: row.get(3)?,
                link_ref_id: row.get(4)?,
                stop_area_id: row.get(5)?,
                is_blocking: row.get::<_, i32>(6)? != 0,
                attributes_blob: row.get(7)?,
            })
        })
        .map_err(sqlite_err)?;
    for row in rows {
        let r = row.map_err(sqlite_err)?;
        let native = projector.unproject_lng_lat(r.lng, r.lat)?;
        let mut attrs = format!(
            "id=\"{}\" x=\"{}\" y=\"{}\"",
            xml_escape(&r.id),
            format_decimal(native.x),
            format_decimal(native.y),
        );
        if let Some(name) = r.name.as_deref() {
            attrs.push_str(&format!(" name=\"{}\"", xml_escape(name)));
        }
        if let Some(link) = r.link_ref_id.as_deref() {
            attrs.push_str(&format!(" linkRefId=\"{}\"", xml_escape(link)));
        }
        if let Some(area) = r.stop_area_id.as_deref() {
            attrs.push_str(&format!(" stopAreaId=\"{}\"", xml_escape(area)));
        }
        if r.is_blocking {
            attrs.push_str(" isBlocking=\"true\"");
        }
        match r.attributes_blob.as_deref() {
            None => {
                write!(writer, "    <stopFacility {attrs}/>\n").map_err(io_write_err)?;
            }
            Some(blob) => {
                write!(writer, "    <stopFacility {attrs}>\n").map_err(io_write_err)?;
                write_indented(writer, blob.trim(), "      ")?;
                writer.write_all(b"    </stopFacility>\n").map_err(io_write_err)?;
            }
        }
    }
    Ok(())
}

pub(super) fn write_transfer_times<W: Write>(
    writer: &mut W,
    connection: &Connection,
) -> Result<(), EzError> {
    let mut stmt = connection
        .prepare(
            "SELECT from_stop, to_stop, transfer_time FROM minimal_transfer_times
             ORDER BY from_stop, to_stop",
        )
        .map_err(sqlite_err)?;
    let rows = stmt
        .query_map([], |row| {
            Ok((
                row.get::<_, String>(0)?,
                row.get::<_, String>(1)?,
                row.get::<_, f64>(2)?,
            ))
        })
        .map_err(sqlite_err)?;
    for row in rows {
        let (from, to, time) = row.map_err(sqlite_err)?;
        write!(
            writer,
            "    <relation fromStop=\"{}\" toStop=\"{}\" transferTime=\"{}\"/>\n",
            xml_escape(&from),
            xml_escape(&to),
            format_decimal(time),
        )
        .map_err(io_write_err)?;
    }
    Ok(())
}
