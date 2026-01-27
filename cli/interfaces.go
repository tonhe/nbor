package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"nbor/platform"
	"nbor/tui"
	"nbor/types"
)

// FindInterface searches for an interface by name (case-insensitive)
func FindInterface(interfaces []types.InterfaceInfo, name string) *types.InterfaceInfo {
	nameLower := strings.ToLower(name)
	for _, iface := range interfaces {
		if strings.ToLower(iface.Name) == nameLower {
			return &iface
		}
	}
	return nil
}

// PrintInterfaceError prints a colored error message for interface not found
func PrintInterfaceError(name string, interfaces []types.InterfaceInfo) {
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

// PrintInterfaces prints the list of available interfaces
func PrintInterfaces(interfaces []types.InterfaceInfo) {
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

// PrintAllInterfaces prints all interfaces including filtered ones with reasons
func PrintAllInterfaces(usable, all []types.InterfaceInfo) {
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

// PrintFilterWarning prints a warning when using a filtered interface
func PrintFilterWarning(name, reason string) {
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
