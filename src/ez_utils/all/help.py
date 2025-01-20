HELP_TEXT = """
Process files using all modules (network, population, index, pt) together.

Usage:
    ez-utils all [OPTIONS] INPUT_DIR OUTPUT_DIR

Arguments:
    INPUT_DIR   Directory containing input files:
                - network.xml: Network configuration file
                - population.xml: Population data file
                - gtfs/ (optional): Directory containing GTFS files for PT processing

    OUTPUT_DIR  Directory to store output files. Will create:
                - network/: Scaled network files (01-10)
                - population/: Scaled population files (01-10)
                - index/: Index files for each scale
                - pt/ (if GTFS provided): Transit schedule and vehicles

Options:
    --help  Show this message and exit
"""
