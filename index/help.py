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
    
    # Example index files with syntax highlighting
    start_example = '''# index-start-05
1:agent1,agent4,agent7    # Agents starting at link 1
2:agent2,agent5           # Agents starting at link 2
3:agent3,agent6          # Agents starting at link 3'''

    pass_example = '''# index-pass-05
1:agent1,agent2,agent4    # Agents passing through link 1
2:agent1,agent3,agent5    # Agents passing through link 2
3:agent2,agent4,agent6    # Agents passing through link 3'''
    
    overview = (
        "[bold yellow]Overview[/bold yellow]\n"
        "The Index module creates two types of indices for each scale (01-10):\n"
        "1. [yellow]Start Index[/yellow]: Maps which agents start their journey from each link\n"
        "2. [yellow]Pass Index[/yellow]: Maps which agents pass through each link in their routes\n\n"
        "[bold yellow]Command Usage[/bold yellow]\n"
        "[yellow]python main.py index create <network_dir> <population_dir>[/yellow]\n\n"
        "[bold yellow]Required Input Directories[/bold yellow]\n"
        "You must provide two directory paths containing the input files:\n\n"
        "1. [yellow]Network Directory[/yellow]:\n"
        "   • Must contain network XML files\n"
        "   • Files [bold red]MUST[/bold red] be named exactly: [yellow]network-XX.xml[/yellow]\n"
        "   • Where XX is a two-digit number from 01 to 10 (e.g., network-01.xml, network-02.xml)\n\n"
        "2. [yellow]Population Directory[/yellow]:\n"
        "   • Must contain population XML files\n"
        "   • Files [bold red]MUST[/bold red] be named exactly: [yellow]population-XX.xml[/yellow]\n"
        "   • Where XX is a two-digit number from 01 to 10 (e.g., population-01.xml, population-02.xml)\n\n"
        "[bold yellow]Important Note[/bold yellow]\n"
        "For each scale XX, both [yellow]network-XX.xml[/yellow] and [yellow]population-XX.xml[/yellow] must exist.\n"
        "For example: if [yellow]network-05.xml[/yellow] exists, [yellow]population-05.xml[/yellow] must also exist.\n\n"
        "[bold yellow]Output Files[/bold yellow]\n"
        "For each scale XX (01-10), creates:\n"
        "1. [yellow]index/start/index-start-XX[/yellow]: Start location indices\n"
        "2. [yellow]index/pass/index-pass-XX[/yellow]: Route path indices\n\n"
        "[bold yellow]Start Index Format[/bold yellow]\n"
    )
    
    pass_format = (
        "\n[bold yellow]Pass Index Format[/bold yellow]\n"
    )
    
    help_content = Group(
        overview,
        Syntax(start_example, "ini", theme="monokai", line_numbers=True),
        pass_format,
        Syntax(pass_example, "ini", theme="monokai", line_numbers=True)
    )
    
    console.print(Panel(help_content, title="[bold]Index Module[/bold]", 
                       border_style="yellow", padding=(1, 2)))
    
    return ""
