use anyhow::{Context, Result};
use std::fs::File;
use std::io::{BufRead, BufReader};
use std::path::Path;
use crate::population::core::models::*;
use super::chunk;
use crate::population::utils::{file, display};
use crate::population::utils::display::ProcessProgress;

pub async fn process_population_file(
    input_path: &Path,
    paths: &ProcessingPaths,
    scale_config: &ScaleConfig,
) -> Result<()> {
    let file = File::open(input_path)?;
    let reader = BufReader::new(file);
    let total_steps = 6;

    // Step 1: Breaking population file into chunks
    let mut current_chunk = ChunkData {
        chunk_id: 0,
        agents: Vec::new(),
        content: String::new(),
        separator_comment_count: 0,
        chunk_number: 1,
    };

    let mut total_separators = 0;
    display::print_step(1, total_steps, current_chunk.chunk_number, total_separators);
    
    for line in reader.lines() {
        let line = line?;
        
        if line.contains("<!--") {
            current_chunk.separator_comment_count += 1;
            total_separators += 1;
            display::print_step(1, total_steps, current_chunk.chunk_number, total_separators);
            
            if current_chunk.separator_comment_count == 2000 {
                chunk::save_chunk(&current_chunk, &paths.chunks_dir)?;
                current_chunk.content.clear();
                current_chunk.chunk_number += 1;
                current_chunk.separator_comment_count = 0;
            }
        } else {
            current_chunk.content.push_str(&line);
            current_chunk.content.push('\n');
        }
    }

    if !current_chunk.content.is_empty() {
        chunk::save_chunk(&current_chunk, &paths.chunks_dir)?;
    }

    // Fix first and last chunks
    chunk::fix_first_chunk(&paths.chunks_dir)?;
    chunk::fix_last_chunk(&paths.chunks_dir, current_chunk.chunk_number)?;

    display::step_success(1, total_steps, current_chunk.chunk_number, total_separators - 1);

    // Step 2: Creating indexes
    let mut progress = ProcessProgress::new(2, total_steps, current_chunk.chunk_number as u64)?;
    for chunk_num in 1..=current_chunk.chunk_number {
        let chunk_path = paths.chunks_dir.join(format!("chunk_{:04}.xml", chunk_num));
        let content = std::fs::read_to_string(&chunk_path)
            .with_context(|| format!("Failed to read chunk file: {}", chunk_path.display()))?;
        
        let chunk = ChunkData {
            chunk_id: chunk_num - 1,
            agents: Vec::new(),
            content,
            separator_comment_count: if chunk_num == current_chunk.chunk_number { 
                current_chunk.separator_comment_count 
            } else { 
                2000 
            },
            chunk_number: chunk_num,
        };

        super::index::process_chunk(&chunk, &paths.index_dir)
            .with_context(|| format!("Failed to process chunk {}", chunk_num))?;
        progress.update_progress(chunk_num as u64, current_chunk.chunk_number as u64, "Processing Chunks");
    }
    progress.finish();

    // Verify indexes were created
    let index_count = std::fs::read_dir(&paths.index_dir)?
        .filter_map(|e| e.ok())
        .filter(|e| e.path().extension().map_or(false, |ext| ext == "idx"))
        .count();

    if index_count == 0 {
        println!("[ERROR] No index files were created");
        return Err(anyhow::anyhow!("Check the chunk processing logs for errors"));
    }

    // Step 3: Creating zones
    let mut progress = ProcessProgress::new(3, total_steps, 1)?;
    progress.print_step_3_establishing();
    let mut _persons = file::load_persons_from_indexes(&paths.index_dir)?;
    let total_population = total_separators - 1;
    progress.print_step_3_established();
    _persons = super::zone::process_step_3(_persons, paths)?;
    progress.print_step_3_complete();
    progress.finish();

    // Step 4: Assigning agents to bins
    let mut progress = ProcessProgress::new(4, total_steps, 1)?;
    progress.print_step_4_progress();
    
    let bins_path = paths.temp_dir.join("population/03_create_area_bins/bins.json");
    if !bins_path.exists() {
        anyhow::bail!("Bins file not found at {}", bins_path.display());
    }

    super::assignment::assign_locations_to_bins(
        &paths.index_dir,
        &bins_path,
        &paths.assigned_dir,
    )?;

    progress.print_step_4_complete();
    progress.finish();

    let population_data = PopulationData {
        persons: _persons,
        current_scale: 100.0,
        total_population,
        chunk_count: current_chunk.chunk_number,
        last_chunk_size: total_separators % 2000,
        total_separator_comments: total_separators,
    };

    // Step 5: Scaling down
    let mut progress = ProcessProgress::new(5, total_steps, 1)?;
    progress.print_step_5_calculating();
    
    if scale_config.use_db {
        progress.print_step_5_calculated();
        crate::population::utils::db::process_for_database(paths, scale_config, &population_data).await?;
        progress.print_step_5_scaled();
    } else {
        let _persons = file::load_persons_from_indexes(&paths.index_dir)?;
        progress.print_step_5_calculated();
        super::scale::process_for_files(paths, scale_config, &population_data)?;
        progress.print_step_5_scaled();
    }
    
    progress.print_step_5_complete();
    progress.finish();

    // Step 6: Creating output
    let mut progress = ProcessProgress::new(6, total_steps, 1)?;
    super::output::process_step_6(paths, &population_data, &mut progress)?;
    progress.finish();

    if scale_config.delete_temp {
        file::cleanup_temp_files(paths)?;
    }

    Ok(())
}
