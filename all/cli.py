import typer
from pathlib import Path
from .help import HELP_TEXT
from .processors import process_files

app = typer.Typer(help="Process files using all modules together")

@app.command(help=HELP_TEXT)
def process(
    input_files: list[Path] = typer.Argument(..., help="Input files to process"),
    output_dir: Path = typer.Argument(..., help="Directory to store output files")
):
    process_files(input_files, output_dir)

if __name__ == "__main__":
    app()
