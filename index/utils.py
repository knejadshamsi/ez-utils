from pathlib import Path
import typer
from typing import Dict, List

def validate_network_dir(network_dir: Path) -> Path:
    """Validate network directory exists and contains required files"""
    if not network_dir.exists():
        raise typer.Exit(f"Network directory not found: {network_dir}")
    
    # Check for at least one network file
    has_network_file = False
    for i in range(1, 11):
        if (network_dir / f"network-{i:02d}.xml").exists():
            has_network_file = True
            break
    
    if not has_network_file:
        raise typer.Exit(f"No network-XX.xml files found in: {network_dir}")
    
    return network_dir

def validate_population_dir(population_dir: Path) -> Path:
    """Validate population directory exists and contains required files"""
    if not population_dir.exists():
        raise typer.Exit(f"Population directory not found: {population_dir}")
    
    # Check for at least one population file
    has_population_file = False
    for i in range(1, 11):
        if (population_dir / f"population-{i:02d}.xml").exists():
            has_population_file = True
            break
    
    if not has_population_file:
        raise typer.Exit(f"No population-XX.xml files found in: {population_dir}")
    
    return population_dir

def write_index_file(index_data: Dict[str, List[str]], output_file: Path) -> None:
    """Write index data to file in format: link_id:agent1,agent2,..."""
    with open(output_file, 'w', encoding='utf-8') as f:
        for link_id, agent_ids in sorted(index_data.items()):
            unique_agents = sorted(set(agent_ids))
            f.write(f"{link_id}:{','.join(unique_agents)}\n")
