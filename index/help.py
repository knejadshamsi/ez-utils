import sys
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console
from rich.text import Text
from rich.console import Group

def print_index_help(args=None):
    args = args or sys.argv
    if not (("--help" in args or "-h" in args) and len(args) > 1 and args[1] == "index"):
        return ""

    console = Console()
    
    # Example index file with syntax highlighting
    index_example = '''# link-agents-XX.idx
[metadata]
scale=5
created=2024-01-20T14:30:00
version=1.0

[mappings]
1,"1,4,7,9"    # Agents on link 1
2,"2,5,8"      # Agents on link 2
3,"3,6,10"     # Agents on link 3
4,"11,12,15"   # Agents on link 4'''
    
    overview = (
        "[bold yellow]Overview[/bold yellow]\n"
        "The Index module creates optimized lookup structures that map relationships between agents and network elements. "
        "These indexes enable efficient simulation processing by providing quick access to spatial relationships, "
        "enabling quick testing with [yellow]python main.py index create ./networks ./populations -s 2,4,6,8[/yellow] for custom scales.\n\n"
        "[bold yellow]Expected Output[/bold yellow]\n"
        "Generates [yellow]link-agents-XX.idx[/yellow] files where XX is the scale percentage (e.g. [yellow]link-agents-05.idx[/yellow] for 5% scale). "
        "Each file provides optimized lookups with O(1) link access and O(log n) agent access.\n\n"
        "[bold yellow]Required Input[/bold yellow]\n"
        "1. [yellow]Network Files[/yellow]: Scaled network XMLs ([yellow]network-XX.xml[/yellow])\n"
        "2. [yellow]Population Files[/yellow]: Scaled population XMLs ([yellow]population-XX.xml[/yellow])\n\n"
        "Index file format example:\n\n"
    )
    
    env_vars = (
        "\n[bold yellow]Environment Variables[/bold yellow]\n"
        "• [yellow]BATCH_SIZE[/yellow]: Processing batch size (default: 1000)\n"
        "• [yellow]MAX_MEMORY[/yellow]: Memory limit in MB (default: 4096)"
    )
    
    help_content = Group(
        overview,
        Syntax(index_example, "ini", theme="monokai", line_numbers=True),
        env_vars
    )
    
    console.print(Panel(help_content, title="[bold]Index Module[/bold]", 
                       border_style="yellow", padding=(1, 2)))
    
    return ""
