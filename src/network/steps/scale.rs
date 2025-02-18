use anyhow::Result;
use crossterm::style::{Color, SetForegroundColor, ResetColor};
use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{BufRead, BufReader, Write};
use std::path::Path;
use crate::network::utils::display;
use super::validate::NetworkConfig;

fn read_population_index(path: &Path, total_steps: usize) -> Result<HashMap<String, Vec<u8>>> {
    let file = File::open(path)?;
    let mut reader = BufReader::new(file);
    let mut agent_scales = HashMap::new();
    let mut lines = Vec::new();
    let mut line = String::new();

    while reader.read_line(&mut line)? > 0 {
        lines.push(line.clone());
        line.clear();
    }
    let total_lines = lines.len();

    let cyan = SetForegroundColor(Color::Cyan);
    let green = SetForegroundColor(Color::Green);
    let reset = ResetColor;

    for (processed, line) in lines.into_iter().enumerate() {
        if let Some((agent_id, scale_str)) = line.split_once(':') {
            let scales: Vec<u8> = scale_str
                .split(',')
                .filter_map(|s| s.trim().parse::<u8>().ok())
                .collect();
            if !scales.is_empty() {
                agent_scales.insert(agent_id.to_string(), scales);
            }
        }
        display::print_step_progress(3, total_steps, processed + 1, total_lines, &format!(
            "{green}Combined{reset} chunks {cyan}→{reset} Mapping population to scale"
        ));
    }

    Ok(agent_scales)
}

fn create_raw_index(input_dir: &Path, output_file: &Path, total_steps: usize) -> Result<()> {
    let mut writer = File::create(output_file)?;
    let mut processed = 0;
    
    let idx_files: Vec<_> = fs::read_dir(input_dir)?
        .filter_map(|entry| entry.ok())
        .filter(|e| e.path().extension().map_or(false, |ext| ext == "idx"))
        .collect();

    let total_files = idx_files.len();

    for entry in idx_files {
        processed += 1;
        let content = fs::read_to_string(entry.path())?;
        write!(writer, "{}", content)?;
        display::print_step_progress(3, total_steps, processed, total_files, "Combining chunks");
    }

    Ok(())
}

pub fn process_scaled_indexes(chunk_dir: &Path, population_idx: &Path, config: &NetworkConfig, path_manager: &crate::util::PathManager) -> Result<()> {
    let total_steps = 8;
    let scale_dir = path_manager.get_network_path("03_scale_traffic_index");
    let raw_index = scale_dir.join("traffic-index-raw.idx");
    path_manager.ensure_dir_exists(&scale_dir)?;
    let cyan = SetForegroundColor(Color::Cyan);
    let green = SetForegroundColor(Color::Green);
    let reset = ResetColor;

    display::print_step_start(3, total_steps, "Combining chunks", None);
    create_raw_index(chunk_dir, &raw_index, total_steps)?;

    display::print_step_start(3, total_steps, &format!("{green}Combined{reset} chunks {cyan}→{reset} Mapping population to scale"), None);
    let agent_scales = read_population_index(population_idx, total_steps)?;

    let file = File::open(&raw_index)?;
    let mut reader = BufReader::new(file);
    let mut lines = Vec::new();
    let mut line = String::new();

    while reader.read_line(&mut line)? > 0 {
        lines.push(line.clone());
        line.clear();
    }
    let total_lines = lines.len();

    display::print_step_start(3, total_steps, &format!("{green}Combined{reset} chunks {cyan}→{reset} {green}Mapped{reset} population to scale {cyan}→{reset} Scaling traffic flow"), None);
    let mut scale_files = HashMap::new();
    
    for (processed, line) in lines.into_iter().enumerate() {
        if let Some((agent_id, _)) = line.split_once(':') {
            if let Some(scales) = agent_scales.get(agent_id) {
                for &scale in scales {
                    let scale_file = scale_dir.join(format!("traffic-index-{:02}.idx", scale));
                    let file = scale_files.entry(scale).or_insert_with(|| {
                        fs::OpenOptions::new()
                            .create(true)
                            .append(true)
                            .open(&scale_file)
                            .unwrap()
                    });
                    writeln!(file, "{}", line.trim())?;
                }
            }
        }
        display::print_step_progress(3, total_steps, processed + 1, total_lines, &format!(
            "{green}Combined{reset} chunks {cyan}→{reset} {green}Mapped{reset} population to scale {cyan}→{reset} Scaling traffic flow"
        ));
    }

    display::print_step_complete(3, total_steps, &format!(
        "{green}Combined{reset} chunks {cyan}→{reset} {green}Mapped{reset} population to scale {cyan}→{reset} {green}Scaled{reset} traffic flow"
    ));

    if config.delete_temp {
        std::fs::remove_dir_all(chunk_dir)?;
    }

    Ok(())
}
