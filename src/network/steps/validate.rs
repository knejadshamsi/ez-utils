use anyhow::Result;
use std::path::Path;
use crate::network::utils::display;

pub struct NetworkConfig {
    pub use_db: bool,
    pub delete_temp: bool,
}

pub struct ProcessingPaths {
    pub scaled_population_idx: String,
    pub population_chunks_dir: String,
}

pub fn validate_input_data(paths: &ProcessingPaths, input_path: &Path, total_steps: usize) -> Result<()> {
    verify_required_files(paths)?;
    display::print_step_complete(1, total_steps, "Population data and XML chunks");
    
    if !input_path.exists() {
        anyhow::bail!("Input network file not found: {}", input_path.display());
    }

    Ok(())
}

fn verify_required_files(paths: &ProcessingPaths) -> Result<()> {
    let index_path = Path::new(&paths.scaled_population_idx);
    if !index_path.exists() {
        anyhow::bail!("No population data found in temp/population. Please run the population module first.");
    }

    if !std::fs::metadata(index_path)?.is_file() {
        anyhow::bail!("No population data found in temp/population. Please run the population module first.");
    }

    let chunks_dir = Path::new(&paths.population_chunks_dir);
    if !chunks_dir.exists() {
        anyhow::bail!(
            "Required population chunks directory not found at: {}",
            chunks_dir.display()
        );
    }

    let has_xml = std::fs::read_dir(chunks_dir)?
        .filter_map(|e| e.ok())
        .any(|e| e.path().extension().map_or(false, |ext| ext == "xml"));

    if !has_xml {
        anyhow::bail!(
            "No XML files found in chunks directory: {}",
            chunks_dir.display()
        );
    }

    Ok(())
}
