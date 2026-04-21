mod network;
mod population;
mod projection;
mod transit;
mod ez;
mod utils;

use network::{
    apply_link_tag_edits_cmd, create_link, create_node, delete_network_link, delete_network_node,
    export_network, get_network_link, query_network_bbox, search_network_links,
    search_network_nodes, update_link_attributes, update_node_attributes, update_node_position,
};
use transit::{
    apply_route_departures_edits_cmd, apply_route_path_links_cmd, apply_route_profile_edits_cmd,
    create_line_cmd, create_route_cmd, create_stop_facility_cmd, delete_line_cmd,
    delete_route_cmd, delete_stop_facility_cmd, delete_transit_transfer_cmd,
    detach_transit_vehicles_cmd, export_transit, find_nearby_network_links_cmd,
    import_transit_vehicles_cmd, list_line_profile_stops_cmd, list_stop_line_usage_cmd,
    list_stop_route_usage_cmd, list_stop_transfers_cmd, list_transit_departures_cmd,
    list_transit_lines_cmd, list_transit_path_links_cmd, list_transit_profile_stops_cmd,
    list_transit_routes_cmd, list_transit_vehicles_cmd, locate_transit_stop_cmd,
    preview_delete_line_cmd, preview_delete_route_cmd, preview_delete_stop_facility_cmd,
    query_transit_stops_bbox_cmd, resolve_route_path_geometry_cmd, search_transit_stops_cmd,
    update_line_cmd, update_route_cmd, update_stop_facility_cmd, upsert_transit_transfer_cmd,
};
use population::{
    edit::{
        apply_plan_edits, create_person, create_plan, delete_person, delete_plan,
        set_plan_selected, update_person_attributes, update_plan_blob,
    },
    export::export_population,
    queries::{get_person, query_population_bbox, search_population},
};
use ez::{
    close_ez, get_state, import_source, load_ez, new_from_xml, remove_source,
    rename_source, save_ez, save_ui_state, unpack_ez, validate_xml, SessionManager,
};
use projection::{list_crs_presets, set_crs};
use utils::check_paths_exist;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .manage(SessionManager::default())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            validate_xml,
            new_from_xml,
            unpack_ez,
            load_ez,
            save_ez,
            close_ez,
            get_state,
            import_source,
            rename_source,
            remove_source,
            save_ui_state,
            query_population_bbox,
            search_population,
            get_person,
            update_person_attributes,
            update_plan_blob,
            apply_plan_edits,
            create_person,
            create_plan,
            delete_plan,
            delete_person,
            set_plan_selected,
            export_population,
            query_network_bbox,
            search_network_links,
            search_network_nodes,
            get_network_link,
            update_node_attributes,
            update_node_position,
            update_link_attributes,
            apply_link_tag_edits_cmd,
            create_node,
            create_link,
            delete_network_node,
            delete_network_link,
            export_network,
            import_transit_vehicles_cmd,
            detach_transit_vehicles_cmd,
            list_transit_lines_cmd,
            list_transit_routes_cmd,
            list_transit_profile_stops_cmd,
            list_transit_path_links_cmd,
            list_transit_departures_cmd,
            query_transit_stops_bbox_cmd,
            list_stop_line_usage_cmd,
            list_stop_route_usage_cmd,
            list_stop_transfers_cmd,
            locate_transit_stop_cmd,
            list_line_profile_stops_cmd,
            resolve_route_path_geometry_cmd,
            create_line_cmd,
            update_line_cmd,
            preview_delete_line_cmd,
            delete_line_cmd,
            create_route_cmd,
            update_route_cmd,
            preview_delete_route_cmd,
            delete_route_cmd,
            apply_route_profile_edits_cmd,
            create_stop_facility_cmd,
            preview_delete_stop_facility_cmd,
            delete_stop_facility_cmd,
            update_stop_facility_cmd,
            search_transit_stops_cmd,
            apply_route_departures_edits_cmd,
            list_transit_vehicles_cmd,
            find_nearby_network_links_cmd,
            apply_route_path_links_cmd,
            upsert_transit_transfer_cmd,
            delete_transit_transfer_cmd,
            export_transit,
            list_crs_presets,
            set_crs,
            check_paths_exist
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
