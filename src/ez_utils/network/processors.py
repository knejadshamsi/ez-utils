from bs4 import BeautifulSoup
import pandas as pd
import os
from pathlib import Path
from typing import Dict, Tuple, List, Union
from .models import NetworkData, ScaledNetwork, Node, Link
from .utils import create_point_from_coordinates
from .base_processor import BaseProcessor

def validate_scales(current_scale: float, requested_scales: Union[List[float], str, None] = None) -> List[float]:
    if requested_scales is None:
        scales = list(range(1, 11))
    else:
        # Handle comma-separated string input
        if isinstance(requested_scales, str):
            try:
                requested_scales = [float(s.strip()) for s in requested_scales.split(',')]
            except ValueError:
                raise ValueError("Invalid scale format. Use comma-separated numbers between 1 and 10")
        
        scales = [float(s) for s in requested_scales if 1 <= float(s) <= 10]
        if not scales:
            raise ValueError("No valid scales provided. Scales must be between 1 and 10.")
    
    valid_scales = [s for s in scales if s <= current_scale]
    if not valid_scales:
        raise ValueError(f"No valid scales below current scale ({current_scale}%). This module only supports scaling down.")
    return sorted(valid_scales)

def validate_network_xml(soup: BeautifulSoup) -> None:
    if not soup.find('network'):
        raise ValueError("Invalid network XML: Missing 'network' root element")
    if not soup.find_all('node'):
        raise ValueError("Invalid network XML: No nodes found")
    
    links = soup.find_all('link')
    if not links:
        raise ValueError("Invalid network XML: No links found")
        
    for link in links:
        if 'freespeed' not in link.attrs:
            raise ValueError(f"Invalid link: Missing freespeed attribute for link {link.get('id', 'unknown')}")
        try:
            float(link['freespeed'])
        except ValueError:
            raise ValueError(f"Invalid link: freespeed must be a number for link {link.get('id', 'unknown')}")

def process_events(events_file: Path) -> Dict[str, Dict]:
    if not events_file or not events_file.exists():
        return {}
        
    traffic_data = {}
    with open(events_file, 'r') as f:
        soup = BeautifulSoup(f, 'lxml-xml')
    
    for event in soup.find_all('event'):
        if event['type'] == 'entered link':
            link_id = event['link']
            time = float(event['time'])
            
            if link_id not in traffic_data:
                traffic_data[link_id] = {'count': 0, 'times': []}
            
            traffic_data[link_id]['count'] += 1
            traffic_data[link_id]['times'].append(time)
    
    return traffic_data

def calculate_adjusted_speed(original_speed: float, traffic_count: int, scale_factor: float) -> float:
    if traffic_count == 0:
        return original_speed
        
    density_factor = (traffic_count * scale_factor) / traffic_count
    if density_factor >= 1:
        return original_speed
        
    min_speed = original_speed * 0.2
    adjusted_speed = original_speed * (0.2 + (0.8 * density_factor))
    return max(min_speed, adjusted_speed)

def process_nodes(nodes_soup: BeautifulSoup) -> pd.DataFrame:
    nodes_list = []
    for node in nodes_soup:
        if 'x' not in node.attrs or 'y' not in node.attrs:
            raise ValueError(f"Invalid node: Missing coordinates for node {node.get('id', 'unknown')}")
        x, y = float(node['x']), float(node['y'])
        nodes_list.append({
            'id': node['id'],
            'x': x,
            'y': y,
            'coordinates': create_point_from_coordinates(x, y)
        })
    return pd.DataFrame(nodes_list)

