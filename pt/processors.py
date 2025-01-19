from pathlib import Path
from bs4 import BeautifulSoup
import pandas as pd
from typing import Dict
from .models import GTFSData, TransitSchedule, VehicleDefinition
from .utils import validate_gtfs_files, parse_gtfs_time
from .base_processor import BaseProcessor

def process_gtfs_data(gtfs_path: Path) -> GTFSData:
    dfs = validate_gtfs_files(gtfs_path)
    
    stop_times_df = dfs['stop_times']
    stop_times_df['arrival_time'] = stop_times_df['arrival_time'].apply(parse_gtfs_time)
    stop_times_df['departure_time'] = stop_times_df['departure_time'].apply(parse_gtfs_time)
    
    return GTFSData(
        routes=dfs['routes'],
        trips=dfs['trips'],
        stop_times=stop_times_df,
        stops=dfs['stops'],
        calendar=dfs['calendar'],
        calendar_dates=dfs['calendar_dates']
    )

def create_transit_schedule(gtfs_data: GTFSData, mode: str) -> str:
    valid_types = {
        'metro': [1, 401],
        'bus': [3, 700, 701, 702, 703]
    }
    
    route_types = valid_types.get(mode, [])
    routes_df = gtfs_data.routes[gtfs_data.routes['route_type'].isin(route_types)]
    
    soup = BeautifulSoup('<transitSchedule></transitSchedule>', 'lxml-xml')
    schedule = soup.find('transitSchedule')
    
    stops = soup.new_tag('transitStops')
    for _, stop in gtfs_data.stops.iterrows():
        stop_facility = soup.new_tag('stopFacility')
        stop_facility['id'] = str(stop['stop_id'])
        stop_facility['x'] = str(stop['stop_lon'])
        stop_facility['y'] = str(stop['stop_lat'])
        stops.append(stop_facility)
    schedule.append(stops)
    
    for _, route in routes_df.iterrows():
        transit_line = soup.new_tag('transitLine')
        transit_line['id'] = str(route['route_id'])
        
        route_trips = gtfs_data.trips[gtfs_data.trips['route_id'] == route['route_id']]
        
        for _, trip in route_trips.iterrows():
            transit_route = create_transit_route(soup, trip, gtfs_data, mode)
            transit_line.append(transit_route)
        
        schedule.append(transit_line)
    
    return str(soup)

def create_transit_route(soup: BeautifulSoup, trip: pd.Series, gtfs_data: GTFSData, mode: str) -> BeautifulSoup:
    transit_route = soup.new_tag('transitRoute')
    transit_route['id'] = str(trip['trip_id'])
    
    transport_mode = soup.new_tag('transportMode')
    transport_mode.string = mode
    transit_route.append(transport_mode)
    
    route_profile = soup.new_tag('routeProfile')
    trip_stops = gtfs_data.stop_times[
        gtfs_data.stop_times['trip_id'] == trip['trip_id']
    ].sort_values('stop_sequence')
    
    first_departure = pd.to_datetime(trip_stops.iloc[0]['departure_time'])
    
    for _, stop_time in trip_stops.iterrows():
        stop = soup.new_tag('stop')
        stop['refId'] = str(stop_time['stop_id'])
        
        arrival = pd.to_datetime(stop_time['arrival_time'])
        departure = pd.to_datetime(stop_time['departure_time'])
        
        arrival_offset = (arrival - first_departure).total_seconds()
        departure_offset = (departure - first_departure).total_seconds()
        
        stop['arrivalOffset'] = f"{int(arrival_offset//3600):02d}:{int((arrival_offset%3600)//60):02d}:{int(arrival_offset%60):02d}"
        stop['departureOffset'] = f"{int(departure_offset//3600):02d}:{int((departure_offset%3600)//60):02d}:{int(departure_offset%60):02d}"
        
        route_profile.append(stop)
    
    transit_route.append(route_profile)
    
    departures = soup.new_tag('departures')
    departure = soup.new_tag('departure')
    departure['id'] = "1"
    departure['departureTime'] = trip_stops.iloc[0]['departure_time']
    departures.append(departure)
    transit_route.append(departures)
    
    return transit_route

def create_vehicles(gtfs_data: GTFSData, mode: str) -> str:
    valid_types = {
        'metro': [1, 401],
        'bus': [3, 700, 701, 702, 703]
    }
    
    route_types = valid_types.get(mode, [])
    routes_df = gtfs_data.routes[gtfs_data.routes['route_type'].isin(route_types)]
    
    soup = BeautifulSoup('<vehicleDefinitions></vehicleDefinitions>', 'lxml-xml')
    vehicles = soup.find('vehicleDefinitions')
    
    vehicle_type = create_vehicle_type(soup, mode)
    vehicles.append(vehicle_type)
    
    used_ids = set()
    for _, trip in gtfs_data.trips[gtfs_data.trips['route_id'].isin(routes_df['route_id'])].iterrows():
        vehicle_id = f"{mode}_{trip['trip_id']}"
        if vehicle_id in used_ids:
            vehicle_id = f"{vehicle_id}_{len(used_ids)}"
        used_ids.add(vehicle_id)
        
        vehicle = soup.new_tag('vehicle')
        vehicle['id'] = vehicle_id
        vehicle['type'] = mode
        vehicles.append(vehicle)
    
    return str(soup)

def create_vehicle_type(soup: BeautifulSoup, mode: str) -> BeautifulSoup:
    vehicle_type = soup.new_tag('vehicleType')
    vehicle_type['id'] = mode
    
    capacity = soup.new_tag('capacity')
    seats = soup.new_tag('seats')
    standing = soup.new_tag('standingRoom')
    
    if mode == 'metro':
        seats.string = "400"
        standing.string = "800"
    else:
        seats.string = "50"
        standing.string = "30"
    
    capacity.append(seats)
    capacity.append(standing)
    vehicle_type.append(capacity)
    
    length = soup.new_tag('length')
    length.string = "12.0" if mode == 'bus' else "150.0"
    vehicle_type.append(length)
    
    width = soup.new_tag('width')
    width.string = "2.5" if mode == 'bus' else "3.2"
    vehicle_type.append(width)
    
    max_speed = soup.new_tag('maximumVelocity')
    max_speed.string = "13.89" if mode == 'bus' else "22.22"
    vehicle_type.append(max_speed)
    
    return vehicle_type

class PtProcessor(BaseProcessor):
    def __init__(self, mode: str = 'bus'):
        super().__init__()
        self.mode = mode

    def process_chunk(self, chunk_file: Path) -> Path:
        output_chunk = self.temp_dir / f"output_{chunk_file.name}"
        
        gtfs_data = process_gtfs_data(chunk_file)
        schedule = create_transit_schedule(gtfs_data, self.mode)
        vehicles = create_vehicles(gtfs_data, self.mode)
        
        with open(output_chunk, 'w') as f:
            f.write(schedule)
            f.write('\n')
            f.write(vehicles)
        
        return output_chunk

def process_file(input_file: Path, output_file: Path, mode: str = 'bus'):
    processor = PtProcessor(mode=mode)
    processor.process_file(input_file, output_file)
