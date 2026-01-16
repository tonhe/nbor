package platform

import (
	"fmt"
	"os"
	"runtime"
)

// Bell sends a terminal bell/alert
func Bell() {
	// Use \a (ASCII bell) which works on most terminals
	// On Windows Terminal and modern terminals this works fine
	// On legacy cmd.exe it may not work but won't crash
	fmt.Fprint(os.Stderr, "\a")
}

// IsBellSupported returns true if the terminal likely supports bell
func IsBellSupported() bool {
	// On Windows, check if we're in Windows Terminal (modern) vs cmd.exe
	if runtime.GOOS == "windows" {
		// Windows Terminal sets WT_SESSION environment variable
		if os.Getenv("WT_SESSION") != "" {
			return true
		}
		// PowerShell and cmd.exe may or may not support it
		// Return true and let it fail silently if it doesn't work
		return true
	}

	// On Unix-like systems, check for a TTY
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}
