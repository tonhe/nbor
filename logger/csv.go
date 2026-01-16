package logger

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"nbor/types"
)

// CSVLogger handles logging neighbor discoveries to a CSV file
type CSVLogger struct {
	mu       sync.Mutex
	file     *os.File
	writer   *csv.Writer
	filepath string
}

// NewCSVLogger creates a new CSV logger with a timestamped filename
func NewCSVLogger() (*CSVLogger, error) {
	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := fmt.Sprintf("nbor-%s.csv", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	writer := csv.NewWriter(file)

	logger := &CSVLogger{
		file:     file,
		writer:   writer,
		filepath: filename,
	}

	// Write header row
	header := []string{
		"Timestamp",
		"Interface",
		"Protocol",
		"Hostname",
		"Port ID",
		"Port Description",
		"Management IP",
		"Platform",
		"Description",
		"Location",
		"Capabilities",
		"Source MAC",
	}

	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	writer.Flush()

	return logger, nil
}

// Log writes a neighbor record to the CSV file
func (l *CSVLogger) Log(n *types.Neighbor) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer == nil {
		return fmt.Errorf("logger is closed")
	}

	// Format capabilities as comma-separated string
	caps := make([]string, len(n.Capabilities))
	for i, cap := range n.Capabilities {
		caps[i] = string(cap)
	}

	// Format management IP
	mgmtIP := ""
	if n.ManagementIP != nil {
		mgmtIP = n.ManagementIP.String()
	}

	// Format source MAC
	srcMAC := ""
	if n.SourceMAC != nil {
		srcMAC = n.SourceMAC.String()
	}

	record := []string{
		n.LastSeen.Format(time.RFC3339),
		n.Interface,
		string(n.Protocol),
		n.Hostname,
		n.PortID,
		n.PortDescription,
		mgmtIP,
		n.Platform,
		sanitizeForCSV(n.Description),
		n.Location,
		strings.Join(caps, ","),
		srcMAC,
	}

	if err := l.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	l.writer.Flush()
	return l.writer.Error()
}

// Close flushes and closes the CSV file
func (l *CSVLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		l.writer.Flush()
		l.writer = nil
	}

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}

	return nil
}

// Filepath returns the path to the log file
func (l *CSVLogger) Filepath() string {
	return l.filepath
}

// sanitizeForCSV removes or replaces characters that might cause issues in CSV
func sanitizeForCSV(s string) string {
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Trim excessive whitespace
	s = strings.TrimSpace(s)
	return s
}

// FormatMAC formats a MAC address for display
func FormatMAC(mac net.HardwareAddr) string {
	if mac == nil {
		return ""
	}
	return mac.String()
}

// FormatIP formats an IP address for display
func FormatIP(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

// FormatCapabilities formats capabilities for display
func FormatCapabilities(caps []types.Capability) string {
	if len(caps) == 0 {
		return ""
	}

	strs := make([]string, len(caps))
	for i, cap := range caps {
		strs[i] = string(cap)
	}
	return strings.Join(strs, ", ")
}

// FormatTime formats a time for display
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// FormatDuration formats the time since a timestamp
func FormatDuration(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}
