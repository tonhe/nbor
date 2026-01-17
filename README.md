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
- **Solarized Dark Theme**: Clean, readable interface with Base16 theming support

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

The tool requires elevated privileges for packet capture:

### Linux/macOS

```bash
sudo ./nbor
```

### Windows

Run from an elevated Command Prompt or PowerShell (Run as Administrator):

```powershell
.\nbor.exe
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
├── main.go           # Application entry point
├── capture/          # Packet capture with gopacket/libpcap
├── parser/           # CDP and LLDP protocol parsing
├── platform/         # OS-specific abstractions (Linux/macOS/Windows)
├── tui/              # Terminal UI with bubbletea/lipgloss
├── logger/           # CSV logging
└── types/            # Shared data types
```

## Theming

The TUI uses a Base16 color scheme (Solarized Dark by default). The theme system is designed to support future configuration file-based theme switching.

## License

MIT
