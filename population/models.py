from dataclasses import dataclass
from typing import Dict, List, Optional
import pandas as pd
from shapely.geometry import Point

@dataclass
class Activity:
    id: str
    activity_order: int
    activity_type: str
    facility: Optional[str]
    start_time: Optional[str]
    end_time: Optional[str]
    coordinates: Optional[Point]

@dataclass
class Person:
    id: str
    activities: List[Activity]

@dataclass
class PopulationData:
    persons: List[Person]
    activities_df: pd.DataFrame
    total_population: int
    current_scale: float

@dataclass
class ScaledPopulation:
    scale: float
    persons: List[Person]
    activities_df: pd.DataFrame
