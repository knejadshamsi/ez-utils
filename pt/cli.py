import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_file
from .help import print_pt_help

console = Console()
app = typer.Typer(help="Process GTFS data to create standardized public transportation schedules")

@app.command(
    help="Create public transportation schedules and vehicle configurations from GTFS data",
    name="create-from-gtfs",
    rich_help_panel="Public Transportation Commands"
)
@app.command(print_pt_help(sys.argv))
def create_pt_from_gtfs(
    gtfs_path: Annotated[Path, typer.Argument(help="Path to GTFS directory containing routes.txt, trips.txt, stops.txt, etc.")] = None,
    output: Annotated[Path, typer.Option("--output", "-o", help="Output directory for generated schedule and vehicle files")] = None,
    mode: Annotated[str, typer.Option("--mode", "-m", help="Transport mode: 'bus' or 'metro'")] = "bus",
    day: Annotated[str, typer.Option("--day", "-d", help="Service day (monday-friday). If not provided, will prompt for selection.")] = None,
):
    if not gtfs_path or not gtfs_path.exists():
        raise typer.Exit("GTFS directory path is required and must exist")
    
    if mode not in ["bus", "metro"]:
        raise typer.Exit("Mode must be either 'bus' or 'metro'")
    
    if day and day.lower() not in ["monday", "tuesday", "wednesday", "thursday", "friday"]:
        raise typer.Exit("Day must be one of: monday, tuesday, wednesday, thursday, friday")
    
    output_dir = output or Path.cwd()
    output_dir.mkdir(parents=True, exist_ok=True)
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        task = progress.add_task(f"[cyan]Processing {mode} schedules...", total=1)
        
        try:
            process_file(
                input_file=gtfs_path,
                output_file=output_dir,
                mode=mode,
                selected_day=day.lower() if day else None
            )
            progress.advance(task)
            
            console.print(f"[green]Created {mode} schedule and vehicle files in {output_dir}")
            console.print(f"[green]Files created:")
            console.print(f"[green]- transitSchedule.xml")
            console.print(f"[green]- vehicles.xml")
            
        except Exception as e:
            console.print(f"[red]Error: {str(e)}")
            raise typer.Exit(1)

if __name__ == "__main__":
    app()
