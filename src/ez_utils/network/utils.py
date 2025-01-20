from pathlib import Path
import typer
from bs4 import BeautifulSoup
import pandas as pd
from typing import Dict, Tuple
import pyproj
from shapely.geometry import Point

def validate_network_file(file_path: Path) -> BeautifulSoup:
    if not file_path.exists():
        raise typer.Exit(f"Network file not found: {file_path}")
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
            if not soup.find('network'):
                raise typer.Exit("Invalid network file: missing network element")
            return soup
    except Exception as e:
        raise typer.Exit(f"Failed to read network file: {e}")

def transform_coordinates(x: float, y: float) -> Tuple[float, float]:
    transformer = pyproj.Transformer.from_crs('EPSG:2950', 'EPSG:4326', always_xy=True)
    return transformer.transform(x, y)

def validate_output_dir(output: Path) -> Path:
    output_dir = output or Path.cwd()
    output_dir.mkdir(parents=True, exist_ok=True)
    network_dir = output_dir / "network-inputs"
    network_dir.mkdir(exist_ok=True)
    return network_dir

def parse_scale_list(scales: str) -> list[int]:
    try:
        return [int(s.strip()) for s in scales.split(",")]
    except ValueError:
        raise typer.Exit("Invalid scale format. Use comma-separated integers (e.g., 1,5,10)")

def create_point_from_coordinates(x: float, y: float) -> Point:
    lon, lat = transform_coordinates(x, y)
    return Point(lon, lat)
