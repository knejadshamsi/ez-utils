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
    
    # Example XML code with syntax highlighting
    xml_example = '''<?xml version="1.0" encoding="utf-8"?>
<population>
    <person id="1">
        <attributes>
            <attribute name="age" class="java.lang.Integer">30</attribute>
            <attribute name="sex" class="java.lang.Integer">1</attribute>
        </attributes>
        <plan selected="yes">
            <activity type="home" x="322952.87" y="5084462.54" end_time="07:00:00"/>
            <leg mode="pt" dep_time="07:00:00" trav_time="00:41:00"/>
            <activity type="work" x="323456.78" y="5084789.12"/>
        </plan>
    </person>
</population>'''
    
    overview = (
        "[bold magenta]Overview[/bold magenta]\n"
        "The Population module processes real-world population data to generate scaled versions for transportation simulations. "
        "It maintains geographic distribution and demographic characteristics while creating different population sizes, "
        "enabling efficient simulation testing with [magenta]python main.py population create input.xml -s 2,4,6,8[/magenta] for custom scales.\n\n"
        "[bold magenta]Expected Output[/bold magenta]\n"
        "Generates [magenta]population-XX.xml[/magenta] files where XX is the scale percentage (e.g. [magenta]population-05.xml[/magenta] for 5% scale). "
        "Each file contains scaled population data with preserved geographic and demographic ratios.\n\n"
        "[bold magenta]Required Input[/bold magenta]\n"
        "1. [magenta]input.xml[/magenta]: Base population file with format:\n\n"
    )
    
    env_vars = (
        "\n[bold magenta]Environment Variables[/bold magenta]\n"
        "• [magenta]TOTAL_POPULATION[/magenta]: Real-world population size\n"
        "• [magenta]CHUNK_SIZE[/magenta]: Processing batch size (default: 10000)"
    )
    
    help_content = Group(
        overview,
        Syntax(xml_example, "xml", theme="monokai", line_numbers=True),
        env_vars
    )
    
    console.print(Panel(help_content, title="[bold]Population Module[/bold]", 
                       border_style="magenta", padding=(1, 2)))
    
    return ""
