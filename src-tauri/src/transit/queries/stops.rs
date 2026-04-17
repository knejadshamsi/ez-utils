use std::path::{Path, PathBuf};

use rusqlite::Connection;

use crate::ez::{fs::db_file_path, types::EzError};

use super::types::{
    parse_modes_csv, ListedStop, StopLineUsage, StopLocation, StopRouteUsage, StopTransfer,
    StopsPage,
};

/// Subquery expression returning the CSV of modes for `stop_facilities.id` via live JOIN.
/// Used in SELECT projections to populate `ListedStop.modes` without a denormalized cache.
const MODES_SUBQUERY: &str = "(SELECT GROUP_CONCAT(DISTINCT LOWER(r.transport_mode))
     FROM route_profile_stops rps
     JOIN routes r ON r.line_id = rps.line_id AND r.id = rps.route_id
     WHERE rps.stop_ref_id = stop_facilities.id)";

pub(super) fn search_stops(
    work_dir: &Path,
    source_name: &str,
    query: Option<&str>,
    limit: i64,
) -> Result<Vec<ListedStop>, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    let pattern = query
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .map(|s| format!("%{}%", s.to_lowercase()));

    let (sql, bind_pattern) = if pattern.is_some() {
        (
            format!(
                "SELECT id, name, lng, lat, link_ref_id, stop_area_id, is_blocking,
                        COALESCE({MODES_SUBQUERY}, '') AS modes
                 FROM stop_facilities
                 WHERE LOWER(id) LIKE ?1 OR LOWER(COALESCE(name, '')) LIKE ?1
                 ORDER BY id
                 LIMIT ?2"
            ),
            true,
        )
    } else {
        (
            format!(
                "SELECT id, name, lng, lat, link_ref_id, stop_area_id, is_blocking,
                        COALESCE({MODES_SUBQUERY}, '') AS modes
                 FROM stop_facilities
                 ORDER BY id
                 LIMIT ?1"
            ),
            false,
        )
    };

    let mut stmt = connection.prepare(&sql).map_err(|err| EzError::Io {
        message: format!("Failed to prepare stop search query: {err}"),
    })?;

    let row_to_stop = |row: &rusqlite::Row<'_>| -> rusqlite::Result<ListedStop> {
        let modes_raw: String = row.get(7)?;
        let modes_opt = if modes_raw.is_empty() { None } else { Some(modes_raw) };
        Ok(ListedStop {
            id: row.get(0)?,
            name: row.get(1)?,
            lng: row.get(2)?,
            lat: row.get(3)?,
            link_ref_id: row.get(4)?,
            stop_area_id: row.get(5)?,
            is_blocking: row.get::<_, i32>(6)? != 0,
            modes: parse_modes_csv(modes_opt),
        })
    };

    let rows = if bind_pattern {
        stmt.query_map(
            rusqlite::params![pattern.as_deref().unwrap(), limit],
            row_to_stop,
        )
    } else {
        stmt.query_map(rusqlite::params![limit], row_to_stop)
    }
    .map_err(|err| EzError::Io {
        message: format!("Failed to run stop search: {err}"),
    })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read stop row: {err}"),
        })?);
    }
    Ok(out)
}

