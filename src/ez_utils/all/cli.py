import typer
from pathlib import Path
from .help import HELP_TEXT
from .processors import process_files

app = typer.Typer(help="Process files using all modules together")

@app.command(help=HELP_TEXT)
def process(
    input_dir: Path = typer.Argument(..., help="Directory containing input files"),
    output_dir: Path = typer.Argument(..., help="Directory to store output files")
):
    if not input_dir.exists():
        typer.echo(f"Error: Input directory '{input_dir}' does not exist")
        raise typer.Exit(1)

    required_files = ['network.xml', 'population.xml']
    missing_files = [f for f in required_files if not (input_dir / f).exists()]
    
    if missing_files:
        typer.echo(f"Error: Missing required files in input directory: {', '.join(missing_files)}")
        raise typer.Exit(1)

    process_files(input_dir, output_dir)

if __name__ == "__main__":
    app()
