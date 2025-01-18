import sys
from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_gtfs_data, create_transit_schedule, create_vehicles
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
):
    if not gtfs_path or not gtfs_path.exists():
        raise typer.Exit("GTFS directory path is required and must exist")
    
    output_dir = output or Path.cwd()
    output_dir.mkdir(parents=True, exist_ok=True)
    pt_dir = output_dir / "pt"
    pt_dir.mkdir(exist_ok=True)
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        # Process GTFS data
        gtfs_task = progress.add_task("[cyan]Processing GTFS data...", total=1)
        gtfs_data = process_gtfs_data(gtfs_path)
        progress.advance(gtfs_task)
        
        # Create schedule files
        schedule_task = progress.add_task("[cyan]Creating schedule files...", total=4)
        
        # Bus schedule
        progress.update(schedule_task, description="Creating bus schedule...")
        bus_schedule = create_transit_schedule(gtfs_data, "bus")
        bus_schedule_file = pt_dir / "bus_schedule.xml"
        bus_schedule_file.write_text(bus_schedule, encoding='utf-8')
        progress.advance(schedule_task)
        
        # Metro schedule
        progress.update(schedule_task, description="Creating metro schedule...")
        metro_schedule = create_transit_schedule(gtfs_data, "metro")
        metro_schedule_file = pt_dir / "metro_schedule.xml"
        metro_schedule_file.write_text(metro_schedule, encoding='utf-8')
        progress.advance(schedule_task)
        
        # Bus vehicles
        progress.update(schedule_task, description="Creating bus vehicles...")
        bus_vehicles = create_vehicles(gtfs_data, "bus")
        bus_vehicles_file = pt_dir / "bus_vehicles.xml"
        bus_vehicles_file.write_text(bus_vehicles, encoding='utf-8')
        progress.advance(schedule_task)
        
        # Metro vehicles
        progress.update(schedule_task, description="Creating metro vehicles...")
        metro_vehicles = create_vehicles(gtfs_data, "metro")
        metro_vehicles_file = pt_dir / "metro_vehicles.xml"
        metro_vehicles_file.write_text(metro_vehicles, encoding='utf-8')
        progress.advance(schedule_task)
    
    console.print("[green]Public transportation schedules and vehicles created successfully from GTFS data")

if __name__ == "__main__":
    app()
