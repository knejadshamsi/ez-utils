from typing import List, Dict
from .models import GTFSData, VehicleDefinition, VehicleCapacity, TransportMode
from .xml_utils import create_xml_root, add_element, format_number

def create_vehicle_type(mode: TransportMode) -> VehicleDefinition:
    """Create vehicle type definition.
    
    Args:
        mode: Transport mode (bus/metro)
        
    Returns:
        VehicleDefinition object
    """
    capacity: VehicleCapacity = {
        'seats': 400 if mode == 'metro' else 50,
        'standingRoom': 800 if mode == 'metro' else 30
    }
    
    return VehicleDefinition(
        id=mode,
        type=mode,
        capacity=capacity,
        length=150.0 if mode == 'metro' else 12.0,
        width=3.2 if mode == 'metro' else 2.5,
        max_speed=22.22 if mode == 'metro' else 13.89  # 80 km/h for metro, 50 km/h for bus
    )

def get_vehicle_ids(
    gtfs_data: GTFSData,
    mode: TransportMode,
    route_types: List[int]
) -> List[str]:
    """Get list of vehicle IDs for mode.
    
    Args:
        gtfs_data: GTFS data object
        mode: Transport mode (bus/metro)
        route_types: Valid route types for mode
        
    Returns:
        List of vehicle IDs
    """
    routes_df = gtfs_data.routes[gtfs_data.routes['route_type'].isin(route_types)]
    
    vehicle_ids = []
    used_ids = set()
    
    for _, trip in gtfs_data.trips[
        gtfs_data.trips['route_id'].isin(routes_df['route_id'])
    ].iterrows():
        vehicle_id = f"{mode}_{trip['trip_id']}"
        if vehicle_id in used_ids:
            vehicle_id = f"{vehicle_id}_{len(used_ids)}"
        used_ids.add(vehicle_id)
        vehicle_ids.append(vehicle_id)
    
    return vehicle_ids

def generate_vehicles(gtfs_data: GTFSData, mode: TransportMode) -> str:
    """Generate MATSim vehicles XML.
    
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
    
    # Create vehicle type
    vehicle_type = create_vehicle_type(mode)
    
    # Get vehicle IDs
    vehicle_ids = get_vehicle_ids(gtfs_data, mode, route_types)
    
    # Generate XML
    soup = create_xml_root('vehicleDefinitions')
    root = soup.find('vehicleDefinitions')
    
    # Add vehicle type
    vtype = add_element(root, 'vehicleType', {'id': vehicle_type.id})
    
    # Add capacity
    capacity = add_element(vtype, 'capacity')
    add_element(capacity, 'seats', text=str(vehicle_type.capacity['seats']))
    add_element(capacity, 'standingRoom', text=str(vehicle_type.capacity['standingRoom']))
    
    # Add dimensions
    add_element(vtype, 'length', text=format_number(vehicle_type.length))
    add_element(vtype, 'width', text=format_number(vehicle_type.width))
    add_element(vtype, 'maximumVelocity', text=format_number(vehicle_type.max_speed))
    
    # Add vehicles
    for vehicle_id in vehicle_ids:
        add_element(
            root,
            'vehicle',
            {
                'id': vehicle_id,
                'type': mode
            }
        )
    
    return str(soup)
