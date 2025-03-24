use std::path::Path;
use std::fs;
use std::collections::HashMap;
use csv::Reader;
use serde_json;
use crate::pt::utils::display::ProcessProgress;
use crate::pt::core::models::{PtError, SelectedServices, TripMapping, StopSequence, StopInfo, StopDetails};

pub fn handle_trip_collection(input_path: &Path) -> Result<(), PtError> {
    let mut progress = ProcessProgress::new(3, "Collecting", "Collected", "trip ids");
    
    let trip_mappings = collect_trips(input_path, &mut progress)?;
    let path = Path::new("temp/pt/03-trips/trip_mapping.json");
    fs::create_dir_all(path.parent().ok_or_else(|| PtError::IoError("Missing parent directory".to_string()))?)?;
    save_json(path, &trip_mappings)?;

    progress.advance("Sequencing", "Sequenced", "stops");
    let stop_sequences = collect_stop_sequences(input_path, &trip_mappings, trip_mappings.len(), &mut progress)?;
    save_stop_sequences(&stop_sequences)?;

    let total_stops = stop_sequences.iter()
        .flat_map(|seq| &seq.stops)
        .map(|stop| &stop.stop_id)
        .collect::<std::collections::HashSet<_>>()
        .len();

    progress.advance("Gathering", "Gathered", "stop details");
    let stop_details = collect_stop_details(input_path, &stop_sequences, total_stops, &mut progress)?;
    let path = Path::new("temp/pt/03-trips/stop_details.json");
    fs::create_dir_all(path.parent().ok_or_else(|| PtError::IoError("Missing parent directory".to_string()))?)?;
    save_json(path, &stop_details)?;
    progress.complete();
    Ok(())
}

fn collect_trips(input_path: &Path, progress: &mut ProcessProgress) -> Result<Vec<TripMapping>, PtError> {
    // Get total number of trips for progress
    let total = {
        let mut temp_reader = Reader::from_path(input_path.join("trips.txt"))
            .map_err(|e| PtError::IoError(e.to_string()))?;
        temp_reader.deserialize::<HashMap<String, String>>().count()
    };

    let selected: SelectedServices = serde_json::from_str(
        &fs::read_to_string(Path::new("temp/pt/02-services/selected_services.json"))
            .map_err(|e| PtError::IoError(e.to_string()))?,
    )
    .map_err(|e| PtError::JsonError(e.to_string()))?;

    let mut reader = Reader::from_path(input_path.join("trips.txt"))
        .map_err(|e| PtError::IoError(e.to_string()))?;
    
    let mut processed = 0;
    let mut trips = Vec::new();
    for result in reader
        .deserialize::<HashMap<String, String>>() {
            processed += 1;
            progress.update_target(&format!("trip ids [{}/{}]", processed, total));
            
            if let Ok(record) = result {
                let record: HashMap<String, String> = record;
                if let Some(service_id) = record.get("service_id") {
                    if service_id == &selected.bus_service || service_id == &selected.metro_service {
                        if let (Some(trip_id), Some(route_id)) = (record.get("trip_id"), record.get("route_id")) {
                            trips.push(TripMapping {
                                trip_id: trip_id.clone(),
                                route_id: route_id.clone(),
                            });
                        }
                    }
                }
            }
        }

    Ok(trips)
}

