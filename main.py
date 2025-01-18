import typer
from typing_extensions import Annotated
from pathlib import Path
from typing import Optional
from rich.console import Console

from pt import pt_cli
from network import network_cli
from population import population_cli
from index import index_cli

console = Console()
app = typer.Typer(rich_markup_mode="rich")

# Subcommands
app.add_typer(pt_cli, name="pt", help="Public transportation commands for processing GTFS data and creating schedules")
app.add_typer(network_cli, name="network", help="Network commands for processing and scaling network data")
app.add_typer(population_cli, name="population", help="Population commands for processing and scaling population data")
app.add_typer(index_cli, name="index", help="Index commands for creating agent-network relationships")

if __name__ == "__main__":
    app()
