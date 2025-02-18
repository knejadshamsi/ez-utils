use anyhow::Result;
use std::collections::HashMap;
use std::fs::File;
use std::io::{BufRead, BufReader, Write};
use std::path::Path;
use crossterm::{
    cursor::{Hide, MoveToColumn},
    execute,
    style::{Color, Print, ResetColor, SetForegroundColor},
    terminal::{Clear, ClearType},
};
use crate::network::models::{NetworkPaths, LinkRatio};
use std::io::stdout;

use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct LinkAdjustment {
    pub speed_factor: f64,
    pub capacity_factor: f64,
}

const MIN_SCALE: u8 = 1;
const MAX_SCALE: u8 = 10;
const RATIO_THRESHOLD: f64 = 0.8;

fn read_flow_index(path: &Path) -> Result<HashMap<String, usize>> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let mut link_counts = HashMap::new();

    for line in reader.lines() {
        let line = line?;
        if let Some((link_id, agents_str)) = line.split_once(':') {
            let count = agents_str.trim().split(',').count();
            link_counts.insert(link_id.to_string(), count);
        }
    }

    Ok(link_counts)
}

fn print_current_state(step: usize, total_steps: usize, part1_done: bool, message: &str, current: Option<(usize, usize)>) {
    let mut stdout = stdout();
    
    let base = format!("{}[STEP {}/{}]{} ", 
        SetForegroundColor(Color::Magenta),
        step,
        total_steps,
        ResetColor
    );

    let content = if !part1_done {
        if let Some((current, total)) = current {
            format!("Calculating raw flow {}[{}/{}]{}", 
                SetForegroundColor(Color::Yellow),
                current,
                total,
                ResetColor
            )
        } else {
            "Calculating raw flow".to_string()
        }
    } else {
        let prefix = format!("{}Calculated{} raw flow {}→{} ", 
            SetForegroundColor(Color::Green),
            ResetColor,
            SetForegroundColor(Color::Cyan),
            ResetColor
        );

        if message.starts_with("Calculated") {
            format!("{}{}Calculated{} flow ratio",
                prefix,
                SetForegroundColor(Color::Green),
                ResetColor
            )
        } else if message.contains("[") {
            let (_text, count) = message.split_once('[')
                .expect("Invalid message format: missing [");
            format!("{}Calculating flow ratio {}[{}{}",
                prefix,
                SetForegroundColor(Color::Yellow),
                count,
                ResetColor
            )
        } else {
            format!("{}Calculating flow ratio",
                prefix
            )
        }
    };

    execute!(
        stdout,
        Hide,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!("{}{}", base, content))
    ).unwrap();
    stdout.flush().unwrap();
}

fn process_raw_flows(flow_index_dir: &Path) -> Result<HashMap<String, usize>> {
    let raw_path = flow_index_dir.join("flow-index-raw.idx");
    print_current_state(5, 8, false, "", Some((1, 1)));
    let result = read_flow_index(&raw_path)?;
    Ok(result)
}

fn process_scaled_flows(
    scale: u8,
    raw_flows: &HashMap<String, usize>,
    flow_index_dir: &Path,
    ratio_dir: &Path,
) -> Result<()> {
    let scale_path = flow_index_dir.join(format!("flow-index-{:02}.idx", scale));
    let scaled_flows = read_flow_index(&scale_path)?;
    let mut ratios = HashMap::new();
    
    for (link_id, raw_count) in raw_flows {
        if let Some(scaled_count) = scaled_flows.get(link_id) {
            if *raw_count == 0 || *scaled_count == 0 {
                continue;
            }

            let ratio = *scaled_count as f64 / *raw_count as f64;
            if ratio < RATIO_THRESHOLD || ratio >= 1.0 {
                ratios.insert(link_id.clone(), LinkRatio { ratio });
            }
        }
    }

    let output_path = ratio_dir.join(format!("ratio-{:02}.json", scale));
    let output_file = File::create(output_path)?;
    serde_json::to_writer_pretty(output_file, &ratios)?;

    Ok(())
}

pub fn calculate_network_adjustments(paths: &NetworkPaths) -> Result<()> {
    let flow_index_dir = paths.temp_dir.join("network").join("04_flow_index");
    let ratio_dir = paths.temp_dir.join("network").join("05_ratio_index");
    
    if !flow_index_dir.exists() {
        anyhow::bail!("Flow index directory not found: {}", flow_index_dir.display());
    }
    
    let raw_flows = process_raw_flows(&flow_index_dir)?;

    print_current_state(5, 8, true, "Calculating flow ratio [0/10]", None);
    
    for scale in MIN_SCALE..=MAX_SCALE {
        process_scaled_flows(scale, &raw_flows, &flow_index_dir, &ratio_dir)?;
        print_current_state(5, 8, true, &format!("Calculating flow ratio [{}/10]", scale), None);
    }

    print_current_state(5, 8, true, "Calculated", None);
    println!();
    
    Ok(())
}
