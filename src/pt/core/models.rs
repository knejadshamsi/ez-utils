use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum PtError {
    #[error("IO error: {0}")]
    IoError(String),
    #[error("Missing GTFS files: {0}")]
    MissingFiles(String),
    #[error("Invalid service selection")]
    InvalidService,
    #[error("JSON error: {0}")]
    JsonError(String),
    #[error("XML error: {0}")]
    XmlError(String),
    #[error("Invalid day selection")]
    InvalidDay,
    #[error("Failed to parse XML: {0}")]
    XmlParseError(String),
    #[error("UTF-8 conversion error: {0}")]
    FromUtf8Error(String),
}

impl From<std::io::Error> for PtError {
    fn from(error: std::io::Error) -> Self {
        PtError::IoError(error.to_string())
    }
}

impl From<quick_xml::Error> for PtError {
    fn from(error: quick_xml::Error) -> Self {
        PtError::XmlParseError(error.to_string())
    }
}

impl From<std::string::FromUtf8Error> for PtError {
    fn from(error: std::string::FromUtf8Error) -> Self {
        PtError::FromUtf8Error(error.to_string())
    }
}

#[derive(Debug, Deserialize)]
pub struct GTFSData {
    pub calendar: String,
    pub trips: String,
    pub stop_times: String,
    pub stops: String,
    pub routes: String,
}

#[derive(Debug)]
pub struct TransitSchedule {
    pub stops: Vec<TransitStop>,
    pub routes: Vec<TransitRoute>,
}

#[derive(Debug)]
pub struct TransitStop {
    pub id: String,
    pub x: f64,
    pub y: f64,
    pub name: String,
}

#[derive(Debug)]
pub struct TransitRoute {
    pub id: String,
    pub vehicle: VehicleDefinition,
    pub stops: Vec<String>,
}

#[derive(Debug)]
pub struct VehicleDefinition {
    pub route_type: i32,
    pub capacity: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ServicePattern {
    pub service_id: String,
    pub monday: bool,
    pub tuesday: bool,
    pub wednesday: bool,
    pub thursday: bool,
    pub friday: bool,
    pub saturday: bool,
    pub sunday: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RouteInfo {
    pub route_id: String,
    pub route_type: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ServiceRouteMapping {
    pub service_id: String,
    pub routes: Vec<RouteInfo>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SelectedServices {
    pub metro_service: String,
    pub bus_service: String,
    pub date: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TripMapping {
    pub trip_id: String,
    pub route_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StopSequence {
    pub trip_id: String,
    pub stops: Vec<StopInfo>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct StopInfo {
    pub stop_id: String,
    pub arrival_time: String,
    pub departure_time: String,
    pub stop_sequence: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StopDetails {
    pub stop_id: String,
    pub stop_lat: f64,
    pub stop_lon: f64,
    pub stop_name: String,
}

#[derive(Debug, Clone)]
pub enum ServiceDay {
    Weekday,
    Weekend,
    Holiday,
    Monday,
    Tuesday,
    Wednesday,
    Thursday,
    Friday,
    Saturday,
    Sunday,
}
