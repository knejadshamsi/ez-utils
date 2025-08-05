# EZ-Utils 🚀

A comprehensive toolkit for MATSim simulation input file processing, featuring both command-line and graphical interfaces for population, transit and network analysis.

## ✨ Features

### Dual Interface Design
- **CLI/TUI Mode**: Command-line tools with elegant terminal interfaces (thanks to [Charm](https://charm.sh))
- **GUI Mode**: Desktop application with interactive map-based editing (thanks to [Wails](https://wails.io))

### Data Processing
- **GTFS Processing**: Convert and process General Transit Feed Specification data and edit them
- **Population Scaling**: Scale MATSim population data from 1-10% with parallel processing or edit them directly
- **Network Editing**: Interactive road network topology editing
- **Transit Vehicle Management**: Create and manage transit vehicle definitions

### Performance Optimized
- **Parallel Processing**: Configurable worker pools for large dataset handling
- **Database Support**: SQLite (local) and PostgreSQL (external) integration
- **Chunked Processing**: Memory-efficient processing of large files
- **Spatial Indexing**: Optimized geographic queries and CRUD operationation

## 🛠️ Technology Stack

- **Backend**: Go 1.23+ with Wails v2 framework
- **CLI/TUI**: Charm libraries (Bubble Tea, Lipgloss, Huh)
- **Frontend**: Svelte 5 + TypeScript + Vite
- **Mapping**: Leaflet.js with interactive editing capabilities
- **Database**: SQLite (local), PostgreSQL (optional)
- **UI**: Tailwind CSS + Flowbite components

## 📋 Prerequisites

### For Linux
```bash
# Install WebKit2GTK development libraries (required for GUI)
sudo apt install libwebkit2gtk-4.1-dev  # Ubuntu/Debian
# or
sudo dnf install webkit2gtk4.1-devel    # Fedora/RHEL

# Go 1.23 or later
# Node.js 18+ and npm (for frontend development)
```

## 🚀 Installation

### Install from Source
```bash
# Clone the repository
git clone https://github.com/your-username/ez-utils.git
cd ez-utils

# Build the application
wails build -tags=wails,production,webkit2_41

# Install to system PATH
go install -tags=wails,production,webkit2_41

# Verify installation
ez-utils --help
```

## 🎯 Quick Start

### Test Commands
```bash
# Navigate to test directory
cd test

# Test CLI commands (no GUI)
ez-utils scale
ez-utils create

# Test Transit Vehicle TUI
ez-utils create tv transitVehicles.xml

# Test GUI application
ez-utils edit population test_pop.xml
```

## 📚 Usage

### Population Management
```bash
# Scale population data with database integration
ez-utils scale population population.xml --db

# Scale to 10% with cleanup
ez-utils scale population population.xml --db -s 0.1 --clean

# GUI-based population editing
ez-utils edit population test_pop.xml
```

### Transit Data Processing
```bash
# Process GTFS data for all time periods
ez-utils create pt ./gtfs-data

# Process for work days only
ez-utils create pt ./gtfs-data work

# Process for specific date
ez-utils create pt ./gtfs-data 15-04-25

# GUI-based transit editing
ez-utils edit pt transit_mtl.xml
```

### Network Editing
```bash
# Launch network editor
ez-utils edit network network_mtl.xml
```

### Transit Vehicle Management
```bash
# Create/edit transit vehicles (TUI interface)
ez-utils create tv transitVehicles.xml
```

## 🔧 Development

### Development Server
```bash
# Start development server with hot reload
wails dev -tags='wails,!production,webkit2_41'
```

### Configuration
Edit `config.yaml` to customize:
- Database connections
- Worker pool settings
- Processing parameters
- Output directories

## 🌍 GUI Workflows

### Population Editing
1. **View Mode**: Browse population data with map visualization
2. **Edit Mode**: Modify person locations and attributes
3. **Zone Management**: Create geographic zones for analysis
4. **Plan Editing**: Modify individual person travel plans

### Network Editing
1. **View Mode**: Display road network topology
2. **Edit Mode**: Modify nodes and links
3. **Create Mode**: Add new network elements

### Public Transport Editing
1. **Mode Selection**: Choose BUS, METRO, or TRAM
2. **Line Management**: Edit transit lines and routes
3. **Stop Management**: Modify stop locations
4. **Schedule Editing**: Update departure times

## 🐛 Troubleshooting

### Common Issues

1. **WebKit2GTK Missing**: Install required development libraries
2. **Permission Denied**: Ensure binary is in PATH and executable
3. **GUI Won't Launch**: Check WebKit dependencies and display configuration
4. **Database Connection**: Verify PostgreSQL settings in config.yaml