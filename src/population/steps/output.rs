use anyhow::Result;
use std::{
    collections::HashMap,
    fs::{self, File},
    io::{BufWriter, Write},
};
use crate::population::{core::models::{ProcessingPaths, PopulationData}, utils::display::ProcessProgress};

fn extract_agent_id(line: &str) -> Option<&str> {
    if line.trim().starts_with("<person id=\"") {
        let start = line.find('"')? + 1;
        let end = line[start..].find('"')?;
        Some(&line[start..start + end])
    } else {
        None
    }
}

pub fn process_step_6(paths: &ProcessingPaths, population_data: &PopulationData, progress: &mut ProcessProgress) -> Result<()> {
    progress.print_step_6_indexing(0)?;
    let _total_agents = create_scale_index(paths)?;
    progress.print_step_6_indexing(100)?;

    let source_chunks_dir = paths.temp_dir.join("population").join("01_split_population_chunks").join("chunks");
    let total_chunks = fs::read_dir(&source_chunks_dir)?.count();
    if total_chunks == 0 {
        return Err(anyhow::anyhow!("No source chunks found"));
    }
    progress.print_step_6_indexed(0, total_chunks as u64)?;

    process_scaled_chunks(paths, population_data, progress)?;
    progress.print_step_6_created(0)?;
    combine_scaled_chunks(paths, population_data, progress)?;

    progress.print_step_6_complete();
    Ok(())
}

fn create_scale_index(paths: &ProcessingPaths) -> Result<usize> {
    let output_dir = paths.temp_dir.join("population").join("06_generate_scaled_population");
    fs::create_dir_all(&output_dir)?;

    let mut agent_scales: HashMap<String, Vec<u32>> = HashMap::new();
    
    for scale in 1..=10u32 {
        let scale_str = format!("{:02}", scale);
        let dense_path = paths.dense_dir.join(&scale_str);
        
        if !dense_path.exists() { continue; }

        for entry in fs::read_dir(&dense_path)? {
            let entry = entry?;
            let content = fs::read_to_string(entry.path())?;
            
            for line in content.lines() {
                let agent_id = line.trim();
                if !agent_id.is_empty() {
                    agent_scales.entry(agent_id.to_string())
                        .or_insert_with(Vec::new)
                        .push(scale);
                }
            }
        }
    }

    let index_file = output_dir.join("new_population.idx");
    let mut writer = BufWriter::new(File::create(&index_file)?);

    let mut entries: Vec<_> = agent_scales.into_iter().collect();
    entries.sort_by(|a, b| a.0.cmp(&b.0));

    let total_agents = entries.len();
    for (agent_id, mut scales) in entries {
        scales.sort_unstable();
        scales.dedup();
        writeln!(writer, "{}:{}", agent_id, scales.iter()
            .map(|s| s.to_string())
            .collect::<Vec<_>>()
            .join(","))?;
    }

    if total_agents == 0 {
        return Err(anyhow::anyhow!("No agents found in dense directories"));
    }

    Ok(total_agents)
}

