use anyhow::Result;
use std::fs;
use std::collections::HashMap;
use geo::Point;
use crate::population::core::models::*;
use crate::population::core::config::Database;
use crate::population::steps::assignment::calculate_location_density;

pub async fn process_for_database(paths: &ProcessingPaths, scale_config: &ScaleConfig, population_data: &PopulationData) -> Result<()> {
    println!("Initializing database connection...");
    let db = Database::new().await?;
    db.ensure_tables_exist().await?;

    println!("Loading agents from index files...");
    let pb = indicatif::ProgressBar::new_spinner();
    pb.set_style(
        indicatif::ProgressStyle::default_spinner()
            .template("{spinner:.green} Loading indexes {msg}...")?
            .tick_chars("⠁⠂⠄⡀⢀⠠⠐⠈")
    );

    let mut all_agents = Vec::new();
    let index_files = fs::read_dir(&paths.index_dir)?;
    
    for entry in index_files {
        let entry = entry?;
        pb.set_message(entry.file_name().to_string_lossy().to_string());
        let content = fs::read_to_string(entry.path())?;
        let index_entries: HashMap<String, Point<f64>> = serde_json::from_str(&content)?;
        let mut agents = Vec::new();
        for (id, coordinates) in index_entries {
            agents.push(AgentIndex {
                agent_id: id,
                coordinates,
                xml_content: String::new(),
            });
        }
        all_agents.extend(agents);
    }

    pb.finish_with_message(format!("Loaded {} agents", all_agents.len()));

    println!("\nCalculating location density...");
    // Use the total population data for density calculations
    if population_data.total_population > all_agents.len() {
        println!("Warning: Processing subset of total population ({} of {})", 
            all_agents.len(), population_data.total_population);
    }

    // Parse activities and coordinates for density calculation
    let mut persons = Vec::with_capacity(all_agents.len());
    for agent in &all_agents {
        // Use simplified parser for faster density calculation
        if let Ok(activities) = crate::population::steps::scale::parse_basic_activities(&agent.xml_content) {
            persons.push(Person {
                id: agent.agent_id.clone(),
                activities,
            });
        }
    }

    let location_density = calculate_location_density(&persons);
    println!("Identified {} unique areas with proper density distribution", location_density.len());

    println!("\nCalculating agent scales...");
    let agent_scales = crate::population::steps::scale::calculate_agent_scales(
        100.0,
        scale_config.target_scale,
        all_agents.len()
    )?;

    // Store density info
    println!("\nSaving density information...");
    let scale_str = format!("{:02}", scale_config.target_scale as i32);
    let dense_file = paths.dense_dir.join(format!("density-{}.json", scale_str));
    let density_data = serde_json::to_string_pretty(&location_density)?;
    fs::write(dense_file, density_data)?;

    // Insert agents into database
    println!("\nInserting agents into database...");
    let pb = indicatif::ProgressBar::new(all_agents.len() as u64);
    pb.set_style(
        indicatif::ProgressStyle::default_bar()
            .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {pos}/{len} agents ({msg})")?
            .progress_chars("#>-")
    );

    for (agent, scale) in all_agents.into_iter().zip(agent_scales.iter()) {
        let scales = vec![*scale];
        let point = format!(
            "ST_SetSRID(ST_MakePoint({}, {}), 4326)",
            agent.coordinates.x(),
            agent.coordinates.y()
        );

        db.insert_agent(
            &agent.agent_id,
            &point,
            &scales.iter().map(|&s| s as i32).collect::<Vec<_>>(),
            &agent.xml_content,
        ).await?;

        pb.inc(1);
        if pb.position() % 1000 == 0 {
            pb.set_message(format!("{}% complete", (pb.position() * 100) / pb.length().unwrap_or(1)));
        }
    }

    pb.finish_with_message("All agents inserted into database");

    Ok(())
}
