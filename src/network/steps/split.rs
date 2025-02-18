use anyhow::Result;
use std::fs::{self, File};
use std::io::{BufRead, BufReader, Write};
use std::path::Path;
use crate::network::utils::display::{self, SplitState};

#[derive(Debug)]
struct NetworkSeparators {
    nodes_start: usize,
    links_boundary: usize,
    links_end: usize,
}

pub fn split_network_file(
    input_file: &Path,
    output_dir: &Path,
    chunk_size: usize,
) -> Result<()> {
    fs::create_dir_all(output_dir)?;
    let chunks_dir = output_dir.join("chunks");
    fs::create_dir_all(&chunks_dir)?;

    let separators = find_separators(input_file)?;
    display::print_split_progress(6, 8, SplitState::FoundSeparators, None);

    let file = File::open(input_file)?;
    let reader = BufReader::new(file);
    let lines: Vec<String> = reader.lines().collect::<Result<_, _>>()?;

    let nodes_section = format!(
        "{}\n{}\n{}\n",
        "<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
        "<network>",
        lines[separators.nodes_start..separators.links_boundary]
            .join("\n")
    );
    fs::write(output_dir.join("raw-nodes.xml"), nodes_section)?;

    let links_section = format!(
        "{}\n{}\n{}\n{}\n",
        "<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
        "<network>",
        lines[separators.links_boundary..separators.links_end].join("\n"),
        "</network>"
    );
    fs::write(output_dir.join("raw-links.xml"), &links_section)?;
    display::print_split_progress(6, 8, SplitState::Splitting(false), None);

    let links: Vec<_> = links_section
        .split("</link>")
        .filter(|s| s.contains("<link "))
        .map(|s| format!("{}</link>", s))
        .collect();

    let total_chunks = (links.len() + chunk_size - 1) / chunk_size;
    for (chunk_index, chunk) in links.chunks(chunk_size).enumerate() {
        display::print_split_progress(6, 8, SplitState::Chunking, Some((chunk_index + 1, total_chunks)));

        let chunk_content = format!(
            "{}\n{}\n{}\n{}\n",
            "<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
            "<links>",
            chunk.join("\n"),
            "</links>"
        );

        let chunk_path = chunks_dir.join(format!("chunk_{:04}.xml", chunk_index + 1));
        let mut file = fs::File::create(chunk_path)?;
        file.write_all(chunk_content.as_bytes())?;
    }

    display::print_split_progress(6, 8, SplitState::Complete, None);
    println!();

    Ok(())
}

fn find_separators(input_file: &Path) -> Result<NetworkSeparators> {
    let file = File::open(input_file)?;
    let reader = BufReader::new(file);
    let mut separator_count = 0;
    let mut separators = NetworkSeparators {
        nodes_start: 0,
        links_boundary: 0,
        links_end: 0,
    };

    for (idx, line) in reader.lines().enumerate() {
        let line = line?;
        if line.contains("<!--") {
            display::print_split_progress(6, 8, SplitState::Finding, Some((separator_count + 1, 3)));
            separator_count += 1;
            match separator_count {
                1 => separators.nodes_start = idx,
                2 => separators.links_boundary = idx,
                3 => {
                    separators.links_end = idx;
                    break;
                }
                _ => {}
            }
        }
    }

    if separator_count != 3 {
        anyhow::bail!("Network file must contain exactly 3 separator comments");
    }

    Ok(separators)
}
