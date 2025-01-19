import pandas as pd
from pathlib import Path
import typer
from .models import GTFSData
from typing import Dict
import numpy as np
from datetime import datetime

def read_gtfs_file(gtfs_path: Path, filename: str) -> pd.DataFrame:
    """Read and clean GTFS file.
    
    Args:
        gtfs_path: Path to GTFS directory
        filename: Name of file to read
        
    Returns:
        DataFrame with file contents
        
    Raises:
        typer.Exit: If file cannot be read
    """
    try:
        df = pd.read_csv(gtfs_path / filename)
        df = df.replace(r'^\s*$', pd.NA, regex=True)
        return df
    except Exception as e:
        raise typer.Exit(f"Failed to read {filename}: {e}")

def parse_gtfs_time(time_str: str) -> pd.Timestamp:
    """Parse GTFS time string.
    
    Args:
        time_str: Time string in HH:MM:SS format
        
    Returns:
        Pandas Timestamp
        
    Raises:
        typer.Exit: If time format is invalid
    """
    if not isinstance(time_str, str) or not time_str.strip():
        raise typer.Exit(f"Invalid time format: {time_str}")
    
    try:
        hours, minutes, seconds = map(int, time_str.strip().split(':'))
        if not (0 <= minutes <= 59 and 0 <= seconds <= 59):
            raise typer.Exit(f"Invalid minutes/seconds in time: {time_str}")
        
        if hours >= 24:
            hours = hours % 24
        return pd.Timestamp(f"{hours:02d}:{minutes:02d}:{seconds:02d}")
    except ValueError:
        raise typer.Exit(f"Invalid time format: {time_str}")

def validate_coordinates(df: pd.DataFrame, file: str) -> None:
    """Validate coordinate values.
    
    Args:
        df: DataFrame with stop coordinates
        file: Filename for error messages
        
    Raises:
        typer.Exit: If coordinates are invalid
    """
    # Validate data types
    if not (pd.to_numeric(df['stop_lat'], errors='coerce').notna().all() and 
            pd.to_numeric(df['stop_lon'], errors='coerce').notna().all()):
        raise typer.Exit(f"Non-numeric coordinates in {file}")
    
    # Convert to float for range check
    df['stop_lat'] = df['stop_lat'].astype(float)
    df['stop_lon'] = df['stop_lon'].astype(float)
    
    if not all(df['stop_lat'].between(-90, 90)) or not all(df['stop_lon'].between(-180, 180)):
        raise typer.Exit(f"Invalid coordinates in {file}")

def validate_route_types(df: pd.DataFrame, file: str) -> None:
    """Validate route types.
    
    Args:
        df: DataFrame with route types
        file: Filename for error messages
        
    Raises:
        typer.Exit: If route types are invalid
    """
    # Validate data type
    if not pd.to_numeric(df['route_type'], errors='coerce').notna().all():
        raise typer.Exit(f"Non-numeric route types in {file}")
    
    df['route_type'] = df['route_type'].astype(int)
    valid_types = [1, 3, 401, 700, 701, 702, 703]  # metro and bus types
    if not all(df['route_type'].isin(valid_types)):
        raise typer.Exit(f"Invalid route types in {file}. Must be one of {valid_types}")

def validate_service_dates(calendar_df: pd.DataFrame, dates_df: pd.DataFrame) -> None:
    """Validate service dates and exceptions.
    
    Args:
        calendar_df: Calendar DataFrame
        dates_df: Calendar dates DataFrame
        
    Raises:
        typer.Exit: If service dates are invalid
    """
    # Check if at least one weekday is active
    weekdays = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday']
    if not any(calendar_df[weekdays].any()):
        raise typer.Exit("No weekday services found in calendar.txt")
    
    # Validate binary values for weekdays
    for day in weekdays:
        if not all(calendar_df[day].isin([0, 1])):
            raise typer.Exit(f"Invalid value in calendar.txt for {day}. Must be 0 or 1")
    
    # Validate date format in calendar_dates
    try:
        dates_df['date'] = pd.to_datetime(dates_df['date'], format='%Y%m%d')
    except ValueError:
        raise typer.Exit("Invalid date format in calendar_dates.txt. Must be YYYYMMDD")
    
    # Validate date range
    today = pd.Timestamp.now().normalize()
    if dates_df['date'].max() < today:
        raise typer.Exit("All service dates are in the past")
    
    # Validate exception types
    if not pd.to_numeric(dates_df['exception_type'], errors='coerce').notna().all():
        raise typer.Exit("Non-numeric exception types in calendar_dates.txt")
    
    dates_df['exception_type'] = dates_df['exception_type'].astype(int)
    if not all(dates_df['exception_type'].isin([1, 2])):
        raise typer.Exit("Invalid exception_type in calendar_dates.txt. Must be 1 (added) or 2 (removed)")
    
    # Validate service_id consistency
    calendar_services = set(calendar_df['service_id'])
    dates_services = set(dates_df['service_id'])
    if not dates_services.issubset(calendar_services):
        invalid_services = dates_services - calendar_services
        raise typer.Exit(f"Service IDs in calendar_dates.txt not found in calendar.txt: {invalid_services}")

