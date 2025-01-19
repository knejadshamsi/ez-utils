from typing import Tuple
import pyproj

def convert_coordinates(lat: float, lon: float) -> Tuple[float, float]:
    """Convert latitude/longitude to UTM coordinates.
    
    Args:
        lat: Latitude in degrees
        lon: Longitude in degrees
        
    Returns:
        Tuple of (x, y) coordinates in UTM
        
    Raises:
        ValueError: If coordinate conversion fails
    """
    try:
        proj = pyproj.Transformer.from_crs("EPSG:4326", "EPSG:32632", always_xy=True)
        x, y = proj.transform(lon, lat)
        return x, y
    except Exception as e:
        raise ValueError(f"Failed to convert coordinates ({lat}, {lon}): {str(e)}")
