from pathlib import Path
from typing import List
from index.processors import IndexProcessor
from network.processors import NetworkProcessor
from population.processors import PopulationProcessor
from pt.processors import PtProcessor
from .base_processor import BaseProcessor

class AllProcessor(BaseProcessor):
    def __init__(self):
        super().__init__()
        self.processors = [
            IndexProcessor(),
            NetworkProcessor(),
            PopulationProcessor(),
            PtProcessor()
        ]

    def process_chunk(self, chunk_file: Path) -> Path:
        output_chunk = self.temp_dir / f"output_{chunk_file.name}"
        intermediate_file = chunk_file
        
        for processor in self.processors:
            temp_output = processor.process_chunk(intermediate_file)
            intermediate_file = temp_output
        
        return intermediate_file

def process_files(input_files: List[Path], output_dir: Path):
    processor = AllProcessor()
    output_dir.mkdir(parents=True, exist_ok=True)
    
    for input_file in input_files:
        output_file = output_dir / f"processed_{input_file.name}"
        processor.process_file(input_file, output_file)
