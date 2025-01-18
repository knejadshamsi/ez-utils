from pathlib import Path
import typer
from bs4 import BeautifulSoup
import pyproj
from shapely.geometry import Point, LineString
from typing import Dict, List, Tuple

def validate_network_dir(network_dir: Path) -> Path:
    if not network_dir.exists():
        raise typer.Exit(f"Network directory not found: {network_dir}")
    return network_dir

def validate_population_dir(population_dir: Path) -> Path:
    if not population_dir.exists():
        raise typer.Exit(f"Population directory not found: {population_dir}")
    return population_dir

def validate_output_dir(output: Path) -> Path:
    output_dir = output or Path.cwd()
    output_dir.mkdir(parents=True, exist_ok=True)
    index_dir = output_dir / "indexes"
    index_dir.mkdir(exist_ok=True)
    return index_dir

def parse_scale_list(scales: str) -> list[int]:
    try:
        return [int(s.strip()) for s in scales.split(",")]
    except ValueError:
        raise typer.Exit("Invalid scale format. Use comma-separated integers (e.g., 1,5,10)")

def transform_coordinates(x: float, y: float) -> Tuple[float, float]:
    transformer = pyproj.Transformer.from_crs('EPSG:2950', 'EPSG:4326', always_xy=True)
    return transformer.transform(x, y)

def create_point_from_coordinates(x: float, y: float) -> Point:
    lon, lat = transform_coordinates(x, y)
    return Point(lon, lat)

def calculate_distance_to_link(point: Point, from_point: Point, to_point: Point) -> float:
    line = LineString([from_point, to_point])
    return point.distance(line)

def write_index_file(index_data: Dict[str, List[str]], output_file: Path) -> None:
    with open(output_file, 'w', encoding='utf-8') as f:
        for link_id, agent_ids in index_data.items():
            f.write(f"{link_id}:{','.join(agent_ids)}\n")
