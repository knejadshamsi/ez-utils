use std::path::Path;
use std::fs;
use std::collections::HashMap;
use csv::Reader;
use serde_json;
use chrono::Local;
use crate::pt::utils::display::ProcessProgress;
use crate::pt::core::models::{ServiceDay, ServicePattern, ServiceRouteMapping, SelectedServices, RouteInfo, PtError};

pub fn handle_service_selection(
    input_path: &Path,
    service_day: ServiceDay,
) -> Result<(), PtError> {
    let temp_dir = Path::new("temp").join("pt").join("02-services");
    fs::create_dir_all(&temp_dir).map_err(|e| PtError::IoError(e.to_string()))?;

    let mut progress = ProcessProgress::new(2, "Identifying", "Identified", "Services");
    let service_patterns = identify_service_patterns(input_path, &service_day)?;
    save_json(
        &temp_dir.join("service_patterns.json"),
        &service_patterns,
    )?;

    progress.advance("Mapping", "Mapped", "routes to services");
    let service_mappings = create_service_route_mapping(input_path, &service_patterns, &mut progress)?;
    save_json(
        &temp_dir.join("service_mapping.json"),
        &service_mappings,
    )?;

    progress.advance("Selecting", "Selected", "service for metro and bus");
    let selected = select_services(&service_mappings)?;
    save_json(
        &temp_dir.join("selected_services.json"),
        &selected,
    )?;

    progress.complete();

    Ok(())
}

fn identify_service_patterns(
    input_path: &Path,
    service_day: &ServiceDay,
) -> Result<Vec<ServicePattern>, PtError> {
    let mut reader = Reader::from_path(input_path.join("calendar.txt"))
        .map_err(|e| PtError::IoError(e.to_string()))?;

    let patterns: Vec<ServicePattern> = reader
        .deserialize()
        .filter_map(|result| {
            let record: HashMap<String, String> = result.ok()?;
            let matches = match service_day {
                ServiceDay::Weekday => record.get("monday").unwrap_or(&"0".to_string()) == "1" ||
                                     record.get("tuesday").unwrap_or(&"0".to_string()) == "1" ||
                                     record.get("wednesday").unwrap_or(&"0".to_string()) == "1" ||
                                     record.get("thursday").unwrap_or(&"0".to_string()) == "1" ||
                                     record.get("friday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Weekend => record.get("saturday").unwrap_or(&"0".to_string()) == "1" ||
                                     record.get("sunday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Monday => record.get("monday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Tuesday => record.get("tuesday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Wednesday => record.get("wednesday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Thursday => record.get("thursday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Friday => record.get("friday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Saturday => record.get("saturday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Sunday => record.get("sunday").unwrap_or(&"0".to_string()) == "1",
                ServiceDay::Holiday => false,
            };

            if matches {
                Some(ServicePattern {
                    service_id: record.get("service_id")?.clone(),
                    monday: record.get("monday")? == "1",
                    tuesday: record.get("tuesday")? == "1",
                    wednesday: record.get("wednesday")? == "1",
                    thursday: record.get("thursday")? == "1",
                    friday: record.get("friday")? == "1",
                    saturday: record.get("saturday")? == "1",
                    sunday: record.get("sunday")? == "1",
                })
            } else {
                None
            }
        })
        .collect();

    Ok(patterns)
}

fn create_service_route_mapping(
    input_path: &Path,
    service_patterns: &[ServicePattern],
    progress: &mut ProcessProgress,
) -> Result<Vec<ServiceRouteMapping>, PtError> {
    // Get total number of services for progress
    let total = service_patterns.len();
    let mut processed = 0;

    let mut route_reader = Reader::from_path(input_path.join("routes.txt"))
        .map_err(|e| PtError::IoError(e.to_string()))?;

    let routes: HashMap<String, i32> = route_reader
        .deserialize()
        .filter_map(|result| {
            let record: HashMap<String, String> = result.ok()?;
            Some((
                record.get("route_id")?.clone(),
                record.get("route_type")?.parse().ok()?,
            ))
        })
        .collect();

    let mut mappings: Vec<ServiceRouteMapping> = Vec::new();
    
    for pattern in service_patterns {
        let mut trip_reader = Reader::from_path(input_path.join("trips.txt"))
            .map_err(|e| PtError::IoError(e.to_string()))?;

        let mut service_routes = ServiceRouteMapping {
            service_id: pattern.service_id.clone(),
            routes: Vec::new(),
        };

        for result in trip_reader.deserialize() {
            let record: HashMap<String, String> = result
                .map_err(|e| PtError::IoError(e.to_string()))?;
            if let (Some(service_id), Some(route_id)) = (record.get("service_id"), record.get("route_id")) {
                if service_id == &pattern.service_id {
                    if let Some(route_type) = routes.get(route_id) {
                        service_routes.routes.push(RouteInfo {
                            route_id: route_id.clone(),
                            route_type: *route_type,
                        });
                    }
                }
            }
        }

        processed += 1;
        progress.update_target(&format!("routes [{}/{}]", processed, total));

        if !service_routes.routes.is_empty() {
            mappings.push(service_routes);
        }
    }

    Ok(mappings)
}

fn select_services(mappings: &[ServiceRouteMapping]) -> Result<SelectedServices, PtError> {
    use rand::seq::SliceRandom;
    let mut rng = rand::thread_rng();

    let bus_services: Vec<&ServiceRouteMapping> = mappings
        .iter()
        .filter(|m| m.routes.iter().any(|r| r.route_type == 3))
        .collect();

    let metro_services: Vec<&ServiceRouteMapping> = mappings
        .iter()
        .filter(|m| m.routes.iter().any(|r| r.route_type == 1))
        .collect();

    let selected = SelectedServices {
        bus_service: bus_services
            .choose(&mut rng)
            .ok_or(PtError::InvalidService)?
            .service_id
            .clone(),
        metro_service: metro_services
            .choose(&mut rng)
            .ok_or(PtError::InvalidService)?
            .service_id
            .clone(),
        date: Local::now().format("%Y-%m-%d").to_string(),
    };

    Ok(selected)
}

fn save_json<T: serde::Serialize>(path: &Path, data: &T) -> Result<(), PtError> {
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).map_err(|e| PtError::IoError(e.to_string()))?;
    }
    let json = serde_json::to_string_pretty(data)
        .map_err(|e| PtError::JsonError(e.to_string()))?;
    fs::write(path, json).map_err(|e| PtError::IoError(e.to_string()))?;
    Ok(())
}