fn process_scaled_chunks(paths: &ProcessingPaths, _population_data: &PopulationData, progress: &mut ProcessProgress) -> Result<()> {
    let index_path = paths.temp_dir
        .join("population")
        .join("06_generate_scaled_population")
        .join("new_population.idx");
    
    let chunks_base = paths.temp_dir
        .join("population")
        .join("06_generate_scaled_population")
        .join("chunks");

    let content = fs::read_to_string(&index_path)?;
    
    let mut agent_scales: HashMap<String, Vec<u32>> = HashMap::new();
    for line in content.lines() {
        if let Some((agent_id, scales)) = line.split_once(':') {
            let scales: Vec<u32> = scales
                .split(',')
                .filter_map(|s| s.trim().parse().ok())
                .collect();
            if !scales.is_empty() {
                agent_scales.insert(agent_id.trim().to_string(), scales);
            }
        }
    }

    for scale in 1..=10 {
        let scale_str = format!("{:02}", scale);
        let scale_dir = chunks_base.join(&scale_str);
        fs::create_dir_all(&scale_dir)?;
    }

    let source_chunks_dir = paths.temp_dir
        .join("population")
        .join("01_split_population_chunks")
        .join("chunks");

    if !source_chunks_dir.exists() {
        return Err(anyhow::anyhow!("Source chunks directory not found"));
    }

    let mut chunks = Vec::new();
    for entry in fs::read_dir(&source_chunks_dir)? {
        chunks.push(entry?);
    }

    let total_chunks = chunks.len();
    if total_chunks == 0 {
        return Err(anyhow::anyhow!("No source chunks found"));
    }

    for (i, dir_entry) in chunks.into_iter().enumerate() {
        progress.print_step_6_indexed(i as u64 + 1, total_chunks as u64)?;
        let content = fs::read_to_string(dir_entry.path())?;
        
        let mut writers: HashMap<String, BufWriter<File>> = HashMap::new();
        for scale in 1..=10 {
            let scale_str = format!("{:02}", scale);
            let scale_dir = chunks_base.join(&scale_str);
            let chunk_path = scale_dir.join(dir_entry.file_name());
            writers.insert(scale_str, BufWriter::new(File::create(chunk_path)?));
        }

        let mut current_lines = Vec::new();
        let mut current_agent_id = None;

        for line in content.lines() {
            let trimmed = line.trim();
            if trimmed.is_empty() { continue; }

            if trimmed.starts_with("<person") {
                if let Some(id) = extract_agent_id(trimmed) {
                    // Write previous person if exists
                    if let Some(agent_id) = current_agent_id.take() {
                        if let Some(scales) = agent_scales.get(&agent_id) {
                            let person_content = current_lines.join("\n") + "\n"; // Ensure newline after each person
                            for &scale in scales {
                                if let Some(writer) = writers.get_mut(&format!("{:02}", scale)) {
                                    write!(writer, "{}", person_content)?;
                                }
                            }
                        }
                    }
                    current_lines.clear();
                    current_agent_id = Some(id.to_string());
                }
            }
            current_lines.push(line.to_string());
        }

        // Write last person if exists
        if let Some(agent_id) = current_agent_id {
            if let Some(scales) = agent_scales.get(&agent_id) {
                let person_content = current_lines.join("\n") + "\n";
                for &scale in scales {
                    if let Some(writer) = writers.get_mut(&format!("{:02}", scale)) {
                        write!(writer, "{}", person_content)?;
                    }
                }
            }
        }

        for mut writer in writers.into_values() {
            writer.flush()?;
        }
    }

    Ok(())
}

fn combine_scaled_chunks(paths: &ProcessingPaths, _population_data: &PopulationData, progress: &mut ProcessProgress) -> Result<()> {
    fs::create_dir_all(&paths.output_dir)?;
    let total_scales = 10;
    let mut processed = 0;

    for scale in 1..=10 {
        let scale_str = format!("{:02}", scale);
        let scale_dir = paths.temp_dir
            .join("population")
            .join("06_generate_scaled_population")
            .join("chunks")
            .join(&scale_str);

        let output_file = paths.output_dir.join(format!("population-{}.xml", scale_str));
        
        if !scale_dir.exists() || fs::read_dir(&scale_dir)?.next().is_none() {
            processed += 1;
            progress.print_step_6_created((processed * 100 / total_scales) as u32)?;
            continue;
        }

        let mut writer = BufWriter::new(File::create(&output_file)?);

        writeln!(writer, "<?xml version=\"1.0\" encoding=\"utf-8\"?>")?;
        writeln!(writer, "<!DOCTYPE population SYSTEM \"http://www.matsim.org/files/dtd/population_v6.dtd\">")?;
        writeln!(writer, "<population desc=\"Switzerland Baseline\">")?;

        let mut entries: Vec<_> = fs::read_dir(&scale_dir)?.collect::<Result<_, _>>()?;
        entries.sort_by_key(|e| e.path());

        for entry in entries {
            let chunk_path = entry.path();
            let content = fs::read_to_string(&chunk_path)?;
            
            for line in content.lines() {
                let trimmed = line.trim();
                if trimmed.is_empty() { continue; }

                if trimmed.starts_with("<?xml") || 
                   trimmed.starts_with("<!DOCTYPE") || 
                   trimmed == "<population desc=\"Switzerland Baseline\">" ||
                   trimmed == "</population>" {
                    continue;
                }
                writeln!(writer, "    {}", line)?;
            }
        }

        writeln!(writer, "</population>")?;
        writer.flush()?;

        processed += 1;
        progress.print_step_6_created((processed * 100 / total_scales) as u32)?;
    }

    Ok(())
}
