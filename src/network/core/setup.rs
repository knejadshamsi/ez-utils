use std::collections::HashMap;
use std::path::Path;
use std::fs;
use anyhow::Result;
use quick_xml::events::Event;
use quick_xml::Reader;

use super::models::{NetworkPaths, LinkFlow, NetworkData, ScaledNetwork, Node, Link, LinkCoordinates};

pub fn setup_processing_directories(base_path: &Path, path_manager: &crate::util::PathManager) -> Result<NetworkPaths> {
    let raw_chunks_dir = path_manager.get_network_path("02_traffic_index");
    let population_chunks_dir = path_manager.get_population_path("chunks");
    let network_index_dir = path_manager.get_network_path("03_scale_traffic_index");
    let split_dir = path_manager.get_network_path("04_flow_index");
    let ratio_dir = path_manager.get_network_path("05_ratio_index");
    let network_split_dir = path_manager.get_network_path("06_network");
    let scaled_network_dir = path_manager.get_network_path("07_scaled_network");

    for dir in [
        &raw_chunks_dir,
        &network_index_dir,
        &split_dir,
        &ratio_dir,
        &network_split_dir,
        &scaled_network_dir
    ] {
        if dir.exists() {
            std::fs::remove_dir_all(dir)?;
        }
    }

    path_manager.ensure_dir_exists(&raw_chunks_dir)?;
    path_manager.ensure_dir_exists(&split_dir)?;
    path_manager.ensure_dir_exists(&network_index_dir)?;
    path_manager.ensure_dir_exists(&ratio_dir)?;
    path_manager.ensure_dir_exists(&network_split_dir)?;
    path_manager.ensure_dir_exists(&scaled_network_dir)?;

    let output_dir = base_path.join("output").join("network");
    fs::create_dir_all(&output_dir)?;
    let temp_dir = path_manager.get_temp_dir().clone();

    Ok(NetworkPaths {
        raw_chunks_dir,
        population_chunks_dir,
        network_index_dir,
        temp_dir,
        output_dir,
        split_dir,
        network_split_dir,
    })
}

pub fn process_link_usage(chunks_dir: &Path) -> Result<HashMap<String, Vec<String>>> {
    let mut link_usage = HashMap::new();

    for entry in fs::read_dir(chunks_dir)? {
        let entry = entry?;
        let content = fs::read_to_string(entry.path())?;
        
        let mut reader = Reader::from_str(&content);
        reader.trim_text(true);
        let mut buf = Vec::new();

        let mut current_agent = None;
        let mut current_plan = None;

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(e)) => match e.name().as_ref() {
                    b"person" => {
                        if let Some(id) = e.attributes()
                            .find(|a| a.as_ref().unwrap().key.as_ref() == b"id")
                            .and_then(|a| a.ok())
                            .map(|a| String::from_utf8_lossy(&a.value).into_owned())
                        {
                            current_agent = Some(id);
                        }
                    },
                    b"plan" => {
                        if let Some(selected) = e.attributes()
                            .find(|a| a.as_ref().unwrap().key.as_ref() == b"selected")
                            .and_then(|a| a.ok())
                            .map(|a| String::from_utf8_lossy(&a.value).into_owned())
                        {
                            if selected == "yes" {
                                current_plan = Some(Vec::new());
                            }
                        }
                    },
                    b"leg" => {
                        if let Some(ref mut plan) = current_plan {
                            if let Some(route) = e.attributes()
                                .find(|a| a.as_ref().unwrap().key.as_ref() == b"route")
                                .and_then(|a| a.ok())
                                .map(|a| String::from_utf8_lossy(&a.value).into_owned())
                            {
                                for link_id in route.split(' ') {
                                    plan.push(link_id.to_string());
                                }
                            }
                        }
                    },
                    _ => {}
                },
                Ok(Event::End(e)) => match e.name().as_ref() {
                    b"person" => {
                        if let (Some(agent), Some(plan)) = (current_agent.take(), current_plan.take()) {
                            for link_id in plan {
                                link_usage.entry(link_id)
                                    .or_insert_with(Vec::new)
                                    .push(agent.clone());
                            }
                        }
                    },
                    _ => {}
                },
                Ok(Event::Eof) => break,
                Err(e) => return Err(e.into()),
                _ => {}
            }
            buf.clear();
        }
    }

    Ok(link_usage)
}

