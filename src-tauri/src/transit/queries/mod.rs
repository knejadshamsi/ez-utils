use std::path::PathBuf;

use tauri::State;

use crate::ez::types::{EzError, SessionManager, SourceKind};

mod lines;
mod route_detail;
mod stops;
mod types;

pub use types::{
    DepartureEditInput, LineDeletePreview, LinePathPoint, LinkedRoutePath, ListedDeparture,
    ListedLine, ListedPathLink, ListedProfileStop, ListedRoute, ListedVehicle,
    ProfileStopEditInput, RouteDeletePreview, StopFacilityDeletePreview, StopLineUsage,
    StopLocation, StopRouteUsage, StopTransfer, StopsPage,
};

pub(super) fn resolve_transit_work_dir(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<PathBuf, EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;

    let entry = session
        .ui_state
        .sources
        .iter()
        .find(|entry| entry.name == source_name)
        .ok_or_else(|| EzError::SourceNotFound {
            name: source_name.to_string(),
        })?;

    if entry.kind != SourceKind::Transit {
        return Err(EzError::UnsupportedSource {
            message: format!("Source '{source_name}' is not a transit source."),
        });
    }

    Ok(session.work_dir.clone())
}

#[tauri::command]
pub fn list_transit_lines_cmd(
    source_name: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedLine>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    lines::list_lines(&work_dir, &source_name)
}

#[tauri::command]
pub fn list_transit_routes_cmd(
    source_name: String,
    line_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedRoute>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    lines::list_routes(&work_dir, &source_name, &line_id)
}

#[tauri::command]
pub fn list_transit_profile_stops_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedProfileStop>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    route_detail::list_profile_stops(&work_dir, &source_name, &line_id, &route_id)
}

#[tauri::command]
pub fn list_line_profile_stops_cmd(
    source_name: String,
    line_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<LinePathPoint>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    route_detail::list_line_profile_stops(&work_dir, &source_name, &line_id)
}

#[tauri::command]
pub fn list_transit_path_links_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedPathLink>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    route_detail::list_path_links(&work_dir, &source_name, &line_id, &route_id)
}

#[tauri::command]
pub fn list_transit_departures_cmd(
    source_name: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<ListedDeparture>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    route_detail::list_departures(&work_dir, &source_name, &line_id, &route_id)
}

#[tauri::command]
#[allow(clippy::too_many_arguments)]
pub fn query_transit_stops_bbox_cmd(
    source_name: String,
    min_lng: f64,
    max_lng: f64,
    min_lat: f64,
    max_lat: f64,
    search: Option<String>,
    modes: Option<Vec<String>>,
    page: i64,
    page_size: i64,
    manager: State<'_, SessionManager>,
) -> Result<StopsPage, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::query_stops_bbox(
        &work_dir,
        &source_name,
        min_lng,
        max_lng,
        min_lat,
        max_lat,
        search.as_deref(),
        modes.as_deref(),
        page,
        page_size,
    )
}

#[tauri::command]
#[allow(clippy::too_many_arguments)]
pub fn locate_transit_stop_cmd(
    source_name: String,
    min_lng: f64,
    max_lng: f64,
    min_lat: f64,
    max_lat: f64,
    search: Option<String>,
    modes: Option<Vec<String>>,
    stop_id: String,
    page_size: i64,
    manager: State<'_, SessionManager>,
) -> Result<StopLocation, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::locate_stop(
        &work_dir,
        &source_name,
        min_lng,
        max_lng,
        min_lat,
        max_lat,
        search.as_deref(),
        modes.as_deref(),
        &stop_id,
        page_size,
    )
}

#[tauri::command]
pub fn resolve_route_path_geometry_cmd(
    transit_source: String,
    network_source: String,
    line_id: String,
    route_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Option<LinkedRoutePath>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &transit_source)?;
    // Validate the named network source exists and is actually a network source.
    {
        let guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
        let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
        let entry = session
            .ui_state
            .sources
            .iter()
            .find(|entry| entry.name == network_source)
            .ok_or_else(|| EzError::SourceNotFound {
                name: network_source.clone(),
            })?;
        if entry.kind != SourceKind::Network {
            return Err(EzError::UnsupportedSource {
                message: format!("Source '{network_source}' is not a network source."),
            });
        }
    }
    route_detail::resolve_route_path_geometry(
        &work_dir,
        &transit_source,
        &network_source,
        &line_id,
        &route_id,
    )
}

