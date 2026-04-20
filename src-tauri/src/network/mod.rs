use std::{fs, path::Path};

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::{EzError, NetworkMetadata},
};

pub(crate) mod capture;
pub(crate) mod detail;
pub(crate) mod edit;
pub(crate) mod export;
pub(crate) mod parser;
pub(crate) mod patch;
pub(crate) mod queries;
pub(crate) mod schema;
pub(crate) mod types;

pub use detail::get_network_link;
pub use edit::{
    apply_link_tag_edits_cmd, create_link, create_node, delete_network_link, delete_network_node,
    update_link_attributes, update_node_attributes, update_node_position,
};
pub use export::export_network;
pub use queries::{query_network_bbox, search_network_links, search_network_nodes};

use crate::projection::CrsConfig;

pub(crate) fn import_network(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<NetworkMetadata, EzError> {
    schema::create_network_db(work_dir, source_name)?;

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
