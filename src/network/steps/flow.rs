use anyhow::Result;
use std::collections::{HashMap, HashSet};
use std::fs::{self, File};
use std::io::{BufRead, BufReader, Write};
use std::path::Path;
use crate::network::utils::display;
use crate::network::models::NetworkPaths;

fn process_index_file(input_path: &Path, output_path: &Path) -> Result<()> {
    let file = File::open(input_path)?;
    let reader = BufReader::new(file);
    let mut link_agents: HashMap<String, HashSet<String>> = HashMap::new();

    for line in reader.lines() {
        let line = line?;
        if let Some((agent_id, link_ids_str)) = line.split_once(':') {
            let agent_id = agent_id.trim().to_string();
            for link_id in link_ids_str.trim().split(',').filter(|s| !s.is_empty()) {
                link_agents
                    .entry(link_id.to_string())
                    .or_default()
                    .insert(agent_id.clone());
            }
        }
    }

    let mut output = File::create(output_path)?;
    for (link_id, agents) in link_agents {
        writeln!(output, "{}: {}", link_id, agents.into_iter().collect::<Vec<_>>().join(","))?;
    }

    Ok(())
}

pub fn create_flow_indexes(_paths: &NetworkPaths, path_manager: &crate::util::PathManager) -> Result<()> {
    let input_dir = path_manager.get_network_path("03_scale_traffic_index");
    let output_dir = path_manager.get_network_path("04_flow_index");

    if !input_dir.exists() {
        anyhow::bail!("Input directory not found: {}", input_dir.display());
    }

    path_manager.ensure_dir_exists(&output_dir)?;

    let index_files: Vec<_> = fs::read_dir(&input_dir)?
        .filter_map(|entry| entry.ok())
        .map(|entry| entry.path())
        .filter(|path| path.is_file())
        .collect();

    let total_files = index_files.len();
    for (i, input_path) in index_files.iter().enumerate() {
        display::print_step_progress(4, 8, i + 1, total_files, "Assigning agents to link ids");
        
        let file_name = input_path.file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("");
            
        let scale = if file_name.contains("raw") {
            "raw".to_string()
        } else if let Some(num) = file_name.chars().filter(|c| c.is_ascii_digit()).collect::<String>().parse::<u8>().ok() {
            format!("{:02}", num)
        } else {
            continue;
        };
        
        let output_path = output_dir.join(format!("flow-index-{}.idx", scale));
        process_index_file(input_path, &output_path)?;
    }

    display::print_step_complete(4, 8, "Assigned agents to link ids");
    Ok(())
}