/// Find network links near a point (for the "click a nearby link" picker in
/// stop-facility editing). Queries the network DB by joining links against
/// nodes and filtering where either endpoint falls inside a lat/lng box of
/// approx radius_m meters. Cheap + good enough for the sparse-UX case; no
/// spatial index needed.
#[tauri::command]
pub fn find_nearby_network_links_cmd(
    network_source: String,
    lng: f64,
    lat: f64,
    radius_m: f64,
    manager: State<'_, SessionManager>,
) -> Result<Vec<types::NearbyNetworkLink>, EzError> {
    let work_dir = {
        let guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
        let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
        let entry = session
            .ui_state
            .sources
            .iter()
            .find(|e| e.name == network_source)
            .ok_or_else(|| EzError::SourceNotFound {
                name: network_source.clone(),
            })?;
        if entry.kind != SourceKind::Network {
            return Err(EzError::UnsupportedSource {
                message: format!("Source '{network_source}' is not a network source."),
            });
        }
        session.work_dir.clone()
    };

    let path = crate::ez::fs::db_file_path(&work_dir, &network_source);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: network_source,
        });
    }
    let connection =
        rusqlite::Connection::open(&path).map_err(|err| EzError::Io {
            message: format!("Failed to open network SQLite database: {err}"),
        })?;

    // 1 deg latitude ~= 111_320 m. Longitude scales by cos(lat).
    let dlat = radius_m / 111_320.0;
    let dlng = radius_m / (111_320.0 * lat.to_radians().cos().abs().max(1e-6));
    let min_lat = lat - dlat;
    let max_lat = lat + dlat;
    let min_lng = lng - dlng;
    let max_lng = lng + dlng;

    let mut stmt = connection
        .prepare(
            "SELECT l.id, fn.lng, fn.lat, tn.lng, tn.lat
             FROM links AS l
             JOIN nodes AS fn ON fn.id = l.from_node
             JOIN nodes AS tn ON tn.id = l.to_node
             WHERE (fn.lng BETWEEN ?1 AND ?2 AND fn.lat BETWEEN ?3 AND ?4)
                OR (tn.lng BETWEEN ?1 AND ?2 AND tn.lat BETWEEN ?3 AND ?4)
             LIMIT 500",
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to prepare nearby-links query: {err}"),
        })?;
    let rows = stmt
        .query_map(
            rusqlite::params![min_lng, max_lng, min_lat, max_lat],
            |row| {
                Ok(types::NearbyNetworkLink {
                    id: row.get(0)?,
                    from_lng: row.get(1)?,
                    from_lat: row.get(2)?,
                    to_lng: row.get(3)?,
                    to_lat: row.get(4)?,
                })
            },
        )
        .map_err(|err| EzError::Io {
            message: format!("Failed to run nearby-links query: {err}"),
        })?;
    let mut out = Vec::new();
    for row in rows {
        out.push(row.map_err(|err| EzError::Io {
            message: format!("Failed to read nearby-link row: {err}"),
        })?);
    }
    Ok(out)
}

#[tauri::command]
pub fn search_transit_stops_cmd(
    source_name: String,
    query: Option<String>,
    limit: i64,
    manager: State<'_, SessionManager>,
) -> Result<Vec<types::ListedStop>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::search_stops(&work_dir, &source_name, query.as_deref(), limit.max(1).min(500))
}

#[tauri::command]
pub fn list_stop_line_usage_cmd(
    source_name: String,
    stop_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<StopLineUsage>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::list_stop_line_usage(&work_dir, &source_name, &stop_id)
}

#[tauri::command]
pub fn list_stop_route_usage_cmd(
    source_name: String,
    stop_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<StopRouteUsage>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::list_stop_route_usage(&work_dir, &source_name, &stop_id)
}

#[tauri::command]
pub fn list_stop_transfers_cmd(
    source_name: String,
    stop_id: String,
    manager: State<'_, SessionManager>,
) -> Result<Vec<StopTransfer>, EzError> {
    let work_dir = resolve_transit_work_dir(&manager, &source_name)?;
    stops::list_stop_transfers(&work_dir, &source_name, &stop_id)
}
