use anyhow::Result;
use crate::population::core::models::*;
use rand::{seq::SliceRandom, thread_rng};
use std::{
    collections::HashMap,
    fs,
    path::Path,
};

pub fn parse_basic_activities(xml_content: &str) -> Result<Vec<Activity>> {
    let mut activities = Vec::new();
    let mut activity_order = 0;
    
    for line in xml_content.lines() {
        if line.contains("<activity ") {
            if let Some(start_idx) = line.find("type=\"") {
                if let Some(end_idx) = line[start_idx + 6..].find('\"') {
                    let activity_type = line[start_idx + 6..start_idx + 6 + end_idx].to_string();
                    activities.push(Activity {
                        id: format!("act_{}", activity_order),
                        activity_type,
                        activity_order,
                        facility: None,
                        start_time: None,
                        end_time: None,
                        coordinates: None,
                    });
                    activity_order += 1;
                }
            }
        }
    }
    Ok(activities)
}

pub fn calculate_agent_scales(
    current_scale: f64,
    target_scale: f64,
    total_agents: usize,
) -> Result<Vec<f64>> {
    if target_scale > current_scale {
        anyhow::bail!("Target scale cannot be larger than current scale");
    }
    
    let scale_ratios = vec![target_scale / current_scale; total_agents];
    Ok(scale_ratios)
}

pub fn validate_scales(current_scale: f64, requested_scales: Option<Vec<f64>>) -> Result<Vec<f64>> {
    if current_scale < 1.0 {
        anyhow::bail!("Input scale must be at least 1% for processing");
    }

    let scales = match requested_scales {
        Some(scales) => scales,
        None => (1..=10).map(|i| i as f64).collect()
    };

    let valid_scales = scales.into_iter()
        .filter(|scale| *scale <= current_scale && *scale >= 1.0)
        .collect::<Vec<_>>();

    if valid_scales.is_empty() {
        anyhow::bail!("No valid scales found between 1% and current scale ({}%)", current_scale);
    }

    Ok(valid_scales)
}

fn read_bin_file(path: &Path) -> Result<Vec<String>> {
    let content = fs::read_to_string(path)?;
    Ok(content.lines().map(String::from).collect())
}

fn calculate_bin_stats(assigned_dir: &Path) -> Result<(HashMap<String, usize>, usize)> {
    let mut bin_counts = HashMap::new();
    let mut total_agents = 0;

    for entry in fs::read_dir(assigned_dir)? {
        let path = entry?.path();
        if path.is_file() {
            let bin_name = path.file_name()
                .and_then(|n| n.to_str())
                .map(String::from)
                .ok_or_else(|| anyhow::anyhow!("Invalid bin filename"))?;
            
            let agents = read_bin_file(&path)?;
            let count = agents.len();
            bin_counts.insert(bin_name, count);
            total_agents += count;
        }
    }

    Ok((bin_counts, total_agents))
}

fn calculate_bin_ratios(bin_counts: &HashMap<String, usize>, total_agents: usize) -> HashMap<String, f64> {
    bin_counts.iter()
        .map(|(bin, count)| {
            (bin.clone(), *count as f64 / total_agents as f64)
        })
        .collect()
}

fn calculate_agents_to_remove(
    bin_ratios: &HashMap<String, f64>,
    total_agents: usize,
    target_scale: f64,
    total_population: usize
) -> HashMap<String, usize> {
    let scale_ratio = target_scale / 100.0;
    let target_agents = (total_population as f64 * scale_ratio).round() as usize;
    let agents_to_remove = total_agents.saturating_sub(target_agents);

    bin_ratios.iter()
        .map(|(bin, ratio)| {
            let remove_count = (agents_to_remove as f64 * ratio).round() as usize;
            (bin.clone(), remove_count)
        })
        .collect()
}

fn randomly_select_agents(agents: &[String], retain_count: usize) -> Vec<String> {
    let mut rng = thread_rng();
    let mut agents = agents.to_vec();
    agents.shuffle(&mut rng);
    agents.truncate(retain_count);
    agents
}

pub fn process_for_files(paths: &ProcessingPaths, scale_config: &ScaleConfig, population_data: &PopulationData) -> Result<()> {
    // Step 1 & 2: Calculate bin counts and ratios
    let (bin_counts, total_agents) = calculate_bin_stats(&paths.assigned_dir)?;
    let bin_ratios = calculate_bin_ratios(&bin_counts, total_agents);

    // Validate scales - use empty vector to get default 1-10 scales when no scales specified
    let scales = if scale_config.scales.is_empty() {
        validate_scales(population_data.current_scale, None)?
    } else {
        validate_scales(population_data.current_scale, Some(scale_config.scales.clone()))?
    };

    // Process each scale
    for scale in scales {
        let scale_str = format!("{:02}", scale as u32);
        let scale_dir = paths.dense_dir.join(&scale_str);
        fs::create_dir_all(&scale_dir)?;

        // Step 3-7: Calculate agents to remove for each bin
        let removal_counts = calculate_agents_to_remove(
            &bin_ratios,
            total_agents,
            scale,
            population_data.total_population
        );

        // Step 8: Process each bin
        for (bin_name, remove_count) in removal_counts {
            let input_path = paths.assigned_dir.join(&bin_name);
            let agents = read_bin_file(&input_path)?;
            
            let retain_count = agents.len().saturating_sub(remove_count);
            let selected_agents = randomly_select_agents(&agents, retain_count);
            
            let output_path = scale_dir.join(&bin_name);
            let agents_only = selected_agents.iter()
                .filter_map(|agent| agent.split(':').next())
                .collect::<Vec<_>>();
            fs::write(&output_path, agents_only.join("\n"))?;
        }
    }

    Ok(())
}
