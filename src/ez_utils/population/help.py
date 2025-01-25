import sys
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console
from rich.text import Text
from rich.console import Group

def print_population_help(args=None):
    args = args or sys.argv
    if not (("--help" in args or "-h" in args) and len(args) > 1 and args[1] == "population"):
        return ""

    console = Console()
    
    xml_example = '''<?xml version="1.0" encoding="utf-8"?>
<population>
    <person id="1">
        <plan selected="yes">
            <activity type="home" x="322952.87" y="5084462.54" end_time="07:00:00"/>
            <leg mode="pt" dep_time="07:00:00" trav_time="00:41:00"/>
            <activity type="work" x="323456.78" y="5084789.12"/>
        </plan>
    </person>
</population>'''
    
    overview = (
        "[bold magenta]Overview[/bold magenta]\n"
        "The Population module scales down population data while preserving geographic density. "
        "It analyzes first activity locations to maintain relative population distribution across areas. "
        "Scales range from 1-10 (representing 1% to 10% of original). "
        "Use [magenta]python main.py population create input.xml --scales 2,4,6[/magenta] for specific scales, "
        "or omit --scales for all possible scales.\n\n"
        "[bold magenta]Required Input[/bold magenta]\n"
        "1. [magenta]input.xml[/magenta]: Base population file (format shown below)\n\n"
        "[bold magenta]Expected Output[/bold magenta]\n"
        "Generates [magenta]population/population-XX.xml[/magenta] files where XX is the scale (01-10). "
        "Population is reduced while maintaining geographic density distribution.\n\n"
        "[bold magenta]Scale Behavior[/bold magenta]\n"
        "• Only scales down (1-10)\n"
        "• Preserves population density\n"
        "• Maintains activity patterns\n"
        "• Uses first activity location for density calculation\n\n"
        "[bold magenta]Plan Selection[/bold magenta]\n"
        "• By default, uses plan with selected=\"yes\" attribute\n"
        "• Falls back to first plan if no selected plan exists\n"
        "• Use --include-all-plans flag to always use first plan\n\n"
        "[bold magenta]Database Export[/bold magenta]\n"
        "• Use --export-db flag to store data in PostgreSQL\n"
        "• Creates/updates required database tables:\n"
        "  - agents: Stores agent data with start coordinates and percentages\n"
        "  - link_coordinates: Stores network link data with passing agents\n"
        "• Automatically includes 10% population if not in scales\n\n"
        "[bold magenta]Environment Variables[/bold magenta]\n"
        "• [magenta]TOTAL_POPULATION[/magenta]: Real-world total population size\n"
        "• [magenta]POSTGRES_DB[/magenta]: Database name\n"
        "• [magenta]POSTGRES_USER[/magenta]: Database username\n"
        "• [magenta]POSTGRES_PASSWORD[/magenta]: Database password\n"
        "• [magenta]POSTGRES_HOST[/magenta]: Database host (default: localhost)\n"
        "• [magenta]POSTGRES_PORT[/magenta]: Database port (default: 5432)\n\n"
        "[bold magenta]Population File Format[/bold magenta]\n"
    )
    
    help_content = Group(
        overview,
        Syntax(xml_example, "xml", theme="monokai", line_numbers=True)
    )
    
    console.print(Panel(help_content, title="[bold]Population Module[/bold]", 
                       border_style="magenta", padding=(1, 2)))
    
    return ""
