//go:build windows

package tree_sitter_typed_perl

import "syscall"

// loadLibraryPlatform handles Windows DLL loading
func loadLibraryPlatform(path string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(path)
	return uintptr(handle), err
}
