import typer
import os
from pathlib import Path
from rich.console import Console
from rich.table import Table
from dotenv import load_dotenv, set_key

app = typer.Typer(rich_markup_mode="rich")
console = Console()

def get_env_file_path() -> Path:
    root_dir = Path(__file__).parent.parent.parent.parent
    env_path = root_dir / ".env"
    if not env_path.exists():
        example_env = root_dir / ".env.example"
        if example_env.exists():
            example_env.rename(env_path)
    return env_path

@app.command("--list")
def list_variables():
    env_path = get_env_file_path()
    load_dotenv(env_path)
    
    table = Table(title="Environment Variables")
    table.add_column("Variable", style="cyan")
    table.add_column("Value", style="green")
    
    for key, value in os.environ.items():
        if key in ["POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD", 
                  "POSTGRES_HOST", "POSTGRES_PORT", "TOTAL_POPULATION"]:
            table.add_row(key, value)
    
    console.print(table)

@app.command()
def set_variable(variable_name: str, value: str):
    env_path = get_env_file_path()
    if not env_path.exists():
        console.print(f"[red]Error: .env file not found at {env_path}[/red]")
        raise typer.Exit(1)
        
    valid_vars = ["POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD", 
                 "POSTGRES_HOST", "POSTGRES_PORT", "TOTAL_POPULATION"]
    
    if variable_name not in valid_vars:
        console.print(f"[red]Error: Invalid variable name. Valid options are: {', '.join(valid_vars)}[/red]")
        raise typer.Exit(1)
    
    set_key(env_path, variable_name, value)
    console.print(f"[green]Successfully set {variable_name}={value}[/green]")

@app.callback()
def callback():
    """
    Set or view environment variables used by ez-utils.
    
    Usage:
      ez-utils set --list                    List all environment variables
      ez-utils set VARIABLE_NAME NEW_VALUE   Set a new value for an environment variable
    """
    pass