def process_links(links_soup: BeautifulSoup, nodes_df: pd.DataFrame) -> Tuple[pd.DataFrame, pd.DataFrame]:
    links_list = []
    coords_list = []
    
    for link in links_soup:
        if 'id' not in link.attrs or 'from' not in link.attrs or 'to' not in link.attrs:
            raise ValueError(f"Invalid link: Missing required attributes")
            
        link_attrs = dict(link.attrs)
        links_list.append(link_attrs)
        
        try:
            from_node = nodes_df[nodes_df['id'] == link_attrs['from']].iloc[0]
            to_node = nodes_df[nodes_df['id'] == link_attrs['to']].iloc[0]
        except IndexError:
            raise ValueError(f"Invalid link: Referenced nodes not found for link {link_attrs['id']}")
        
        coords_dict = {
            'link_id': link_attrs['id'],
            'from_node': from_node['coordinates'],
            'to_node': to_node['coordinates'],
            'length': float(link_attrs.get('length', 0)),
            'freespeed': float(link_attrs['freespeed'])
        }
        coords_list.append(coords_dict)
    
    links_df = pd.DataFrame(links_list)
    coords_df = pd.DataFrame(coords_list)
    
    return links_df, coords_df

def process_network(network_soup: BeautifulSoup) -> NetworkData:
    validate_network_xml(network_soup)
    nodes_df = process_nodes(network_soup.find_all('node'))
    links_df, coords_df = process_links(network_soup.find_all('link'), nodes_df)
    
    return NetworkData(
        nodes=nodes_df,
        links=links_df,
        link_coordinates=coords_df
    )

def create_scaled_network(network_data: NetworkData, scale: float, traffic_data: Dict[str, Dict] = None) -> ScaledNetwork:
    scaled_links = network_data.links.copy()
    scale_factor = scale / 100
    
    if traffic_data:
        for idx, link in scaled_links.iterrows():
            link_traffic = traffic_data.get(link['id'], {'count': 0})
            original_speed = float(link['freespeed'])
            scaled_links.at[idx, 'freespeed'] = str(
                calculate_adjusted_speed(original_speed, link_traffic['count'], scale_factor)
            )
    
    return ScaledNetwork(
        scale=scale,
        nodes=network_data.nodes.copy(),
        links=scaled_links,
        link_coordinates=network_data.link_coordinates.copy()
    )

def network_to_xml(network: ScaledNetwork) -> str:
    soup = BeautifulSoup('<network></network>', 'lxml-xml')
    root = soup.find('network')
    
    for _, node in network.nodes.iterrows():
        node_tag = soup.new_tag('node')
        node_tag['id'] = str(node['id'])
        node_tag['x'] = str(node['x'])
        node_tag['y'] = str(node['y'])
        root.append(node_tag)
    
    for _, link in network.links.iterrows():
        link_tag = soup.new_tag('link')
        for col in link.index:
            if pd.notna(link[col]):
                link_tag[col] = str(link[col])
        root.append(link_tag)
    
    return str(soup)

class NetworkProcessor(BaseProcessor):
    def __init__(self, events_file: Path = None, scales: Union[List[float], str, None] = None):
        super().__init__()
        self.events_file = events_file
        self.scales = scales
        self.traffic_data = None
        if events_file:
            self.traffic_data = process_events(events_file)

    def process_chunk(self, chunk_file: Path) -> List[Path]:
        with open(chunk_file, 'r') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
        
        network_data = process_network(soup)
        current_scale = 100
        valid_scales = validate_scales(current_scale, self.scales)
        
        output_chunks = []
        for scale in valid_scales:
            output_chunk = self.temp_dir / f"network-{int(scale):02d}.xml"
            scaled_network = create_scaled_network(network_data, scale, self.traffic_data)
            output_xml = network_to_xml(scaled_network)
            
            with open(output_chunk, 'w') as f:
                f.write(output_xml)
            output_chunks.append(output_chunk)
        
        return output_chunks

def process_file(input_file: Path, output_dir: Path, events_file: Path = None, scales: Union[List[float], str, None] = None):
    if not input_file.exists():
        raise FileNotFoundError(f"Input file not found: {input_file}")
    
    # Ensure network subdirectory exists
    network_dir = output_dir / "network"
    network_dir.mkdir(parents=True, exist_ok=True)
    
    processor = NetworkProcessor(events_file, scales)
    processor.process_file(input_file, network_dir)
