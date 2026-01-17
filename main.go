package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/gopacket"
	"github.com/muesli/termenv"

	"nbor/capture"
	"nbor/config"
	"nbor/logger"
	"nbor/parser"
	"nbor/platform"
	"nbor/tui"
	"nbor/types"
	"nbor/version"
)

func init() {
	// Force true color mode on Windows Terminal which supports it but doesn't
	// set COLORTERM environment variable. This enables proper background colors.
	// Safe to call even on terminals that don't support true color - they'll
	// just display the closest available colors.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// Global channel for interface selection (needed because bubbletea copies the model)
var selectedInterfaceChan = make(chan types.InterfaceInfo, 1)

// cliOptions holds parsed command-line arguments
type cliOptions struct {
	themeName         string
	interfaceName     string
	listThemes        bool
	listInterfaces    bool
	listAllInterfaces bool
	showHelp          bool
	showVersion       bool
}

// parseArgs parses command-line arguments
func parseArgs() cliOptions {
	opts := cliOptions{}
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			opts.showHelp = true
		case arg == "-v" || arg == "--version":
			opts.showVersion = true
		case arg == "--list-themes":
			opts.listThemes = true
		case arg == "-l" || arg == "--list-interfaces":
			opts.listInterfaces = true
		case arg == "--list-all-interfaces":
			opts.listAllInterfaces = true
		case arg == "-t" || arg == "--theme":
			if i+1 < len(args) {
				i++
				opts.themeName = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a theme name\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--theme="):
			opts.themeName = strings.TrimPrefix(arg, "--theme=")
		case strings.HasPrefix(arg, "-t="):
			opts.themeName = strings.TrimPrefix(arg, "-t=")
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Error: unknown option %s\n", arg)
			fmt.Fprintf(os.Stderr, "Run 'nbor --help' for usage\n")
			os.Exit(1)
		default:
			// Positional argument = interface name
			if opts.interfaceName == "" {
				opts.interfaceName = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument %s\n", arg)
				os.Exit(1)
			}
		}
	}

	return opts
}

