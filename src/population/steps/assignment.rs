use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{self, BufRead, BufWriter, Write};
use std::path::Path;
use anyhow::{Result, Context};
use geo::{Point, Polygon, Contains};
use serde::Deserialize;
use crate::population::core::models::*;

pub fn calculate_location_density(persons: &[Person]) -> HashMap<LocationKey, f64> {
    let mut location_counts: HashMap<LocationKey, i32> = HashMap::new();
    let mut total_agents = 0;

    for person in persons {
        if let Some(first_activity) = person.activities.first() {
            if let Some(coords) = &first_activity.coordinates {
                let location = LocationKey::from_coords(coords.x(), coords.y());
                *location_counts.entry(location).or_insert(0) += 1;
                total_agents += 1;
            }
        }
    }

    let total_agents = total_agents as f64;
    location_counts.into_iter()
        .map(|(location, count)| {
            let density = if total_agents > 0.0 {
                count as f64 / total_agents
            } else {
                0.0
            };
            (location, density)
        })
        .collect()
}

pub fn select_persons_by_density(
    persons: &[Person],
    target_count: usize,
    location_density: &HashMap<LocationKey, f64>,
) -> Result<Vec<Person>> {
    if target_count == 0 {
        anyhow::bail!("Target count must be positive");
    }
    if target_count > persons.len() {
        anyhow::bail!("Target count cannot be greater than current population");
    }

    let scale_factor = target_count as f64 / persons.len() as f64;
    let mut persons_by_location: HashMap<LocationKey, Vec<Person>> = HashMap::new();

    // Group persons by their first activity location
    for person in persons {
        if let Some(first_activity) = person.activities.first() {
            if let Some(coords) = &first_activity.coordinates {
                let location = LocationKey::from_coords(coords.x(), coords.y());
                persons_by_location.entry(location)
                    .or_default()
                    .push(person.clone());
            }
        }
    }

    let mut selected_persons = Vec::new();
    for (location_key, density) in location_density {
        let target_location_count = (density * scale_factor * persons.len() as f64) as usize;
        if let Some(available_persons) = persons_by_location.get(location_key) {
            selected_persons.extend(
                available_persons.iter()
                    .take(target_location_count)
                    .cloned()
            );
        }
    }

    if selected_persons.is_empty() {
        anyhow::bail!("No persons could be selected based on density criteria");
    }

    selected_persons.truncate(target_count);
    Ok(selected_persons)
}

#[derive(Debug, Deserialize)]
struct BinConfig {
    bins: Vec<Bin>,
}

#[derive(Debug, Deserialize)]
struct Bin {
    id: u32,
    coordinates: Vec<[f64; 2]>,
}

impl Bin {
    fn contains_point(&self, point: &Point<f64>) -> bool {
        let ring: Vec<_> = self.coordinates.iter()
            .map(|[x, y]| (*x, *y))
            .collect();
        let polygon = Polygon::new(ring.into(), vec![]);
        polygon.contains(point)
    }
}

pub fn assign_locations_to_bins(
    index_dir: &Path,
    bins_path: &Path,
    output_dir: &Path,
) -> Result<()> {
    // Load bins configuration
    let bins_content = fs::read_to_string(bins_path)
        .context("Failed to read bins configuration")?;
    let bin_config: BinConfig = serde_json::from_str(&bins_content)
        .context("Failed to parse bins configuration")?;

    // Create map of bin files
    let mut bin_writers: HashMap<u32, BufWriter<File>> = HashMap::new();
    for bin in &bin_config.bins {
        let file = File::create(output_dir.join(format!("bin-{:02}.txt", bin.id)))
            .with_context(|| format!("Failed to create output file for bin {}", bin.id))?;
        bin_writers.insert(bin.id, BufWriter::new(file));
    }

    // Process each index file
    for entry in fs::read_dir(index_dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.extension().and_then(|s| s.to_str()) == Some("idx") {
            // Extract chunk number from filename
            let chunk_number = path.file_stem()
                .and_then(|s| s.to_str())
                .and_then(|s| s.strip_prefix("index_"))
                .and_then(|s| s.parse::<u32>().ok())
                .context("Failed to parse chunk number from filename")?;

            process_index_file(&path, &bin_config.bins, &mut bin_writers, chunk_number)?;
        }
    }

    // Flush all writers
    for (_, writer) in bin_writers {
        writer.into_inner()?.flush()?;
    }

    Ok(())
}

fn process_index_file(
    path: &Path,
    bins: &[Bin],
    bin_writers: &mut HashMap<u32, BufWriter<File>>,
    chunk_number: u32,
) -> Result<()> {
    let file = File::open(path)?;
    for line in io::BufReader::new(file).lines() {
        let line = line?;
        if let Some((agent_id, coords)) = line.split_once(':') {
            let coords: Vec<f64> = coords
                .split(',')
                .map(|s| s.trim().parse::<f64>())
                .collect::<Result<_, _>>()
                .context("Failed to parse coordinates")?;

            if coords.len() == 2 {
                let point = Point::new(coords[0], coords[1]);
                for bin in bins {
                    if bin.contains_point(&point) {
                        if let Some(writer) = bin_writers.get_mut(&bin.id) {
                            writeln!(writer, "{}: {}", agent_id, chunk_number)?;
                        }
                        break;  // Point can only be in one bin
                    }
                }
            }
        }
    }
    Ok(())
}
