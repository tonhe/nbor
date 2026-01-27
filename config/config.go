// Package config provides configuration loading, saving, and management.
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

	// FilterCapabilities filters which neighbors to display/log based on their capabilities
	// Empty means show all neighbors
	FilterCapabilities []string `toml:"filter_capabilities"`

	// StalenessTimeout is the number of seconds before a neighbor is marked as stale (grayed out)
	StalenessTimeout int `toml:"staleness_timeout"`

	// StaleRemovalTime is the number of seconds before a stale neighbor is removed from display
	// 0 means never remove stale neighbors
	StaleRemovalTime int `toml:"stale_removal_time"`

	// LoggingEnabled controls whether neighbor events are logged to files
	LoggingEnabled bool `toml:"logging_enabled"`

	// LogDirectory is the directory where log files are stored
	LogDirectory string `toml:"log_directory"`

	// AutoSelectInterface automatically selects the interface if only one wired interface is available
	AutoSelectInterface bool `toml:"auto_select_interface"`
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
		FilterCapabilities: []string{}, // Empty means show all
		StalenessTimeout:   180,         // 3 minutes
		StaleRemovalTime:   0,           // Never remove
		LoggingEnabled:      true,
		LogDirectory:        "", // Empty means use default location
		AutoSelectInterface: true,
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
	meta, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return DefaultConfig(), err
	}

	// Fill in defaults for missing values
	defaults := DefaultConfig()
	if cfg.Theme == "" {
		cfg.Theme = defaults.Theme
	}
	// Note: SystemName and SystemDescription empty is valid (means use defaults at runtime)

	// For bool fields, use metadata to check if they were actually defined
	// This allows us to distinguish between "not set" and "explicitly set to false"
	// Check ALL boolean fields for consistency and future-proofing
	if !meta.IsDefined("cdp_listen") {
		cfg.CDPListen = defaults.CDPListen
	}
	if !meta.IsDefined("cdp_broadcast") {
		cfg.CDPBroadcast = defaults.CDPBroadcast
	}
	if !meta.IsDefined("lldp_listen") {
		cfg.LLDPListen = defaults.LLDPListen
	}
	if !meta.IsDefined("lldp_broadcast") {
		cfg.LLDPBroadcast = defaults.LLDPBroadcast
	}
	if !meta.IsDefined("broadcast_on_startup") {
		cfg.BroadcastOnStartup = defaults.BroadcastOnStartup
	}
	if !meta.IsDefined("logging_enabled") {
		cfg.LoggingEnabled = defaults.LoggingEnabled
	}
	if !meta.IsDefined("auto_select_interface") {
		cfg.AutoSelectInterface = defaults.AutoSelectInterface
	}

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

	// Fill in new field defaults
	// FilterCapabilities: empty is valid (means show all), so don't fill default
	if cfg.StalenessTimeout <= 0 {
		cfg.StalenessTimeout = defaults.StalenessTimeout
	}
	// StaleRemovalTime: 0 is valid (means never remove), so don't fill default
	// LogDirectory: empty is valid (means use default location)

	// Validate and fix any out-of-range values
	cfg.ValidateAndFix()

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
		"# Display Filtering",
		"# filter_capabilities limits which neighbors are shown/logged based on capabilities",
		"# Empty array means show all neighbors",
		fmt.Sprintf("filter_capabilities = %s", formatStringSlice(cfg.FilterCapabilities)),
		"",
		"# Staleness Settings",
		"# staleness_timeout is seconds before a neighbor is grayed out (default 180)",
		fmt.Sprintf("staleness_timeout = %d", cfg.StalenessTimeout),
		"# stale_removal_time is seconds before stale neighbors are removed (0 = never)",
		fmt.Sprintf("stale_removal_time = %d", cfg.StaleRemovalTime),
		"",
		"# Logging",
		fmt.Sprintf("logging_enabled = %t", cfg.LoggingEnabled),
		"# log_directory is where log files are stored (empty = default location)",
		fmt.Sprintf("log_directory = %q", cfg.LogDirectory),
		"",
		"# Interface Selection",
		"# auto_select_interface skips the picker when only one wired interface is available",
		fmt.Sprintf("auto_select_interface = %t", cfg.AutoSelectInterface),
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

// Validate checks configuration values and returns any validation errors
// Returns nil if all values are valid
func (c *Config) Validate() []string {
	var errors []string
	defaults := DefaultConfig()

	// AdvertiseInterval: 1-300 seconds
	if c.AdvertiseInterval < 1 || c.AdvertiseInterval > 300 {
		errors = append(errors, fmt.Sprintf("advertise_interval %d out of range (1-300), using default %d",
			c.AdvertiseInterval, defaults.AdvertiseInterval))
	}

	// TTL: 1-65535 seconds
	if c.TTL < 1 || c.TTL > 65535 {
		errors = append(errors, fmt.Sprintf("ttl %d out of range (1-65535), using default %d",
			c.TTL, defaults.TTL))
	}

	// StalenessTimeout: 0-86400 seconds (0 = disable staleness)
	if c.StalenessTimeout < 0 || c.StalenessTimeout > 86400 {
		errors = append(errors, fmt.Sprintf("staleness_timeout %d out of range (0-86400), using default %d",
			c.StalenessTimeout, defaults.StalenessTimeout))
	}

	// StaleRemovalTime: 0-86400 seconds (0 = never remove)
	if c.StaleRemovalTime < 0 || c.StaleRemovalTime > 86400 {
		errors = append(errors, fmt.Sprintf("stale_removal_time %d out of range (0-86400), using default %d",
			c.StaleRemovalTime, defaults.StaleRemovalTime))
	}

	return errors
}

// ValidateAndFix checks configuration values and fixes invalid ones to defaults
// Returns a list of fields that were fixed
func (c *Config) ValidateAndFix() []string {
	var fixed []string
	defaults := DefaultConfig()

	// AdvertiseInterval: 1-300 seconds
	if c.AdvertiseInterval < 1 || c.AdvertiseInterval > 300 {
		fixed = append(fixed, fmt.Sprintf("advertise_interval: %d -> %d", c.AdvertiseInterval, defaults.AdvertiseInterval))
		c.AdvertiseInterval = defaults.AdvertiseInterval
	}

	// TTL: 1-65535 seconds
	if c.TTL < 1 || c.TTL > 65535 {
		fixed = append(fixed, fmt.Sprintf("ttl: %d -> %d", c.TTL, defaults.TTL))
		c.TTL = defaults.TTL
	}

	// StalenessTimeout: 0-86400 seconds
	if c.StalenessTimeout < 0 || c.StalenessTimeout > 86400 {
		fixed = append(fixed, fmt.Sprintf("staleness_timeout: %d -> %d", c.StalenessTimeout, defaults.StalenessTimeout))
		c.StalenessTimeout = defaults.StalenessTimeout
	}

	// StaleRemovalTime: 0-86400 seconds
	if c.StaleRemovalTime < 0 || c.StaleRemovalTime > 86400 {
		fixed = append(fixed, fmt.Sprintf("stale_removal_time: %d -> %d", c.StaleRemovalTime, defaults.StaleRemovalTime))
		c.StaleRemovalTime = defaults.StaleRemovalTime
	}

	return fixed
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
