from rich.padding import Padding
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console

def print_help():
    line1 = (
        "This command creates scaled population files based on real-world population data.\n"
        "It processes agent information to create multiple scale versions while preserving\n"
        "geographic distribution and demographic characteristics."
    )
    
    line2 = (
        "[bold]Generated Files:[/bold]\n"
        "  • Population files scaled from 1% to 10% (e.g., population-01.xml through population-10.xml)\n"
        "  • Each file maintains geographic density and demographic ratios\n\n"
        "[bold]Directory Structure:[/bold]\n"
        "  population-inputs/\n"
        "  ├── population-01.xml\n"
        "  ├── population-05.xml\n"
        "  └── population-10.xml"
    )
    
    line3 = "The population files follow this XML structure:"
    
    code = '''<?xml version="1.0" encoding="utf-8"?>
<population>
    <person id="1">
        <attributes>
            <attribute name="age" class="java.lang.Integer">30</attribute>
            <attribute name="sex" class="java.lang.Integer">1</attribute>
        </attributes>
        <plan selected="yes">
            <activity type="home" x="322952.87" y="5084462.54" 
                      end_time="07:00:00"/>
            <leg mode="pt" dep_time="07:00:00" 
                 trav_time="00:41:00"/>
        </plan>
    </person>
</population>'''
    
    syntax = Padding(Syntax(code, "xml", theme="native", line_numbers=True), (1, 0))
    
    line4 = (
        "\n[bold]Scaling Logic:[/bold]\n"
        "1. Read TOTAL_POPULATION from .env\n"
        "2. Calculate current scale from base population\n"
        "3. Remove/scale agents to match target percentage\n"
        "4. Preserve geographic distribution\n\n"
        "[bold]Required Environment Variables:[/bold]\n"
        "  • TOTAL_POPULATION: Real-world population size\n"
        "  • CHUNK_SIZE: Processing batch size\n\n"
        "[red bold]NOTE:[/red bold] Ensure .env file contains required variables"
    )
    
    help_text = f"{line1}\n\n{line2}\n\n{line3}\n\n{syntax}\n\n{line4}"
    console = Console()
    console.print(Panel(help_text, title="[bold magenta]Population[/bold magenta]", border_style="magenta"))
