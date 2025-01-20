from pathlib import Path
import shutil
from typing import List
from ..index.processors import IndexProcessor, process_scale
from ..network.processors import NetworkProcessor
from ..population.processors import PopulationProcessor
from ..pt.processors import PtProcessor

def process_files(input_dir: Path, output_dir: Path):
    """Process all files in input directory and create scaled outputs.
    
    Args:
        input_dir: Directory containing:
            - network.xml: Network configuration
            - population.xml: Population data
            - gtfs/ (optional): GTFS files
            
        output_dir: Directory to create:
            - network/: Scaled network files (01-10)
            - population/: Scaled population files (01-10) 
            - index/: Index files for each scale
            - pt/ (optional): Transit files
    """
    # Create processors
    network_processor = NetworkProcessor()
    population_processor = PopulationProcessor()
    pt_processor = PtProcessor()

    # Create output directories
    output_dir.mkdir(parents=True, exist_ok=True)
    network_dir = output_dir / "network"
    population_dir = output_dir / "population"
    network_dir.mkdir(parents=True, exist_ok=True)
    population_dir.mkdir(parents=True, exist_ok=True)

    # Process network files
    network_file = input_dir / "network.xml"
    if not network_file.exists():
        raise FileNotFoundError(f"Network file not found: {network_file}")
    network_processor.process_file(network_file, network_dir)

    # Process population files
    population_file = input_dir / "population.xml"
    if not population_file.exists():
        raise FileNotFoundError(f"Population file not found: {population_file}")
    population_processor.process_file(population_file, population_dir)

    # Create indices for each scale
    for scale in range(1, 11):
        process_scale(scale, network_dir, population_dir)

    # Process PT data if available
    gtfs_dir = input_dir / "gtfs"
    if gtfs_dir.exists():
        pt_dir = output_dir / "pt"
        pt_dir.mkdir(parents=True, exist_ok=True)
        pt_processor.process_file(gtfs_dir, pt_dir)
