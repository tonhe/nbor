package cli

import (
	"strings"

	"nbor/config"
)

// ApplyOverrides applies CLI flag overrides to the config
func ApplyOverrides(cfg *config.Config, opts Options) {
	// Identity overrides
	if opts.SystemName != "" {
		cfg.SystemName = opts.SystemName
	}
	if opts.SystemDescription != "" {
		cfg.SystemDescription = opts.SystemDescription
	}

	// Listening overrides
	if opts.CDPListen != nil {
		cfg.CDPListen = *opts.CDPListen
	}
	if opts.LLDPListen != nil {
		cfg.LLDPListen = *opts.LLDPListen
	}

	// Broadcasting overrides
	if opts.BroadcastAll {
		cfg.CDPBroadcast = true
		cfg.LLDPBroadcast = true
	}
	if opts.CDPBroadcast != nil {
		cfg.CDPBroadcast = *opts.CDPBroadcast
	}
	if opts.LLDPBroadcast != nil {
		cfg.LLDPBroadcast = *opts.LLDPBroadcast
	}

	// Timing overrides
	if opts.Interval > 0 {
		cfg.AdvertiseInterval = opts.Interval
	}
	if opts.TTL > 0 {
		cfg.TTL = opts.TTL
	}

	// Capabilities override
	if opts.Capabilities != "" {
		caps := strings.Split(opts.Capabilities, ",")
		var cleanCaps []string
		for _, c := range caps {
			c = strings.TrimSpace(strings.ToLower(c))
			if c != "" {
				cleanCaps = append(cleanCaps, c)
			}
		}
		if len(cleanCaps) > 0 {
			cfg.Capabilities = cleanCaps
		}
	}

	// Auto-select override
	if opts.NoAutoSelect != nil {
		cfg.AutoSelectInterface = !*opts.NoAutoSelect
	}
}
