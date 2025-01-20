from dataclasses import dataclass
from typing import Dict, Optional
import pandas as pd
from shapely.geometry import Point

@dataclass
class Node:
    id: str
    x: float
    y: float
    coordinates: Point

@dataclass
class Link:
    id: str
    from_node: str
    to_node: str
    length: float
    freespeed: float
    capacity: float
    permlanes: float
    oneway: Optional[bool] = True
    modes: Optional[str] = "car"

@dataclass
class NetworkData:
    nodes: pd.DataFrame
    links: pd.DataFrame
    link_coordinates: pd.DataFrame

@dataclass
class ScaledNetwork:
    scale: float
    nodes: pd.DataFrame
    links: pd.DataFrame
    link_coordinates: pd.DataFrame
