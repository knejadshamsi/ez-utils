use serde::{Deserialize, Serialize};

#[derive(Debug)]
pub(crate) struct NodeRecord {
    pub id: String,
    pub lng: f64,
    pub lat: f64,
    pub raw_xml: String,
}

#[derive(Debug)]
pub(crate) struct LinkRecord {
    pub id: String,
    pub from_node: String,
    pub to_node: String,
    pub tag_blob: String,
    pub attributes_blob: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkNodePayload {
    pub id: String,
    pub lng: f64,
    pub lat: f64,
    /// true = node rendered only to keep a link from looking frayed at the
    /// bbox edge. Ghost nodes are non-interactive on the frontend.
    pub ghost: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkLinkPayload {
    pub id: String,
    pub from_node: String,
    pub to_node: String,
    pub from_lng: f64,
    pub from_lat: f64,
    pub to_lng: f64,
    pub to_lat: f64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkBboxResult {
    pub nodes: Vec<NetworkNodePayload>,
    pub links: Vec<NetworkLinkPayload>,
    pub node_total: i64,
    pub link_total: i64,
    pub node_cap_exceeded: bool,
    pub link_cap_exceeded: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkSearchLinkResult {
    pub links: Vec<NetworkLinkPayload>,
    pub connected_nodes: Vec<NetworkNodePayload>,
    pub total: i64,
    pub cap_exceeded: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkSearchNodeResult {
    pub nodes: Vec<NetworkNodePayload>,
    pub connected_links: Vec<NetworkLinkPayload>,
    pub total: i64,
    pub cap_exceeded: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkLinkDetailPayload {
    pub id: String,
    pub from_node: String,
    pub to_node: String,
    pub from_lng: f64,
    pub from_lat: f64,
    pub to_lng: f64,
    pub to_lat: f64,
    pub tag_blob: String,
    pub attributes_blob: Option<String>,
    pub from_node_attributes_blob: Option<String>,
    pub to_node_attributes_blob: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct LinkTagEditInput {
    pub length: Option<String>,
    pub freespeed: Option<String>,
    pub capacity: Option<String>,
    pub permlanes: Option<String>,
    pub oneway: Option<String>,
    pub modes: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CreateNodeInput {
    pub id: String,
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CreateLinkInput {
    pub id: String,
    pub from_node: String,
    pub to_node: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct DeleteNodeResult {
    pub deleted_link_ids: Vec<String>,
}