#[allow(clippy::too_many_arguments)]
pub(super) fn query_stops_bbox(
    work_dir: &Path,
    source_name: &str,
    min_lng: f64,
    max_lng: f64,
    min_lat: f64,
    max_lat: f64,
    search: Option<&str>,
    modes: Option<&[String]>,
    page: i64,
    page_size: i64,
) -> Result<StopsPage, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    // When modes is Some(empty), caller wants zero results (all toggles off).
    if let Some(list) = modes {
        if list.is_empty() {
            return Ok(StopsPage {
                items: Vec::new(),
                total: 0,
                page,
                page_size,
            });
        }
    }

    let search_pattern = search
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .map(|s| format!("%{}%", s.to_lowercase()));

    let mut where_clauses: Vec<String> =
        vec!["lng BETWEEN ? AND ?".into(), "lat BETWEEN ? AND ?".into()];
    if search_pattern.is_some() {
        where_clauses.push("(LOWER(id) LIKE ? OR LOWER(COALESCE(name, '')) LIKE ?)".into());
    }
    if let Some(list) = modes {
        let mode_clauses: Vec<String> = (0..list.len())
            .map(|_| {
                "EXISTS (
                    SELECT 1 FROM route_profile_stops rps
                    JOIN routes r ON r.line_id = rps.line_id AND r.id = rps.route_id
                    WHERE rps.stop_ref_id = stop_facilities.id
                      AND LOWER(r.transport_mode) = ?
                )"
                .to_string()
            })
            .collect();
        where_clauses.push(format!("({})", mode_clauses.join(" OR ")));
    }
    let where_clause = format!("WHERE {}", where_clauses.join(" AND "));

    // Common WHERE params shared by count + select.
    let push_common = |params: &mut Vec<Box<dyn rusqlite::ToSql>>| {
        params.push(Box::new(min_lng));
        params.push(Box::new(max_lng));
        params.push(Box::new(min_lat));
        params.push(Box::new(max_lat));
        if let Some(pattern) = search_pattern.as_deref() {
            params.push(Box::new(pattern.to_string()));
            params.push(Box::new(pattern.to_string()));
        }
        if let Some(list) = modes {
            for m in list {
                params.push(Box::new(m.to_lowercase()));
            }
        }
    };

    let total: i64 = {
        let count_sql = format!("SELECT COUNT(*) FROM stop_facilities {where_clause}");
        let mut stmt = connection.prepare(&count_sql).map_err(|err| EzError::Io {
            message: format!("Failed to prepare stops count query: {err}"),
        })?;
        let mut params: Vec<Box<dyn rusqlite::ToSql>> = Vec::new();
        push_common(&mut params);
        stmt.query_row(rusqlite::params_from_iter(params.iter()), |row| row.get(0))
            .map_err(|err| EzError::Io {
                message: format!("Failed to count stops: {err}"),
            })?
    };

    let offset = page.max(0) * page_size.max(1);
    let limit = page_size.max(1);

    let select_sql = format!(
        "SELECT id, name, lng, lat, link_ref_id, stop_area_id, is_blocking,
                COALESCE({MODES_SUBQUERY}, '') AS modes
         FROM stop_facilities
         {where_clause}
         ORDER BY id
         LIMIT ? OFFSET ?"
    );
    let mut stmt = connection.prepare(&select_sql).map_err(|err| EzError::Io {
        message: format!("Failed to prepare stops query: {err}"),
    })?;

    let mut params: Vec<Box<dyn rusqlite::ToSql>> = Vec::new();
    push_common(&mut params);
    params.push(Box::new(limit));
    params.push(Box::new(offset));

    fn row_to_stop(row: &rusqlite::Row<'_>) -> rusqlite::Result<ListedStop> {
        let modes_raw: String = row.get(7)?;
        let modes_opt = if modes_raw.is_empty() { None } else { Some(modes_raw) };
        Ok(ListedStop {
            id: row.get(0)?,
            name: row.get(1)?,
            lng: row.get(2)?,
            lat: row.get(3)?,
            link_ref_id: row.get(4)?,
            stop_area_id: row.get(5)?,
            is_blocking: row.get::<_, i32>(6)? != 0,
            modes: parse_modes_csv(modes_opt),
        })
    }

    let rows = stmt
        .query_map(rusqlite::params_from_iter(params.iter()), row_to_stop)
        .map_err(|err| EzError::Io {
            message: format!("Failed to run stops query: {err}"),
        })?;
    let mut items = Vec::new();
    for row in rows {
        items.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read stop row: {err}"),
        })?);
    }

    Ok(StopsPage {
        items,
        total,
        page,
        page_size,
    })
}

#[allow(clippy::too_many_arguments)]
pub(super) fn locate_stop(
    work_dir: &Path,
    source_name: &str,
    min_lng: f64,
    max_lng: f64,
    min_lat: f64,
    max_lat: f64,
    search: Option<&str>,
    modes: Option<&[String]>,
    stop_id: &str,
    page_size: i64,
) -> Result<StopLocation, EzError> {
    let path: PathBuf = db_file_path(work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }
    let connection = Connection::open(&path).map_err(|err| EzError::Io {
        message: format!("Failed to open transit SQLite database: {err}"),
    })?;

    if let Some(list) = modes {
        if list.is_empty() {
            return Ok(StopLocation {
                page: 0,
                position_in_page: 0,
            });
        }
    }

    let search_pattern = search
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .map(|s| format!("%{}%", s.to_lowercase()));

    let mut where_clauses: Vec<String> = vec![
        "lng BETWEEN ? AND ?".into(),
        "lat BETWEEN ? AND ?".into(),
        "id < ?".into(),
    ];
    if search_pattern.is_some() {
        where_clauses.push("(LOWER(id) LIKE ? OR LOWER(COALESCE(name, '')) LIKE ?)".into());
    }
    if let Some(list) = modes {
        let mode_clauses: Vec<String> = (0..list.len())
            .map(|_| {
                "EXISTS (
                    SELECT 1 FROM route_profile_stops rps
                    JOIN routes r ON r.line_id = rps.line_id AND r.id = rps.route_id
                    WHERE rps.stop_ref_id = stop_facilities.id
                      AND LOWER(r.transport_mode) = ?
                )"
                .to_string()
            })
            .collect();
        where_clauses.push(format!("({})", mode_clauses.join(" OR ")));
    }
    let where_clause = format!("WHERE {}", where_clauses.join(" AND "));

    let mut params: Vec<Box<dyn rusqlite::ToSql>> = Vec::new();
    params.push(Box::new(min_lng));
    params.push(Box::new(max_lng));
    params.push(Box::new(min_lat));
    params.push(Box::new(max_lat));
    params.push(Box::new(stop_id.to_string()));
    if let Some(pattern) = search_pattern.as_deref() {
        params.push(Box::new(pattern.to_string()));
        params.push(Box::new(pattern.to_string()));
    }
    if let Some(list) = modes {
        for m in list {
            params.push(Box::new(m.to_lowercase()));
        }
    }

    let count_sql = format!("SELECT COUNT(*) FROM stop_facilities {where_clause}");
    let rank: i64 = connection
        .prepare(&count_sql)
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare stop location query: {err}"),
        })?
        .query_row(rusqlite::params_from_iter(params.iter()), |row| row.get(0))
        .map_err(|err| EzError::Io {
            message: format!("Failed to compute stop rank: {err}"),
        })?;

    let effective_page_size = page_size.max(1);
    let page = rank / effective_page_size;
    let position_in_page = rank % effective_page_size;

    Ok(StopLocation {
        page,
        position_in_page,
    })
}

