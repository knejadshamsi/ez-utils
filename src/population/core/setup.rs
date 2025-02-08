use anyhow::{Result, Context};
use std::fs;
use std::path::Path;
use super::models::*;
use crate::population::steps::process::process_population_file;

pub async fn setup_population(input_path: &Path, scale_config: &ScaleConfig) -> Result<()> {
    let paths = setup_processing_directories(input_path.parent().context("Input file must have a parent directory")?)?;
    process_population_file(input_path, &paths, scale_config).await?;
    Ok(())
}

pub fn setup_processing_directories(base_path: &Path) -> Result<ProcessingPaths> {
    let temp_dir = base_path.join("temp");
    let population_dir = temp_dir.join("population");
    let chunks_dir = population_dir.join("01_split_population_chunks/chunks");
    let index_dir = population_dir.join("02_extract_agent_locations/indexes");
    let assigned_dir = population_dir.join("04_assign_location_bins/assigned");
    let dense_dir = population_dir.join("05_scale_population_density");
    let output_dir = base_path.join("output/population");

    crate::population::utils::display::print_step(1, 4, 0, 0);
    for dir in [&chunks_dir, &index_dir, &output_dir, &assigned_dir, &dense_dir] {
        fs::create_dir_all(dir)?;
    }
    crate::population::utils::display::step_success(1, 4, 0, 0);

    Ok(ProcessingPaths {
        base_dir: base_path.to_path_buf(),
        temp_dir,
        raw_dir: chunks_dir.clone(),
        chunks_dir: chunks_dir.clone(),
        index_dir,
        dense_dir: dense_dir.clone(),
        output_dir: output_dir.clone(),
        scaled_chunks_dir: dense_dir.join("scaled_chunks"),
        dense_new_dir: dense_dir,
        raw_chunks_dir: chunks_dir.clone(),
        output_population_dir: output_dir,
        assigned_dir,
    })
}

pub fn cleanup_temp_files(paths: &ProcessingPaths) -> Result<()> {
    if paths.chunks_dir.exists() {
        fs::remove_dir_all(&paths.chunks_dir)?;
    }

    if paths.index_dir.exists() {
        fs::remove_dir_all(&paths.index_dir)?;
    }

    Ok(())
}
