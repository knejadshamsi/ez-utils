import pandas as pd
from typing import Dict, List
from .models import (
    GTFSData, TransitSchedule, TransitStop, TransitRoute,
    StopReference, DepartureReference, TransportMode
)
from .coordinate_utils import convert_coordinates
from .time_utils import format_time, format_time_offset
from .xml_utils import create_xml_root, add_element, format_number

def create_transit_stops(gtfs_data: GTFSData) -> Dict[str, TransitStop]:
    """Create transit stops from GTFS data.
    
    Args:
        gtfs_data: GTFS data object
        
    Returns:
        Dictionary of stop ID to TransitStop object
    """
    stops = {}
    for _, stop in gtfs_data.stops.iterrows():
        x, y = convert_coordinates(stop['stop_lat'], stop['stop_lon'])
        stops[str(stop['stop_id'])] = TransitStop(
            id=str(stop['stop_id']),
            x=x,
            y=y
        )
    return stops

def create_stop_references(
    trip_stops: pd.DataFrame,
    first_departure: pd.Timestamp
) -> List[StopReference]:
    """Create stop references with time offsets.
    
    Args:
        trip_stops: DataFrame of stops for a trip
        first_departure: First departure time
        
    Returns:
        List of StopReference objects
    """
    stops: List[StopReference] = []
    for _, stop_time in trip_stops.iterrows():
        arrival = stop_time['arrival_time']
        departure = stop_time['departure_time']
        
        arrival_offset = (arrival - first_departure).total_seconds()
        departure_offset = (departure - first_departure).total_seconds()
        
        stops.append({
            'refId': str(stop_time['stop_id']),
            'arrivalOffset': format_time_offset(arrival_offset),
            'departureOffset': format_time_offset(departure_offset)
        })
    return stops

def create_transit_routes(
    gtfs_data: GTFSData,
    mode: TransportMode,
    route_types: List[int]
) -> Dict[str, TransitRoute]:
    """Create transit routes from GTFS data.
    
    Args:
        gtfs_data: GTFS data object
        mode: Transport mode (bus/metro)
        route_types: Valid route types for mode
        
    Returns:
        Dictionary of route ID to TransitRoute object
    """
    routes = {}
    routes_df = gtfs_data.routes[gtfs_data.routes['route_type'].isin(route_types)]
    
    for _, route in routes_df.iterrows():
        route_trips = gtfs_data.trips[gtfs_data.trips['route_id'] == route['route_id']]
        
        for _, trip in route_trips.iterrows():
            trip_stops = gtfs_data.stop_times[
                gtfs_data.stop_times['trip_id'] == trip['trip_id']
            ].sort_values('stop_sequence')
            
            if trip_stops.empty:
                continue
            
            first_departure = trip_stops.iloc[0]['departure_time']
            
            # Create stop references
            stops = create_stop_references(trip_stops, first_departure)
            
            # Create departure reference
            departures: List[DepartureReference] = [{
                'id': "1",
                'departureTime': format_time(first_departure),
                'vehicleRefId': f"{mode}_{trip['trip_id']}"
            }]
            
            # Create route
            routes[str(trip['trip_id'])] = TransitRoute(
                id=str(trip['trip_id']),
                mode=mode,
                stops=stops,
                departures=departures
            )
    
    return routes

def generate_transit_schedule(gtfs_data: GTFSData, mode: TransportMode) -> str:
    """Generate MATSim transit schedule XML.
    
    Args:
        gtfs_data: GTFS data object
        mode: Transport mode (bus/metro)
        
    Returns:
        XML string
        
    Raises:
        ValueError: If no routes found for mode
    """
    valid_types = {
        'metro': [1, 401],
        'bus': [3, 700, 701, 702, 703]
    }
    
    route_types = valid_types[mode]
    routes_df = gtfs_data.routes[gtfs_data.routes['route_type'].isin(route_types)]
    
    if routes_df.empty:
        raise ValueError(f"No {mode} routes found")
    
    # Create schedule structure
    schedule = TransitSchedule(
        stops=create_transit_stops(gtfs_data),
        routes=create_transit_routes(gtfs_data, mode, route_types)
    )
    
    # Generate XML
    soup = create_xml_root('transitSchedule')
    root = soup.find('transitSchedule')
    
    # Add stops
    stops_elem = add_element(root, 'transitStops')
    for stop in schedule.stops.values():
        stop_facility = add_element(
            stops_elem,
            'stopFacility',
            {
                'id': stop.id,
                'x': format_number(stop.x),
                'y': format_number(stop.y),
                'linkRefId': stop.linkRefId
            }
        )
        if 'stop_name' in gtfs_data.stops.columns:
            stop_name = gtfs_data.stops[
                gtfs_data.stops['stop_id'] == stop.id
            ]['stop_name'].iloc[0]
            add_element(stop_facility, 'name', text=str(stop_name))
    
    # Add routes grouped by route_id
    for route_id in routes_df['route_id'].unique():
        transit_line = add_element(root, 'transitLine', {'id': str(route_id)})
        
        # Add route name if available
        if 'route_short_name' in routes_df.columns:
            route_name = routes_df[
                routes_df['route_id'] == route_id
            ]['route_short_name'].iloc[0]
            add_element(transit_line, 'name', text=str(route_name))
        
        # Add all routes for this line
        for route in schedule.routes.values():
            if route.id in gtfs_data.trips[
                gtfs_data.trips['route_id'] == route_id
            ]['trip_id'].values:
                transit_route = add_element(
                    transit_line,
                    'transitRoute',
                    {'id': route.id}
                )
                
                add_element(transit_route, 'transportMode', text=route.mode)
                
                # Add route profile
                route_profile = add_element(transit_route, 'routeProfile')
                for stop in route.stops:
                    add_element(
                        route_profile,
                        'stop',
                        {
                            'refId': stop['refId'],
                            'arrivalOffset': stop['arrivalOffset'],
                            'departureOffset': stop['departureOffset']
                        }
                    )
                
                # Add departures
                departures = add_element(transit_route, 'departures')
                for dep in route.departures:
                    add_element(
                        departures,
                        'departure',
                        {
                            'id': dep['id'],
                            'departureTime': dep['departureTime'],
                            'vehicleRefId': dep['vehicleRefId']
                        }
                    )
    
    return str(soup)
