import typer
from typing_extensions import Annotated
from pathlib import Path
from typing import Optional
from rich.console import Console
from rich.panel import Panel

from pt import pt_cli
from network import network_cli
from population import population_cli
from index import index_cli

console = Console()
app = typer.Typer(rich_markup_mode="rich", no_args_is_help=True)

def print_root_help():
    help_text = (
        "[bold]Available Commands:[/bold]\n\n"
        "  [cyan]pt[/cyan]         Public transportation commands for processing GTFS data\n"
        "  [cyan]network[/cyan]    Network commands for processing and scaling network data\n"
        "  [cyan]population[/cyan] Population commands for processing and scaling population data\n"
        "  [cyan]index[/cyan]      Index commands for creating agent-network relationships\n\n"
        "[bold]Usage:[/bold]\n"
        "  ez-utils [command] --help     Show help for specific command\n"
        "  ez-utils [command] [options]  Run command with options"
    )
    console.print(Panel(help_text, title="[bold magenta]EZ-Utils CLI[/bold magenta]", border_style="magenta"))
    raise typer.Exit()

@app.callback(invoke_without_command=True)
def main(ctx: typer.Context):
    if ctx.invoked_subcommand is None:
        print_root_help()

# Subcommands
app.add_typer(pt_cli, name="pt", help="Public transportation commands for processing GTFS data and creating schedules")
app.add_typer(network_cli, name="network", help="Network commands for processing and scaling network data")
app.add_typer(population_cli, name="population", help="Population commands for processing and scaling population data")
app.add_typer(index_cli, name="index", help="Index commands for creating agent-network relationships")

if __name__ == "__main__":
    app()
