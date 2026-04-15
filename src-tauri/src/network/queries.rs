//! Read-side commands for the network module: viewport bbox and id searches.
//! Single-entity detail fetches live in `detail.rs`.

use std::collections::HashSet;

use rusqlite::{params, Connection};
use tauri::State;

use crate::ez::{
    fs::db_file_path,
    types::{EzError, SessionManager},
};

use super::types::{
    NetworkBboxResult, NetworkLinkPayload, NetworkNodePayload, NetworkSearchLinkResult,
    NetworkSearchNodeResult,
};

const NETWORK_BBOX_NODE_CAP: i64 = 5000;
const NETWORK_BBOX_LINK_CAP: i64 = 10000;
const NETWORK_SEARCH_CAP: i64 = 500;

#[tauri::command]
pub fn query_network_bbox(
    source_name: String,
    min_lng: f64,
    min_lat: f64,
    max_lng: f64,
    max_lat: f64,
    manager: State<'_, SessionManager>,
) -> Result<NetworkBboxResult, EzError> {
    let connection = open_network_connection(&manager, &source_name)?;

    // Count-first: determine whether either cap is exceeded before fetching.
    let node_total: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM nodes
             WHERE lng BETWEEN ?1 AND ?2 AND lat BETWEEN ?3 AND ?4",
            params![min_lng, max_lng, min_lat, max_lat],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;

    let link_total: i64 = connection
        .query_row(
            "SELECT COUNT(*) FROM links
             WHERE from_node IN (SELECT id FROM nodes
                                 WHERE lng BETWEEN ?1 AND ?2 AND lat BETWEEN ?3 AND ?4)
                OR to_node   IN (SELECT id FROM nodes
                                 WHERE lng BETWEEN ?1 AND ?2 AND lat BETWEEN ?3 AND ?4)",
            params![min_lng, max_lng, min_lat, max_lat],
            |row| row.get(0),
        )
        .map_err(sqlite_err)?;

    let node_cap_exceeded = node_total > NETWORK_BBOX_NODE_CAP;
    let link_cap_exceeded = link_total > NETWORK_BBOX_LINK_CAP;

    // Fetch in-bbox nodes (capped to NETWORK_BBOX_NODE_CAP — flag still tells frontend to warn).
    let mut nodes_stmt = connection
        .prepare(
            "SELECT id, lng, lat FROM nodes
             WHERE lng BETWEEN ?1 AND ?2 AND lat BETWEEN ?3 AND ?4
             LIMIT ?5",
        )
        .map_err(sqlite_err)?;
    let mut nodes: Vec<NetworkNodePayload> = nodes_stmt
        .query_map(
            params![min_lng, max_lng, min_lat, max_lat, NETWORK_BBOX_NODE_CAP],
            |row| {
                Ok(NetworkNodePayload {
                    id: row.get(0)?,
                    lng: row.get(1)?,
                    lat: row.get(2)?,
                    ghost: false,
                })
            },
        )
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    // Fetch links that touch any in-bbox node (capped to NETWORK_BBOX_LINK_CAP).
    let mut links_stmt = connection
        .prepare(
            "SELECT l.id, l.from_node, l.to_node,
                    nf.lng, nf.lat, nt.lng, nt.lat
             FROM links l
             JOIN nodes nf ON nf.id = l.from_node
             JOIN nodes nt ON nt.id = l.to_node
             WHERE (nf.lng BETWEEN ?1 AND ?2 AND nf.lat BETWEEN ?3 AND ?4)
                OR (nt.lng BETWEEN ?1 AND ?2 AND nt.lat BETWEEN ?3 AND ?4)
             LIMIT ?5",
        )
        .map_err(sqlite_err)?;
    let links: Vec<NetworkLinkPayload> = links_stmt
        .query_map(
            params![min_lng, max_lng, min_lat, max_lat, NETWORK_BBOX_LINK_CAP],
            |row| {
                Ok(NetworkLinkPayload {
                    id: row.get(0)?,
                    from_node: row.get(1)?,
                    to_node: row.get(2)?,
                    from_lng: row.get(3)?,
                    from_lat: row.get(4)?,
                    to_lng: row.get(5)?,
                    to_lat: row.get(6)?,
                })
            },
        )
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    // Derive ghost nodes: endpoints referenced by fetched links but outside bbox.
    let in_bbox_ids: HashSet<&str> = nodes.iter().map(|n| n.id.as_str()).collect();
    let mut ghost_ids: HashSet<String> = HashSet::new();
    for link in &links {
        if !in_bbox_ids.contains(link.from_node.as_str()) {
            ghost_ids.insert(link.from_node.clone());
        }
        if !in_bbox_ids.contains(link.to_node.as_str()) {
            ghost_ids.insert(link.to_node.clone());
        }
    }

    if !ghost_ids.is_empty() {
        // Coords for ghosts are already carried on each link payload; derive
        // the unique node payloads from that to avoid an extra query.
        let mut emitted: HashSet<String> = HashSet::new();
        for link in &links {
            if ghost_ids.contains(&link.from_node) && emitted.insert(link.from_node.clone()) {
                nodes.push(NetworkNodePayload {
                    id: link.from_node.clone(),
                    lng: link.from_lng,
                    lat: link.from_lat,
                    ghost: true,
                });
            }
            if ghost_ids.contains(&link.to_node) && emitted.insert(link.to_node.clone()) {
                nodes.push(NetworkNodePayload {
                    id: link.to_node.clone(),
                    lng: link.to_lng,
                    lat: link.to_lat,
                    ghost: true,
                });
            }
        }
    }

    Ok(NetworkBboxResult {
        nodes,
        links,
        node_total,
        link_total,
        node_cap_exceeded,
        link_cap_exceeded,
    })
}

