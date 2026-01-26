package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	// Theme is the slug name of the theme to use (e.g., "tokyo-night", "catppuccin-mocha")
	Theme string `toml:"theme"`

	// SystemName is the name advertised in CDP/LLDP broadcasts (defaults to hostname)
	SystemName string `toml:"system_name"`

	// SystemDescription is the description advertised in CDP/LLDP broadcasts
	SystemDescription string `toml:"system_description"`

	// CDPListen enables listening for CDP packets
	CDPListen bool `toml:"cdp_listen"`

	// CDPBroadcast enables broadcasting CDP packets
	CDPBroadcast bool `toml:"cdp_broadcast"`

	// LLDPListen enables listening for LLDP packets
	LLDPListen bool `toml:"lldp_listen"`

	// LLDPBroadcast enables broadcasting LLDP packets
	LLDPBroadcast bool `toml:"lldp_broadcast"`

	// BroadcastOnStartup enables broadcasting immediately when the application starts
	// If false, broadcasting must be manually enabled with the 'b' key
	BroadcastOnStartup bool `toml:"broadcast_on_startup"`

	// AdvertiseInterval is the interval between broadcast packets in seconds
	AdvertiseInterval int `toml:"advertise_interval"`

	// TTL is the time-to-live for advertised information in seconds
	TTL int `toml:"ttl"`

	// Capabilities is the list of capabilities to advertise (router, bridge, station, etc.)
	Capabilities []string `toml:"capabilities"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme:              "solarized-dark",
		SystemName:         "", // Empty means use hostname
		SystemDescription:  "", // Empty means use default "nbor vX.Y.Z"
		CDPListen:          true,
		CDPBroadcast:       false,
		LLDPListen:         true,
		LLDPBroadcast:      false,
		BroadcastOnStartup: false,
		AdvertiseInterval:  5,
		TTL:                20,
		Capabilities:       []string{"station"},
	}
}

// GetConfigDir returns the configuration directory path for the current platform
// Linux/macOS: $XDG_CONFIG_HOME/nbor or ~/.config/nbor
// Windows: %APPDATA%\nbor
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Use %APPDATA% on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback to user profile if APPDATA not set
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "nbor")
	default:
		// Use XDG_CONFIG_HOME on Linux/macOS, default to ~/.config
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(xdgConfig, "nbor")
	}

	return configDir, nil
}

// GetConfigPath returns the full path to the configuration file
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the configuration from the config file
// Returns default config if file doesn't exist
func Load() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return DefaultConfig(), err
	}

	// Fill in defaults for missing values
	defaults := DefaultConfig()
	if cfg.Theme == "" {
		cfg.Theme = defaults.Theme
	}
	// Note: SystemName and SystemDescription empty is valid (means use defaults at runtime)
	// For bool fields, we can't distinguish between "not set" and "set to false"
	// so we rely on the TOML decoder's behavior (missing = zero value)
	// For new configs, the defaults will be written on first save

	// Fill in missing numeric defaults (0 means not set for these)
	if cfg.AdvertiseInterval <= 0 {
		cfg.AdvertiseInterval = defaults.AdvertiseInterval
	}
	if cfg.TTL <= 0 {
		cfg.TTL = defaults.TTL
	}
	if len(cfg.Capabilities) == 0 {
		cfg.Capabilities = defaults.Capabilities
	}

	return cfg, nil
}

// Save writes the configuration to the config file
// Creates the config directory if it doesn't exist
func Save(cfg Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create the config file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write config with comments
	lines := []string{
		"# nbor configuration",
		"# Run `nbor --list-themes` to see available themes",
		"",
		"# Visual theme (use slug format with hyphens, e.g., tokyo-night, catppuccin-mocha)",
		fmt.Sprintf("theme = %q", cfg.Theme),
		"",
		"# System Identity",
		"# system_name defaults to hostname if empty",
		fmt.Sprintf("system_name = %q", cfg.SystemName),
		fmt.Sprintf("system_description = %q", cfg.SystemDescription),
		"",
		"# Protocol Listening",
		fmt.Sprintf("cdp_listen = %t", cfg.CDPListen),
		fmt.Sprintf("lldp_listen = %t", cfg.LLDPListen),
		"",
		"# Protocol Broadcasting",
		fmt.Sprintf("cdp_broadcast = %t", cfg.CDPBroadcast),
		fmt.Sprintf("lldp_broadcast = %t", cfg.LLDPBroadcast),
		"# broadcast_on_startup controls whether broadcasting starts automatically",
		fmt.Sprintf("broadcast_on_startup = %t", cfg.BroadcastOnStartup),
		"",
		"# Broadcasting Settings",
		"# advertise_interval is the time between broadcasts in seconds",
		fmt.Sprintf("advertise_interval = %d", cfg.AdvertiseInterval),
		"# ttl is the time-to-live for advertised information in seconds",
		fmt.Sprintf("ttl = %d", cfg.TTL),
		"",
		"# Capabilities to advertise (router, bridge, station, switch, phone, etc.)",
		fmt.Sprintf("capabilities = %s", formatStringSlice(cfg.Capabilities)),
		"",
	}

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// formatStringSlice formats a string slice as a TOML array
func formatStringSlice(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	quoted := make([]string, len(s))
	for i, v := range s {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// EnsureConfigExists creates the default config file if it doesn't exist
func EnsureConfigExists() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File exists, nothing to do
	}

	// Create default config
	return Save(DefaultConfig())
}
