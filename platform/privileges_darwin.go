//go:build darwin

package platform

import (
	"errors"
	"os"
)

// ErrNotPrivileged is returned when the application lacks required privileges
var ErrNotPrivileged = errors.New("nbor requires root privileges. Run with sudo")

// CheckPrivileges verifies the application has necessary privileges for packet capture
func CheckPrivileges() error {
	if os.Geteuid() != 0 {
		return ErrNotPrivileged
	}
	return nil
}

// GetPrivilegeHint returns a hint for how to gain privileges
func GetPrivilegeHint() string {
	return "Try running: sudo nbor"
}
