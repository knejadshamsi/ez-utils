from pathlib import Path
from typing import Dict, Tuple, List, Set
from .models import GTFSData, TransitSchedule, TransitStop, TransitRoute, TransportMode
from .utils import validate_gtfs_files
from .base_processor import BaseProcessor
from .service_processor import process_gtfs_data
from .schedule_generator import generate_transit_schedule
from .vehicle_generator import generate_vehicles
import shutil

class PtProcessor(BaseProcessor):
    def __init__(self, mode: TransportMode = 'bus', selected_day: str = None):
        """Initialize PT processor.
        
        Args:
            mode: Transport mode (bus/metro)
            selected_day: Optional day of week (monday-friday)
            
        Raises:
            ValueError: If mode is invalid
        """
        super().__init__()
        if mode not in ('bus', 'metro'):
            raise ValueError("Mode must be either 'bus' or 'metro'")
        self.mode = mode
        self.selected_day = selected_day

    def process_chunk(self, chunk_file: Path) -> Path:
        """Process GTFS data chunk.
        
        Args:
            chunk_file: Path to GTFS directory
            
        Returns:
            Path to output directory containing generated files
            
        Note:
            GTFS processing doesn't use chunking as all files need to be processed together
        """
        # Create output directory
        output_dir = self.temp_dir / chunk_file.stem
        output_dir.mkdir(exist_ok=True)
        
        # Copy all GTFS files to temp directory
        for file in chunk_file.parent.glob('*.txt'):
            shutil.copy2(file, output_dir)
        
        # Process GTFS data
        gtfs_data = process_gtfs_data(output_dir, self.selected_day)
        
        # Generate output files
        schedule_file = output_dir / "transitSchedule.xml"
        vehicles_file = output_dir / "vehicles.xml"
        
        schedule = generate_transit_schedule(gtfs_data, self.mode)
        vehicles = generate_vehicles(gtfs_data, self.mode)
        
        schedule_file.write_text(schedule)
        vehicles_file.write_text(vehicles)
        
        return output_dir

def process_file(input_file: Path, output_file: Path, mode: TransportMode = 'bus', selected_day: str = None):
    """Process GTFS files to generate MATSim transit files.
    
    Args:
        input_file: Path to GTFS directory
        output_file: Path to output directory
        mode: Transport mode (bus/metro)
        selected_day: Optional day of week (monday-friday)
    """
    processor = PtProcessor(mode=mode, selected_day=selected_day)
    processor.process_file(input_file, output_file)