#[tauri::command]
pub fn search_network_links(
    source_name: String,
    query: String,
    exact: bool,
    manager: State<'_, SessionManager>,
) -> Result<NetworkSearchLinkResult, EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let param = if exact { query.clone() } else { format!("%{}%", query) };
    let op = if exact { "=" } else { "LIKE" };

    let total_sql = format!("SELECT COUNT(*) FROM links WHERE id {} ?1", op);
    let total: i64 = connection
        .query_row(&total_sql, params![param], |row| row.get(0))
        .map_err(sqlite_err)?;

    if total > NETWORK_SEARCH_CAP {
        return Ok(NetworkSearchLinkResult {
            links: Vec::new(),
            connected_nodes: Vec::new(),
            total,
            cap_exceeded: true,
        });
    }

    let links_sql = format!(
        "SELECT l.id, l.from_node, l.to_node,
                nf.lng, nf.lat, nt.lng, nt.lat
         FROM links l
         JOIN nodes nf ON nf.id = l.from_node
         JOIN nodes nt ON nt.id = l.to_node
         WHERE l.id {} ?1
         ORDER BY l.id ASC
         LIMIT ?2",
        op
    );
    let mut stmt = connection.prepare(&links_sql).map_err(sqlite_err)?;
    let links: Vec<NetworkLinkPayload> = stmt
        .query_map(params![param, NETWORK_SEARCH_CAP], |row| {
            Ok(NetworkLinkPayload {
                id: row.get(0)?,
                from_node: row.get(1)?,
                to_node: row.get(2)?,
                from_lng: row.get(3)?,
                from_lat: row.get(4)?,
                to_lng: row.get(5)?,
                to_lat: row.get(6)?,
            })
        })
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    // Derive connected nodes without an extra query.
    let mut emitted: HashSet<String> = HashSet::new();
    let mut connected_nodes: Vec<NetworkNodePayload> = Vec::new();
    for link in &links {
        if emitted.insert(link.from_node.clone()) {
            connected_nodes.push(NetworkNodePayload {
                id: link.from_node.clone(),
                lng: link.from_lng,
                lat: link.from_lat,
                ghost: false,
            });
        }
        if emitted.insert(link.to_node.clone()) {
            connected_nodes.push(NetworkNodePayload {
                id: link.to_node.clone(),
                lng: link.to_lng,
                lat: link.to_lat,
                ghost: false,
            });
        }
    }

    Ok(NetworkSearchLinkResult {
        links,
        connected_nodes,
        total,
        cap_exceeded: false,
    })
}

#[tauri::command]
pub fn search_network_nodes(
    source_name: String,
    query: String,
    exact: bool,
    manager: State<'_, SessionManager>,
) -> Result<NetworkSearchNodeResult, EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let param = if exact { query.clone() } else { format!("%{}%", query) };
    let op = if exact { "=" } else { "LIKE" };

    let total_sql = format!("SELECT COUNT(*) FROM nodes WHERE id {} ?1", op);
    let total: i64 = connection
        .query_row(&total_sql, params![param], |row| row.get(0))
        .map_err(sqlite_err)?;

    if total > NETWORK_SEARCH_CAP {
        return Ok(NetworkSearchNodeResult {
            nodes: Vec::new(),
            connected_links: Vec::new(),
            total,
            cap_exceeded: true,
        });
    }

    let nodes_sql = format!(
        "SELECT id, lng, lat FROM nodes
         WHERE id {} ?1
         ORDER BY id ASC
         LIMIT ?2",
        op
    );
    let mut stmt = connection.prepare(&nodes_sql).map_err(sqlite_err)?;
    let nodes: Vec<NetworkNodePayload> = stmt
        .query_map(params![param, NETWORK_SEARCH_CAP], |row| {
            Ok(NetworkNodePayload {
                id: row.get(0)?,
                lng: row.get(1)?,
                lat: row.get(2)?,
                ghost: false,
            })
        })
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    let mut connected_links: Vec<NetworkLinkPayload> = Vec::new();
    if !nodes.is_empty() {
        let links_sql = format!(
            "SELECT l.id, l.from_node, l.to_node,
                    nf.lng, nf.lat, nt.lng, nt.lat
             FROM links l
             JOIN nodes nf ON nf.id = l.from_node
             JOIN nodes nt ON nt.id = l.to_node
             WHERE l.from_node IN (SELECT id FROM nodes WHERE id {op} ?1)
                OR l.to_node   IN (SELECT id FROM nodes WHERE id {op} ?1)",
            op = op
        );
        let mut stmt = connection.prepare(&links_sql).map_err(sqlite_err)?;
        connected_links = stmt
            .query_map(params![param], |row| {
                Ok(NetworkLinkPayload {
                    id: row.get(0)?,
                    from_node: row.get(1)?,
                    to_node: row.get(2)?,
                    from_lng: row.get(3)?,
                    from_lat: row.get(4)?,
                    to_lng: row.get(5)?,
                    to_lat: row.get(6)?,
                })
            })
            .map_err(sqlite_err)?
            .collect::<Result<Vec<_>, _>>()
            .map_err(sqlite_err)?;
    }

    Ok(NetworkSearchNodeResult {
        nodes,
        connected_links,
        total,
        cap_exceeded: false,
    })
}


pub(crate) fn open_network_connection(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<Connection, EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
    let path = db_file_path(&session.work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }

    Connection::open(path).map_err(|err| EzError::Io {
        message: format!("Failed to open network SQLite database: {err}"),
    })
}

pub(crate) fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Network query failed: {err}"),
    }
}
