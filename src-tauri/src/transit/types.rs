#[derive(Debug)]
pub(crate) struct StopFacilityRecord {
    pub id: String,
    pub lng: f64,
    pub lat: f64,
    pub name: Option<String>,
    pub link_ref_id: Option<String>,
    pub stop_area_id: Option<String>,
    pub is_blocking: bool,
    pub attributes_blob: Option<String>,
}

#[derive(Debug)]
#[allow(dead_code)]
pub(crate) struct LineRecord {
    pub id: String,
    pub name: Option<String>,
    pub attributes_blob: Option<String>,
}

#[derive(Debug)]
#[allow(dead_code)]
pub(crate) struct RouteRecord {
    pub id: String,
    pub line_id: String,
    pub transport_mode: String,
    pub description: Option<String>,
    pub attributes_blob: Option<String>,
}

#[derive(Debug)]
pub(crate) struct RouteProfileStopRecord {
    pub line_id: String,
    pub route_id: String,
    pub sequence: i32,
    pub stop_ref_id: String,
    pub arrival_offset: Option<String>,
    pub departure_offset: Option<String>,
    pub allow_boarding: bool,
    pub allow_alighting: bool,
    pub await_departure: bool,
}

#[derive(Debug)]
pub(crate) struct RoutePathLinkRecord {
    pub line_id: String,
    pub route_id: String,
    pub sequence: i32,
    pub link_id: String,
}

#[derive(Debug)]
pub(crate) struct DepartureRecord {
    pub id: String,
    pub line_id: String,
    pub route_id: String,
    pub departure_time: String,
    pub vehicle_ref_id: Option<String>,
    pub attributes_blob: Option<String>,
}

#[derive(Debug)]
pub(crate) struct MinimalTransferTimeRecord {
    pub from_stop: String,
    pub to_stop: String,
    pub transfer_time: f64,
}
