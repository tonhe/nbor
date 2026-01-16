//go:build !windows

package platform

// CheckNpcap is a no-op on non-Windows platforms
func CheckNpcap() error {
	return nil
}
