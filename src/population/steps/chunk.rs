use anyhow::Result;
use std::fs;
use std::path::Path;
use crate::population::core::models::*;

pub fn fix_first_chunk(chunks_dir: &Path) -> Result<()> {
    let first_chunk_path = chunks_dir.join("chunk_0001.xml");
    let content = fs::read_to_string(&first_chunk_path)?;
    
    let lines: Vec<_> = content.lines().collect();
    let mut xml_ver_idx = None;
    let mut doctype_idx = None;
    let mut pop_start_idx = None;
    
    for (idx, line) in lines.iter().enumerate() {
        let trimmed = line.trim();
        if trimmed.starts_with("<?xml") && xml_ver_idx.is_none() {
            xml_ver_idx = Some(idx);
        } else if trimmed.starts_with("<!DOCTYPE") && doctype_idx.is_none() {
            doctype_idx = Some(idx);
        } else if trimmed.starts_with("<population") && pop_start_idx.is_none() {
            pop_start_idx = Some(idx);
        }
    }
    
    let mut new_content = String::new();
    new_content.push_str("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n");
    new_content.push_str("<!DOCTYPE population SYSTEM \"http://www.matsim.org/files/dtd/population_v6.dtd\">\n");
    new_content.push_str("<population desc=\"Switzerland Baseline\">\n");
    
    let start_idx = pop_start_idx.map(|idx| idx + 1).unwrap_or(0);
    for line in &lines[start_idx..] {
        let trimmed = line.trim();
        if trimmed.is_empty() || trimmed.starts_with("<?xml") || 
           trimmed.starts_with("<!DOCTYPE") || trimmed.starts_with("<population") || 
           trimmed == "</population>" {
            continue;
        }
        new_content.push_str(line);
        new_content.push('\n');
    }
    new_content.push_str("</population>\n");
    
    fs::write(first_chunk_path, new_content.trim())?;
    Ok(())
}

pub fn fix_last_chunk(chunks_dir: &Path, chunk_num: usize) -> Result<()> {
    let last_chunk_path = chunks_dir.join(format!("chunk_{:04}.xml", chunk_num));
    let content = fs::read_to_string(&last_chunk_path)?;
    
    let mut new_content = String::new();
    let mut header_done = false;
    let mut pop_end_seen = false;
    
    for line in content.lines() {
        let trimmed = line.trim();
        
        if !header_done {
            new_content.push_str("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n");
            new_content.push_str("<!DOCTYPE population SYSTEM \"http://www.matsim.org/files/dtd/population_v6.dtd\">\n");
            new_content.push_str("<population desc=\"Switzerland Baseline\">\n");
            header_done = true;
        }
        
        if trimmed.is_empty() || trimmed.starts_with("<?xml") || 
           trimmed.starts_with("<!DOCTYPE") || trimmed.starts_with("<population") {
            continue;
        }
        
        if trimmed == "</population>" {
            if !pop_end_seen {
                pop_end_seen = true;
            }
            continue;
        }
        
        new_content.push_str(line);
        new_content.push('\n');
    }
    
    new_content.push_str("</population>\n");
    fs::write(last_chunk_path, new_content.trim())?;
    Ok(())
}

pub fn save_chunk(chunk: &ChunkData, chunks_dir: &Path) -> Result<()> {
    let chunk_path = chunks_dir.join(format!("chunk_{:04}.xml", chunk.chunk_number));
    let mut full_content = String::new();
    
    full_content.push_str("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n");
    full_content.push_str("<!DOCTYPE population SYSTEM \"http://www.matsim.org/files/dtd/population_v6.dtd\">\n");
    full_content.push_str("<population desc=\"Switzerland Baseline\">\n");
    
    for line in chunk.content.lines() {
        let trimmed = line.trim();
        if trimmed.is_empty() || trimmed.starts_with("<?xml") ||
           trimmed.starts_with("<!DOCTYPE") || trimmed.starts_with("<population") ||
           trimmed == "</population>" {
            continue;
        }
        full_content.push_str(line);
        full_content.push('\n');
    }
    
    full_content.push_str("</population>\n");
    fs::write(chunk_path, full_content)?;
    Ok(())
}
