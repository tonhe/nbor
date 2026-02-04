package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/muesli/termenv"

	"nbor/broadcast"
	"nbor/capture"
	"nbor/cli"
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

// Global channels for TUI-to-main communication
var restartLogChan = make(chan struct{}, 1)
var restartCaptureChan = make(chan struct{}, 1)
var broadcastToggleChan = make(chan bool, 1)
var configUpdateChan = make(chan *config.Config, 1)

func main() {
	// Parse CLI arguments
	opts := cli.ParseArgs()

	// Handle help flag
	if opts.ShowHelp {
		cli.PrintHelp()
		os.Exit(0)
	}

	// Handle version flag
	if opts.ShowVersion {
		fmt.Printf("nbor version %s\n", version.Version)
		os.Exit(0)
	}

	// Handle list-themes flag
	if opts.ListThemes {
		cli.PrintThemes()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Apply CLI overrides to config
	cli.ApplyOverrides(&cfg, opts)

	// Determine theme: CLI flag overrides config
	themeName := cfg.Theme
	if opts.ThemeName != "" {
		themeName = opts.ThemeName
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
	if opts.ListInterfaces {
		cli.PrintInterfaces(interfaces)
		os.Exit(0)
	}

	// Handle list-all-interfaces flag
	if opts.ListAllInterfaces {
		allInterfaces, err := platform.GetAllInterfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing all interfaces: %v\n", err)
			os.Exit(1)
		}
		cli.PrintAllInterfaces(interfaces, allInterfaces)
		os.Exit(0)
	}

	if len(interfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No suitable Ethernet interfaces found.\n")
		fmt.Fprintf(os.Stderr, "Make sure you have wired network adapters available.\n")
		os.Exit(1)
	}

	// Check for interface argument
	var preselectedInterface *types.InterfaceInfo
	if opts.InterfaceName != "" {
		preselectedInterface = cli.FindInterface(interfaces, opts.InterfaceName)
		if preselectedInterface == nil {
			// Not found in usable interfaces, check filtered interfaces
			allInterfaces, _ := platform.GetAllInterfaces()
			if filteredIface := cli.FindInterface(allInterfaces, opts.InterfaceName); filteredIface != nil {
				// Found but was filtered - warn and allow
				reason := platform.GetFilterReason(filteredIface.Name)
				if reason == "" {
					reason = "filtered interface"
				}
				cli.PrintFilterWarning(filteredIface.Name, reason)
				preselectedInterface = filteredIface
			} else {
				// Truly not found
				cli.PrintInterfaceError(opts.InterfaceName, interfaces)
			}
		}
	}

	// Auto-select interface if only one is available and up
	if preselectedInterface == nil && cfg.AutoSelectInterface {
		var upInterfaces []types.InterfaceInfo
		for _, iface := range interfaces {
			if iface.IsUp {
				upInterfaces = append(upInterfaces, iface)
			}
		}
		if len(upInterfaces) == 1 {
			preselectedInterface = &upInterfaces[0]
		}
	}

	// Create neighbor store
	store := types.NewNeighborStore()

	// Create the TUI application
	// If interface is preselected, start at interface picker, otherwise show main menu
	var app tui.AppModel
	if preselectedInterface != nil {
		app = tui.NewAppAtInterfacePicker(interfaces, store, &cfg, selectedInterfaceChan, restartLogChan, restartCaptureChan, broadcastToggleChan, configUpdateChan)
	} else {
		app = tui.NewApp(interfaces, store, &cfg, selectedInterfaceChan, restartLogChan, restartCaptureChan, broadcastToggleChan, configUpdateChan)
	}

	// Create program with options
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Variables for capture state
	var capturer *capture.Capturer
	var csvLogger *logger.CSVLogger
	var broadcaster *broadcast.Broadcaster
	var pcapHandle *pcap.Handle

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cleanupAll(capturer, csvLogger, broadcaster)
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

		// Open pcap handle for both capture and broadcast
		// Use 100ms timeout instead of BlockForever to allow clean shutdown on Linux
		handle, err := pcap.OpenLive(internalName, 65535, true, 100*time.Millisecond)
		if err != nil {
			p.Send(tui.ErrorMsg{Err: fmt.Errorf("failed to open interface: %w", err)})
			return
		}
		pcapHandle = handle

		// Set BPF filter for capture
		filter := "ether dst 01:00:0c:cc:cc:cc or ether dst 01:80:c2:00:00:0e"
		if err := handle.SetBPFFilter(filter); err != nil {
			handle.Close()
			p.Send(tui.ErrorMsg{Err: fmt.Errorf("failed to set BPF filter: %w", err)})
			return
		}

		// Create capturer using existing handle
		cap := capture.NewCapturerWithHandle(handle, internalName)
		capturer = cap

		// Create CSV logger (if enabled)
		if cfg.LoggingEnabled {
			csvLog, err := logger.NewCSVLogger(cfg.LogDirectory, cfg.FilterCapabilities)
			if err != nil {
				p.Send(tui.ErrorMsg{Err: fmt.Errorf("failed to create log file: %w", err)})
				cap.Stop()
				return
			}
			csvLogger = csvLog
		}

		// Create broadcaster
		bc := broadcast.NewBroadcaster(handle, &cfg, &ifaceInfo)
		broadcaster = bc

		// Start broadcaster only if BroadcastOnStartup is enabled AND a protocol is configured
		if cfg.BroadcastOnStartup && (cfg.CDPBroadcast || cfg.LLDPBroadcast) {
			bc.Start()
		}

		// Set up neighbor callback - only log first-seen neighbors
		store.OnNewNeighbor = func(n *types.Neighbor) {
			// Ring terminal bell
			platform.Bell()

			// Log to CSV (only new neighbors, not updates) if logging is enabled
			if csvLogger != nil {
				if err := csvLogger.Log(n); err != nil {
					// Log error but don't crash
					fmt.Fprintf(os.Stderr, "Warning: failed to log neighbor: %v\n", err)
				}
			}

			// Notify TUI
			p.Send(tui.NewNeighborMsg{Neighbor: n})
		}
		// Note: OnUpdate not set - we only log first-seen neighbors

		// Determine log path for display
		logPath := ""
		if csvLogger != nil {
			logPath = csvLogger.Filepath()
		}

		// Signal TUI to transition to capture view
		p.Send(tui.StartCaptureMsg{
			Interface: ifaceInfo,
			LogPath:   logPath,
		})

		// Start capturing
		packets := cap.Start()

		// Process packets (pass local MAC to filter out own broadcasts)
		localMAC := ""
		if ifaceInfo.MAC != nil {
			localMAC = ifaceInfo.MAC.String()
		}
		processPackets(packets, store, ifaceInfo.Name, localMAC, &cfg)
	}()

	// Goroutine to handle broadcast toggle messages from TUI
	go func() {
		for enabled := range broadcastToggleChan {
			if broadcaster != nil {
				if enabled {
					broadcaster.Start()
				} else {
					broadcaster.Stop()
				}
			}
		}
	}()

	// Goroutine to handle config updates from TUI
	go func() {
		for newCfg := range configUpdateChan {
			// Update local config reference
			cfg = *newCfg
			// Update broadcaster config
			if broadcaster != nil {
				broadcaster.UpdateConfig(newCfg)
			}
		}
	}()

	// Goroutine to handle log restart requests
	go func() {
		for range restartLogChan {
			// Only restart if logging is enabled
			if cfg.LoggingEnabled {
				// Close old log file if exists
				if csvLogger != nil {
					csvLogger.Close()
				}

				// Create new log file with current config
				newLogger, err := logger.NewCSVLogger(cfg.LogDirectory, cfg.FilterCapabilities)
				if err != nil {
					// Log error but continue with old logger
					continue
				}
				csvLogger = newLogger

				// Notify TUI of new log path
				p.Send(tui.LogRestartedMsg{LogPath: csvLogger.Filepath()})
			}
		}
	}()

	// Run the TUI
	if _, err := p.Run(); err != nil {
		cleanupAll(capturer, csvLogger, broadcaster)
		if pcapHandle != nil {
			pcapHandle.Close()
		}
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}

	// Check if we should restart (interface change requested)
	select {
	case <-restartCaptureChan:
		// Clean up current session
		cleanupAll(capturer, csvLogger, broadcaster)
		if pcapHandle != nil {
			pcapHandle.Close()
		}
		// Re-exec the program to restart fresh, with --no-auto-select to force interface picker
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error restarting: %v\n", err)
			os.Exit(1)
		}
		// Build args, adding --no-auto-select if not already present
		args := os.Args[1:] // Skip program name for exec.Command
		if !slices.Contains(args, "--no-auto-select") {
			args = append(args, "--no-auto-select")
		}
		// Use os/exec instead of syscall.Exec for cross-platform support (Windows)
		cmd := exec.Command(exe, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error restarting: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		// Normal exit
	}

	// Clean up on exit
	cleanupAll(capturer, csvLogger, broadcaster)
	if pcapHandle != nil {
		pcapHandle.Close()
	}
}

