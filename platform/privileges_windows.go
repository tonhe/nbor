//go:build windows

package platform

import (
	"errors"

	"golang.org/x/sys/windows"
)

// ErrNotPrivileged is returned when the application lacks required privileges
var ErrNotPrivileged = errors.New("nbor requires Administrator privileges. Run from an elevated prompt")

// ErrNpcapNotFound is returned when Npcap is not installed
var ErrNpcapNotFound = errors.New("Npcap not found. Please install Npcap from https://npcap.com (check 'WinPcap API-compatible Mode' during install)")

// CheckPrivileges verifies the application has necessary privileges for packet capture
func CheckPrivileges() error {
	// First check for Npcap by trying to find devices
	// This is done in capture package, but we check admin rights here
	if !isAdmin() {
		return ErrNotPrivileged
	}
	return nil
}

// isAdmin checks if the current process is running with administrator privileges
func isAdmin() bool {
	var sid *windows.SID

	// Create a SID for the BUILTIN\Administrators group
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// Check if the current process token is a member of the Administrators group
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

// GetPrivilegeHint returns a hint for how to gain privileges
func GetPrivilegeHint() string {
	return "Right-click Command Prompt or PowerShell and select 'Run as administrator'"
}
