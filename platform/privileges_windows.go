//go:build windows

package platform

import (
	"errors"

	"golang.org/x/sys/windows"
)

// ErrNpcapNotFound is returned when Npcap is not installed
var ErrNpcapNotFound = errors.New("Npcap not found. Please install Npcap from https://npcap.com (check 'WinPcap API-compatible Mode' during install)")

// CheckPrivileges verifies the application has necessary privileges for packet capture
func CheckPrivileges() error {
	if !isAdmin() {
		return errors.New("nbor requires Administrator privileges.\nRight-click Command Prompt or PowerShell and select 'Run as administrator'")
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

