import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_scale
from .utils import validate_network_dir, validate_population_dir
from .help import print_index_help

console = Console()
app = typer.Typer(help="Create indices for agent start locations and route paths")

@app.command(
    help="Generate start and pass indices for each scale",
    name="create"
)
@app.command(print_index_help(sys.argv))
def create_index(
    network_dir: Annotated[Path, typer.Argument(help="Directory containing network XML files (network-XX.xml)")] = None,
    population_dir: Annotated[Path, typer.Argument(help="Directory containing population XML files (population-XX.xml)")] = None,
):
    if not network_dir or not population_dir:
        console.print("[red]Network and population directories are required")
        raise typer.Exit(1)
    
    # Validate directories and their contents
    network_dir = validate_network_dir(network_dir)
    population_dir = validate_population_dir(population_dir)

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        task = progress.add_task("[cyan]Processing scales...", total=10)
        
        for scale in range(1, 11):
            progress.update(task, description=f"Processing scale {scale:02d}...")
            process_scale(scale, network_dir, population_dir)
            progress.advance(task)
    
    console.print("[green]Index files created successfully")

if __name__ == "__main__":
    app()
