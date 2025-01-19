from bs4 import BeautifulSoup
from pathlib import Path
from typing import Dict, List, Tuple
from .models import ChunkResult
from .base_processor import BaseProcessor

class IndexProcessor(BaseProcessor):
    def __init__(self):
        super().__init__()
        self.start_dir = Path("index/start")
        self.pass_dir = Path("index/pass")
        self.start_dir.mkdir(parents=True, exist_ok=True)
        self.pass_dir.mkdir(parents=True, exist_ok=True)

    def process_chunk(self, chunk_file: Path) -> Tuple[Path, Path]:
        """Process XML chunk and return paths to start and pass index chunks"""
        result = self._process_xml_chunk(chunk_file)
        
        # Write both start and pass indices
        start_chunk = self.temp_dir / f"start_{chunk_file.name}"
        pass_chunk = self.temp_dir / f"pass_{chunk_file.name}"
        
        self._write_index_file(start_chunk, result.start_indices)
        self._write_index_file(pass_chunk, result.pass_indices)
        
        return start_chunk, pass_chunk

    def _process_xml_chunk(self, chunk_file: Path) -> ChunkResult:
        start_indices: Dict[str, List[str]] = {}
        pass_indices: Dict[str, List[str]] = {}
        
        with open(chunk_file, 'r', encoding='utf-8') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
            
        for person in soup.find_all('person'):
            agent_id = person['id']
            plan = person.find('plan', selected="yes") or person.find('plan')
            if not plan:
                continue

            # Find first home activity
            activities = plan.find_all('act')
            first_home = None
            for act in activities:
                if act.get('type') == 'home' and 'link' in act.attrs:
                    first_home = act
                    break

            if first_home:
                start_link = first_home['link']
                if start_link not in start_indices:
                    start_indices[start_link] = []
                start_indices[start_link].append(agent_id)

            # Process route links
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

def process_scale(scale: int, network_dir: Path, population_dir: Path):
    """Process a specific scale number (01-10)"""
    network_file = network_dir / f"network-{scale:02d}.xml"
    population_file = population_dir / f"population-{scale:02d}.xml"
    
    if not network_file.exists() or not population_file.exists():
        print(f"Skipping scale {scale:02d}: Missing required pair of files")
        return
    
    # Process both indices in one pass
    processor = IndexProcessor()
    start_output = Path("index/start") / f"index-start-{scale:02d}"
    pass_output = Path("index/pass") / f"index-pass-{scale:02d}"
    processor.process_file(population_file, start_output, pass_output)
