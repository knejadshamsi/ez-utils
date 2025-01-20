from dataclasses import dataclass
from typing import Dict, List, Optional, TypedDict, Literal
import pandas as pd
import re

class VehicleCapacity(TypedDict):
    seats: int
    standingRoom: int

class StopReference(TypedDict):
    refId: str
    arrivalOffset: str
    departureOffset: str

class DepartureReference(TypedDict):
    id: str
    departureTime: str
    vehicleRefId: str

TransportMode = Literal['bus', 'metro']

@dataclass
class GTFSData:
    routes: pd.DataFrame
    trips: pd.DataFrame
    stop_times: pd.DataFrame
    stops: pd.DataFrame
    calendar: pd.DataFrame
    calendar_dates: pd.DataFrame
    
    def __post_init__(self):
        # Validate DataFrame types
        if not all(isinstance(df, pd.DataFrame) for df in [
            self.routes, self.trips, self.stop_times,
            self.stops, self.calendar, self.calendar_dates
        ]):
            raise TypeError("All GTFS data must be pandas DataFrames")
        
        # Validate required columns
        required_columns = {
            'routes': ['route_id', 'route_type'],
            'trips': ['trip_id', 'route_id', 'service_id'],
            'stop_times': ['trip_id', 'stop_id', 'stop_sequence', 'arrival_time', 'departure_time'],
            'stops': ['stop_id', 'stop_lat', 'stop_lon'],
            'calendar': ['service_id', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
            'calendar_dates': ['service_id', 'date', 'exception_type']
        }
        
        for name, df in [
            ('routes', self.routes),
            ('trips', self.trips),
            ('stop_times', self.stop_times),
            ('stops', self.stops),
            ('calendar', self.calendar),
            ('calendar_dates', self.calendar_dates)
        ]:
            missing = [col for col in required_columns[name] if col not in df.columns]
            if missing:
                raise ValueError(f"Missing required columns in {name}: {missing}")

@dataclass
class TransitStop:
    id: str
    x: float
    y: float
    linkRefId: str = "dummy"
    
    def __post_init__(self):
        if not isinstance(self.id, str):
            raise TypeError("Stop ID must be a string")
        if not isinstance(self.x, (int, float)):
            raise TypeError("X coordinate must be a number")
        if not isinstance(self.y, (int, float)):
            raise TypeError("Y coordinate must be a number")
        if not isinstance(self.linkRefId, str):
            raise TypeError("Link reference ID must be a string")

@dataclass
class TransitRoute:
    id: str
    mode: TransportMode
    stops: List[StopReference]
    departures: List[DepartureReference]
    
    def __post_init__(self):
        if not isinstance(self.id, str):
            raise TypeError("Route ID must be a string")
        if self.mode not in ('bus', 'metro'):
            raise ValueError("Mode must be 'bus' or 'metro'")
        
        time_pattern = re.compile(r'^\d{2}:\d{2}:\d{2}$')
        
        # Validate stops
        for stop in self.stops:
            if not isinstance(stop, dict):
                raise TypeError("Stop must be a dictionary")
            if not all(k in stop for k in ('refId', 'arrivalOffset', 'departureOffset')):
                raise ValueError("Stop missing required fields")
            if not isinstance(stop['refId'], str):
                raise TypeError("Stop refId must be a string")
            if not time_pattern.match(stop['arrivalOffset']):
                raise ValueError(f"Invalid arrival offset format: {stop['arrivalOffset']}")
            if not time_pattern.match(stop['departureOffset']):
                raise ValueError(f"Invalid departure offset format: {stop['departureOffset']}")
        
        # Validate departures
        for departure in self.departures:
            if not isinstance(departure, dict):
                raise TypeError("Departure must be a dictionary")
            if not all(k in departure for k in ('id', 'departureTime', 'vehicleRefId')):
                raise ValueError("Departure missing required fields")
            if not isinstance(departure['id'], str):
                raise TypeError("Departure ID must be a string")
            if not time_pattern.match(departure['departureTime']):
                raise ValueError(f"Invalid departure time format: {departure['departureTime']}")
            if not isinstance(departure['vehicleRefId'], str):
                raise TypeError("Vehicle reference ID must be a string")

@dataclass
class TransitSchedule:
    stops: Dict[str, TransitStop]
    routes: Dict[str, TransitRoute]
    
    def __post_init__(self):
        if not isinstance(self.stops, dict):
            raise TypeError("Stops must be a dictionary")
        if not isinstance(self.routes, dict):
            raise TypeError("Routes must be a dictionary")
        
        # Validate stop references
        stop_ids = set(self.stops.keys())
        for route in self.routes.values():
            for stop in route.stops:
                if stop['refId'] not in stop_ids:
                    raise ValueError(f"Invalid stop reference: {stop['refId']}")

@dataclass
class VehicleDefinition:
    id: str
    type: TransportMode
    capacity: VehicleCapacity
    length: float
    width: float
    max_speed: float
    
    def __post_init__(self):
        if not isinstance(self.id, str):
            raise TypeError("Vehicle ID must be a string")
        
        # Validate capacity
        if not isinstance(self.capacity, dict) or \
           'seats' not in self.capacity or \
           'standingRoom' not in self.capacity:
            raise ValueError("Capacity must be a dict with 'seats' and 'standingRoom'")
        
        # Validate numeric fields
        if not all(isinstance(v, (int, float)) for v in [
            self.capacity['seats'],
            self.capacity['standingRoom'],
            self.length,
            self.width,
            self.max_speed
        ]):
            raise ValueError("Numeric fields must be numbers")
        
        # Validate positive values
        if not all(v > 0 for v in [
            self.capacity['seats'],
            self.capacity['standingRoom'],
            self.length,
            self.width,
            self.max_speed
        ]):
            raise ValueError("All numeric values must be positive")
        
        # Validate vehicle type
        if self.type not in ('bus', 'metro'):
            raise ValueError("Vehicle type must be 'bus' or 'metro'")
        
        # Validate capacity ranges
        if self.type == 'bus':
            if not (30 <= self.capacity['seats'] <= 80):
                raise ValueError("Bus seated capacity must be between 30 and 80")
            if not (20 <= self.capacity['standingRoom'] <= 50):
                raise ValueError("Bus standing capacity must be between 20 and 50")
        else:  # metro
            if not (200 <= self.capacity['seats'] <= 600):
                raise ValueError("Metro seated capacity must be between 200 and 600")
            if not (400 <= self.capacity['standingRoom'] <= 1000):
                raise ValueError("Metro standing capacity must be between 400 and 1000")
