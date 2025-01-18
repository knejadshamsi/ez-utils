from rich.padding import Padding
from rich.panel import Panel
from rich.syntax import Syntax
from rich.console import Console

def print_help():
    line1 = (
        "This command generates index files that link agents to network elements.\n"
        "These indexes are crucial for efficient simulation processing and analysis,\n"
        "providing quick lookups for agent-network relationships."
    )
    
    line2 = (
        "[bold]Generated Files:[/bold]\n"
        "  • Index files for each population scale (e.g., link-agents-01.idx through link-agents-10.idx)\n"
        "  • Each file maps agents to their relevant network links\n\n"
        "[bold]Directory Structure:[/bold]\n"
        "  indexes/\n"
        "  ├── link-agents-01.idx\n"
        "  ├── link-agents-05.idx\n"
        "  └── link-agents-10.idx"
    )
    
    line3 = "The index files follow this structure:"
    
    code = '''# link-agents-XX.idx
link_id,agent_ids
1,"1,4,7,9"
2,"2,5,8"
3,"3,6,10"'''
    
    syntax = Padding(Syntax(code, "text", theme="native", line_numbers=True), (1, 0))
    
    line4 = (
        "\n[bold]Processing Steps:[/bold]\n"
        "1. Read population and network files\n"
        "2. Map agents to network links\n"
        "3. Generate optimized index structure\n"
        "4. Create scale-specific index files\n\n"
        "[red bold]NOTE:[/red bold] Requires corresponding population and network files"
    )
    
    help_text = f"{line1}\n\n{line2}\n\n{line3}\n\n{syntax}\n\n{line4}"
    console = Console()
    console.print(Panel(help_text, title="[bold yellow]Index[/bold yellow]", border_style="yellow"))
