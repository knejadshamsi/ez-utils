//! Detail fetch for a single link, including bundled endpoint-node attributes.

use rusqlite::params;
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

use super::{
    capture::extract_attributes_blob,
    queries::{open_network_connection, sqlite_err},
    types::NetworkLinkDetailPayload,
};

#[tauri::command]
pub fn get_network_link(
    source_name: String,
    link_id: String,
    manager: State<'_, SessionManager>,
) -> Result<NetworkLinkDetailPayload, EzError> {
    let connection = open_network_connection(&manager, &source_name)?;

    let row = connection
        .query_row(
            "SELECT l.from_node, l.to_node, l.tag_blob, l.attributes_blob,
                    nf.lng, nf.lat, nf.raw_xml,
                    nt.lng, nt.lat, nt.raw_xml
             FROM links l
             JOIN nodes nf ON nf.id = l.from_node
             JOIN nodes nt ON nt.id = l.to_node
             WHERE l.id = ?1",
            params![link_id],
            |row| {
                Ok((
                    row.get::<_, String>(0)?,
                    row.get::<_, String>(1)?,
                    row.get::<_, String>(2)?,
                    row.get::<_, Option<String>>(3)?,
                    row.get::<_, f64>(4)?,
                    row.get::<_, f64>(5)?,
                    row.get::<_, String>(6)?,
                    row.get::<_, f64>(7)?,
                    row.get::<_, f64>(8)?,
                    row.get::<_, String>(9)?,
                ))
            },
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::NetworkLinkNotFound {
                link_id: link_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let (
        from_node,
        to_node,
        tag_blob,
        attributes_blob,
        from_lng,
        from_lat,
        from_raw,
        to_lng,
        to_lat,
        to_raw,
    ) = row;

    // Extract just the <attributes> child of each endpoint node's raw XML.
    // The raw XML itself is not sent to the frontend; it's only needed here
    // to mine out the attributes blob.
    let from_node_attributes_blob = extract_attributes_blob(&from_raw)?;
    let to_node_attributes_blob = extract_attributes_blob(&to_raw)?;

    Ok(NetworkLinkDetailPayload {
        id: link_id,
        from_node,
        to_node,
        from_lng,
        from_lat,
        to_lng,
        to_lat,
        tag_blob,
        attributes_blob,
        from_node_attributes_blob,
        to_node_attributes_blob,
    })
}
