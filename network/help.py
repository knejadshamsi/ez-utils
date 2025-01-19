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
        "The Network module scales down transportation networks while preserving traffic behavior. "
        "It analyzes traffic flow and adjusts link speeds to maintain similar travel times across different scales. "
        "Scales range from 1-10 (representing 1% to 10% of original). "
        "Use [green]python main.py network create input.xml --scales 2,4,6[/green] for specific scales, "
        "or omit --scales for all possible scales.\n\n"
        "[bold green]Required Input[/bold green]\n"
        "1. [green]input.xml[/green]: Base network file (format shown below)\n"
        "2. [green]events.xml[/green]: Optional MATSim events file for traffic analysis\n\n"
        "[bold green]Expected Output[/bold green]\n"
        "Generates [green]network/network-XX.xml[/green] files where XX is the scale (01-10). "
        "Link speeds are adjusted based on traffic density to maintain similar travel times.\n\n"
        "[bold green]Scale Behavior[/bold green]\n"
        "• Only scales down (1-10)\n"
        "• Adjusts link speeds based on traffic density\n"
        "• Preserves network topology\n"
        "• Maintains relative travel times\n\n"
        "[bold green]Network File Format[/bold green]\n"
    )
    
    help_content = Group(
        overview,
        Syntax(xml_example, "xml", theme="monokai", line_numbers=True)
    )
    
    console.print(Panel(help_content, title="[bold]Network Module[/bold]", 
                       border_style="green", padding=(1, 2)))
    
    return ""
