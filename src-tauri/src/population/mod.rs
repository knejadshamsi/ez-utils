use std::{fs, path::Path};

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::EzError,
};

pub(crate) mod capture;
pub(crate) mod edit;
pub(crate) mod export;
pub(crate) mod parser;
pub(crate) mod patch;
pub(crate) mod queries;
pub(crate) mod schema;
pub(crate) mod types;

use crate::projection::CrsConfig;

pub(crate) fn import_population(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<(), EzError> {
    schema::create_population_db(work_dir, source_name)?;

    if let Err(err) = parser::run(work_dir, file_path, source_name, crs_config) {
        let _ = fs::remove_file(db_file_path(work_dir, source_name));
        let _ = fs::remove_file(preamble_file_path(work_dir, source_name));
        let _ = fs::remove_file(postamble_file_path(work_dir, source_name));
        return Err(err);
    }

    Ok(())
}
