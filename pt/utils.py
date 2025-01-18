import pandas as pd
from pathlib import Path
import typer
from .models import GTFSData
from typing import Dict

def read_gtfs_file(gtfs_path: Path, filename: str) -> pd.DataFrame:
    try:
        df = pd.read_csv(gtfs_path / filename)
        df = df.replace(r'^\s*$', pd.NA, regex=True)
        return df
    except Exception as e:
        raise typer.Exit(f"Failed to read {filename}: {e}")

def parse_gtfs_time(time_str: str) -> pd.Timestamp:
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

def validate_gtfs_files(gtfs_path: Path) -> Dict[str, pd.DataFrame]:
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
    
    for file, columns in required_columns.items():
        df = read_gtfs_file(gtfs_path, file)
        if not all(col in df.columns for col in columns):
            raise typer.Exit(f"Missing required columns in {file}")
        dfs[file.replace('.txt', '')] = df
    
    return dfs
