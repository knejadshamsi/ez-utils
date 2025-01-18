from dataclasses import dataclass
from typing import Dict, List, Optional
import pandas as pd

@dataclass
class GTFSData:
    routes: pd.DataFrame
    trips: pd.DataFrame
    stop_times: pd.DataFrame
    stops: pd.DataFrame
    calendar: pd.DataFrame
    calendar_dates: pd.DataFrame

@dataclass
class TransitSchedule:
    stops: Dict[str, dict]
    routes: Dict[str, dict]
    trips: Dict[str, dict]

@dataclass
class VehicleDefinition:
    id: str
    type: str
    capacity: dict
    length: float
    width: float
    max_speed: float
