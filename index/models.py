from dataclasses import dataclass
from typing import Dict, List

@dataclass
class IndexEntry:
    link_id: str
    agent_ids: List[str]

@dataclass
class ChunkResult:
    start_indices: Dict[str, List[str]]  # link_id -> agent_ids
    pass_indices: Dict[str, List[str]]   # link_id -> agent_ids
