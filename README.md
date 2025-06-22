# ez-utils

A fast utility for scaling MATSim population XML files.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ez-utils.git
cd ez-utils

# Install dependencies
bun install

# Install globally
bun link
```

## Usage

```bash
ez-utils scale population <input-file> [options]
```

### Options

- `--clean` - Clean up SQLite database before processing
- `--scales, -s` - Comma-separated percentage values 1-10 (default: 1,5,10)
- `--help, -h` - Show help message

### Examples

```bash
# Scale to default percentages (1%, 5%, 10%)
ez-utils scale population population.xml

# Scale to custom percentages
ez-utils scale population population.xml --scales 1,2,3,4,5

# Clean database before processing
ez-utils scale population population.xml --clean
```

