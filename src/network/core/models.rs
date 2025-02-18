use std::path::PathBuf;
use std::collections::HashMap;
use geo_types::Point;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: String,
    pub x: f64,
    pub y: f64,
    pub coordinates: Point<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Link {
    pub id: String,
    pub from_node: String,
    pub to_node: String,
    pub length: f64,
    pub freespeed: f64,
    pub capacity: f64,
    pub permlanes: f64,
    pub oneway: Option<bool>,
    pub modes: Option<String>,
}

#[derive(Debug)]
pub struct NetworkChunk {
    pub content: String,
    pub chunk_number: usize,
}

#[derive(Debug)]
pub struct NetworkPaths {
    pub raw_chunks_dir: PathBuf,
    pub population_chunks_dir: PathBuf,
    pub temp_dir: PathBuf,
    pub output_dir: PathBuf,
    pub network_index_dir: PathBuf,
    pub split_dir: PathBuf,
    pub network_split_dir: PathBuf,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LinkAdjustment {
    pub speed_factor: f64,
    pub capacity_factor: f64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LinkRatio {
    pub ratio: f64,
}

#[derive(Debug)]
pub struct NetworkData {
    pub nodes: Vec<Node>,
    pub links: Vec<Link>,
    pub link_coordinates: Vec<LinkCoordinates>,
}

#[derive(Debug)]
pub struct ScaledNetwork {
    pub nodes: Vec<Node>,
    pub links: Vec<Link>,
    pub scale: f64,
}

#[derive(Debug)]
pub struct LinkFlow {
    pub link_id: String,
    pub flow_ratio: f64,
    pub base_flow: f64,
    pub scaled_flow: f64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LinkCoordinates {
    pub link_id: String,
    pub from_node: Point<f64>,
    pub to_node: Point<f64>,
    pub length: f64,
    pub freespeed: f64,
    pub agent_usage: HashMap<String, Vec<String>>,
}
