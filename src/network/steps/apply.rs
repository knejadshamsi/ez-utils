use anyhow::Result;
use quick_xml::{
    events::{Event, BytesStart},
    reader::Reader,
    writer::Writer
};
use std::{
    collections::HashMap,
    fs::{self, File},
    io::{BufReader, Write},
    path::Path
};
use serde_json::from_reader;
use crossterm::style::{Color, SetForegroundColor, ResetColor};
use crate::network::{
    utils::display::print_step_progress,
    steps::output::get_raw_nodes_content,
    models::{NetworkPaths, LinkRatio}
};

fn read_ratios(ratio_dir: &Path, scale: u8) -> Result<HashMap<String, LinkRatio>> {
    let path = ratio_dir.join(format!("ratio-{:02}.json", scale));
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    Ok(from_reader(reader)?)
}

fn process_link_attributes(element: &BytesStart, ratios: &HashMap<String, LinkRatio>) -> Event<'static> {
    let mut elem = element.to_owned();
    let mut id = None;
    let mut freespeed = None;
    let mut capacity = None;

    for attr in element.attributes() {
        if let Ok(attr) = attr {
            match attr.key.as_ref() {
                b"id" => id = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                b"freespeed" => freespeed = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                b"capacity" => capacity = Some(String::from_utf8_lossy(&attr.value).into_owned()),
                _ => continue,
            }
        }
    }

    if let (Some(id), Some(speed_str), Some(cap_str)) = (id, freespeed, capacity) {
        if let (Ok(speed), Ok(cap)) = (speed_str.parse::<f64>(), cap_str.parse::<f64>()) {
            if let Some(ratio) = ratios.get(&id) {
                let new_speed = speed * ratio.ratio;
                let new_capacity = cap * ratio.ratio;

                elem = BytesStart::new("link");
                for attr in element.attributes() {
                    if let Ok(attr) = attr {
                        match attr.key.as_ref() {
                            b"freespeed" => elem.push_attribute(("freespeed", new_speed.to_string().as_str())),
                            b"capacity" => elem.push_attribute(("capacity", new_capacity.to_string().as_str())),
                            _ => elem.push_attribute(attr.clone()),
                        }
                    }
                }
            }
        }
    }

    Event::Start(elem)
}

fn process_chunk(content: &str, ratios: &HashMap<String, LinkRatio>) -> Result<String> {
    let mut reader = Reader::from_str(content);
    reader.trim_text(true);

    let mut writer = Writer::new(Vec::new());
    let mut buf = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(ref e)) if e.name().as_ref() == b"link" => {
                writer.write_event(process_link_attributes(e, ratios))?;
                writer.write_event(Event::End(BytesStart::new("link").to_end()))?;
            }
            Ok(Event::End(e)) if e.name().as_ref() == b"link" => {
            }
            Ok(Event::Eof) => break,
            Ok(event) => writer.write_event(event)?,
            Err(e) => return Err(anyhow::anyhow!("Error processing XML: {}", e)),
        }
        buf.clear();
    }

    Ok(String::from_utf8(writer.into_inner())?)
}

pub fn apply_network_ratios(
    paths: &NetworkPaths,
    scale_index: u8,
    total_scales: u8
) -> Result<()> {
    let ratio_dir = paths.temp_dir.join("network").join("05_ratio_index");
    let ratios = read_ratios(&ratio_dir, scale_index)?;
    
    let raw_nodes = get_raw_nodes_content(paths)?;
    let chunks_dir = paths.network_split_dir.join("chunks");
    let entries: Vec<_> = fs::read_dir(&chunks_dir)?
        .filter_map(|e| e.ok())
        .filter(|e| e.path().is_file() && e.path().extension().and_then(|s| s.to_str()) == Some("xml"))
        .collect();

    let total_chunks = entries.len();
    let scaled_dir = paths.temp_dir.join("network").join("07_scaled_network");
    let output_file = scaled_dir.join(format!("scaled-network-{:02}.xml", scale_index));
    let mut file = File::create(&output_file)?;

    file.write_all(raw_nodes.as_bytes())?;
    file.write_all(b"<links>\n")?;

    for (chunk_index, entry) in entries.iter().enumerate() {
        print_step_progress(7, 8, chunk_index, total_chunks, 
            &format!("Applying ratio for scale {}[{}/{}]{}. Chunk processed:", 
                SetForegroundColor(Color::Yellow),
                scale_index, total_scales,
                ResetColor));

        let content = fs::read_to_string(entry.path())?;
        let processed = process_chunk(&content, &ratios)?;
        file.write_all(processed.as_bytes())?;
    }

    file.write_all(b"</links>\n</network>")?;
    Ok(())
}
