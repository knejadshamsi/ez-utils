import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_population, create_scaled_population, population_to_xml
from .utils import validate_population_file, validate_output_dir, parse_scale_list
from .help import print_population_help

console = Console()
app = typer.Typer(help="Create and manage scaled population files for transportation simulations")

@app.command(
    help="Generate scaled population files while preserving geographic and demographic distributions",
    name="create",
    rich_help_panel="Population Commands"
)
@app.command(print_population_help(sys.argv))
def create_population(
    input: Annotated[Path, typer.Argument(help="Input MATSim population XML file with agent plans and attributes")] = None,
    output: Annotated[Path, typer.Option("--output", "-o", help="Output directory for scaled population files")] = None,
    scales: Annotated[str, typer.Option("--scales", "-s", help="Comma-separated list of scale percentages (e.g., 1,5,10)")] = "1,5,10",
):
    if not input:
        raise typer.Exit("Input population file is required")
    
    population_soup = validate_population_file(input)
    population_dir = validate_output_dir(output)
    scale_list = parse_scale_list(scales)
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        # Process base population
        base_task = progress.add_task("[cyan]Processing base population...", total=2)
        
        # Read file and calculate current scale
        progress.update(base_task, description="Reading population file...")
        population_data = process_population(population_soup)
        progress.advance(base_task)
        
        # Calculate current scale
        progress.update(base_task, description="Calculating population scale...")
        console.print(f"[cyan]Current population scale: {population_data.current_scale:.1f}%")
        progress.advance(base_task)
        
        # Create scaled versions
        scale_task = progress.add_task("[cyan]Creating scaled populations...", total=len(scale_list))
        for scale in scale_list:
            progress.update(scale_task, description=f"Creating population-{scale:02d}.xml...")
            
            # Create scaled population
            scaled_population = create_scaled_population(population_data, scale)
            population_xml = population_to_xml(scaled_population)
            
            # Write scaled population file
            output_file = population_dir / f"population-{scale:02d}.xml"
            output_file.write_text(population_xml, encoding='utf-8')
            
            progress.advance(scale_task)
    
    console.print("[green]Population files created successfully")

if __name__ == "__main__":
    app()
