use std::{fs, path::Path};

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, TransitMetadata},
};
use crate::projection::CrsConfig;

pub(crate) mod edit;
pub(crate) mod export;
pub(crate) mod parser;
pub(crate) mod queries;
pub(crate) mod schema;
pub(crate) mod types;
pub(crate) mod vehicles;
pub(crate) mod xml_attrs;

pub use edit::{
    apply_route_departures_edits_cmd, apply_route_path_links_cmd, apply_route_profile_edits_cmd,
    create_line_cmd, create_route_cmd, create_stop_facility_cmd, delete_line_cmd,
    delete_route_cmd, delete_stop_facility_cmd, delete_transit_transfer_cmd,
    list_transit_vehicles_cmd, preview_delete_line_cmd, preview_delete_route_cmd,
    preview_delete_stop_facility_cmd, update_line_cmd, update_route_cmd, update_stop_facility_cmd,
    upsert_transit_transfer_cmd,
};
pub use export::export_transit;
pub use queries::{
    find_nearby_network_links_cmd, list_line_profile_stops_cmd, list_stop_line_usage_cmd,
    list_stop_route_usage_cmd, list_stop_transfers_cmd, list_transit_departures_cmd,
    list_transit_lines_cmd, list_transit_path_links_cmd, list_transit_profile_stops_cmd,
    list_transit_routes_cmd, locate_transit_stop_cmd, query_transit_stops_bbox_cmd,
    resolve_route_path_geometry_cmd, search_transit_stops_cmd,
};
pub use vehicles::{detach_transit_vehicles_cmd, import_transit_vehicles_cmd};

pub(crate) fn import_transit(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<TransitMetadata, EzError> {
    schema::create_transit_db(work_dir, source_name)?;

    match parser::run(work_dir, file_path, source_name, crs_config) {
        Ok(metadata) => Ok(metadata),
        Err(err) => {
            let _ = fs::remove_file(db_file_path(work_dir, source_name));
            let _ = fs::remove_file(preamble_file_path(work_dir, source_name));
            let _ = fs::remove_file(postamble_file_path(work_dir, source_name));
            Err(err)
        }
    }
}
