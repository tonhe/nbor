package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	// Theme is the slug name of the theme to use (e.g., "tokyo-night", "catppuccin-mocha")
	Theme string `toml:"theme"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme: "solarized-dark",
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
	if cfg.Theme == "" {
		cfg.Theme = "solarized-dark"
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

	// Write header comment
	if _, err := file.WriteString("# nbor configuration\n"); err != nil {
		return err
	}
	if _, err := file.WriteString("# Run `nbor --list-themes` to see available themes\n\n"); err != nil {
		return err
	}
	if _, err := file.WriteString("# Theme name (use slug format with hyphens, e.g., tokyo-night, catppuccin-mocha)\n"); err != nil {
		return err
	}

	// Encode config as TOML
	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
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
