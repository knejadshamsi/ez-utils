from pathlib import Path
import typer
from typing_extensions import Annotated
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from rich.console import Console
from .processors import process_network, create_scaled_network, network_to_xml
from .utils import validate_network_file, validate_output_dir, parse_scale_list
from .help import print_help

console = Console()
app = typer.Typer()

@app.command("create", help=print_help())
def create_network(
    input: Annotated[Path, typer.Argument(help="Input MATSim network XML file containing nodes and links")] = None,
    output: Annotated[Path, typer.Option("--output", "-o", help="Output directory for scaled network files and PostGIS data")] = None,
    scales: Annotated[str, typer.Option("--scales", "-s", help="Comma-separated list of scale percentages (e.g., 1,5,10)")] = "1,5,10",
):
    if not input:
        raise typer.Exit("Input network file is required")
    
    network_soup = validate_network_file(input)
    network_dir = validate_output_dir(output)
    scale_list = parse_scale_list(scales)
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TaskProgressColumn(),
        console=console
    ) as progress:
        # Process base network
        base_task = progress.add_task("[cyan]Processing base network...", total=1)
        network_data = process_network(network_soup)
        progress.advance(base_task)
        
        # Create scaled versions
        scale_task = progress.add_task("[cyan]Creating scaled networks...", total=len(scale_list))
        for scale in scale_list:
            progress.update(scale_task, description=f"Creating network-{scale:02d}.xml...")
            
            # Scale network
            scaled_network = create_scaled_network(network_data, scale)
            network_xml = network_to_xml(scaled_network)
            
            # Write scaled network file
            output_file = network_dir / f"network-{scale:02d}.xml"
            output_file.write_text(network_xml, encoding='utf-8')
            
            progress.advance(scale_task)
    
    console.print("[green]Network files created successfully")

if __name__ == "__main__":
    app()
