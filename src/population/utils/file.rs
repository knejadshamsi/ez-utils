use anyhow::{Context, Result, anyhow};
use std::fs;
use std::path::Path;
use geo::Point;
use crate::population::core::models::*;
use crate::util::get_temp_dir;

pub fn create_point_from_coordinates(lon: f64, lat: f64) -> Result<Point<f64>> {
    Ok(Point::new(lon, lat))
}

pub fn parse_time(time_str: &str) -> Option<String> {
    let parts: Vec<&str> = time_str.split(':').collect();
    if parts.len() >= 2 {
        Some(time_str.to_string())
    } else {
        None
    }
}

pub fn mtm8_to_latlon(x: f64, y: f64) -> Result<(f64, f64)> {
    use proj4rs::Proj;
    
    let from_proj = Proj::from_proj_string("+proj=tmerc +lat_0=0 +lon_0=-73.5 +k=0.9999 +x_0=304800 +y_0=0 +ellps=GRS80 +units=m +no_defs")
        .map_err(|e| anyhow::anyhow!("Failed to create source projection: {}", e))?;
    let to_proj = Proj::from_proj_string("+proj=longlat +datum=WGS84 +no_defs")
        .map_err(|e| anyhow::anyhow!("Failed to create destination projection: {}", e))?;

    use proj4rs::transform::transform;
    let mut point = (x, y, 0.0);

    transform(&from_proj, &to_proj, &mut point)
        .map_err(|e| anyhow!("Failed to transform coordinates: {}", e))?;

    let (lon_rad, lat_rad, _) = point;
    let lon = lon_rad.to_degrees();
    let lat = lat_rad.to_degrees();

    Ok((lon, lat))
}

pub fn setup_processing_directories(input_path: &Path) -> Result<ProcessingPaths> {
    let base_path = input_path.parent().ok_or_else(|| anyhow::anyhow!("Input file must have a parent directory"))?;
    let output_dir = base_path.join("output/population");
    
    let temp_dir = get_temp_dir();
    let population_dir = temp_dir.join("population");
    let raw_dir = population_dir.join("01_split_population_chunks/chunks");
    let chunks_dir = raw_dir.clone();
    let index_dir = population_dir.join("02_extract_agent_locations/indexes");
    let dense_dir = population_dir.join("05_scale_population_density/dense");
    let raw_chunks_dir = raw_dir.clone();
    let output_population_dir = output_dir.clone();

    let paths_to_clean = [
        base_path.join("temp"),
        base_path.join("output"),
        temp_dir.clone(),
        population_dir.clone(),
        raw_dir.clone(),
        chunks_dir.clone(),
        index_dir.parent().unwrap().to_path_buf(),
        index_dir.clone(),
        dense_dir.clone(),
        output_dir.parent().unwrap().to_path_buf(),
        output_dir.clone(),
    ];

    for path in &paths_to_clean {
        if path.exists() {
            if path.is_file() {
                fs::remove_file(path)?;
            } else {
                fs::remove_dir_all(path)?;
            }
        }
    }

    let assigned_dir = population_dir.join("04_assign_location_bins/assigned");
    let scaled_chunks_dir = population_dir.join("06_generate_scaled_population/chunks");
    let dense_new_dir = population_dir.join("05_scale_population_density");

    for dir in &[&chunks_dir, &index_dir, &dense_dir, &output_dir, 
                 &scaled_chunks_dir, &dense_new_dir, &assigned_dir] {
        fs::create_dir_all(dir)?;
    }

    Ok(ProcessingPaths {
        base_dir: base_path.to_path_buf(),
        temp_dir,
        raw_dir,
        chunks_dir,
        index_dir,
        dense_dir,
        output_dir,
        scaled_chunks_dir,
        dense_new_dir,
        raw_chunks_dir,
        output_population_dir,
        assigned_dir,
    })
}

pub fn cleanup_temp_files(paths: &ProcessingPaths) -> Result<()> {
    if paths.chunks_dir.exists() {
        fs::remove_dir_all(&paths.chunks_dir)
            .context("Failed to remove chunks directory")?;
    }

    if paths.index_dir.exists() {
        fs::remove_dir_all(&paths.index_dir)
            .context("Failed to remove index directory")?;
    }

    if paths.raw_dir.exists() {
        fs::remove_dir_all(&paths.raw_dir)
            .context("Failed to remove raw directory")?;
    }

    Ok(())
}

pub fn load_persons_from_indexes(index_dir: &Path) -> Result<Vec<Person>> {
    let mut persons = Vec::new();
    let index_files = fs::read_dir(index_dir)?;
    
    for entry in index_files {
        let entry = entry?;
        let content = fs::read_to_string(entry.path())?;
        
        for line in content.lines() {
            if let Some((id, coords)) = line.split_once(':') {
                let coords = coords.trim();
                let mut parts = coords.split(',').map(|s| s.trim().parse::<f64>());
                if let (Some(Ok(x)), Some(Ok(y))) = (parts.next(), parts.next()) {
                    let coordinates = Point::new(x, y);
                    let activities = vec![Activity {
                        id: id.to_string(),
                        activity_type: "home".to_string(),
                        activity_order: 1,
                        facility: None,
                        start_time: None,
                        end_time: None,
                        coordinates: Some(coordinates),
                    }];
                    persons.push(Person {
                        id: id.to_string(),
                        activities,
                    });
                }
            }
        }
    }

    Ok(persons)
}
