import sys
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console
from rich.text import Text
from rich.console import Group

def print_network_help(args=None):
    args = args or sys.argv
    if not (("--help" in args or "-h" in args) and len(args) > 1 and args[1] == "network"):
        return ""

    console = Console()
    
    # Example XML code with syntax highlighting
    xml_example = '''<?xml version="1.0" encoding="utf-8"?>
<network>
    <nodes>
        <node id="1" x="346519.0" y="5053098.2"/>
        <node id="2" x="346620.3" y="5053187.9"/>
    </nodes>
    <links capperiod="01:00:00">
        <link id="1" from="1" to="2" length="100.0"
              freespeed="13.89" capacity="600.0"
              permlanes="1.0" oneway="1"
              modes="car,bike"/>
        <link id="2" from="2" to="1" length="100.0"
              freespeed="13.89" capacity="600.0"
              permlanes="1.0" oneway="1"
              modes="car,bike"/>
    </links>
</network>'''
    
    overview = (
        "[bold green]Overview[/bold green]\n"
        "The Network module manages transportation network data for simulation environments. "
        "It processes and scales network files while maintaining topology and connectivity, "
        "enabling efficient simulation testing with [green]python main.py network create input.xml -s 2,4,6,8[/green] for custom scales.\n\n"
        "[bold green]Expected Output[/bold green]\n"
        "Generates [green]network-XX.xml[/green] files where XX is the scale percentage (e.g. [green]network-05.xml[/green] for 5% scale). "
        "Each file maintains topology and connectivity while scaling network properties.\n\n"
        "[bold green]Required Input[/bold green]\n"
        "1. [green]input.xml[/green]: Base network file with format:\n\n"
    )
    
    env_vars = (
        "\n[bold green]Environment Variables[/bold green]\n"
        "• [green]DB_HOST[/green]: PostGIS database host\n"
        "• [green]DB_PORT[/green]: Database port (default: 5432)\n"
        "• [green]DB_NAME[/green]: Database name\n"
        "• [green]DB_USER[/green]: Database username\n"
        "• [green]DB_PASS[/green]: Database password"
    )
    
    help_content = Group(
        overview,
        Syntax(xml_example, "xml", theme="monokai", line_numbers=True),
        env_vars
    )
    
    console.print(Panel(help_content, title="[bold]Network Module[/bold]", 
                       border_style="green", padding=(1, 2)))
    
    return ""
