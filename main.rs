use anyhow::Result;
use clap::Parser;
use indicatif::{ProgressBar, ProgressStyle};
use quick_xml::{events::Event, Reader};
use std::io::Cursor;
use std::{fs, path::PathBuf};
use tokio;

mod population;
use population::{
    models::{Activity, Person, PopulationData, ScaledPopulation},
    processors::{process_activities, create_scaled_population, population_to_xml, validate_scales},
    utils::{get_total_population, validate_output_dir},
    db::Database,
};

#[derive(Parser)]
#[command(author, version, about)]
struct Args {
    #[arg(help = "Input large XML file")]
    input: PathBuf,
    
    #[arg(short, long, help = "Output directory for chunks")]
    output: Option<PathBuf>,
    
    #[arg(short, long, default_value = "2000", help = "Agents per chunk")]
    chunk_size: usize,

    #[arg(short, long, help = "Comma-separated list of scales (1-10)")]
    scales: Option<String>,

    #[arg(long, help = "Export to PostgreSQL database")]
    export_db: bool,
}

fn process_xml_in_chunks(input_path: &PathBuf, output_dir: &PathBuf, chunk_size: usize) -> Result<()> {
    // Read entire file into memory
    let file_content = fs::read(input_path)?;
    let mut reader = Reader::from_reader(Cursor::new(&file_content));
     reader.trim_text(true);
    
    fs::create_dir_all(output_dir)?;
    
    let mut buf = Vec::new();
    let mut agent_count = 0;
    let mut chunk_num = 1;
    let mut current_chunk = Vec::new();
    
    let pb = ProgressBar::new_spinner();
    pb.set_style(ProgressStyle::default_spinner()
        .template("{spinner:.green} Processing chunk {msg}")
        .unwrap());
    
    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) if e.name().as_ref() == b"person" => {
                // Store start position of person element
                let start_pos = reader.buffer_position();
                let mut depth = 1;
                let mut inner_buf = Vec::new();

                // Read until matching end tag
                while depth > 0 {
                    match reader.read_event_into(&mut inner_buf) {
                        Ok(Event::Start(_)) => depth += 1,
                        Ok(Event::End(_)) => depth -= 1,
                        Err(e) => return Err(e.into()),
                        _ => ()
                    }
                    inner_buf.clear();
                }

                // Get end position
                let end_pos = reader.buffer_position();
                
                // Extract raw bytes between start and end positions
                let raw_xml = file_content.get(start_pos..end_pos)
                    .ok_or_else(|| anyhow::anyhow!("Invalid byte range"))?.to_vec();
                
                current_chunk.push(raw_xml);
                agent_count += 1;
                
                if agent_count >= chunk_size {
                    // Write chunk to file
                    let chunk_path = output_dir.join(format!("chunk_{:04}.xml", chunk_num));
                    let mut output = Vec::new();
                    output.extend_from_slice(b"<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<population>\n");
                    for agent in &current_chunk {
                        output.extend_from_slice(agent);
                        output.extend_from_slice(b"\n");
                    }
                    output.extend_from_slice(b"</population>");
                    fs::write(chunk_path, output)?;
                    
                    pb.set_message(format!("{}", chunk_num));
                    chunk_num += 1;
                    agent_count = 0;
                    current_chunk.clear();
                }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            _ => (),
        }
        buf.clear();
    }
    
    // Write remaining agents if any
    if !current_chunk.is_empty() {
        let chunk_path = output_dir.join(format!("chunk_{:04}.xml", chunk_num));
        let mut output = Vec::new();
        output.extend_from_slice(b"<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<population>\n");
        for agent in &current_chunk {
            output.extend_from_slice(agent);
            output.extend_from_slice(b"\n");
        }
        output.extend_from_slice(b"</population>");
        fs::write(chunk_path, output)?;
    }
    
    pb.finish_with_message(format!("Created {} chunks", chunk_num));
    Ok(())
}

async fn process_population_chunk(
    chunk_path: PathBuf,
    scales: Option<Vec<f64>>,
    db: Option<&Database>,
) -> Result<()> {
    let content = fs::read(&chunk_path)?;
    let mut reader = Reader::from_reader(Cursor::new(&content));
    let mut buf = Vec::new();
    let mut persons = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) if e.name().as_ref() == b"person" => {
                let element = e.try_to_owned()?;
                let id = element
                    .attributes()
                    .find(|a| a.as_ref().map(|attr| attr.key.as_ref() == b"id").unwrap_or(false))
                    .ok_or_else(|| anyhow::anyhow!("Person missing id attribute"))?
                    .map(|attr| String::from_utf8_lossy(&attr.value).into_owned())?;

                let start_pos = reader.buffer_position();
                let mut depth = 1;
                let mut inner_buf = Vec::new();

                while depth > 0 {
                    match reader.read_event_into(&mut inner_buf) {
                        Ok(Event::Start(_)) => depth += 1,
                        Ok(Event::End(_)) => depth -= 1,
                        Err(e) => return Err(e.into()),
                        _ => ()
                    }
                    inner_buf.clear();
                }

                let end_pos = reader.buffer_position();
                let person_content = content[start_pos..end_pos].to_vec();
                
                let activities = process_activities(&person_content, &id)?;
                persons.push(Person { id, activities });
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            _ => (),
        }
        buf.clear();
    }

    let total_population = get_total_population()?;
    let current_scale = (persons.len() as f64 / total_population as f64) * 100.0;

    let population_data = PopulationData {
        persons,
        total_population,
        current_scale,
    };

    let valid_scales = validate_scales(current_scale, scales)?;
    
    for scale in valid_scales {
        let scaled_population = create_scaled_population(&population_data, scale)?;
        let output_xml = population_to_xml(&scaled_population)?;
        
        let output_path = chunk_path.with_file_name(format!("population_{:02}.xml", scale as i32));
        fs::write(output_path, output_xml)?;

        if let Some(db) = db {
            for person in &scaled_population.persons {
                if let Some(first_activity) = person.activities.first() {
                    if let Some(coords) = &first_activity.coordinates {
                        let start_coord = format!("POINT({} {})", coords.x(), coords.y());
                        db.insert_agent(
                            &person.id,
                            &start_coord,
                            &[scale as i32],
                            &format!("<person id=\"{}\">{}</person>", person.id, output_xml),
                        ).await?;
                    }
                }
            }
        }
    }

    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();
    let output_dir = args.output.unwrap_or_else(|| PathBuf::from("chunks"));
    validate_output_dir(&output_dir)?;

    let scales = args.scales.map(|s| {
        s.split(',')
            .filter_map(|scale| scale.trim().parse::<f64>().ok())
            .collect::<Vec<_>>()
    });

    let db = if args.export_db {
        Some(Database::new().await?)
    } else {
        None
    };

    if let Some(ref db) = db {
        db.ensure_tables_exist().await?;
    }

    process_xml_in_chunks(&args.input, &output_dir, args.chunk_size)?;

    let pb = ProgressBar::new_spinner();
    pb.set_style(ProgressStyle::default_spinner()
        .template("{spinner:.green} Processing population chunks")
        .unwrap());

    for entry in fs::read_dir(&output_dir)? {
        let path = entry?.path();
        if path.is_file() && path.extension().map_or(false, |ext| ext == "xml") {
            process_population_chunk(path, scales.clone(), db.as_ref()).await?;
            pb.inc(1);
        }
    }

    pb.finish_with_message("Population processing complete");
    Ok(())
}
