use anyhow::Result;
use std::path::Path;
use std::fs::File;
use std::io::{BufWriter, Write};
use std::collections::HashSet;
use quick_xml::Reader;
use quick_xml::events::Event;
use crate::network::utils::display;

pub struct ChunkData {
    pub chunk_id: usize,
    pub file_path: String,
    pub chunk_number: usize,
}

fn create_chunk_index(input_path: &Path, output_dir: &Path) -> Result<()> {
    let mut reader = Reader::from_file(input_path)?;
    reader.trim_text(true);

    let file_number = input_path
        .file_name()
        .and_then(|n| n.to_str())
        .and_then(|s| s.strip_prefix("chunk_"))
        .and_then(|s| s.strip_suffix(".xml"))
        .ok_or(anyhow::anyhow!("Invalid chunk filename format"))?;

    let output_path = output_dir.join(format!("chunk_{:04}.idx", file_number));

    let output_file = File::create(&output_path)?;
    let mut writer = BufWriter::new(output_file);
    let mut buf = Vec::new();

    let mut current_agent_id = String::new();
    let mut in_person = false;
    let mut in_selected_plan = false;
    let mut found_selected = false;
    let mut link_ids = Vec::new();
    let mut in_route = false;
    let mut is_valid_route = true;

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(ref e)) => match e.name().as_ref() {
                b"person" => {
                    in_person = true;
                    current_agent_id.clear();
                    for attr in e.attributes() {
                        if let Ok(attr) = attr {
                            if attr.key.as_ref() == b"id" {
                                current_agent_id = String::from_utf8_lossy(&attr.value).into_owned();
                                break;
                            }
                        }
                    }
                }
                b"plan" => {
                    if in_person && !found_selected {
                        for attr in e.attributes() {
                            if let Ok(attr) = attr {
                                if attr.key.as_ref() == b"selected" && attr.value.as_ref() == b"yes" {
                                    in_selected_plan = true;
                                    found_selected = true;
                                    break;
                                }
                            }
                        }
                        if !found_selected {
                            in_selected_plan = true;
                        }
                    }
                }
                b"route" => {
                    if in_selected_plan {
                        in_route = true;
                        is_valid_route = true;
                        link_ids.clear();
                    }
                }
                _ => (),
            },
            Ok(Event::Text(e)) if in_route => {
                let text = String::from_utf8_lossy(&e.into_inner()).trim().to_string();
                if !text.is_empty() {
                    if text.contains('{') || text.contains('}') {
                        is_valid_route = false;
                    } else {
                        let unique_links: HashSet<_> = text.split_whitespace()
                            .map(String::from)
                            .collect();
                        link_ids.extend(unique_links);
                    }
                }
            }
            Ok(Event::End(ref e)) => match e.name().as_ref() {
                b"person" => {
                    if !current_agent_id.is_empty() && !link_ids.is_empty() && is_valid_route {
                        writeln!(writer, "{}: {}", current_agent_id, link_ids.join(","))?;
                    }
                    in_person = false;
                    in_selected_plan = false;
                    found_selected = false;
                    link_ids.clear();
                    is_valid_route = true;
                }
                b"plan" => {
                    in_selected_plan = false;
                }
                b"route" => {
                    in_route = false;
                }
                _ => (),
            },
            Ok(Event::Eof) => break,
            Err(e) => return Err(anyhow::anyhow!("Error parsing XML: {}", e)),
            _ => (),
        }
        buf.clear();
    }

    Ok(())
}

pub fn process_chunks(chunks_dir: &Path, path_manager: &crate::util::PathManager) -> Result<()> {
    let total_steps = 8;
    let output_dir = path_manager.get_network_path("02_traffic_index");
    
    if !output_dir.exists() {
        path_manager.ensure_dir_exists(&output_dir)?;
    }

    let chunks: Vec<_> = std::fs::read_dir(chunks_dir)?
        .filter_map(|entry| entry.ok())
        .filter(|entry| {
            entry.path().extension().map_or(false, |ext| ext == "xml")
        })
        .enumerate()
        .map(|(idx, entry)| ChunkData {
            chunk_id: idx,
            file_path: entry.path().to_string_lossy().into_owned(),
            chunk_number: idx + 1,
        })
        .collect();

    let total_chunks = chunks.len();

    display::print_step_start(2, total_steps, "Indexing traffic flow", None);
    for (i, chunk) in chunks.into_iter().enumerate() {
        create_chunk_index(Path::new(&chunk.file_path), &output_dir)?;
        display::print_step_progress(2, total_steps, i + 1, total_chunks, "Indexing traffic flow");
    }
    display::print_step_complete(2, total_steps, "Indexed traffic flow");

    Ok(())
}