// processPackets processes incoming packets and updates the store
// localMAC is used to filter out our own broadcast packets
// cfg is used to check listen settings (CDPListen, LLDPListen)
func processPackets(packets <-chan gopacket.Packet, store *types.NeighborStore, ifaceName string, localMAC string, cfg *config.Config) {
	for packet := range packets {
		// Filter out our own broadcasts by checking source MAC
		srcMAC := capture.GetSourceMAC(packet)
		if srcMAC != nil && srcMAC.String() == localMAC {
			// This is our own broadcast, skip it
			continue
		}

		var neighbor *types.Neighbor
		var err error

		// Determine packet type and parse (respecting listen settings)
		if capture.IsCDPPacket(packet) {
			if !cfg.CDPListen {
				continue // CDP listening disabled
			}
			neighbor, err = parser.ParseCDP(packet, ifaceName)
		} else if capture.IsLLDPPacket(packet) {
			if !cfg.LLDPListen {
				continue // LLDP listening disabled
			}
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

// cleanupAll handles graceful shutdown of all components
func cleanupAll(cap *capture.Capturer, log *logger.CSVLogger, bc *broadcast.Broadcaster) {
	if bc != nil {
		bc.Stop()
	}
	if cap != nil {
		cap.Stop()
	}
	if log != nil {
		log.Close()
	}
}
