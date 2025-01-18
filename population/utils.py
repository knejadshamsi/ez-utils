from pathlib import Path
import typer
from bs4 import BeautifulSoup
import pandas as pd
import pyproj
from shapely.geometry import Point
import os
from typing import Dict, Tuple

def validate_population_file(file_path: Path) -> BeautifulSoup:
    if not file_path.exists():
        raise typer.Exit(f"Population file not found: {file_path}")
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
            if not soup.find('population'):
                raise typer.Exit("Invalid population file: missing population element")
            return soup
    except Exception as e:
        raise typer.Exit(f"Failed to read population file: {e}")

def validate_output_dir(output: Path) -> Path:
    output_dir = output or Path.cwd()
    output_dir.mkdir(parents=True, exist_ok=True)
    population_dir = output_dir / "population-inputs"
    population_dir.mkdir(exist_ok=True)
    return population_dir

def parse_scale_list(scales: str) -> list[int]:
    try:
        return [int(s.strip()) for s in scales.split(",")]
    except ValueError:
        raise typer.Exit("Invalid scale format. Use comma-separated integers (e.g., 1,5,10)")

def get_total_population() -> int:
    total_population = int(os.getenv("TOTAL_POPULATION", 0))
    if not total_population:
        raise typer.Exit("TOTAL_POPULATION not set in .env file")
    return total_population

def transform_coordinates(x: float, y: float) -> Tuple[float, float]:
    transformer = pyproj.Transformer.from_crs('EPSG:2950', 'EPSG:4326', always_xy=True)
    return transformer.transform(x, y)

def create_point_from_coordinates(x: float, y: float) -> Point:
    lon, lat = transform_coordinates(x, y)
    return Point(lon, lat)

def parse_time(time_str: str) -> str:
    if not time_str:
        return None
    try:
        hours, minutes, seconds = map(int, time_str.strip().split(':'))
        if not (0 <= minutes <= 59 and 0 <= seconds <= 59):
            return None
        return f"{hours:02d}:{minutes:02d}:{seconds:02d}"
    except (ValueError, AttributeError):
        return None
