//go:build windows

// ABOUTME: Windows-specific terminal dimension detection.
// ABOUTME: Falls back to environment variables since Unix ioctl is unavailable.

package cli

// getTerminalHeight returns the height of the terminal in rows.
// On Windows, falls back to environment variables.
func getTerminalHeight() int {
	return getTerminalHeightFromEnv()
}
