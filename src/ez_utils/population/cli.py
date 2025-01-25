import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_file
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
    include_all_plans: Annotated[bool, typer.Option("--include-all-plans", help="Include all plans regardless of selected attribute")] = False,
    export_db: Annotated[bool, typer.Option("--export-db", help="Export processed data to PostgreSQL database")] = False,
):
    if not input:
        raise typer.Exit("Input population file is required")
    
    input_file = validate_population_file(input)
    population_dir = validate_output_dir(output)
    scale_list = parse_scale_list(scales)
    
    if export_db and "10" not in scales:
        scale_list.append(10)

    # Process file with chunking and progress tracking
    process_file(input_file, population_dir, scale_list, include_all_plans, export_db)
    
    console.print("[green]Population files created successfully")
    if export_db:
        console.print("[green]Data exported to database successfully")

if __name__ == "__main__":
    app()
