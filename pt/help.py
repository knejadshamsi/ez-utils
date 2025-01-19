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
    
    schedule_example = '''<?xml version="1.0" encoding="utf-8"?>
<transitSchedule>
    <transitStops>
        <!-- Stop facility attributes:
             - id: Unique identifier (from GTFS stop_id)
             - x,y: UTM coordinates (converted from GTFS lat/lon)
             - linkRefId: Reference to network link (default: "dummy") -->
        <stopFacility id="stop1" x="346519.0" y="5053098.2" linkRefId="dummy">
            <name>Central Station</name>
        </stopFacility>
    </transitStops>
    <transitLine id="line1">
        <!-- Transit route attributes:
             - id: Unique identifier (from GTFS trip_id)
             - transportMode: "bus" or "metro" -->
        <transitRoute id="route1">
            <transportMode>bus</transportMode>
            <routeProfile>
                <!-- Stop attributes:
                     - refId: Reference to stopFacility
                     - arrivalOffset: Time offset from route start (HH:MM:SS)
                     - departureOffset: Time offset from route start (HH:MM:SS) -->
                <stop refId="stop1" arrivalOffset="00:00:00" departureOffset="00:00:00"/>
                <stop refId="stop2" arrivalOffset="00:10:00" departureOffset="00:10:30"/>
            </routeProfile>
            <departures>
                <!-- Departure attributes:
                     - id: Unique identifier
                     - departureTime: First stop departure (HH:MM:SS)
                     - vehicleRefId: Reference to vehicle -->
                <departure id="1" departureTime="06:00:00" vehicleRefId="bus_1"/>
            </departures>
        </transitRoute>
    </transitLine>
</transitSchedule>'''

    vehicles_example = '''<?xml version="1.0" encoding="utf-8"?>
<vehicleDefinitions>
    <!-- Vehicle type attributes:
         - id: Mode identifier ("bus" or "metro")
         - seats: Seated capacity
         - standingRoom: Standing capacity
         - length/width: Vehicle dimensions (meters)
         - maximumVelocity: Max speed (meters/second) -->
    <vehicleType id="bus">
        <capacity>
            <seats>50</seats>
            <standingRoom>30</standingRoom>
        </capacity>
        <length>12.0</length>
        <width>2.5</width>
        <maximumVelocity>13.89</maximumVelocity>
    </vehicleType>
    
    <!-- Vehicle instance attributes:
         - id: Unique identifier (mode_tripId)
         - type: Reference to vehicleType -->
    <vehicle id="bus_1" type="bus"/>
</vehicleDefinitions>'''
    
    overview = (
        "[bold cyan]Overview[/bold cyan]\n"
        "The Public Transportation (PT) module converts GTFS data into MATSim-compatible transit schedules. "
        "It processes routes, stops, and schedules for a single weekday, generating XML files for either bus or metro services.\n\n"
        "[bold cyan]Usage[/bold cyan]\n"
        "Basic command: [cyan]python main.py pt create-from-gtfs ./gtfs_data -o ./output -m bus -d monday[/cyan]\n\n"
        "Options:\n"
        "• [cyan]-m, --mode[/cyan]: Transport mode ('bus' or 'metro')\n"
        "  - bus: GTFS route_type 3, 700-703\n"
        "  - metro: GTFS route_type 1, 401\n"
        "• [cyan]-d, --day[/cyan]: Service day (monday-friday). If not provided, will prompt for selection\n"
        "• [cyan]-o, --output[/cyan]: Output directory (default: current directory)\n\n"
        "[bold cyan]Output Files[/bold cyan]\n"
        "1. transitSchedule.xml:\n"
        "   • Transit stops with UTM coordinates\n"
        "   • Route definitions and profiles\n"
        "   • Time-based schedules\n"
        "   • Vehicle references\n\n"
        "2. vehicles.xml:\n"
        "   • Vehicle type definitions\n"
        "   • Capacity specifications\n"
        "   • Physical dimensions\n"
        "   • Speed limitations\n\n"
        "[bold cyan]Vehicle Specifications[/bold cyan]\n"
        "Bus:\n"
        "• Capacity: 50 seats, 30 standing\n"
        "• Dimensions: 12.0m x 2.5m\n"
        "• Max speed: 50 km/h (13.89 m/s)\n\n"
        "Metro:\n"
        "• Capacity: 400 seats, 800 standing\n"
        "• Dimensions: 150.0m x 3.2m\n"
        "• Max speed: 80 km/h (22.22 m/s)\n\n"
        "[bold cyan]Required Input[/bold cyan]\n"
        "GTFS files required in input directory:\n"
        "• routes.txt: Transit routes with route_type\n"
        "• trips.txt: Vehicle trips linking routes and services\n"
        "• stop_times.txt: Stop arrival/departure times\n"
        "• stops.txt: Stop locations (lat/lon)\n"
        "• calendar.txt: Weekly service patterns\n"
        "• calendar_dates.txt: Service exceptions\n\n"
        "[bold cyan]transitSchedule.xml Format[/bold cyan]\n"
    )
    
    vehicles_section = (
        "\n[bold cyan]vehicles.xml Format[/bold cyan]\n"
    )
    
    help_content = Group(
        overview,
        Syntax(schedule_example, "xml", theme="monokai", line_numbers=True),
        vehicles_section,
        Syntax(vehicles_example, "xml", theme="monokai", line_numbers=True)
    )
    
    console.print(Panel(help_content, title="[bold]Public Transportation Module[/bold]", 
                       border_style="cyan", padding=(1, 2)))
    
    return ""
