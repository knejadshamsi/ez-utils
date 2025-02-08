use anyhow::{anyhow, Result};
use quick_xml::events::{Event, BytesStart, BytesEnd};
use quick_xml::Reader;
use quick_xml::Writer;
use std::io::Cursor;
use std::path::Path;
use std::fs;
use crate::population::core::models::*;
use crate::population::utils::file::{create_point_from_coordinates, parse_time, mtm8_to_latlon};
use std::fmt::Write;

pub fn process_chunk(chunk: &ChunkData, index_dir: &Path) -> Result<()> {
    let mut reader = Reader::from_str(&chunk.content);
    reader.trim_text(true);
    reader.check_end_names(false);
    let mut buf = Vec::new();
    let mut agents = Vec::new();
    let mut current_agent_id = None;
    let mut found_coordinates = None;
    let mut in_plan = false;
    let mut is_selected_plan = false;
    let mut first_plan = true;
    
    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => {
                match e.name().as_ref() {
                    b"person" => {
                        current_agent_id = None;
                        found_coordinates = None;
                        first_plan = true;

                        let attrs = e.attributes().collect::<Result<Vec<_>, _>>()?;
                        if let Some(id_attr) = attrs.iter().find(|attr| attr.key.as_ref() == b"id") {
                            current_agent_id = Some(String::from_utf8_lossy(&id_attr.value).into_owned());
                        }
                    },
                    b"plan" => {
                        in_plan = true;
                        let attrs = e.attributes().collect::<Result<Vec<_>, _>>()?;
                        if let Some(selected_attr) = attrs.iter().find(|attr| attr.key.as_ref() == b"selected") {
                            is_selected_plan = String::from_utf8_lossy(&selected_attr.value) == "yes";
                        }
                    },
                    b"activity" if current_agent_id.is_some() && found_coordinates.is_none() && in_plan && (first_plan || is_selected_plan) => {
                        let attrs = e.attributes().collect::<Result<Vec<_>, _>>()?;
                        let mut x = None;
                        let mut y = None;

                        for attr in attrs {
                            match attr.key.as_ref() {
                                b"x" => x = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                                b"y" => y = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                                _ => {}
                            }
                        }

                        if let (Some(x), Some(y)) = (x, y) {
                            let (lon, lat) = mtm8_to_latlon(x, y)?;
                            found_coordinates = Some(create_point_from_coordinates(lon, lat)?);
                        }
                    },
                    _ => {}
                }
            },
            Ok(Event::Empty(e)) => {
                if e.name().as_ref() == b"activity" && current_agent_id.is_some() && 
                   found_coordinates.is_none() && in_plan && (first_plan || is_selected_plan) {
                    let attrs = e.attributes().collect::<Result<Vec<_>, _>>()?;
                    let mut x = None;
                    let mut y = None;

                    for attr in attrs {
                        match attr.key.as_ref() {
                            b"x" => x = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                            b"y" => y = Some(String::from_utf8_lossy(&attr.value).parse::<f64>()?),
                            _ => {}
                        }
                    }

                    if let (Some(x), Some(y)) = (x, y) {
                        let (lon, lat) = mtm8_to_latlon(x, y)?;
                        found_coordinates = Some(create_point_from_coordinates(lon, lat)?);
                    }
                }
            },
            Ok(Event::End(e)) => {
                match e.name().as_ref() {
                    b"person" => {
                        if let (Some(agent_id), Some(coordinates)) = (&current_agent_id, found_coordinates) {
                            agents.push(AgentIndex {
                                agent_id: agent_id.clone(),
                                coordinates,
                                xml_content: String::new(),
                            });
                        }
                        current_agent_id = None;
                        found_coordinates = None;
                        in_plan = false;
                        is_selected_plan = false;
                        first_plan = true;
                    },
                    b"plan" => {
                        in_plan = false;
                        if found_coordinates.is_none() {
                            first_plan = false;
                        }
                    },
                    _ => {}
                }
            },
            Ok(Event::Eof) => break,
            Err(e) => {
                // Log the error but continue processing
                eprintln!("Warning: XML parsing error in chunk {}: {}", chunk.chunk_number, e);
                continue;
            },
            _ => {}
        }
        buf.clear();
    }

    let index_path = index_dir.join(format!("index_{:04}.idx", chunk.chunk_number));
    if agents.is_empty() {
        return Err(anyhow!("No valid agents found in chunk {}", chunk.chunk_number));
    }

    let mut index_content = String::new();
    for agent in &agents {
        writeln!(
            index_content,
            "{}: {}, {}",
            agent.agent_id,
            agent.coordinates.x(),
            agent.coordinates.y()
        )?;
    }

    fs::write(&index_path, &index_content)
        .map_err(|e| anyhow!("Failed to write index file {}: {}", index_path.display(), e))?;

    Ok(())
}

pub fn parse_activities_from_xml(xml_content: &str, person_id: &str) -> Result<Vec<Activity>> {
    let mut reader = Reader::from_str(xml_content);
    reader.trim_text(true);
    reader.check_end_names(false); // Disable end tag name validation
    let mut activities = Vec::new();
    let mut buf = Vec::new();
    let mut activity_order = 1;

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Empty(e)) | Ok(Event::Start(e)) if e.name().as_ref() == b"activity" => {
                let element = e.to_owned();
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
                    (Some(x), Some(y)) => {
                        let (lon, lat) = mtm8_to_latlon(x, y)?;
                        Some(create_point_from_coordinates(lon, lat)?)
                    },
                    _ => None,
                };

                let activity_type = activity_type.ok_or_else(|| anyhow!("Activity missing type attribute"))?;
                let start_time = start_time.and_then(|t| parse_time(&t));
                let end_time = end_time.and_then(|t| parse_time(&t));

                activities.push(Activity {
                    id: person_id.to_string(),
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
            Err(e) => {
                // Log the error but continue processing
                eprintln!("Warning: XML parsing error for person {}: {}", person_id, e);
                continue;
            },
            _ => {}
        }
        buf.clear();
    }

    if activities.is_empty() {
        anyhow::bail!("No activities found for person {}", person_id);
    }

    Ok(activities)
}

pub fn population_to_xml(population: &ScaledPopulation) -> Result<String> {
    let mut writer = Writer::new(Cursor::new(Vec::new()));
    writer.write_event(Event::Start(BytesStart::new("population")))?;

    for person in &population.persons {
        let mut person_elem = BytesStart::new("person");
        person_elem.push_attribute(("id", person.id.as_str()));
        writer.write_event(Event::Start(person_elem))?;

        let mut plan_elem = BytesStart::new("plan");
        plan_elem.push_attribute(("selected", "yes"));
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

        writer.write_event(Event::End(BytesEnd::new("plan")))?;
        writer.write_event(Event::End(BytesEnd::new("person")))?;
    }

    writer.write_event(Event::End(BytesEnd::new("population")))?;

    let result = writer.into_inner().into_inner();
    String::from_utf8(result).map_err(|e| anyhow!("Failed to convert XML to string: {}", e))
}
