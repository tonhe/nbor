package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Options holds parsed command-line arguments
type Options struct {
	ThemeName         string
	InterfaceName     string
	ListThemes        bool
	ListInterfaces    bool
	ListAllInterfaces bool
	ShowHelp          bool
	ShowVersion       bool

	// CDP/LLDP options
	SystemName        string
	SystemDescription string
	CDPListen         *bool // nil = use config, true/false = override
	LLDPListen        *bool
	CDPBroadcast      *bool
	LLDPBroadcast     *bool
	BroadcastAll      bool // --broadcast enables both
	Interval          int  // 0 = use config
	TTL               int  // 0 = use config
	Capabilities      string

	// Interface selection
	NoAutoSelect *bool // nil = use config, true/false = override
}

// ParseArgs parses command-line arguments
func ParseArgs() Options {
	opts := Options{}
	args := os.Args[1:]

	// Helper for bool pointer flags
	boolTrue := true
	boolFalse := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			opts.ShowHelp = true
		case arg == "-v" || arg == "--version":
			opts.ShowVersion = true
		case arg == "--list-themes":
			opts.ListThemes = true
		case arg == "-l" || arg == "--list-interfaces":
			opts.ListInterfaces = true
		case arg == "--list-all-interfaces":
			opts.ListAllInterfaces = true
		case arg == "-t" || arg == "--theme":
			if i+1 < len(args) {
				i++
				opts.ThemeName = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a theme name\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--theme="):
			opts.ThemeName = strings.TrimPrefix(arg, "--theme=")
		case strings.HasPrefix(arg, "-t="):
			opts.ThemeName = strings.TrimPrefix(arg, "-t=")

		// CDP/LLDP flags
		case arg == "--name":
			if i+1 < len(args) {
				i++
				opts.SystemName = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a system name\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--name="):
			opts.SystemName = strings.TrimPrefix(arg, "--name=")

		case arg == "--description":
			if i+1 < len(args) {
				i++
				opts.SystemDescription = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a description\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--description="):
			opts.SystemDescription = strings.TrimPrefix(arg, "--description=")

		case arg == "--cdp-listen":
			opts.CDPListen = &boolTrue
		case arg == "--no-cdp-listen":
			opts.CDPListen = &boolFalse
		case arg == "--lldp-listen":
			opts.LLDPListen = &boolTrue
		case arg == "--no-lldp-listen":
			opts.LLDPListen = &boolFalse

		case arg == "--cdp-broadcast":
			opts.CDPBroadcast = &boolTrue
		case arg == "--no-cdp-broadcast":
			opts.CDPBroadcast = &boolFalse
		case arg == "--lldp-broadcast":
			opts.LLDPBroadcast = &boolTrue
		case arg == "--no-lldp-broadcast":
			opts.LLDPBroadcast = &boolFalse
		case arg == "--broadcast":
			opts.BroadcastAll = true

		case arg == "--interval":
			if i+1 < len(args) {
				i++
				val, err := strconv.Atoi(args[i])
				if err != nil || val <= 0 {
					fmt.Fprintf(os.Stderr, "Error: %s requires a positive integer\n", arg)
					os.Exit(1)
				}
				opts.Interval = val
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires an interval in seconds\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--interval="):
			val, err := strconv.Atoi(strings.TrimPrefix(arg, "--interval="))
			if err != nil || val <= 0 {
				fmt.Fprintf(os.Stderr, "Error: --interval requires a positive integer\n")
				os.Exit(1)
			}
			opts.Interval = val

		case arg == "--ttl":
			if i+1 < len(args) {
				i++
				val, err := strconv.Atoi(args[i])
				if err != nil || val <= 0 {
					fmt.Fprintf(os.Stderr, "Error: %s requires a positive integer\n", arg)
					os.Exit(1)
				}
				opts.TTL = val
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a TTL in seconds\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--ttl="):
			val, err := strconv.Atoi(strings.TrimPrefix(arg, "--ttl="))
			if err != nil || val <= 0 {
				fmt.Fprintf(os.Stderr, "Error: --ttl requires a positive integer\n")
				os.Exit(1)
			}
			opts.TTL = val

		case arg == "--capabilities":
			if i+1 < len(args) {
				i++
				opts.Capabilities = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a comma-separated list\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--capabilities="):
			opts.Capabilities = strings.TrimPrefix(arg, "--capabilities=")

		case arg == "--auto-select":
			opts.NoAutoSelect = &boolFalse // auto-select enabled (noAutoSelect = false)
		case arg == "--no-auto-select":
			opts.NoAutoSelect = &boolTrue // auto-select disabled (noAutoSelect = true)

		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Error: unknown option %s\n", arg)
			fmt.Fprintf(os.Stderr, "Run 'nbor --help' for usage\n")
			os.Exit(1)
		default:
			// Positional argument = interface name
			if opts.InterfaceName == "" {
				opts.InterfaceName = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument %s\n", arg)
				os.Exit(1)
			}
		}
	}

	return opts
}
