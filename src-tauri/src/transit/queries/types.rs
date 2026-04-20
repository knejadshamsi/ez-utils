use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedLine {
    pub id: String,
    pub name: Option<String>,
    pub route_count: i64,
    pub modes: Vec<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedRoute {
    pub id: String,
    pub line_id: String,
    pub transport_mode: String,
    pub description: Option<String>,
    pub stop_count: i64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedProfileStop {
    pub sequence: i32,
    pub stop_ref_id: String,
    pub stop_name: Option<String>,
    pub stop_lng: Option<f64>,
    pub stop_lat: Option<f64>,
    pub stop_link_ref_id: Option<String>,
    pub stop_area_id: Option<String>,
    pub stop_is_blocking: Option<bool>,
    pub arrival_offset: Option<String>,
    pub departure_offset: Option<String>,
    pub allow_boarding: bool,
    pub allow_alighting: bool,
    pub await_departure: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedPathLink {
    pub sequence: i32,
    pub link_id: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedDeparture {
    pub id: String,
    pub departure_time: String,
    pub vehicle_ref_id: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedStop {
    pub id: String,
    pub name: Option<String>,
    pub lng: f64,
    pub lat: f64,
    pub link_ref_id: Option<String>,
    pub stop_area_id: Option<String>,
    pub is_blocking: bool,
    pub modes: Vec<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopLineUsage {
    pub line_id: String,
    pub line_name: Option<String>,
    pub route_count: i64,
    pub modes: Vec<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopRouteUsage {
    pub line_id: String,
    pub route_id: String,
    pub transport_mode: String,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopTransfer {
    pub other_stop_id: String,
    pub other_stop_name: Option<String>,
    pub transfer_time: f64,
    pub direction: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopsPage {
    pub items: Vec<ListedStop>,
    pub total: i64,
    pub page: i64,
    pub page_size: i64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopLocation {
    pub page: i64,
    pub position_in_page: i64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LinePathPoint {
    pub route_id: String,
    pub sequence: i32,
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct GeoPoint {
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LinkedRoutePath {
    pub line_id: String,
    pub route_id: String,
    pub network_source: String,
    pub link_count: usize,
    pub points: Vec<GeoPoint>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LineDeletePreview {
    pub routes: i64,
    pub profile_stops: i64,
    pub path_links: i64,
    pub departures: i64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct RouteDeletePreview {
    pub profile_stops: i64,
    pub path_links: i64,
    pub departures: i64,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StopFacilityDeletePreview {
    pub stop_id: String,
    pub profile_stops: i64,
    pub routes: i64,
    pub transfers: i64,
}

#[derive(Debug, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProfileStopEditInput {
    pub stop_ref_id: String,
    pub arrival_offset: Option<String>,
    pub departure_offset: Option<String>,
    pub allow_boarding: bool,
    pub allow_alighting: bool,
    pub await_departure: bool,
}

#[derive(Debug, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DepartureEditInput {
    pub id: String,
    pub departure_time: String,
    pub vehicle_ref_id: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ListedVehicle {
    pub id: String,
    pub vehicle_type: Option<String>,
    pub referenced: bool,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NearbyNetworkLink {
    pub id: String,
    pub from_lng: f64,
    pub from_lat: f64,
    pub to_lng: f64,
    pub to_lat: f64,
}

pub(super) fn parse_modes_csv(raw: Option<String>) -> Vec<String> {
    raw.map(|csv| {
        csv.split(',')
            .filter_map(|piece| {
                let t = piece.trim();
                if t.is_empty() {
                    None
                } else {
                    Some(t.to_string())
                }
            })
            .collect::<Vec<_>>()
    })
    .unwrap_or_default()
}
