use std::collections::HashMap;
use std::path::{Path, PathBuf};

use rusqlite::{params, Connection};

use crate::ez::{fs::db_file_path, types::EzError};

use super::types::{GeoPoint, LinePathPoint, LinkedRoutePath, ListedDeparture, ListedPathLink, ListedProfileStop};

pub(super) fn list_profile_stops(
    work_dir: &Path,
    source_name: &str,
    line_id: &str,
    route_id: &str,
) -> Result<Vec<ListedProfileStop>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT rps.sequence,
                    rps.stop_ref_id,
                    sf.name,
                    sf.lng,
                    sf.lat,
                    sf.link_ref_id,
                    sf.stop_area_id,
                    sf.is_blocking,
                    rps.arrival_offset,
                    rps.departure_offset,
                    rps.allow_boarding,
                    rps.allow_alighting,
                    rps.await_departure
             FROM route_profile_stops AS rps
             LEFT JOIN stop_facilities AS sf ON sf.id = rps.stop_ref_id
             WHERE rps.line_id = ?1 AND rps.route_id = ?2
             ORDER BY rps.sequence",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare profile stops query: {err}"),
        })?;

    let rows = stmt
        .query_map([line_id, route_id], |row| {
            let is_blocking_raw: Option<i32> = row.get(7)?;
            Ok(ListedProfileStop {
                sequence: row.get(0)?,
                stop_ref_id: row.get(1)?,
                stop_name: row.get(2)?,
                stop_lng: row.get(3)?,
                stop_lat: row.get(4)?,
                stop_link_ref_id: row.get(5)?,
                stop_area_id: row.get(6)?,
                stop_is_blocking: is_blocking_raw.map(|v| v != 0),
                arrival_offset: row.get(8)?,
                departure_offset: row.get(9)?,
                allow_boarding: row.get::<_, i32>(10)? != 0,
                allow_alighting: row.get::<_, i32>(11)? != 0,
                await_departure: row.get::<_, i32>(12)? != 0,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run profile stops query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read profile stop row: {err}"),
        })?);
    }
    Ok(out)
}

pub(super) fn list_path_links(
    work_dir: &Path,
    source_name: &str,
    line_id: &str,
    route_id: &str,
) -> Result<Vec<ListedPathLink>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT sequence, link_id
             FROM route_path_links
             WHERE line_id = ?1 AND route_id = ?2
             ORDER BY sequence",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare path links query: {err}"),
        })?;

    let rows = stmt
        .query_map([line_id, route_id], |row| {
            Ok(ListedPathLink {
                sequence: row.get(0)?,
                link_id: row.get(1)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run path links query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read path link row: {err}"),
        })?);
    }
    Ok(out)
}

pub(super) fn list_line_profile_stops(
    work_dir: &Path,
    source_name: &str,
    line_id: &str,
) -> Result<Vec<LinePathPoint>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT rps.route_id, rps.sequence, sf.lng, sf.lat
             FROM route_profile_stops AS rps
             JOIN stop_facilities AS sf ON sf.id = rps.stop_ref_id
             WHERE rps.line_id = ?1
             ORDER BY rps.route_id, rps.sequence",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare line profile stops query: {err}"),
        })?;

    let rows = stmt
        .query_map([line_id], |row| {
            Ok(LinePathPoint {
                route_id: row.get(0)?,
                sequence: row.get(1)?,
                lng: row.get(2)?,
                lat: row.get(3)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run line profile stops query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read line profile stop row: {err}"),
        })?);
    }
    Ok(out)
}

