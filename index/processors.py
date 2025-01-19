from bs4 import BeautifulSoup
from pathlib import Path
from typing import Dict, List, Tuple
from .models import ChunkResult
from .base_processor import BaseProcessor

class IndexProcessor(BaseProcessor):
    def __init__(self, output_type: str):
        super().__init__()
        self.output_type = output_type  # 'start' or 'pass'
        self.start_dir = Path("index/start")
        self.pass_dir = Path("index/pass")
        self.start_dir.mkdir(parents=True, exist_ok=True)
        self.pass_dir.mkdir(parents=True, exist_ok=True)

    def process_chunk(self, chunk_file: Path) -> Path:
        result = self._process_xml_chunk(chunk_file)
        output_chunk = self.temp_dir / f"processed_{chunk_file.name}"
        
        # Write only the relevant index type
        indices = result.start_indices if self.output_type == 'start' else result.pass_indices
        self._write_index_file(output_chunk, indices)
        
        return output_chunk

    def _process_xml_chunk(self, chunk_file: Path) -> ChunkResult:
        start_indices: Dict[str, List[str]] = {}
        pass_indices: Dict[str, List[str]] = {}
        
        with open(chunk_file, 'r', encoding='utf-8') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
            
        for person in soup.find_all('person'):
            agent_id = person['id']
            plan = person.find('plan')
            if not plan:
                continue

            # Process start location (first home activity)
            if self.output_type == 'start':
                first_home = plan.find('act', type='home')
                if first_home and 'link' in first_home.attrs:
                    start_link = first_home['link']
                    if start_link not in start_indices:
                        start_indices[start_link] = []
                    start_indices[start_link].append(agent_id)

            # Process route links
            if self.output_type == 'pass':
                for route in plan.find_all('route', type='links'):
                    if route.string:
                        link_ids = route.string.strip().split()
                        for link_id in link_ids:
                            if link_id not in pass_indices:
                                pass_indices[link_id] = []
                            pass_indices[link_id].append(agent_id)

        return ChunkResult(start_indices=start_indices, pass_indices=pass_indices)

    def _write_index_file(self, output_file: Path, indices: Dict[str, List[str]]):
        with open(output_file, 'w', encoding='utf-8') as f:
            for link_id, agent_ids in sorted(indices.items()):
                unique_agents = sorted(set(agent_ids))  # Remove duplicates and sort
                f.write(f"{link_id}:{','.join(unique_agents)}\n")

    def merge_chunks(self, chunk_files: List[Path], output_file: Path):
        combined_indices: Dict[str, List[str]] = {}
        
        for chunk_file in chunk_files:
            with open(chunk_file, 'r', encoding='utf-8') as f:
                for line in f:
                    if ':' not in line:
                        continue
                    link_id, agents_str = line.strip().split(':', 1)
                    if not agents_str:
                        continue
                    
                    agent_ids = agents_str.split(',')
                    if link_id not in combined_indices:
                        combined_indices[link_id] = []
                    combined_indices[link_id].extend(agent_ids)

        # Write final combined file with unique sorted agents per link
        with open(output_file, 'w', encoding='utf-8') as f:
            for link_id, agent_ids in sorted(combined_indices.items()):
                unique_agents = sorted(set(agent_ids))
                f.write(f"{link_id}:{','.join(unique_agents)}\n")

def process_scale(scale: int, network_dir: Path, population_dir: Path):
    """Process a specific scale number (01-10)"""
    network_file = network_dir / f"network-{scale:02d}.xml"
    population_file = population_dir / f"population-{scale:02d}.xml"
    
    if not network_file.exists() or not population_file.exists():
        print(f"Skipping scale {scale:02d}: Missing required pair of files")
        return
    
    # Process start index
    start_processor = IndexProcessor('start')
    start_output = Path("index/start") / f"index-start-{scale:02d}"
    start_processor.process_file(population_file, start_output)
    
    # Process pass index
    pass_processor = IndexProcessor('pass')
    pass_output = Path("index/pass") / f"index-pass-{scale:02d}"
    pass_processor.process_file(population_file, pass_output)
