use std::path::PathBuf;
use anyhow::Result;
use std::fs;
use std::path::Path;

use super::core::setup::setup_processing_directories;
use super::steps;
use super::utils::display::{print_step_start, print_step_complete};

pub async fn run(input: PathBuf, use_db: bool, delete_temp: bool) -> Result<()> {
    let path_manager = crate::util::PathManager::new(&input);
    let current_dir = std::env::current_dir()?;
    let paths = setup_processing_directories(&current_dir, &path_manager)?;
    let config = steps::validate::NetworkConfig { 
        use_db, 
        delete_temp 
    };

    let processing_paths = steps::validate::ProcessingPaths {
        scaled_population_idx: path_manager
            .get_population_path("06_generate_scaled_population/new_population.idx")
            .to_string_lossy()
            .into_owned(),
        population_chunks_dir: path_manager
            .get_population_path("01_split_population_chunks/chunks")
            .to_string_lossy()
            .into_owned(),
    };

    steps::validate_input_data(&processing_paths, &input, 8)?;

    let chunks_dir = Path::new(&processing_paths.population_chunks_dir);
    steps::process_chunks(chunks_dir, &path_manager)?;

    steps::process_scaled_indexes(chunks_dir, Path::new(&processing_paths.scaled_population_idx), &config, &path_manager)?;

    steps::flow::create_flow_indexes(&paths, &path_manager)?;

    steps::adjustment::calculate_network_adjustments(&paths)?;

    steps::split::split_network_file(&input, &paths.network_split_dir, 1000)?;

    print_step_start(7, 8, "Applying network ratios", None);
    let total_scales = 10;
    for scale in 1..=total_scales {
        steps::apply::apply_network_ratios(&paths, scale, total_scales)?;
    }
    print_step_complete(7, 8, "Applied network ratios");

    steps::output::compose_scaled_networks(&paths)?;

    if config.delete_temp {
        fs::remove_dir_all(path_manager.get_temp_dir())?;
    }

    Ok(())
}
