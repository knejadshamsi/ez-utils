import sys
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console
from rich.text import Text
from rich.console import Group

def print_pt_help(args=None):
    args = args or sys.argv
    if not (("--help" in args or "-h" in args) and len(args) > 1 and args[1] == "pt"):
        return ""

    console = Console()
    
    # Example XML code with syntax highlighting
    xml_example = '''<?xml version="1.0" encoding="utf-8"?>
<transitSchedule>
    <transitStops>
        <stopFacility id="stop1" x="346519.0" y="5053098.2">
            <name>Central Station</name>
        </stopFacility>
    </transitStops>
    <transitLine id="line1">
        <transitRoute id="route1">
            <transportMode>bus</transportMode>
            <routeProfile>
                <stop refId="stop1" departureOffset="00:00:00"/>
            </routeProfile>
            <departures>
                <departure id="1" departureTime="06:00:00"/>
            </departures>
        </transitRoute>
    </transitLine>
</transitSchedule>'''
    
    overview = (
        "[bold cyan]Overview[/bold cyan]\n"
        "The Public Transportation (PT) module processes GTFS data to create standardized schedule files for bus and metro services. "
        "It enables realistic public transit modeling with accurate routes, schedules, and vehicle configurations, "
        "with quick setup using [cyan]python main.py pt create-from-gtfs ./gtfs_data -o ./transit_files[/cyan].\n\n"
        "[bold cyan]Expected Output[/bold cyan]\n"
        "Generates four XML files:\n"
        "• [cyan]bus_schedule.xml[/cyan]: Bus routes, stops, and schedules\n"
        "• [cyan]metro_schedule.xml[/cyan]: Metro lines, stations, and schedules\n"
        "• [cyan]bus_vehicles.xml[/cyan]: Bus types and capacities\n"
        "• [cyan]metro_vehicles.xml[/cyan]: Metro types and capacities\n\n"
        "[bold cyan]Required Input[/bold cyan]\n"
        "1. [cyan]GTFS Data[/cyan]: Standard transit files with format:\n\n"
    )
    
    env_vars = (
        "\n[bold cyan]Environment Variables[/bold cyan]\n"
        "• [cyan]DB_HOST[/cyan]: Database server hostname\n"
        "• [cyan]DB_PORT[/cyan]: Database server port\n"
        "• [cyan]DB_NAME[/cyan]: Transit database name\n"
        "• [cyan]DB_USER[/cyan]: Database username\n"
        "• [cyan]DB_PASS[/cyan]: Database password"
    )
    
    help_content = Group(
        overview,
        Syntax(xml_example, "xml", theme="monokai", line_numbers=True),
        env_vars
    )
    
    console.print(Panel(help_content, title="[bold]Public Transportation Module[/bold]", 
                       border_style="cyan", padding=(1, 2)))
    
    return ""