pub fn process_network_data(network_path: &Path) -> Result<NetworkData> {
    let content = fs::read_to_string(network_path)?;
    let mut reader = Reader::from_str(&content);
    reader.trim_text(true);
    let mut buf = Vec::new();

    let mut nodes = Vec::new();
    let mut links = Vec::new();
    let mut link_coordinates = Vec::new();

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => match e.name().as_ref() {
                b"node" => {
                    let mut id = String::new();
                    let mut x = 0.0;
                    let mut y = 0.0;

                    for attr in e.attributes() {
                        let attr = attr?;
                        match attr.key.as_ref() {
                            b"id" => id = String::from_utf8_lossy(&attr.value).into_owned(),
                            b"x" => x = String::from_utf8_lossy(&attr.value).parse()?,
                            b"y" => y = String::from_utf8_lossy(&attr.value).parse()?,
                            _ => {}
                        }
                    }

                    nodes.push(Node {
                        id: id.clone(),
                        x,
                        y,
                        coordinates: geo_types::Point::new(x, y),
                    });
                },
                b"link" => {
                    let mut link = Link {
                        id: String::new(),
                        from_node: String::new(),
                        to_node: String::new(),
                        length: 0.0,
                        freespeed: 0.0,
                        capacity: 0.0,
                        permlanes: 1.0,
                        oneway: None,
                        modes: None,
                    };

                    for attr in e.attributes() {
                        let attr = attr?;
                        match attr.key.as_ref() {
                            b"id" => link.id = String::from_utf8_lossy(&attr.value).into_owned(),
                            b"from" => link.from_node = String::from_utf8_lossy(&attr.value).into_owned(),
                            b"to" => link.to_node = String::from_utf8_lossy(&attr.value).into_owned(),
                            b"length" => link.length = String::from_utf8_lossy(&attr.value).parse()?,
                            b"freespeed" => link.freespeed = String::from_utf8_lossy(&attr.value).parse()?,
                            b"capacity" => link.capacity = String::from_utf8_lossy(&attr.value).parse()?,
                            b"permlanes" => link.permlanes = String::from_utf8_lossy(&attr.value).parse()?,
                            _ => {}
                        }
                    }

                    links.push(link);
                },
                _ => {}
            },
            Ok(Event::Eof) => break,
            Err(e) => return Err(e.into()),
            _ => {}
        }
        buf.clear();
    }

    for link in &links {
        if let (Some(from_node), Some(to_node)) = (
            nodes.iter().find(|n| n.id == link.from_node),
            nodes.iter().find(|n| n.id == link.to_node)
        ) {
            link_coordinates.push(LinkCoordinates {
                link_id: link.id.clone(),
                from_node: geo_types::Point::new(from_node.x, from_node.y),
                to_node: geo_types::Point::new(to_node.x, to_node.y),
                length: link.length,
                freespeed: link.freespeed,
                agent_usage: HashMap::new(),
            });
        }
    }

    Ok(NetworkData {
        nodes,
        links,
        link_coordinates,
    })
}

pub fn create_scaled_network(
    network_data: &NetworkData,
    flows: &[LinkFlow],
    scale: f64,
) -> Result<ScaledNetwork> {
    let flow_ratios: HashMap<_, _> = flows.iter()
        .map(|f| (f.link_id.clone(), f.flow_ratio))
        .collect();

    let scaled_links = network_data.links.iter()
        .map(|link| {
            let mut scaled_link = link.clone();
            if let Some(ratio) = flow_ratios.get(&link.id) {
                scaled_link.capacity *= ratio * scale;
            }
            scaled_link
        })
        .collect();

    Ok(ScaledNetwork {
        scale,
        nodes: network_data.nodes.clone(),
        links: scaled_links,
    })
}

pub fn analyze_traffic_flow(
    base_usage: &HashMap<String, Vec<String>>,
    scaled_usage: &HashMap<String, Vec<String>>,
) -> Vec<LinkFlow> {
    let mut flows = Vec::new();

    for (link_id, base_agents) in base_usage {
        let base_flow = base_agents.len();
        let scaled_flow = scaled_usage.get(link_id)
            .map(|agents| agents.len())
            .unwrap_or(0);

        let flow_ratio = if base_flow > 0 {
            scaled_flow as f64 / base_flow as f64
        } else {
            1.0
        };

        flows.push(LinkFlow {
            link_id: link_id.clone(),
            base_flow: base_flow as f64,
            scaled_flow: scaled_flow as f64,
            flow_ratio,
        });
    }

    flows
}
