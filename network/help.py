from rich.padding import Padding
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console

def print_help():
    line1 = (
        "This command generates scaled network files and updates the link_coordinates database table.\n"
        "It processes network data to create multiple scale versions while maintaining network topology\n"
        "and updates a PostGIS-enabled database with geometric information for network analysis."
    )
    
    line2 = (
        "[bold]Generated Files:[/bold]\n"
        "  • Network files scaled from 1% to 10% (e.g., network-01.xml through network-10.xml)\n"
        "  • Each file maintains network connectivity and properties\n\n"
        "[bold]Directory Structure:[/bold]\n"
        "  network-inputs/\n"
        "  ├── network-01.xml\n"
        "  ├── network-05.xml\n"
        "  └── network-10.xml"
    )
    
    line3 = "The network files follow this XML structure:"
    
    code = '''<?xml version="1.0" encoding="utf-8"?>
<network>
    <nodes>
        <node id="1" x="346519.0" y="5053098.2"/>
    </nodes>
    <links capperiod="01:00:00">
        <link id="1" from="1" to="2" length="100.0"
              freespeed="13.89" capacity="600.0"
              permlanes="1.0" oneway="1"
              modes="car"/>
    </links>
</network>'''
    
    syntax = Padding(Syntax(code, "xml", theme="native", line_numbers=True), (1, 0))
    
    line4 = "\n[bold]Database Table Structure:[/bold]"
    
    table_sql = '''CREATE TABLE link_coordinates (
    link_id TEXT PRIMARY KEY,
    from_node GEOMETRY(Point, 4326) NOT NULL,
    to_node GEOMETRY(Point, 4326) NOT NULL,
    length DOUBLE PRECISION,
    freespeed DOUBLE PRECISION
);'''
    
    sql_syntax = Padding(Syntax(table_sql, "sql", theme="native"), (1, 0))
    
    line5 = (
        "\n[bold]Processing Steps:[/bold]\n"
        "1. Parse and validate network XML\n"
        "2. Process nodes and create geometry\n"
        "3. Process links and attributes\n"
        "4. Update database with geometric data\n\n"
        "[red bold]NOTE:[/red bold] Requires PostGIS extension and valid database credentials in [bright_cyan].env[/bright_cyan]"
    )
    
    help_text = f"{line1}\n\n{line2}\n\n{line3}\n\n{syntax}\n\n{line4}\n\n{sql_syntax}\n\n{line5}"
    console = Console()
    console.print(Panel(help_text, title="[bold green]Network[/bold green]", border_style="green"))
