from bs4 import BeautifulSoup
from pathlib import Path
import pandas as pd
from typing import Dict, List, Tuple
from .models import NetworkNode, NetworkLink, Agent, AgentLinkIndex
from .utils import create_point_from_coordinates, calculate_distance_to_link

def process_network_file(network_file: Path) -> Tuple[List[NetworkNode], List[NetworkLink]]:
    with open(network_file, 'r', encoding='utf-8') as f:
        soup = BeautifulSoup(f, 'lxml-xml')
    
    nodes = []
    node_dict = {}
    
    # Process nodes
    for node in soup.find_all('node'):
        x, y = float(node['x']), float(node['y'])
        network_node = NetworkNode(
            id=node['id'],
            x=x,
            y=y,
            coordinates=create_point_from_coordinates(x, y)
        )
        nodes.append(network_node)
        node_dict[node['id']] = network_node
    
    # Process links
    links = []
    for link in soup.find_all('link'):
        network_link = NetworkLink(
            id=link['id'],
            from_node=node_dict[link['from']],
            to_node=node_dict[link['to']],
            length=float(link.get('length', 0)),
            freespeed=float(link.get('freespeed', 0))
        )
        links.append(network_link)
    
    return nodes, links

def process_population_file(population_file: Path) -> List[Agent]:
    with open(population_file, 'r', encoding='utf-8') as f:
        soup = BeautifulSoup(f, 'lxml-xml')
    
    agents = []
    for person in soup.find_all('person'):
        activities = []
        plan = person.find('plan')
        
        for activity in plan.find_all('activity'):
            if 'x' in activity.attrs and 'y' in activity.attrs:
                point = create_point_from_coordinates(
                    float(activity['x']), 
                    float(activity['y'])
                )
                activities.append(point)
        
        if activities:
            agent = Agent(
                id=person['id'],
                location=activities[0],  # Use first activity location as agent location
                activities=activities
            )
            agents.append(agent)
    
    return agents

def create_agent_index(agents: List[Agent], links: List[NetworkLink]) -> Dict[str, List[str]]:
    link_agents = {}
    
    for link in links:
        link_agents[link.id] = []
        
        for agent in agents:
            # Calculate distance from agent's location to link
            distance = calculate_distance_to_link(
                agent.location,
                link.from_node.coordinates,
                link.to_node.coordinates
            )
            
            # If agent is close to link (using a threshold)
            if distance < 0.001:  # Adjust threshold as needed
                link_agents[link.id].append(agent.id)
    
    return link_agents

def create_index_for_scale(scale: int, network_dir: Path, population_dir: Path) -> AgentLinkIndex:
    network_file = network_dir / f"network-{scale:02d}.xml"
    population_file = population_dir / f"population-{scale:02d}.xml"
    
    nodes, links = process_network_file(network_file)
    agents = process_population_file(population_file)
    link_agents = create_agent_index(agents, links)
    
    return AgentLinkIndex(
        scale=scale,
        link_agents=link_agents
    )
