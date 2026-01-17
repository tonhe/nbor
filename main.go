package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/gopacket"
	"github.com/muesli/termenv"

	"nbor/capture"
	"nbor/logger"
	"nbor/parser"
	"nbor/platform"
	"nbor/tui"
	"nbor/types"
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

func main() {
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

	if len(interfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No suitable Ethernet interfaces found.\n")
		fmt.Fprintf(os.Stderr, "Make sure you have wired network adapters available.\n")
		os.Exit(1)
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
		ifaceInfo := <-selectedInterfaceChan

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