def validate_times(df: pd.DataFrame, file: str) -> None:
    """Validate arrival and departure times.
    
    Args:
        df: DataFrame with times
        file: Filename for error messages
        
    Raises:
        typer.Exit: If times are invalid
    """
    try:
        # Check time format
        arrival_times = df['arrival_time'].dropna()
        departure_times = df['departure_time'].dropna()
        
        for times in [arrival_times, departure_times]:
            if not all(times.str.match(r'^\d{2}:\d{2}:\d{2}$')):
                raise typer.Exit(f"Invalid time format in {file}. Must be HH:MM:SS")
        
        # Check if departure time is not before arrival time
        df['arr'] = pd.to_datetime(df['arrival_time'], format='%H:%M:%S', errors='coerce')
        df['dep'] = pd.to_datetime(df['departure_time'], format='%H:%M:%S', errors='coerce')
        if any(df['dep'] < df['arr']):
            raise typer.Exit(f"Departure time before arrival time in {file}")
        
        # Clean up temporary columns
        df.drop(['arr', 'dep'], axis=1, inplace=True)
    except Exception as e:
        raise typer.Exit(f"Time validation failed in {file}: {str(e)}")

def validate_stop_sequence(df: pd.DataFrame, file: str) -> None:
    """Validate stop sequence numbers.
    
    Args:
        df: DataFrame with stop sequences
        file: Filename for error messages
        
    Raises:
        typer.Exit: If sequences are invalid
    """
    # Validate data type
    if not pd.to_numeric(df['stop_sequence'], errors='coerce').notna().all():
        raise typer.Exit(f"Non-numeric stop sequence values in {file}")
    
    df['stop_sequence'] = df['stop_sequence'].astype(int)
    
    # Check for each trip if stop_sequence is sequential
    for trip_id in df['trip_id'].unique():
        trip_stops = df[df['trip_id'] == trip_id].sort_values('stop_sequence')
        expected_sequence = range(trip_stops['stop_sequence'].min(), 
                                trip_stops['stop_sequence'].max() + 1)
        if not all(trip_stops['stop_sequence'] == expected_sequence):
            raise typer.Exit(f"Non-sequential stop_sequence for trip_id {trip_id} in {file}")

def validate_references(dfs: Dict[str, pd.DataFrame]) -> None:
    """Validate cross-file references.
    
    Args:
        dfs: Dictionary of DataFrames
        
    Raises:
        typer.Exit: If references are invalid
    """
    # Check for duplicate IDs
    for file, id_col in [
        ('routes', 'route_id'),
        ('trips', 'trip_id'),
        ('stops', 'stop_id'),
        ('calendar', 'service_id')
    ]:
        if dfs[file][id_col].duplicated().any():
            raise typer.Exit(f"Duplicate {id_col} found in {file}.txt")
    
    # Validate trip references
    trip_ids = set(dfs['trips']['trip_id'])
    stop_time_trips = set(dfs['stop_times']['trip_id'])
    if not stop_time_trips.issubset(trip_ids):
        invalid_trips = stop_time_trips - trip_ids
        raise typer.Exit(f"Trip IDs in stop_times.txt not found in trips.txt: {invalid_trips}")
    
    # Validate route references
    route_ids = set(dfs['routes']['route_id'])
    trip_routes = set(dfs['trips']['route_id'])
    if not trip_routes.issubset(route_ids):
        invalid_routes = trip_routes - route_ids
        raise typer.Exit(f"Route IDs in trips.txt not found in routes.txt: {invalid_routes}")
    
    # Validate stop references
    stop_ids = set(dfs['stops']['stop_id'])
    stop_time_stops = set(dfs['stop_times']['stop_id'])
    if not stop_time_stops.issubset(stop_ids):
        invalid_stops = stop_time_stops - stop_ids
        raise typer.Exit(f"Stop IDs in stop_times.txt not found in stops.txt: {invalid_stops}")

def validate_gtfs_files(gtfs_path: Path) -> Dict[str, pd.DataFrame]:
    """Validate and load GTFS files.
    
    Args:
        gtfs_path: Path to GTFS directory
        
    Returns:
        Dictionary of validated DataFrames
        
    Raises:
        typer.Exit: If validation fails
    """
    required_files = [
        'routes.txt', 'trips.txt', 'stop_times.txt',
        'stops.txt', 'calendar.txt', 'calendar_dates.txt'
    ]
    
    for file in required_files:
        if not (gtfs_path / file).exists():
            raise typer.Exit(f"Missing required file: {file}")
    
    dfs = {}
    required_columns = {
        'routes.txt': ['route_id', 'route_type'],
        'trips.txt': ['trip_id', 'route_id', 'service_id'],
        'stop_times.txt': ['trip_id', 'stop_id', 'stop_sequence', 'arrival_time', 'departure_time'],
        'stops.txt': ['stop_id', 'stop_lat', 'stop_lon'],
        'calendar.txt': ['service_id', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'],
        'calendar_dates.txt': ['service_id', 'date', 'exception_type']
    }
    
    optional_columns = {
        'routes.txt': ['route_short_name', 'route_long_name'],
        'stops.txt': ['stop_name']
    }
    
    for file, columns in required_columns.items():
        df = read_gtfs_file(gtfs_path, file)
        
        # Check required columns
        if not all(col in df.columns for col in columns):
            raise typer.Exit(f"Missing required columns in {file}")
        
        # Check optional columns
        if file in optional_columns:
            missing_optional = [
                col for col in optional_columns[file]
                if col not in df.columns
            ]
            if missing_optional:
                print(f"Warning: Optional columns missing in {file}: {missing_optional}")
        
        # Additional validations
        if file == 'stops.txt':
            validate_coordinates(df, file)
        elif file == 'routes.txt':
            validate_route_types(df, file)
        elif file == 'stop_times.txt':
            validate_times(df, file)
            validate_stop_sequence(df, file)
        
        dfs[file.replace('.txt', '')] = df
    
    # Cross-file validations
    validate_service_dates(dfs['calendar'], dfs['calendar_dates'])
    validate_references(dfs)
    
    return dfs