// printHelp prints the help message
func printHelp() {
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

Theme names use hyphens, not spaces (e.g., tokyo-night, catppuccin-mocha)

Examples:
  nbor                              # Interactive interface picker
  nbor eth0                         # Start on eth0 directly
  nbor --theme dracula              # Use dracula theme
  nbor --theme tokyo-night          # Multi-word theme (use hyphens)
  nbor -t catppuccin-mocha eth0     # Theme + interface
  nbor "Ethernet 2"                 # Windows interface with spaces
  nbor -l                           # List available interfaces

Configuration:
  Config file location:
    Linux/macOS: $XDG_CONFIG_HOME/nbor/config.toml (default: ~/.config/nbor/config.toml)
    Windows:     %%APPDATA%%\nbor\config.toml

  Example config.toml:
    # Theme name (use slug format with hyphens)
    theme = "tokyo-night"
`
	fmt.Print(help)
}

// printThemes prints available themes
func printThemes() {
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

// findInterface searches for an interface by name (case-insensitive)
func findInterface(interfaces []types.InterfaceInfo, name string) *types.InterfaceInfo {
	nameLower := strings.ToLower(name)
	for _, iface := range interfaces {
		if strings.ToLower(iface.Name) == nameLower {
			return &iface
		}
	}
	return nil
}

// printInterfaceError prints a colored error message for interface not found
func printInterfaceError(name string, interfaces []types.InterfaceInfo) {
	theme := tui.DefaultTheme
	errorStyle := lipgloss.NewStyle().Foreground(theme.Base08).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(theme.Base03)
	nameStyle := lipgloss.NewStyle().Foreground(theme.Base0A)

	fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: Interface '%s' not found", name)))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, hintStyle.Render("Available interfaces:"))
	for _, iface := range interfaces {
		status := "down"
		if iface.IsUp {
			status = "up"
		}
		fmt.Fprintf(os.Stderr, "  %s (%s)\n", nameStyle.Render(iface.Name), status)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, hintStyle.Render("Falling back to interface picker..."))
	fmt.Fprintln(os.Stderr)
}

// printInterfaces prints the list of available interfaces
func printInterfaces(interfaces []types.InterfaceInfo) {
	theme := tui.DefaultTheme
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Base0B)
	nameStyle := lipgloss.NewStyle().Foreground(theme.Base0A).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base03)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	upStyle := lipgloss.NewStyle().Foreground(theme.Base0B)
	downStyle := lipgloss.NewStyle().Foreground(theme.Base08)

	fmt.Println(headerStyle.Render("Available interfaces:"))
	fmt.Println()

	if len(interfaces) == 0 {
		fmt.Println("  No suitable Ethernet interfaces found.")
		return
	}

	for _, iface := range interfaces {
		fmt.Printf("  %s\n", nameStyle.Render(iface.Name))

		if len(iface.MAC) > 0 {
			fmt.Printf("    %s %s\n", labelStyle.Render("MAC:"), valueStyle.Render(iface.MAC.String()))
		}

		for _, ip := range iface.IPv4Addrs {
			fmt.Printf("    %s %s\n", labelStyle.Render("IPv4:"), valueStyle.Render(ip.String()))
		}

		for _, ip := range iface.IPv6Addrs {
			fmt.Printf("    %s %s\n", labelStyle.Render("IPv6:"), valueStyle.Render(ip.String()))
		}

		status := downStyle.Render("down")
		if iface.IsUp {
			status = upStyle.Render("up")
		}
		fmt.Printf("    %s %s\n", labelStyle.Render("Status:"), status)
		fmt.Println()
	}
}

// printAllInterfaces prints all interfaces including filtered ones with reasons
func printAllInterfaces(usable, all []types.InterfaceInfo) {
	theme := tui.DefaultTheme
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Base0B)
	filteredHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Base09)
	nameStyle := lipgloss.NewStyle().Foreground(theme.Base0A).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base03)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	reasonStyle := lipgloss.NewStyle().Foreground(theme.Base0E)
	upStyle := lipgloss.NewStyle().Foreground(theme.Base0B)
	downStyle := lipgloss.NewStyle().Foreground(theme.Base08)

	// Build map of usable interfaces for quick lookup
	usableMap := make(map[string]bool)
	for _, iface := range usable {
		usableMap[iface.Name] = true
	}

	// Print usable interfaces
	fmt.Println(headerStyle.Render("Available interfaces:"))
	fmt.Println()

	if len(usable) == 0 {
		fmt.Println("  No suitable Ethernet interfaces found.")
	} else {
		for _, iface := range usable {
			fmt.Printf("  %s\n", nameStyle.Render(iface.Name))

			if len(iface.MAC) > 0 {
				fmt.Printf("    %s %s\n", labelStyle.Render("MAC:"), valueStyle.Render(iface.MAC.String()))
			}

			for _, ip := range iface.IPv4Addrs {
				fmt.Printf("    %s %s\n", labelStyle.Render("IPv4:"), valueStyle.Render(ip.String()))
			}

			status := downStyle.Render("down")
			if iface.IsUp {
				status = upStyle.Render("up")
			}
			fmt.Printf("    %s %s\n", labelStyle.Render("Status:"), status)
			fmt.Println()
		}
	}

	// Print filtered interfaces
	var filtered []types.InterfaceInfo
	for _, iface := range all {
		if !usableMap[iface.Name] {
			filtered = append(filtered, iface)
		}
	}

	if len(filtered) > 0 {
		fmt.Println(filteredHeaderStyle.Render("Filtered interfaces:"))
		fmt.Println()

		for _, iface := range filtered {
			reason := platform.GetFilterReason(iface.Name)
			if reason == "" {
				reason = "unknown"
			}
			fmt.Printf("  %s (%s)\n", nameStyle.Render(iface.Name), reasonStyle.Render(reason))

			if len(iface.MAC) > 0 {
				fmt.Printf("    %s %s\n", labelStyle.Render("MAC:"), valueStyle.Render(iface.MAC.String()))
			}

			for _, ip := range iface.IPv4Addrs {
				fmt.Printf("    %s %s\n", labelStyle.Render("IPv4:"), valueStyle.Render(ip.String()))
			}

			status := downStyle.Render("down")
			if iface.IsUp {
				status = upStyle.Render("up")
			}
			fmt.Printf("    %s %s\n", labelStyle.Render("Status:"), status)
			fmt.Println()
		}
	}
}

// printFilterWarning prints a warning when using a filtered interface
func printFilterWarning(name, reason string) {
	theme := tui.DefaultTheme
	warnStyle := lipgloss.NewStyle().Foreground(theme.Base09).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	hintStyle := lipgloss.NewStyle().Foreground(theme.Base03)
	promptStyle := lipgloss.NewStyle().Foreground(theme.Base0C)

	fmt.Fprintln(os.Stderr, warnStyle.Render(fmt.Sprintf("Warning: '%s' appears to be a %s", name, reason)))
	fmt.Fprintln(os.Stderr, textStyle.Render("CDP/LLDP protocols are typically only used on wired networks."))
	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, promptStyle.Render("Press Enter to continue (or Ctrl+C to cancel)... "))

	// Wait for user to press Enter
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	fmt.Fprintln(os.Stderr, hintStyle.Render("Continuing..."))
	fmt.Fprintln(os.Stderr)
}

func main() {
	// Parse CLI arguments
	opts := parseArgs()

	// Handle help flag
	if opts.showHelp {
		printHelp()
		os.Exit(0)
	}

	// Handle version flag
	if opts.showVersion {
		fmt.Printf("nbor version %s\n", version.Version)
		os.Exit(0)
	}

	// Handle list-themes flag
	if opts.listThemes {
		printThemes()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Determine theme: CLI flag overrides config
	themeName := cfg.Theme
	if opts.themeName != "" {
		themeName = opts.themeName
	}

	// Apply theme
	if theme := tui.GetThemeByName(themeName); theme != nil {
		tui.SetTheme(*theme)
	} else if themeName != "" && themeName != "solarized-dark" {
		// Only warn if user explicitly specified an invalid theme
		fmt.Fprintf(os.Stderr, "Warning: unknown theme '%s', using default\n", themeName)
		fmt.Fprintf(os.Stderr, "Run 'nbor --list-themes' to see available themes\n")
	}

	// Check for Npcap on Windows
	if err := platform.CheckNpcap(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check privileges
	if err := platform.CheckPrivileges(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "%s\n", platform.GetPrivilegeHint())
		os.Exit(1)
	}

	// Get available Ethernet interfaces
	interfaces, err := platform.GetEthernetInterfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing interfaces: %v\n", err)
		os.Exit(1)
	}

	// Handle list-interfaces flag
	if opts.listInterfaces {
		printInterfaces(interfaces)
		os.Exit(0)
	}

	// Handle list-all-interfaces flag
	if opts.listAllInterfaces {
		allInterfaces, err := platform.GetAllInterfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing all interfaces: %v\n", err)
			os.Exit(1)
		}
		printAllInterfaces(interfaces, allInterfaces)
		os.Exit(0)
	}

	if len(interfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No suitable Ethernet interfaces found.\n")
		fmt.Fprintf(os.Stderr, "Make sure you have wired network adapters available.\n")
		os.Exit(1)
	}

	// Check for interface argument
	var preselectedInterface *types.InterfaceInfo
	if opts.interfaceName != "" {
		preselectedInterface = findInterface(interfaces, opts.interfaceName)
		if preselectedInterface == nil {
			// Not found in usable interfaces, check filtered interfaces
			allInterfaces, _ := platform.GetAllInterfaces()
			if filteredIface := findInterface(allInterfaces, opts.interfaceName); filteredIface != nil {
				// Found but was filtered - warn and allow
				reason := platform.GetFilterReason(filteredIface.Name)
				if reason == "" {
					reason = "filtered interface"
				}
				printFilterWarning(filteredIface.Name, reason)
				preselectedInterface = filteredIface
			} else {
				// Truly not found
				printInterfaceError(opts.interfaceName, interfaces)
			}
		}
	}

	// Create neighbor store
	store := types.NewNeighborStore()

	// Create the TUI application
	app := tui.NewApp(interfaces, store, selectedInterfaceChan)

	// Create program with options
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Variables for capture state
	var capturer *capture.Capturer
	var csvLogger *logger.CSVLogger

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cleanup(capturer, csvLogger)
		p.Quit()
	}()

	// Goroutine to handle interface selection
	go func() {
		var ifaceInfo types.InterfaceInfo

		// If interface was preselected via CLI, use it directly
		if preselectedInterface != nil {
			ifaceInfo = *preselectedInterface
			// Also send to channel so TUI knows to skip picker
			select {
			case selectedInterfaceChan <- ifaceInfo:
			default:
			}
		} else {
			// Wait for user selection from TUI picker
			ifaceInfo = <-selectedInterfaceChan
		}

		// Get internal name for pcap (important for Windows)
		internalName := platform.GetInterfaceInternalName(ifaceInfo.Name)

		// Create capturer
		cap, err := capture.NewCapturer(internalName)
		if err != nil {
			p.Send(tui.ErrorMsg{Err: fmt.Errorf("failed to start capture: %w", err)})
			return
		}
		capturer = cap

		// Create CSV logger
		csvLog, err := logger.NewCSVLogger()
		if err != nil {
			p.Send(tui.ErrorMsg{Err: fmt.Errorf("failed to create log file: %w", err)})
			cap.Stop()
			return
		}
		csvLogger = csvLog

		// Set up neighbor callback - only log first-seen neighbors
		store.OnNewNeighbor = func(n *types.Neighbor) {
			// Ring terminal bell
			platform.Bell()

			// Log to CSV (only new neighbors, not updates)
			if err := csvLogger.Log(n); err != nil {
				// Log error but don't crash
				fmt.Fprintf(os.Stderr, "Warning: failed to log neighbor: %v\n", err)
			}

			// Notify TUI
			p.Send(tui.NewNeighborMsg{Neighbor: n})
		}
		// Note: OnUpdate not set - we only log first-seen neighbors

		// Signal TUI to transition to capture view
		p.Send(tui.StartCaptureMsg{
			Interface: ifaceInfo,
			LogPath:   csvLogger.Filepath(),
		})

		// Start capturing
		packets := cap.Start()

		// Process packets
		processPackets(packets, store, ifaceInfo.Name)
	}()

	// Run the TUI
	if _, err := p.Run(); err != nil {
		cleanup(capturer, csvLogger)
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}

	// Clean up on exit
	cleanup(capturer, csvLogger)
}

// processPackets processes incoming packets and updates the store
func processPackets(packets <-chan gopacket.Packet, store *types.NeighborStore, ifaceName string) {
	for packet := range packets {
		var neighbor *types.Neighbor
		var err error

		// Determine packet type and parse
		if capture.IsCDPPacket(packet) {
			neighbor, err = parser.ParseCDP(packet, ifaceName)
		} else if capture.IsLLDPPacket(packet) {
			neighbor, err = parser.ParseLLDP(packet, ifaceName)
		} else {
			continue
		}

		if err != nil {
			// Skip malformed packets silently
			continue
		}

		if neighbor != nil {
			neighbor.LastSeen = time.Now()
			store.Update(neighbor)
		}
	}
}

// cleanup handles graceful shutdown
func cleanup(cap *capture.Capturer, log *logger.CSVLogger) {
	if cap != nil {
		cap.Stop()
	}
	if log != nil {
		log.Close()
	}
}
