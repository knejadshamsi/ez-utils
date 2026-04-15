//! Edit-side commands for the network module: positions, attributes, tag
//! patches, create/delete of nodes and links.

use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};
use crate::utils::xml_escape;

use crate::ez::current_crs;
use crate::projection::Projector;

use super::{
    patch::{apply_link_tag_edits, replace_attributes_child, rewrite_node_xy},
    queries::{open_network_connection, sqlite_err},
    types::{CreateLinkInput, CreateNodeInput, DeleteNodeResult, LinkTagEditInput},
};

#[tauri::command]
pub fn update_node_attributes(
    source_name: String,
    node_id: String,
    attributes_blob: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let raw_xml: String = connection
        .query_row(
            "SELECT raw_xml FROM nodes WHERE id = ?1",
            params![node_id],
            |row| row.get(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::NetworkNodeNotFound {
                node_id: node_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let new_xml = replace_attributes_child(&raw_xml, attributes_blob.as_deref())?;

    connection
        .execute(
            "UPDATE nodes SET raw_xml = ?1 WHERE id = ?2",
            params![new_xml, node_id],
        )
        .map_err(sqlite_err)?;
    Ok(())
}

#[tauri::command]
pub fn update_node_position(
    source_name: String,
    node_id: String,
    lng: f64,
    lat: f64,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let raw_xml: String = connection
        .query_row(
            "SELECT raw_xml FROM nodes WHERE id = ?1",
            params![node_id],
            |row| row.get(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::NetworkNodeNotFound {
                node_id: node_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let native = projector.unproject_lng_lat(lng, lat)?;
    let new_xml = rewrite_node_xy(&raw_xml, native.x, native.y)?;

    connection
        .execute(
            "UPDATE nodes SET lng = ?1, lat = ?2, raw_xml = ?3 WHERE id = ?4",
            params![lng, lat, new_xml, node_id],
        )
        .map_err(sqlite_err)?;
    Ok(())
}

#[tauri::command]
pub fn update_link_attributes(
    source_name: String,
    link_id: String,
    attributes_blob: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let affected = connection
        .execute(
            "UPDATE links SET attributes_blob = ?1 WHERE id = ?2",
            params![attributes_blob, link_id],
        )
        .map_err(sqlite_err)?;
    if affected == 0 {
        return Err(EzError::NetworkLinkNotFound { link_id });
    }
    Ok(())
}

#[tauri::command]
pub fn apply_link_tag_edits_cmd(
    source_name: String,
    link_id: String,
    edits: LinkTagEditInput,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let tag_blob: String = connection
        .query_row(
            "SELECT tag_blob FROM links WHERE id = ?1",
            params![link_id],
            |row| row.get(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::NetworkLinkNotFound {
                link_id: link_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let new_blob = apply_link_tag_edits(&tag_blob, &edits)?;
    connection
        .execute(
            "UPDATE links SET tag_blob = ?1 WHERE id = ?2",
            params![new_blob, link_id],
        )
        .map_err(sqlite_err)?;
    Ok(())
}

#[tauri::command]
pub fn create_node(
    source_name: String,
    input: CreateNodeInput,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let exists: i64 = connection
        .query_row("SELECT COUNT(*) FROM nodes WHERE id = ?1", params![input.id], |r| r.get(0))
        .map_err(sqlite_err)?;
    if exists > 0 {
        return Err(EzError::NetworkDuplicateNodeId { node_id: input.id });
    }
    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let native = projector.unproject_lng_lat(input.lng, input.lat)?;
    let raw_xml = format!(
        "<node id=\"{}\" x=\"{:.6}\" y=\"{:.6}\"/>",
        xml_escape(&input.id),
        native.x,
        native.y
    );
    connection
        .execute(
            "INSERT INTO nodes (id, lng, lat, raw_xml) VALUES (?1, ?2, ?3, ?4)",
            params![input.id, input.lng, input.lat, raw_xml],
        )
        .map_err(sqlite_err)?;
    Ok(())
}

#[tauri::command]
pub fn create_link(
    source_name: String,
    input: CreateLinkInput,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let exists: i64 = connection
        .query_row("SELECT COUNT(*) FROM links WHERE id = ?1", params![input.id], |r| r.get(0))
        .map_err(sqlite_err)?;
    if exists > 0 {
        return Err(EzError::NetworkDuplicateLinkId { link_id: input.id });
    }
    for n in [&input.from_node, &input.to_node] {
        let has: i64 = connection
            .query_row("SELECT COUNT(*) FROM nodes WHERE id = ?1", params![n], |r| r.get(0))
            .map_err(sqlite_err)?;
        if has == 0 {
            return Err(EzError::NetworkNodeNotFound { node_id: n.to_string() });
        }
    }
    let tag_blob = format!(
        "id=\"{}\" from=\"{}\" to=\"{}\"",
        xml_escape(&input.id),
        xml_escape(&input.from_node),
        xml_escape(&input.to_node)
    );
    connection
        .execute(
            "INSERT INTO links (id, from_node, to_node, tag_blob, attributes_blob) VALUES (?1, ?2, ?3, ?4, NULL)",
            params![input.id, input.from_node, input.to_node, tag_blob],
        )
        .map_err(sqlite_err)?;
    Ok(())
}

#[tauri::command]
pub fn delete_network_node(
    source_name: String,
    node_id: String,
    manager: State<'_, SessionManager>,
) -> Result<DeleteNodeResult, EzError> {
    let mut connection = open_network_connection(&manager, &source_name)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    let deleted_link_ids: Vec<String> = {
        let mut stmt = tx
            .prepare("SELECT id FROM links WHERE from_node = ?1 OR to_node = ?1")
            .map_err(sqlite_err)?;
        let ids = stmt
            .query_map(params![node_id], |row| row.get::<_, String>(0))
            .map_err(sqlite_err)?
            .collect::<Result<Vec<_>, _>>()
            .map_err(sqlite_err)?;
        ids
    };

    tx.execute(
        "DELETE FROM links WHERE from_node = ?1 OR to_node = ?1",
        params![node_id],
    )
    .map_err(sqlite_err)?;

    let affected = tx
        .execute("DELETE FROM nodes WHERE id = ?1", params![node_id])
        .map_err(sqlite_err)?;
    if affected == 0 {
        return Err(EzError::NetworkNodeNotFound { node_id });
    }
    tx.commit().map_err(sqlite_err)?;

    Ok(DeleteNodeResult { deleted_link_ids })
}

#[tauri::command]
pub fn delete_network_link(
    source_name: String,
    link_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let connection = open_network_connection(&manager, &source_name)?;
    let affected = connection
        .execute("DELETE FROM links WHERE id = ?1", params![link_id])
        .map_err(sqlite_err)?;
    if affected == 0 {
        return Err(EzError::NetworkLinkNotFound { link_id });
    }
    Ok(())
}