fn collect_stop_sequences(
    input_path: &Path,
    trip_mappings: &[TripMapping],
    total: usize,
    progress: &mut ProcessProgress,
) -> Result<Vec<StopSequence>, PtError> {
    let mut reader = Reader::from_path(input_path.join("stop_times.txt"))
        .map_err(|e| PtError::IoError(e.to_string()))?;

    let mut sequences: Vec<StopSequence> = Vec::new();
    let mut current_stops: Vec<StopInfo> = Vec::new();
    let mut current_trip_id = String::new();
    let mut processed = 0;

    for result in reader.deserialize::<HashMap<String, String>>() {
        let record: HashMap<String, String> = result.map_err(|e| PtError::IoError(e.to_string()))?;
        let trip_id = record.get("trip_id").ok_or(PtError::IoError("Missing trip_id".to_string()))?.clone();

        if trip_mappings.iter().any(|tm| tm.trip_id == trip_id) {
            if trip_id != current_trip_id && !current_trip_id.is_empty() {
                sequences.push(StopSequence {
                    trip_id: current_trip_id.clone(),
                    stops: current_stops.clone(),
                });
                current_stops.clear();
                processed += 1;
                progress.update_target(&format!("stops [{}/{}]", processed, total));
            }

            current_trip_id = trip_id;
            current_stops.push(StopInfo {
                stop_id: record.get("stop_id").ok_or(PtError::IoError("Missing stop_id".to_string()))?.clone(),
                arrival_time: record.get("arrival_time").ok_or(PtError::IoError("Missing arrival_time".to_string()))?.clone(),
                departure_time: record.get("departure_time").ok_or(PtError::IoError("Missing departure_time".to_string()))?.clone(),
                stop_sequence: record.get("stop_sequence").ok_or(PtError::IoError("Missing stop_sequence".to_string()))?.parse().map_err(|_| PtError::IoError("Invalid stop_sequence".to_string()))?,
            });
        }
    }

    // Add the last sequence
    if !current_trip_id.is_empty() {
        sequences.push(StopSequence {
            trip_id: current_trip_id,
            stops: current_stops,
        });
        processed += 1;
        progress.update_target(&format!("stops [{}/{}]", processed, total));
    }

    Ok(sequences)
}

fn collect_stop_details(
    input_path: &Path,
    sequences: &[StopSequence],
    total_stops: usize,
    progress: &mut ProcessProgress,
) -> Result<HashMap<String, StopDetails>, PtError> {
    let mut reader = Reader::from_path(input_path.join("stops.txt"))
        .map_err(|e| PtError::IoError(e.to_string()))?;

    let unique_stops: std::collections::HashSet<_> = sequences
        .iter()
        .flat_map(|seq| seq.stops.iter())
        .map(|stop| &stop.stop_id)
        .collect();

    let mut stop_details = HashMap::new();
    let mut processed = 0;

    for result in reader.deserialize::<HashMap<String, String>>() {
        let record: HashMap<String, String> = result.map_err(|e| PtError::IoError(e.to_string()))?;
        if let Some(stop_id) = record.get("stop_id") {
            if unique_stops.contains(stop_id) {
                stop_details.insert(
                    stop_id.clone(),
                    StopDetails {
                        stop_id: stop_id.clone(),
                        stop_lat: record.get("stop_lat").ok_or(PtError::IoError("Missing stop_lat".to_string()))?.parse().map_err(|_| PtError::IoError("Invalid stop_lat".to_string()))?,
                        stop_lon: record.get("stop_lon").ok_or(PtError::IoError("Missing stop_lon".to_string()))?.parse().map_err(|_| PtError::IoError("Invalid stop_lon".to_string()))?,
                        stop_name: record.get("stop_name").ok_or(PtError::IoError("Missing stop_name".to_string()))?.clone(),
                    },
                );
                processed += 1;
                progress.update_target(&format!("stop details [{}/{}]", processed, total_stops));
            }
        }
    }

    Ok(stop_details)
}

fn save_stop_sequences(sequences: &[StopSequence]) -> Result<(), PtError> {
    let dir = Path::new("temp/pt/03-trips/stop_sequences");
    fs::create_dir_all(dir)?;
    for sequence in sequences {
        save_json(
            &dir.join(format!("{}.json", sequence.trip_id)),
            sequence,
        )?;
    }
    Ok(())
}

fn save_json<T: serde::Serialize>(path: &Path, data: &T) -> Result<(), PtError> {
    let json = serde_json::to_string_pretty(data)
        .map_err(|e| PtError::JsonError(e.to_string()))?;
    fs::write(path, json).map_err(|e| PtError::IoError(e.to_string()))?;
    Ok(())
}
