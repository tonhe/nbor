# nbor

A polished TUI tool for discovering network neighbors via CDP (Cisco Discovery Protocol) and LLDP (Link Layer Discovery Protocol).

## Features

- **Interface Selection**: Automatically filters to show only wired Ethernet interfaces
- **Protocol Support**: Listens for both CDP and LLDP packets
- **Rich Neighbor Information**:
  - Switch/device hostname
  - Port ID (the port you're connected to)
  - Management IP address
  - Platform/model
  - System description
  - SNMP Location (if available)
  - Device capabilities (Router, Switch, Bridge, AP, Phone, etc.)
  - Protocol type (CDP or LLDP)
  - Last seen timestamp
- **Visual Alerts**: New neighbors are highlighted and trigger a terminal bell
- **Stale Detection**: Neighbors not seen for 3-4 minutes are grayed out
- **CSV Logging**: All discoveries are logged to a timestamped CSV file
- **20 Built-in Themes**: Solarized, Gruvbox, Dracula, Nord, Tokyo Night, Catppuccin, and more
- **Configuration File**: Persistent settings with XDG support on Linux/macOS and %APPDATA% on Windows

## Platform Support

| Platform | Status |
|----------|--------|
| macOS    | Tested |
| Linux    | Untested |
| Windows  | Tested |

## Requirements

### macOS

```bash
# libpcap is included with macOS
# No additional dependencies needed
```

### Linux (completely untested)

```bash
# Requires libpcap-dev
sudo apt install libpcap-dev  # Debian/Ubuntu
sudo dnf install libpcap-devel  # Fedora/RHEL
```

### Windows (completely untested)

```powershell
# Requires Npcap installed (https://npcap.com)
# During installation, check "WinPcap API-compatible Mode"
```

## Building

### Linux/macOS

```bash
go build -o nbor
```

### Windows

```powershell
go build -o nbor.exe
```

## Usage

The tool requires elevated privileges for packet capture.

### Basic Usage

```bash
# Linux/macOS
sudo ./nbor

# Windows (Run as Administrator)
.\nbor.exe
```

### Command Line Options

```
Usage:
  nbor [options] [interface]

Options:
  -t, --theme <name>      Use specified theme (session only)
  --list-themes           List available themes
  -l, --list-interfaces   List available network interfaces
  --list-all-interfaces   List all interfaces (including filtered)
  -v, --version           Show version
  -h, --help              Show this help
```

### Examples

```bash
# Interactive interface picker (default)
sudo ./nbor

# Start capture directly on a specific interface
sudo ./nbor eth0                    # Linux
sudo ./nbor en0                     # macOS (with warning for WiFi)
.\nbor.exe "Ethernet 2"             # Windows (interface with spaces)

# Use a different theme for this session
sudo ./nbor --theme dracula
sudo ./nbor --theme tokyo-night
sudo ./nbor -t catppuccin-mocha eth0

# List available interfaces
sudo ./nbor -l

# List all interfaces including filtered ones (WiFi, virtual, etc.)
sudo ./nbor --list-all-interfaces

# List available themes
./nbor --list-themes
```

### Filtered Interface Warning

When you specify an interface that would normally be filtered (WiFi, virtual, tunnel, etc.), nbor will warn but allow you to proceed:

```
Warning: 'en0' appears to be a WiFi interface
CDP/LLDP protocols are typically only used on wired networks.
Continuing anyway...
```

## Interface

1. On launch, select a network interface using arrow keys and press Enter
2. The main view shows discovered neighbors in a table
3. Hotkeys:
   - `r` - Refresh display
   - `↑/↓` or `j/k` - Scroll through neighbors
   - `Ctrl+C` or `q` - Quit

## CSV Log Files

Each session creates a CSV log file in the current directory with the format:
```
nbor-YYYY-MM-DD-HHMMSS.csv
```

The log contains all neighbor announcements with timestamps.

## Architecture

The codebase is structured for maintainability and future multi-interface support:

```
nbor/
├── main.go           # Application entry point and CLI parsing
├── capture/          # Packet capture with gopacket/libpcap
├── config/           # Configuration file loading (TOML)
├── logger/           # CSV logging
├── parser/           # CDP and LLDP protocol parsing
├── platform/         # OS-specific abstractions (Linux/macOS/Windows)
├── tui/              # Terminal UI with bubbletea/lipgloss
├── types/            # Shared data types
└── version/          # Version constant
```

## Theming

nbor includes 20 built-in themes based on the Base16 color specification.

### Available Themes

| Slug | Display Name |
|------|--------------|
| `solarized-dark` | Solarized Dark (default) |
| `solarized-light` | Solarized Light |
| `gruvbox-dark` | Gruvbox Dark |
| `gruvbox-light` | Gruvbox Light |
| `dracula` | Dracula |
| `nord` | Nord |
| `one-dark` | One Dark |
| `monokai` | Monokai |
| `tokyo-night` | Tokyo Night |
| `catppuccin-mocha` | Catppuccin Mocha |
| `catppuccin-latte` | Catppuccin Latte |
| `everforest` | Everforest |
| `kanagawa` | Kanagawa |
| `rose-pine` | Rosé Pine |
| `tomorrow-night` | Tomorrow Night |
| `ayu-dark` | Ayu Dark |
| `horizon` | Horizon |
| `zenburn` | Zenburn |
| `palenight` | Palenight |
| `github-dark` | GitHub Dark |

Use `nbor --list-themes` to see available themes. Theme names use hyphens (not spaces), so "Tokyo Night" becomes `tokyo-night`.

**For a single session:**
```bash
sudo ./nbor --theme dracula
```

**To set permanently**, see the Configuration section below.

## Configuration

nbor stores settings in a TOML config file that is created automatically on first run.

### Config File Locations

| Platform | Location |
|----------|----------|
| Linux    | `$XDG_CONFIG_HOME/nbor/config.toml` (default: `~/.config/nbor/config.toml`) |
| macOS    | `$XDG_CONFIG_HOME/nbor/config.toml` (default: `~/.config/nbor/config.toml`) |
| Windows  | `%APPDATA%\nbor\config.toml` |

### Example config.toml

```toml
# Theme name (use slug format with hyphens)
theme = "tokyo-night"
```

## License

MIT
