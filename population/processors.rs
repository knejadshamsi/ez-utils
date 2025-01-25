use anyhow::{Context, Result};
use quick_xml::events::{Event, BytesStart};
use quick_xml::Reader;
use quick_xml::Writer;
use std::collections::HashMap;
use std::io::Cursor;
use std::path::Path;
use std::fs;

use super::models::{Activity, Person, PopulationData, ScaledPopulation};
use super::utils::{create_point_from_coordinates, parse_time, get_total_population};

pub fn validate_scales(current_scale: f64, requested_scales: Option<Vec<f64>>) -> Result<Vec<f64>> {
    if current_scale < 10.0 {
        anyhow::bail!("Current population is less than 10% of total population. Cannot generate scaled versions.");
    }

    let scales = if let Some(scales) = requested_scales {
        scales.into_iter()
            .filter(|&s| s >= 1.0 && s <= 10.0)
            .collect::<Vec<_>>()
    } else {
        (1..=10).map(|i| i as f64).collect()
    };

    if scales.is_empty() {
        anyhow::bail!("No valid scales provided. Scales must be between 1 and 10.");
    }

    let valid_scales = scales.into_iter()
        .filter(|&s| s <= current_scale)
        .collect::<Vec<_>>();

    if valid_scales.is_empty() {
        anyhow::bail!("No valid scales below current scale ({}%). This module only supports scaling down.", current_scale);
    }

    Ok(valid_scales)
}

pub fn process_activities(person_element: &[u8], id: &str) -> Result<Vec<Activity>> {
    let mut reader = Reader::from_reader(Cursor::new(person_element));
    let mut activities = Vec::new();
    let mut buf = Vec::new();
    let mut activity_order = 1;

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Empty(e)) | Ok(Event::Start(e)) if e.name().as_ref() == b"activity" => {
                let element = e.try_to_owned().context("Failed to clone activity element")?;
                let attrs = element.attributes().collect::<Result<Vec<_>, _>>()?;

                let mut activity_type = None;
                let mut facility = None;
                let mut start_time = None;
                let mut end_time = None;
                let mut x = None;
                let mut y = None;

                for attr in attrs {
                    match attr.key.as_ref() {
                        b"type" => activity_type = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                        b"facility" => facility = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                        b"start_time" => start_time = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                        b"end_time" => end_time = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                        b"x" => x = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                        b"y" => y = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                        _ => {}
                    }
                }

                let coordinates = match (x, y) {
                    (Some(x), Some(y)) => Some(create_point_from_coordinates(x, y)?),
                    _ => None,
                };

                let activity_type = activity_type.context("Activity missing type attribute")?;
                let start_time = start_time.and_then(|t| parse_time(&t));
                let end_time = end_time.and_then(|t| parse_time(&t));

                activities.push(Activity {
                    id: id.to_string(),
                    activity_order,
                    activity_type,
                    facility,
                    start_time,
                    end_time,
                    coordinates,
                });

                activity_order += 1;
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            _ => {}
        }
        buf.clear();
    }

    if activities.is_empty() {
        anyhow::bail!("No activities found for person {}", id);
    }

    Ok(activities)
}

pub fn calculate_location_density(persons: &[Person]) -> HashMap<(f64, f64), f64> {
    let mut location_counts = HashMap::new();
    let mut total_agents = 0;

    for person in persons {
        if let Some(first_activity) = person.activities.first() {
            if let Some(coords) = &first_activity.coordinates {
                let location = (coords.x(), coords.y());
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
    location_density: &HashMap<(f64, f64), f64>,
) -> Result<Vec<Person>> {
    if target_count == 0 {
        anyhow::bail!("Target count must be positive");
    }
    if target_count > persons.len() {
        anyhow::bail!("Target count cannot be greater than current population");
    }

    let scale_factor = target_count as f64 / persons.len() as f64;
    let mut persons_by_location: HashMap<(f64, f64), Vec<Person>> = HashMap::new();

    // Group persons by their first activity location
    for person in persons {
        if let Some(first_activity) = person.activities.first() {
            if let Some(coords) = &first_activity.coordinates {
                let location = (coords.x(), coords.y());
                persons_by_location.entry(location)
                    .or_default()
                    .push(person.clone());
            }
        }
    }

    let mut selected_persons = Vec::new();
    for (&location, density) in location_density {
        let target_location_count = (density * scale_factor * persons.len() as f64) as usize;
        if let Some(available_persons) = persons_by_location.get(&location) {
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

pub fn create_scaled_population(
    population_data: &PopulationData,
    target_scale: f64,
) -> Result<ScaledPopulation> {
    if target_scale <= 0.0 {
        anyhow::bail!("Target scale must be positive");
    }
    if target_scale > population_data.current_scale {
        anyhow::bail!(
            "Target scale ({}%) cannot be greater than current scale ({}%)",
            target_scale,
            population_data.current_scale
        );
    }

    let scale_ratio = target_scale / population_data.current_scale;
    let target_count = (population_data.persons.len() as f64 * scale_ratio) as usize;

    let location_density = calculate_location_density(&population_data.persons);
    let selected_persons = select_persons_by_density(
        &population_data.persons,
        target_count,
        &location_density,
    )?;

    Ok(ScaledPopulation {
        scale: target_scale,
        persons: selected_persons,
        location_density,
    })
}

pub fn population_to_xml(population: &ScaledPopulation) -> Result<String> {
    let mut writer = Writer::new(Cursor::new(Vec::new()));
    writer.write_event(Event::Start(BytesStart::new("population")))?;

    for person in &population.persons {
        let mut person_elem = BytesStart::new("person");
        person_elem.push_attribute(("id", person.id.as_str()));
        writer.write_event(Event::Start(person_elem))?;

        let plan_elem = BytesStart::new("plan");
        writer.write_event(Event::Start(plan_elem))?;

        for activity in &person.activities {
            let mut activity_elem = BytesStart::new("activity");
            activity_elem.push_attribute(("type", activity.activity_type.as_str()));

            if let Some(ref facility) = activity.facility {
                activity_elem.push_attribute(("facility", facility.as_str()));
            }
            if let Some(ref start_time) = activity.start_time {
                activity_elem.push_attribute(("start_time", start_time.as_str()));
            }
            if let Some(ref end_time) = activity.end_time {
                activity_elem.push_attribute(("end_time", end_time.as_str()));
            }
            if let Some(ref coords) = activity.coordinates {
                activity_elem.push_attribute(("x", coords.x().to_string().as_str()));
                activity_elem.push_attribute(("y", coords.y().to_string().as_str()));
            }

            writer.write_event(Event::Empty(activity_elem))?;
        }

        writer.write_event(Event::End(BytesStart::new("plan")))?;
        writer.write_event(Event::End(BytesStart::new("person")))?;
    }

    writer.write_event(Event::End(BytesStart::new("population")))?;

    let result = writer.into_inner().into_inner();
    String::from_utf8(result).context("Failed to convert XML to string")
}
