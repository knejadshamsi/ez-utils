from datetime import datetime
import pandas as pd

def format_time_offset(seconds: float) -> str:
    """Format seconds into HH:MM:SS.
    
    Args:
        seconds: Number of seconds
        
    Returns:
        Time string in HH:MM:SS format
    """
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    return f"{hours:02d}:{minutes:02d}:{secs:02d}"

def format_time(timestamp: pd.Timestamp) -> str:
    """Format pandas Timestamp into HH:MM:SS.
    
    Args:
        timestamp: Pandas Timestamp object
        
    Returns:
        Time string in HH:MM:SS format
    """
    return f"{timestamp.hour:02d}:{timestamp.minute:02d}:{timestamp.second:02d}"

def parse_gtfs_time(time_str: str) -> pd.Timestamp:
    """Parse GTFS time string into pandas Timestamp.
    
    Args:
        time_str: Time string in HH:MM:SS format
        
    Returns:
        Pandas Timestamp object
        
    Raises:
        ValueError: If time string is invalid
    """
    if not isinstance(time_str, str) or not time_str.strip():
        raise ValueError(f"Invalid time format: {time_str}")
    
    try:
        hours, minutes, seconds = map(int, time_str.strip().split(':'))
        if not (0 <= minutes <= 59 and 0 <= seconds <= 59):
            raise ValueError(f"Invalid minutes/seconds in time: {time_str}")
        
        if hours >= 24:
            hours = hours % 24
        return pd.Timestamp(f"{hours:02d}:{minutes:02d}:{seconds:02d}")
    except ValueError as e:
        raise ValueError(f"Invalid time format: {time_str}") from e