/// Resolves a transit route's sequence of network link IDs into an ordered polyline,
/// joining against the given network source. Atomic: any unresolved link_id -> Ok(None).
pub(super) fn resolve_route_path_geometry(
    work_dir: &Path,
    transit_source: &str,
    network_source: &str,
    line_id: &str,
    route_id: &str,
) -> Result<Option<LinkedRoutePath>, EzError> {
    let transit_db = db_file_path(work_dir, transit_source);
    if !transit_db.exists() {
        return Err(EzError::SourceNotFound {
            name: transit_source.to_string(),
        });
    }
    let transit_conn = Connection::open(&transit_db).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = transit_conn
        .prepare(
            "SELECT link_id FROM route_path_links
             WHERE line_id = ?1 AND route_id = ?2
             ORDER BY sequence",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare path-links query: {err}"),
        })?;
    let link_ids: Vec<String> = stmt
        .query_map([line_id, route_id], |row| row.get::<_, String>(0))
        .map_err(|err| EzError::Io {
            message: format!("Failed to run path-links query: {err}"),
        })?
        .collect::<Result<Vec<_>, _>>()
        .map_err(|err| EzError::Io {
            message: format!("Failed to read path-link row: {err}"),
        })?;
    if link_ids.is_empty() {
        return Ok(None);
    }

    let network_db = db_file_path(work_dir, network_source);
    if !network_db.exists() {
        return Ok(None);
    }
    let net_conn = Connection::open(&network_db).map_err(|err| EzError::Io {
        message: format!("Failed to open network SQLite database: {err}"),
    })?;
    net_conn
        .execute("CREATE TEMP TABLE t_ids(id TEXT PRIMARY KEY)", [])
        .map_err(|err| EzError::Io {
            message: format!("Failed to create temp table: {err}"),
        })?;
    {
        let mut ins = net_conn
            .prepare("INSERT OR IGNORE INTO t_ids(id) VALUES(?1)")
            .map_err(|err| EzError::Io {
                message: format!("Failed to prepare temp insert: {err}"),
            })?;
        for id in &link_ids {
            ins.execute(params![id]).map_err(|err| EzError::Io {
                message: format!("Failed to insert link id into temp table: {err}"),
            })?;
        }
    }

    let mut link_map: HashMap<String, (f64, f64, f64, f64)> = HashMap::new();
    {
        let mut stmt = net_conn
            .prepare(
                "SELECT l.id, fn.lng, fn.lat, tn.lng, tn.lat
                 FROM links l
                 JOIN nodes fn ON fn.id = l.from_node
                 JOIN nodes tn ON tn.id = l.to_node
                 JOIN t_ids ti ON ti.id = l.id",
            )
            .map_err(|err| EzError::Io {
                message: format!("Failed to prepare link-geom query: {err}"),
            })?;
        let rows = stmt
            .query_map([], |row| {
                Ok((
                    row.get::<_, String>(0)?,
                    row.get::<_, f64>(1)?,
                    row.get::<_, f64>(2)?,
                    row.get::<_, f64>(3)?,
                    row.get::<_, f64>(4)?,
                ))
            })
            .map_err(|err| EzError::Io {
                message: format!("Failed to run link-geom query: {err}"),
            })?;
        for row in rows {
            let (id, fx, fy, tx, ty) = row.map_err(|err| EzError::Io {
                message: format!("Failed to read link-geom row: {err}"),
            })?;
            link_map.insert(id, (fx, fy, tx, ty));
        }
    }

    for id in &link_ids {
        if !link_map.contains_key(id) {
            return Ok(None);
        }
    }

    let mut points: Vec<GeoPoint> = Vec::with_capacity(link_ids.len() * 2);
    for id in &link_ids {
        let (fx, fy, tx, ty) = link_map[id];
        if points.is_empty() {
            points.push(GeoPoint { lng: fx, lat: fy });
        } else {
            let last = points.last().unwrap();
            if (last.lng - fx).abs() > 1e-12 || (last.lat - fy).abs() > 1e-12 {
                points.push(GeoPoint { lng: fx, lat: fy });
            }
        }
        points.push(GeoPoint { lng: tx, lat: ty });
    }

    Ok(Some(LinkedRoutePath {
        line_id: line_id.to_string(),
        route_id: route_id.to_string(),
        network_source: network_source.to_string(),
        link_count: link_ids.len(),
        points,
    }))
}

pub(super) fn list_departures(
    work_dir: &Path,
    source_name: &str,
    line_id: &str,
    route_id: &str,
) -> Result<Vec<ListedDeparture>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let mut stmt = connection
        .prepare(
            "SELECT id, departure_time, vehicle_ref_id
             FROM departures
             WHERE line_id = ?1 AND route_id = ?2
             ORDER BY departure_time, id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare departures query: {err}"),
        })?;

    let rows = stmt
        .query_map([line_id, route_id], |row| {
            Ok(ListedDeparture {
                id: row.get(0)?,
                departure_time: row.get(1)?,
                vehicle_ref_id: row.get(2)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run departures query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read departure row: {err}"),
        })?);
    }
    Ok(out)
}
