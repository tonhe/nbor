package cli

import (
	"fmt"

	"nbor/tui"
)

// PrintHelp prints the help message
func PrintHelp() {
	help := `nbor - Network neighbor discovery tool (CDP/LLDP)

Usage:
  nbor [options] [interface]

Options:
  -t, --theme <name>      Use specified theme (session only)
  --list-themes           List available themes
  -l, --list-interfaces   List available network interfaces
  --list-all-interfaces   List all interfaces (including filtered)
  -v, --version           Show version
  -h, --help              Show this help

Identity Options:
  --name <string>         System name to advertise (default: hostname)
  --description <string>  System description to advertise

Listening Options:
  --cdp-listen            Enable CDP listening (default)
  --no-cdp-listen         Disable CDP listening
  --lldp-listen           Enable LLDP listening (default)
  --no-lldp-listen        Disable LLDP listening

Broadcasting Options:
  --broadcast             Enable both CDP and LLDP broadcasting
  --cdp-broadcast         Enable CDP broadcasting
  --no-cdp-broadcast      Disable CDP broadcasting (default)
  --lldp-broadcast        Enable LLDP broadcasting
  --no-lldp-broadcast     Disable LLDP broadcasting (default)
  --interval <seconds>    Broadcast interval (default: 5)
  --ttl <seconds>         TTL/hold time (default: 20)
  --capabilities <list>   Capabilities to advertise (comma-separated)
                          Options: router, bridge, station, switch, phone

Interface Options:
  --auto-select           Auto-select if only one interface (default)
  --no-auto-select        Always show interface picker

Examples:
  nbor                              # Interactive main menu
  nbor eth0                         # Start on eth0 directly
  nbor --broadcast eth0             # Start broadcasting on eth0
  nbor --broadcast --interval 10    # Broadcast every 10 seconds
  nbor --name "my-host" --broadcast # Custom system name
  nbor --capabilities router,bridge # Advertise as router and bridge

Configuration:
  Config file: ~/.config/nbor/config.toml (Linux/macOS)
               %%APPDATA%%\nbor\config.toml (Windows)

  CLI flags override config file settings.
`
	fmt.Print(help)
}

// PrintThemes prints available themes
func PrintThemes() {
	fmt.Println("Available themes:")
	fmt.Println()
	themes := tui.ListThemes()
	for _, t := range themes {
		fmt.Printf("  %-20s  %s\n", t[0], t[1])
	}
	fmt.Println()
	fmt.Println("Usage: nbor --theme <slug>")
	fmt.Println("Example: nbor --theme tokyo-night")
}
