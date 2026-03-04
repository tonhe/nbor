//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckPrivileges verifies the application has necessary privileges for packet capture.
// If not root, it explains why and re-execs with sudo.
func CheckPrivileges() error {
	if os.Geteuid() == 0 {
		return nil
	}

	fmt.Println("nbor requires root privileges for raw packet capture (CDP/LLDP listening).")
	fmt.Println("Re-running with sudo...")
	fmt.Println()

	return reExecWithSudo()
}

// reExecWithSudo re-executes the current process with sudo, preserving all arguments.
func reExecWithSudo() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	args := append([]string{exe}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	os.Exit(0)
	return nil
}
