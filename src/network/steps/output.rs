use std::fs::{self, File};
use std::io::Write;
use anyhow::Result;
use crossterm::style::{Color, SetForegroundColor, ResetColor};
use crate::network::models::NetworkPaths;
use crate::network::utils::display::{print_step_progress, print_step_without_counter};

pub fn get_raw_nodes_content(network_paths: &NetworkPaths) -> Result<String> {
    let raw_nodes_path = network_paths.temp_dir
        .join("network")
        .join("06_network")
        .join("raw-nodes.xml");
    Ok(fs::read_to_string(raw_nodes_path)?)
}

pub fn compose_scaled_networks(network_paths: &NetworkPaths) -> Result<()> {
    let raw_nodes_content = get_raw_nodes_content(network_paths)?;

    for (scale_idx, scale) in (1..=10).enumerate() {
        let scale_str = format!("{:02}", scale);
        
        if scale_idx == 9 {
            print_step_without_counter(8, 8, &format!("{}Output{} network nodes {}→{} {}Added{} network links",
                SetForegroundColor(Color::Green),
                ResetColor,
                SetForegroundColor(Color::Cyan),
                ResetColor,
                SetForegroundColor(Color::Green),
                ResetColor
            ));
        } else {
            print_step_progress(8, 8, scale_idx + 1, 10, 
                &format!("{}Output{} network nodes {}→{} Adding network links",
                    SetForegroundColor(Color::Green),
                    ResetColor,
                    SetForegroundColor(Color::Cyan),
                    ResetColor
                )
            );
        }

        let output_path = network_paths.output_dir.join(format!("network-{}.xml", scale_str));
        let mut file = File::create(&output_path)?;
        file.write_all(raw_nodes_content.as_bytes())?;

        let scaled_network_path = network_paths.temp_dir
            .join("network")
            .join("07_scaled_network")
            .join(format!("scaled-network-{}.xml", scale_str));
        let scaled_content = fs::read_to_string(scaled_network_path)?;
        file.write_all(scaled_content.as_bytes())?;
    }

    Ok(())
}
