from rich.padding import Padding
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console

def print_help():
    line1 = (
        "This command creates standardized public transportation schedule files for both bus and metro services.\n"
        "The generated XML files follow the MATSim public transportation format and are organized in the pt/ directory.\n"
        "The schedules include detailed route information, stop locations, departure times, and service frequencies."
    )
    
    line2 = (
        "[bold]Generated Files:[/bold]\n"
        "  • [yellow]bus_schedule.xml[/yellow]: Contains bus routes, stops, schedules, and frequencies\n"
        "  • [yellow]metro_schedule.xml[/yellow]: Contains metro lines, stations, schedules, and service patterns\n\n"
        "[bold]Directory Structure:[/bold]\n"
        "  pt/\n"
        "  ├── bus_schedule.xml\n"
        "  └── metro_schedule.xml"
    )
    
    line3 = "The schedule files follow this XML structure:"
    
    code = '''<?xml version="1.0" encoding="utf-8"?>
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
    
    syntax = Padding(Syntax(code, "xml", theme="native", line_numbers=True), (1, 0))
    
    line4 = (
        "\n[bold]Required Environment Variables:[/bold]\n"
        "  • Database connection details in [bright_cyan].env[/bright_cyan] file\n"
        "  • City-specific configuration settings\n\n"
        "[red bold]NOTE:[/red bold] Ensure all required shapefiles and route data are available for the target city."
    )
    
    help_text = f"{line1}\n\n{line2}\n\n{line3}\n\n{syntax}\n\n{line4}"
    console = Console()
    console.print(Panel(help_text, title="[bold cyan]Public Transportation[/bold cyan]", border_style="cyan"))
