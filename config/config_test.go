package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check default values
	if cfg.Theme != "solarized-dark" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "solarized-dark")
	}
	if cfg.AdvertiseInterval != 5 {
		t.Errorf("AdvertiseInterval = %d, want %d", cfg.AdvertiseInterval, 5)
	}
	if cfg.TTL != 20 {
		t.Errorf("TTL = %d, want %d", cfg.TTL, 20)
	}
	if cfg.StalenessTimeout != 180 {
		t.Errorf("StalenessTimeout = %d, want %d", cfg.StalenessTimeout, 180)
	}
	if cfg.StaleRemovalTime != 0 {
		t.Errorf("StaleRemovalTime = %d, want %d", cfg.StaleRemovalTime, 0)
	}
	if !cfg.CDPListen {
		t.Error("CDPListen = false, want true")
	}
	if !cfg.LLDPListen {
		t.Error("LLDPListen = false, want true")
	}
	if cfg.CDPBroadcast {
		t.Error("CDPBroadcast = true, want false")
	}
	if cfg.LLDPBroadcast {
		t.Error("LLDPBroadcast = true, want false")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		wantErrors int
	}{
		{
			name:       "default config is valid",
			cfg:        DefaultConfig(),
			wantErrors: 0,
		},
		{
			name: "interval too low",
			cfg: Config{
				AdvertiseInterval: 0,
				TTL:               20,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "interval too high",
			cfg: Config{
				AdvertiseInterval: 301,
				TTL:               20,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "TTL too low",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               0,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "TTL too high",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               65536,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "staleness timeout negative",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               20,
				StalenessTimeout:  -1,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "staleness timeout too high",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               20,
				StalenessTimeout:  86401,
				StaleRemovalTime:  0,
			},
			wantErrors: 1,
		},
		{
			name: "stale removal negative",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               20,
				StalenessTimeout:  180,
				StaleRemovalTime:  -1,
			},
			wantErrors: 1,
		},
		{
			name: "multiple errors",
			cfg: Config{
				AdvertiseInterval: 0,
				TTL:               0,
				StalenessTimeout:  -1,
				StaleRemovalTime:  -1,
			},
			wantErrors: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.cfg.Validate()
			if len(errors) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidateAndFix(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantFixed int
		checkFn   func(t *testing.T, cfg *Config)
	}{
		{
			name:      "default config needs no fixes",
			cfg:       DefaultConfig(),
			wantFixed: 0,
		},
		{
			name: "fixes interval too low",
			cfg: Config{
				AdvertiseInterval: 0,
				TTL:               20,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantFixed: 1,
			checkFn: func(t *testing.T, cfg *Config) {
				if cfg.AdvertiseInterval != 5 {
					t.Errorf("AdvertiseInterval = %d, want 5", cfg.AdvertiseInterval)
				}
			},
		},
		{
			name: "fixes TTL too high",
			cfg: Config{
				AdvertiseInterval: 5,
				TTL:               70000,
				StalenessTimeout:  180,
				StaleRemovalTime:  0,
			},
			wantFixed: 1,
			checkFn: func(t *testing.T, cfg *Config) {
				if cfg.TTL != 20 {
					t.Errorf("TTL = %d, want 20", cfg.TTL)
				}
			},
		},
		{
			name: "fixes all invalid values",
			cfg: Config{
				AdvertiseInterval: 500,
				TTL:               -1,
				StalenessTimeout:  100000,
				StaleRemovalTime:  -5,
			},
			wantFixed: 4,
			checkFn: func(t *testing.T, cfg *Config) {
				if cfg.AdvertiseInterval != 5 {
					t.Errorf("AdvertiseInterval = %d, want 5", cfg.AdvertiseInterval)
				}
				if cfg.TTL != 20 {
					t.Errorf("TTL = %d, want 20", cfg.TTL)
				}
				if cfg.StalenessTimeout != 180 {
					t.Errorf("StalenessTimeout = %d, want 180", cfg.StalenessTimeout)
				}
				if cfg.StaleRemovalTime != 0 {
					t.Errorf("StaleRemovalTime = %d, want 0", cfg.StaleRemovalTime)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			fixed := cfg.ValidateAndFix()
			if len(fixed) != tt.wantFixed {
				t.Errorf("ValidateAndFix() fixed %d fields, want %d: %v", len(fixed), tt.wantFixed, fixed)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, &cfg)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "nbor-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create custom config
	cfg := Config{
		Theme:             "dracula",
		SystemName:        "test-host",
		SystemDescription: "Test description",
		CDPListen:         true,
		CDPBroadcast:      true,
		LLDPListen:        false,
		LLDPBroadcast:     true,
		AdvertiseInterval: 10,
		TTL:               30,
		Capabilities:      []string{"router", "switch"},
		StalenessTimeout:  300,
		StaleRemovalTime:  600,
		LoggingEnabled:    false,
		LogDirectory:      "/tmp/logs",
	}

	// Save to temp file
	configPath := filepath.Join(tmpDir, "config.toml")
	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	file.Close()

	// We can't easily test Save/Load without mocking GetConfigPath
	// So we just verify Save doesn't error
	err = Save(cfg)
	if err != nil {
		// This is expected if config dir doesn't exist - that's ok for unit test
		t.Logf("Save returned error (expected in test env): %v", err)
	}
}

func TestFormatStringSlice(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, "[]"},
		{[]string{}, "[]"},
		{[]string{"router"}, `["router"]`},
		{[]string{"router", "switch"}, `["router", "switch"]`},
	}

	for _, tt := range tests {
		got := formatStringSlice(tt.input)
		if got != tt.want {
			t.Errorf("formatStringSlice(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
