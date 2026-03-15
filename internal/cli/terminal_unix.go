//go:build !windows

// ABOUTME: Unix-specific terminal dimension detection using ioctl syscall.
// ABOUTME: Falls back to environment variables and tput when syscall fails.

package cli

import (
	"syscall"
	"unsafe"
)

// getTerminalHeight returns the height of the terminal in rows
func getTerminalHeight() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return getTerminalHeightFromEnv()
	}

	if errno != 0 {
		return getTerminalHeightFromEnv()
	}

	return int(ws.Row)
}
