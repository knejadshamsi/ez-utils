use std::path::Path;
use std::fs;
use crate::pt::utils::display::ProcessProgress;
use crate::pt::core::models::PtError;

fn create_temp_directories() -> Result<(), PtError> {
    let temp_dir = Path::new("temp/pt");
    if temp_dir.exists() {
        fs::remove_dir_all(temp_dir).map_err(|e| PtError::IoError(e.to_string()))?;
    }

    let dirs = [
        "temp/pt/02-services",
        "temp/pt/03-trips/stop_sequences",
        "temp/pt/04_schedule_building",
    ];

    for dir in dirs.iter() {
        fs::create_dir_all(dir).map_err(|e| PtError::IoError(e.to_string()))?;
    }

    Ok(())
}

const REQUIRED_FILES: [&str; 5] = [
    "calendar.txt",
    "trips.txt",
    "stop_times.txt",
    "stops.txt",
    "routes.txt",
];

pub fn validate_gtfs_files(input_path: &Path) -> Result<(), PtError> {
    let mut progress = ProcessProgress::new(1, "Validating", "Validated", "GTFS .txt files:");

    for file in REQUIRED_FILES.iter() {
        let file_path = input_path.join(file);
        if !file_path.exists() {
            return Err(PtError::MissingFiles(file.to_string()));
        }
        let file_name = file.trim_end_matches(".txt");
        let target = format!("{} {}", progress.current_part.1, crate::pt::utils::display::format_file_name(file_name));
        progress.update_target(&target);
        std::thread::sleep(std::time::Duration::from_millis(300));
    }
    
    create_temp_directories()?;
    progress.complete();

    Ok(())
}
