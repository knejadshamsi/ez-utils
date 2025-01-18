from dataclasses import dataclass
from typing import Dict, List, Optional
from shapely.geometry import Point

@dataclass
class NetworkNode:
    id: str
    x: float
    y: float
    coordinates: Point

@dataclass
class NetworkLink:
    id: str
    from_node: NetworkNode
    to_node: NetworkNode
    length: float
    freespeed: float

@dataclass
class Agent:
    id: str
    location: Point
    activities: List[Point]

@dataclass
class LinkAgentIndex:
    link_id: str
    agent_ids: List[str]
    distance_to_link: float

@dataclass
class AgentLinkIndex:
    scale: int
    link_agents: Dict[str, List[str]]  # link_id -> list of agent_ids
