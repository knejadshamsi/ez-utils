from typing import Set
import pandas as pd
from .models import GTFSData
from .time_utils import parse_gtfs_time

def get_service_dates(gtfs_data: GTFSData) -> pd.DataFrame:
    """Get available service dates from GTFS calendar.
    
    Args:
        gtfs_data: GTFS data object
        
    Returns:
        DataFrame with service_id and day columns
        
    Raises:
        ValueError: If no weekday services found
    """
    calendar_df = gtfs_data.calendar
    dates = []
    days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday']
    
    for _, row in calendar_df.iterrows():
        for day in days:
            if row[day] == 1:
                dates.append({
                    'service_id': row['service_id'],
                    'day': day
                })
    
    if not dates:
        raise ValueError("No weekday services found in calendar")
    
    return pd.DataFrame(dates)

def get_active_services(gtfs_data: GTFSData, selected_day: str) -> Set[str]:
    """Get active service IDs for the selected day.
    
    Args:
        gtfs_data: GTFS data object
        selected_day: Day of week (monday-friday)
        
    Returns:
        Set of active service IDs
    """
    # Get regular services for the day
    day_services = set(gtfs_data.calendar[gtfs_data.calendar[selected_day] == 1]['service_id'])
    
    # Apply exceptions from calendar_dates
    added_services = set(gtfs_data.calendar_dates[
        gtfs_data.calendar_dates['exception_type'] == 1
    ]['service_id'])
    
    removed_services = set(gtfs_data.calendar_dates[
        gtfs_data.calendar_dates['exception_type'] == 2
    ]['service_id'])
    
    return (day_services | added_services) - removed_services

def process_gtfs_data(gtfs_data: GTFSData, selected_day: str = None) -> GTFSData:
    """Process GTFS data for a specific day.
    
    Args:
        gtfs_data: GTFS data object
        selected_day: Optional day of week (monday-friday)
        
    Returns:
        Processed GTFS data object
        
    Raises:
        ValueError: If no service found for selected day
    """
    # Process stop times
    gtfs_data.stop_times['arrival_time'] = gtfs_data.stop_times['arrival_time'].apply(parse_gtfs_time)
    gtfs_data.stop_times['departure_time'] = gtfs_data.stop_times['departure_time'].apply(parse_gtfs_time)
    
    # Handle day selection
    if not selected_day:
        service_dates = get_service_dates(gtfs_data)
        if len(service_dates) > 1:
            print("\nAvailable service days:")
            for i, day in enumerate(service_dates['day'].unique(), 1):
                print(f"{i}. {day.capitalize()}")
            day_idx = int(input("\nSelect a day (enter number): ")) - 1
            selected_day = service_dates['day'].unique()[day_idx]
        else:
            selected_day = service_dates['day'].iloc[0]
    
    # Get active services including exceptions
    active_services = get_active_services(gtfs_data, selected_day)
    if not active_services:
        raise ValueError(f"No service found for {selected_day}")
    
    # Filter trips for active services
    gtfs_data.trips = gtfs_data.trips[gtfs_data.trips['service_id'].isin(active_services)]
    if gtfs_data.trips.empty:
        raise ValueError(f"No trips found for {selected_day}")
    
    return gtfs_data
