export function printUsage() {
  console.log(`
Usage: ez-utils scale population <input-file> [options]

Arguments:
  <input-file>    Path to MATSim population XML file (required)

Options:
  --clean         Clean up SQLite database before processing
  --scales, -s    Comma-separated percentage values 1-10 (e.g., --scales 1,2,5,10)
  --help, -h      Show this help message

Examples:
  ez-utils scale population population.xml
  ez-utils scale population population.xml --clean
  ez-utils scale population population.xml --scales 1,2,3,4,5 --clean
  ez-utils scale population population.xml -s 1,5,10
`);
}