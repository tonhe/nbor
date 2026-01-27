// Package broadcast provides CDP and LLDP frame generation and transmission.
package broadcast

import (
	"os"
	"sync"
	"time"

	"github.com/google/gopacket/pcap"

	"nbor/config"
	"nbor/types"
)

// Broadcaster handles periodic CDP/LLDP packet transmission
type Broadcaster struct {
	handle     *pcap.Handle
	config     *config.Config
	iface      *types.InterfaceInfo
	systemName string
	stopChan   chan struct{}
	running    bool
	mu         sync.Mutex
}

// NewBroadcaster creates a new broadcaster instance
func NewBroadcaster(handle *pcap.Handle, cfg *config.Config, iface *types.InterfaceInfo) *Broadcaster {
	// Determine system name
	systemName := cfg.SystemName
	if systemName == "" {
		hostname, err := os.Hostname()
		if err == nil {
			systemName = hostname
		} else {
			systemName = "nbor"
		}
	}

	return &Broadcaster{
		handle:     handle,
		config:     cfg,
		iface:      iface,
		systemName: systemName,
		stopChan:   make(chan struct{}),
	}
}

// Start begins periodic packet transmission
func (b *Broadcaster) Start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.stopChan = make(chan struct{})
	b.mu.Unlock()

	go b.run()
}

// Stop stops the broadcaster
func (b *Broadcaster) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}
	b.running = false
	close(b.stopChan)
}

// IsRunning returns whether the broadcaster is currently running
func (b *Broadcaster) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}

// UpdateConfig updates the broadcaster configuration
func (b *Broadcaster) UpdateConfig(cfg *config.Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = cfg

	// Update system name if changed
	if cfg.SystemName != "" {
		b.systemName = cfg.SystemName
	}
}

// run is the main broadcast loop
func (b *Broadcaster) run() {
	// Get interval from config
	b.mu.Lock()
	interval := time.Duration(b.config.AdvertiseInterval) * time.Second
	b.mu.Unlock()

	// Send immediately on start
	b.transmit()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.transmit()
			// Check if interval changed
			b.mu.Lock()
			newInterval := time.Duration(b.config.AdvertiseInterval) * time.Second
			b.mu.Unlock()
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
		case <-b.stopChan:
			return
		}
	}
}

// transmit sends CDP and/or LLDP packets based on configuration
func (b *Broadcaster) transmit() {
	b.mu.Lock()
	cfg := b.config
	iface := b.iface
	systemName := b.systemName
	b.mu.Unlock()

	// Send CDP if enabled
	if cfg.CDPBroadcast {
		frame, err := BuildCDPFrame(cfg, iface, systemName)
		if err == nil {
			_ = b.handle.WritePacketData(frame)
		}
	}

	// Send LLDP if enabled
	if cfg.LLDPBroadcast {
		frame, err := BuildLLDPFrame(cfg, iface, systemName)
		if err == nil {
			_ = b.handle.WritePacketData(frame)
		}
	}
}

// SendNow sends packets immediately (for testing)
func (b *Broadcaster) SendNow() error {
	b.transmit()
	return nil
}
