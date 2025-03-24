use std::path::Path;
use std::fs;

use crate::pt::cli::Cli;
use crate::pt::steps::{validate, select, collect, build};
use super::models::PtError;

pub fn run(cli: Cli) -> Result<(), PtError> {
    // Validate day options
    cli.day_options.validate()
        .map_err(|_| PtError::InvalidDay)?;

    let input_path = Path::new(&cli.input);
    
    // Create output directory if it doesn't exist
    let output_dir = Path::new("output");
    if !output_dir.exists() {
        fs::create_dir_all(output_dir)?;
    }

    // Step 1: Validate GTFS files and setup directory structure
    validate::validate_gtfs_files(input_path)?;

    // Step 2: Service Selection
    select::handle_service_selection(input_path, cli.day_options.get_service_day())?;

    // Step 3: Trip Collection
    collect::handle_trip_collection(input_path)?;

    // Step 4: Schedule Building
    build::handle_schedule_building(output_dir)?;

    // Clean up if requested
    if cli.clean {
        fs::remove_dir_all(Path::new("temp\\pt"))?;
    }

    Ok(())
}
