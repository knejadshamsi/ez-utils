from bs4 import BeautifulSoup
import pandas as pd
from pathlib import Path
from typing import Dict, Tuple, List
from .models import NetworkData, ScaledNetwork, Node, Link
from .utils import create_point_from_coordinates
from .base_processor import BaseProcessor

def process_nodes(nodes_soup: BeautifulSoup) -> pd.DataFrame:
    nodes_list = []
    for node in nodes_soup:
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
        link_attrs = dict(link.attrs)
        links_list.append(link_attrs)
        
        from_node = nodes_df[nodes_df['id'] == link_attrs['from']].iloc[0]
        to_node = nodes_df[nodes_df['id'] == link_attrs['to']].iloc[0]
        
        coords_dict = {
            'link_id': link_attrs['id'],
            'from_node': from_node['coordinates'],
            'to_node': to_node['coordinates'],
            'length': float(link_attrs.get('length', 0)),
            'freespeed': float(link_attrs.get('freespeed', 0))
        }
        coords_list.append(coords_dict)
    
    links_df = pd.DataFrame(links_list)
    coords_df = pd.DataFrame(coords_list)
    
    return links_df, coords_df

def process_network(network_soup: BeautifulSoup) -> NetworkData:
    nodes_df = process_nodes(network_soup.find_all('node'))
    links_df, coords_df = process_links(network_soup.find_all('link'), nodes_df)
    
    return NetworkData(
        nodes=nodes_df,
        links=links_df,
        link_coordinates=coords_df
    )

def create_scaled_network(network_data: NetworkData, scale: float) -> ScaledNetwork:
    scaled_links = network_data.links.copy()
    
    # Scale capacities and speeds
    scaled_links['capacity'] = scaled_links['capacity'].astype(float) * (scale / 100)
    
    # Adjust speeds for heavily scaled down networks
    if scale < 50:
        scaled_links['freespeed'] = scaled_links['freespeed'].astype(float) * 0.8
    
    return ScaledNetwork(
        scale=scale,
        nodes=network_data.nodes.copy(),
        links=scaled_links,
        link_coordinates=network_data.link_coordinates.copy()
    )

def network_to_xml(network: ScaledNetwork) -> str:
    soup = BeautifulSoup('<network></network>', 'lxml-xml')
    root = soup.find('network')
    
    # Add nodes
    for _, node in network.nodes.iterrows():
        node_tag = soup.new_tag('node')
        node_tag['id'] = str(node['id'])
        node_tag['x'] = str(node['x'])
        node_tag['y'] = str(node['y'])
        root.append(node_tag)
    
    # Add links
    for _, link in network.links.iterrows():
        link_tag = soup.new_tag('link')
        for col in link.index:
            if pd.notna(link[col]):
                link_tag[col] = str(link[col])
        root.append(link_tag)
    
    return str(soup)

class NetworkProcessor(BaseProcessor):
    def process_chunk(self, chunk_file: Path) -> Path:
        output_chunk = self.temp_dir / f"output_{chunk_file.name}"
        
        with open(chunk_file, 'r') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
        
        network_data = process_network(soup)
        scaled_network = create_scaled_network(network_data, scale=100)  # Default no scaling
        output_xml = network_to_xml(scaled_network)
        
        with open(output_chunk, 'w') as f:
            f.write(output_xml)
        
        return output_chunk

def process_file(input_file: Path, output_file: Path):
    processor = NetworkProcessor()
    processor.process_file(input_file, output_file)
