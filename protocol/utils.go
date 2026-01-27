package protocol

import "strings"

// CleanString removes null bytes and trims whitespace
func CleanString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.TrimSpace(s)
	return s
}
