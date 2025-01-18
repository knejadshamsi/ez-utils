import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import create_index_for_scale
from .utils import (
    validate_network_dir,
    validate_population_dir,
    validate_output_dir,
    parse_scale_list,
    write_index_file
)
from .help import print_index_help

console = Console()
app = typer.Typer(help="Create and manage optimized lookup indexes for agent-network relationships")

@app.command(
    help="Generate index files mapping relationships between agents and network elements",
    name="create",
    rich_help_panel="Index Commands"
)
@app.command(print_index_help(sys.argv))
def create_index(
    network_dir: Annotated[Path, typer.Argument(help="Directory containing scaled network XML files (network-XX.xml)")] = None,
    population_dir: Annotated[Path, typer.Argument(help="Directory containing scaled population XML files (population-XX.xml)")] = None,
    output: Annotated[Path, typer.Option("--output", "-o", help="Output directory for generated index files")] = None,
    scales: Annotated[str, typer.Option("--scales", "-s", help="Comma-separated list of scale percentages (e.g., 1,5,10)")] = "1,5,10",
):
    if not network_dir or not population_dir:
        raise typer.Exit("Network and population directories are required")
    
    network_dir = validate_network_dir(network_dir)
    population_dir = validate_population_dir(population_dir)
    index_dir = validate_output_dir(output)
    scale_list = parse_scale_list(scales)
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        for scale in scale_list:
            task = progress.add_task(f"[cyan]Creating link-agents-{scale:02d}.idx...", total=3)
            
            # Create index for current scale
            progress.update(task, description=f"Processing scale {scale}%...")
            index = create_index_for_scale(scale, network_dir, population_dir)
            progress.advance(task)
            
            # Write index file
            progress.update(task, description=f"Writing index file...")
            index_file = index_dir / f"link-agents-{scale:02d}.idx"
            write_index_file(index.link_agents, index_file)
            progress.advance(task)
    
    console.print("[green]Index files created successfully")

if __name__ == "__main__":
    app()
