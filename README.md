# ez-utils

Desktop editor for MATSim transit, network, and population XML files. Import, visually edit, and export with full round-trip fidelity.

## What it does

- **Import** MATSim network, population, and transit-schedule XML files into per-source SQLite databases
- **Edit visually** on a map: drag nodes, click to add persons/activities, rename lines and routes, attach link refs by clicking nearby network geometry
- **Round-trip safe**: raw XML blobs preserved byte-for-byte on parse; edits rebuild only the changed fields, leaving unrecognized MATSim attributes intact
- **Map-driven UX**: Leaflet viewport drives paginated DB queries via R*Tree-style bbox filters; drawers update with visible-area data
- **Coordinate system flexibility**: pick from 9 city presets or supply any custom proj4 string
- **Workspace persistence**: all source state saves to a single `.ez` archive (bundled SQLite + UI state JSON + optional sidecar files)

## Supported data

| Group | What's parsed | What's editable |
|---|---|---|
| **Network** | `<node>` and `<link>` elements (graph topology) | Positions, tag-level attributes (length, freespeed, capacity, permlanes, oneway, modes), attribute blocks, add/delete |
| **Population** | `<person>` with plans, activities, legs | Activity coordinates, start/end times, activity type, leg mode/duration, person attributes, add/delete persons and plans |
| **Transit schedule** | `<transitStops>`, `<transitLines>`, `<transitRoutes>`, `<minimalTransferTimes>` | Lines (add, rename, delete), routes, profile stops (reorder, per-stop flags, arrival/departure offsets), path links, transfer times, departures with vehicle picker, stop facilities (id, name, coords, link ref, stop area, blocking) |
| **Transit vehicles** (optional sidecar) | Imported alongside a transit source | Read-only viewer for now |

## Coordinate reference systems

Nine preset cities plus custom proj4 input, selectable in Settings:

- Montreal, Canada (EPSG:32188 - MTM Zone 8)
- Tehran, Iran (EPSG:32639 - UTM 39N)
- Paris, France (EPSG:2154 - Lambert-93)
- Berlin, Germany (EPSG:25833 - ETRS89 UTM 33N)
- London, UK (EPSG:27700 - British National Grid)
- New York, USA (EPSG:32618 - UTM 18N)
- Tokyo, Japan (EPSG:6677 - JGD2011 Zone 9)
- Sydney, Australia (EPSG:28356 - MGA94 Zone 56)
- Zurich, Switzerland (EPSG:2056 - LV95)
- Custom - supply any proj4 definition string + center lng/lat + label

CRS locks once sources are imported; swap CRS with an empty workspace.

## Tech stack

- **Frontend**: [Tauri 2](https://tauri.app/), [Svelte 5](https://svelte.dev/) (runes), TypeScript, [Leaflet](https://leafletjs.com/), [proj4js](http://proj4js.org/), [daisyUI](https://daisyui.com/) + Tailwind CSS, [svelte-i18n](https://github.com/kaisermann/svelte-i18n)
- **Backend**: Rust, [rusqlite](https://github.com/rusqlite/rusqlite) (SQLite), [quick-xml](https://github.com/tafia/quick-xml), [proj4rs](https://github.com/3liz/proj4rs)

## Development

### Prerequisites (all platforms)

- Rust toolchain (stable): https://rustup.rs/
- Node.js 20+
- Yarn (via `corepack enable`, or install standalone)

Plus platform-specific Tauri prerequisites:

- **Windows**: Microsoft C++ Build Tools + WebView2 (pre-installed on Windows 11). See https://tauri.app/start/prerequisites/#windows
- **Linux**: `webkit2gtk-4.1`, `libayatana-appindicator3-dev`, `librsvg2-dev`, `build-essential`, `curl`, `wget`, `file`, `libssl-dev`, `libxdo-dev`, `libgtk-3-dev`. See https://tauri.app/start/prerequisites/#linux
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)

### Run in dev

```bash
yarn install
yarn tauri dev
```

Hot-reloads frontend; rebuilds Rust backend on change.

## Build for your platform

```bash
yarn tauri build
```

Builds release artifacts under `src-tauri/target/release/bundle/`.

### Windows

Produces:
- `src-tauri/target/release/bundle/msi/ez-utils_<version>_x64_en-US.msi` - Windows installer
- `src-tauri/target/release/bundle/nsis/ez-utils_<version>_x64-setup.exe` - NSIS setup executable
- `src-tauri/target/release/ez-utils.exe` - standalone executable (portable)

**Install**: run the `.msi` or `.exe` installer and follow the prompts. Or copy `ez-utils.exe` somewhere on `PATH` and run it directly.

### Linux

Produces:
- `src-tauri/target/release/bundle/deb/ez-utils_<version>_amd64.deb` - Debian/Ubuntu package
- `src-tauri/target/release/bundle/rpm/ez-utils-<version>-1.x86_64.rpm` - RHEL/Fedora package
- `src-tauri/target/release/bundle/appimage/ez-utils_<version>_amd64.AppImage` - portable AppImage

**Install**:
```bash
# Debian / Ubuntu
sudo dpkg -i ez-utils_<version>_amd64.deb

# RHEL / Fedora
sudo rpm -i ez-utils-<version>-1.x86_64.rpm

# AppImage (no install, run directly)
chmod +x ez-utils_<version>_amd64.AppImage
./ez-utils_<version>_amd64.AppImage
```

### macOS

Produces:
- `src-tauri/target/release/bundle/macos/ez-utils.app` - application bundle
- `src-tauri/target/release/bundle/dmg/ez-utils_<version>_x64.dmg` - DMG disk image

**Install**: open the `.dmg` and drag `ez-utils.app` to `/Applications`.

## License

GNU General Public License v3.0. Full license text: [gnu.org/licenses/gpl-3.0](https://www.gnu.org/licenses/gpl-3.0.html).