pub(super) fn list_stop_line_usage(
    work_dir: &Path,
    source_name: &str,
    stop_id: &str,
) -> Result<Vec<StopLineUsage>, EzError> {
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
            "SELECT l.id,
                    l.name,
                    COUNT(DISTINCT r.id) AS route_count,
                    GROUP_CONCAT(DISTINCT r.transport_mode) AS modes
             FROM lines AS l
             JOIN routes AS r ON r.line_id = l.id
             JOIN route_profile_stops AS rps
                    ON rps.line_id = r.line_id AND rps.route_id = r.id
             WHERE rps.stop_ref_id = ?1
             GROUP BY l.id, l.name
             ORDER BY l.id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare stop-line query: {err}"),
        })?;

    let rows = stmt
        .query_map([stop_id], |row| {
            Ok(StopLineUsage {
                line_id: row.get(0)?,
                line_name: row.get(1)?,
                route_count: row.get(2)?,
                modes: parse_modes_csv(row.get(3)?),
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run stop-line query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read stop-line row: {err}"),
        })?);
    }
    Ok(out)
}

pub(super) fn list_stop_route_usage(
    work_dir: &Path,
    source_name: &str,
    stop_id: &str,
) -> Result<Vec<StopRouteUsage>, EzError> {
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
            "SELECT DISTINCT r.line_id, r.id, r.transport_mode, r.description
             FROM routes AS r
             JOIN route_profile_stops AS rps
                    ON rps.line_id = r.line_id AND rps.route_id = r.id
             WHERE rps.stop_ref_id = ?1
             ORDER BY r.line_id, r.id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare stop-route query: {err}"),
        })?;

    let rows = stmt
        .query_map([stop_id], |row| {
            Ok(StopRouteUsage {
                line_id: row.get(0)?,
                route_id: row.get(1)?,
                transport_mode: row.get(2)?,
                description: row.get(3)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run stop-route query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read stop-route row: {err}"),
        })?);
    }
    Ok(out)
}

pub(super) fn list_stop_transfers(
    work_dir: &Path,
    source_name: &str,
    stop_id: &str,
) -> Result<Vec<StopTransfer>, EzError> {
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
            "SELECT to_stop AS other_id,
                    (SELECT name FROM stop_facilities WHERE id = to_stop) AS other_name,
                    transfer_time,
                    'from' AS direction
             FROM minimal_transfer_times
             WHERE from_stop = ?1
             UNION ALL
             SELECT from_stop AS other_id,
                    (SELECT name FROM stop_facilities WHERE id = from_stop) AS other_name,
                    transfer_time,
                    'to' AS direction
             FROM minimal_transfer_times
             WHERE to_stop = ?1
             ORDER BY other_id",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare stop-transfers query: {err}"),
        })?;

    let rows = stmt
        .query_map([stop_id], |row| {
            Ok(StopTransfer {
                other_stop_id: row.get(0)?,
                other_stop_name: row.get(1)?,
                transfer_time: row.get(2)?,
                direction: row.get(3)?,
            })
        })
        .map_err(|err| EzError::Io {
            message: format!("Failed to run stop-transfers query: {err}"),
        })?;

    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read stop-transfer row: {err}"),
        })?);
    }
    Ok(out)
}
